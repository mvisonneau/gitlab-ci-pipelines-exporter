package exporter

import (
	"context"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/heptiolabs/healthcheck"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/vmihailenco/taskq/v3"
	"github.com/vmihailenco/taskq/v3/memqueue"
	"github.com/vmihailenco/taskq/v3/redisq"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/gitlab"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/ratelimit"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/storage"
)

var (
	config        schemas.Config
	gitlabClient  *gitlab.Client
	redisClient   *redis.Client
	taskFactory   taskq.Factory
	pullingQueue  taskq.Queue
	store         storage.Storage
	cfgUpdateLock sync.RWMutex
)

// Configure ..
func Configure(cfg schemas.Config, userAgentVersion string) error {
	cfgUpdateLock.Lock()
	config = cfg
	cfgUpdateLock.Unlock()

	configurePullingQueue()
	configureStore()
	return configureGitlabClient(userAgentVersion)
}

// ConfigureGitlabClient ..
func configureGitlabClient(userAgentVersion string) (err error) {
	cfgUpdateLock.Lock()
	defer cfgUpdateLock.Unlock()

	gitlabClient, err = gitlab.NewClient(gitlab.ClientConfig{
		URL:              config.Gitlab.URL,
		Token:            config.Gitlab.Token,
		DisableTLSVerify: !config.Gitlab.EnableTLSVerify,
		UserAgentVersion: userAgentVersion,
		RateLimiter:      newRateLimiter(),
		ReadinessURL:     config.Gitlab.HealthURL,
	})
	return
}

// ConfigureRedisClient ..
func ConfigureRedisClient(c *redis.Client) error {
	cfgUpdateLock.Lock()
	defer cfgUpdateLock.Unlock()

	redisClient = c
	if _, err := redisClient.Ping(context.Background()).Result(); err != nil {
		return errors.Wrap(err, "connecting to redis")
	}
	return nil
}

// ConfigurePullingQueue ..
func configurePullingQueue() {
	cfgUpdateLock.Lock()
	defer cfgUpdateLock.Unlock()

	pullingQueueOptions := &taskq.QueueOptions{
		Name:                 "pull",
		PauseErrorsThreshold: 0,

		// Disable system resources checks
		MinSystemResources: taskq.SystemResources{
			Load1PerCPU:          -1,
			MemoryFreeMB:         0,
			MemoryFreePercentage: 0,
		},
	}

	if redisClient != nil {
		taskFactory = redisq.NewFactory()
		pullingQueueOptions.Redis = redisClient
	} else {
		taskFactory = memqueue.NewFactory()
	}

	pullingQueue = taskFactory.RegisterQueue(pullingQueueOptions)

	// Purge the queue when we start
	// I am only partially convinced this will not cause issues in HA fashion
	if err := pullingQueue.Purge(); err != nil {
		log.WithField("error", err.Error()).Error("purging the pulling queue")
	}
}

// ConfigureStore ..
func configureStore() {
	cfgUpdateLock.Lock()
	defer cfgUpdateLock.Unlock()

	if redisClient != nil {
		store = storage.NewRedisStorage(redisClient)
	} else {
		store = storage.NewLocalStorage()
	}

	// Load all the configured projects in the store
	for _, p := range config.Projects {
		exists, err := store.ProjectExists(p.Key())
		if err != nil {
			log.WithFields(log.Fields{
				"project-name": p.Name,
				"error":        err.Error(),
			}).Error("reading project from the store")
		}

		if !exists {
			if err = store.SetProject(p); err != nil {
				log.WithFields(log.Fields{
					"project-name": p.Name,
					"error":        err.Error(),
				}).Error("writing project in the store")
			}

			if config.Pull.RefsFromProjects.OnInit {
				go schedulePullRefsFromProject(context.Background(), p)
				go schedulePullRefsFromPipeline(context.Background(), p)
			}

			if config.Pull.EnvironmentsFromProjects.OnInit {
				go schedulePullEnvironmentsFromProject(context.Background(), p)
			}
		}
	}
}

func newRateLimiter() ratelimit.Limiter {
	if redisClient != nil {
		return ratelimit.NewRedisLimiter(context.Background(), redisClient, config.Pull.MaximumGitLabAPIRequestsPerSecond)
	}
	return ratelimit.NewLocalLimiter(config.Pull.MaximumGitLabAPIRequestsPerSecond)
}

func processPullingQueue(ctx context.Context) {
	if redisClient != nil {
		if err := taskFactory.StartConsumers(ctx); err != nil {
			log.Fatal(err)
		}
	}
}

func healthCheckHandler() (h healthcheck.Handler) {
	h = healthcheck.NewHandler()
	if config.Gitlab.EnableHealthCheck {
		h.AddReadinessCheck("gitlab-reachable", gitlabClient.ReadinessCheck())
	} else {
		log.Warn("GitLab health check has been disabled. Readiness checks won't be operated.")
	}

	return
}

// Run executes the http servers supporting the exporter
func Run() {
	// Graceful shutdowns
	onShutdown := make(chan os.Signal, 1)
	signal.Notify(onShutdown, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)

	schedulingContext, stopOrchestratePulling := context.WithCancel(context.Background())
	schedule(schedulingContext)
	processPullingQueue(schedulingContext)

	// HTTP server
	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:    config.Server.ListenAddress,
		Handler: mux,
	}

	// health endpoints
	health := healthCheckHandler()
	mux.HandleFunc("/health/live", health.LiveEndpoint)
	mux.HandleFunc("/health/ready", health.ReadyEndpoint)

	// metrics endpoint
	if config.Server.Metrics.Enabled {
		mux.HandleFunc("/metrics", MetricsHandler)
	}

	// pprof/debug endpoints
	if config.Server.EnablePprof {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	// webhook endpoints
	if config.Server.Webhook.Enabled {
		mux.HandleFunc("/webhook", WebhookHandler)
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	log.WithFields(
		log.Fields{
			"listen-address":           config.Server.ListenAddress,
			"pprof-endpoint-enabled":   config.Server.EnablePprof,
			"metrics-endpoint-enabled": config.Server.Metrics.Enabled,
			"webhook-endpoint-enabled": config.Server.Webhook.Enabled,
		},
	).Info("http server started")

	<-onShutdown
	log.Info("received signal, attempting to gracefully exit..")
	stopOrchestratePulling()

	httpServerContext, forceHTTPServerShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer forceHTTPServerShutdown()

	if err := srv.Shutdown(httpServerContext); err != nil {
		log.Fatalf("metrics server shutdown failed: %+v", err)
	}

	log.Info("stopped!")
}
