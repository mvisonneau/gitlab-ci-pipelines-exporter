package cmd

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/heptiolabs/healthcheck"
	"github.com/mvisonneau/go-helpers/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/xanzy/go-gitlab"
	"go.uber.org/ratelimit"
)

const (
	userAgent = "gitlab-ci-pipelines-exporter"
)

// Client holds a GitLab client
type Client struct {
	*gitlab.Client
	RateLimiter ratelimit.Limiter
}

var (
	coverage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_coverage",
			Help: "Coverage of the most recent pipeline",
		},
		[]string{"project", "topics", "ref"},
	)

	lastRunDuration = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_last_run_duration_seconds",
			Help: "Duration of last pipeline run",
		},
		[]string{"project", "topics", "ref"},
	)

	lastRunJobDuration = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_last_job_run_duration_seconds",
			Help: "Duration of last job run",
		},
		[]string{"project", "topics", "ref", "stage", "job_name"},
	)

	lastRunJobStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_last_job_run_status",
			Help: "Status of the most recent job",
		},
		[]string{"project", "topics", "ref", "stage", "job_name", "status"},
	)

	lastRunJobArtifactSize = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_last_job_run_artifact_size",
			Help: "Filesize of the most recent job artifacts",
		},
		[]string{"project", "topics", "ref", "stage", "job_name"},
	)

	timeSinceLastJobRun = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_time_since_last_job_run_seconds",
			Help: "Elapsed time since most recent GitLab CI job run.",
		},
		[]string{"project", "topics", "ref", "stage", "job_name"},
	)

	jobRunCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gitlab_ci_pipeline_job_run_count",
			Help: "GitLab CI pipeline job run count",
		},
		[]string{"project", "topics", "ref", "stage", "job_name"},
	)

	lastRunID = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_last_run_id",
			Help: "ID of the most recent pipeline",
		},
		[]string{"project", "topics", "ref"},
	)

	lastRunStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_last_run_status",
			Help: "Status of the most recent pipeline",
		},
		[]string{"project", "topics", "ref", "status"},
	)

	runCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gitlab_ci_pipeline_run_count",
			Help: "GitLab CI pipeline run count",
		},
		[]string{"project", "topics", "ref"},
	)

	timeSinceLastRun = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_time_since_last_run_seconds",
			Help: "Elapsed time since most recent GitLab CI pipeline run.",
		},
		[]string{"project", "topics", "ref"},
	)
)

var defaultMetrics = []prometheus.Collector{coverage, lastRunDuration, lastRunID, lastRunStatus, runCount, timeSinceLastJobRun, lastRunDuration, lastRunJobStatus, jobRunCount, timeSinceLastJobRun, lastRunJobArtifactSize}

func newMetricsRegistry(log *log.Logger, metrics ...prometheus.Collector) *prometheus.Registry {
	registry := prometheus.NewRegistry()
	for _, m := range metrics {
		if err := registry.Register(m); err != nil {
			log.Fatalf("could not add provided metric '%v' to the Prometheus registry: %v", m, err)
		}
	}
	return registry
}

// Run launches the exporter
func Run(ctx *cli.Context) error {

	// Configure logger
	lc := &logger.Config{
		Level:  ctx.GlobalString("log-level"),
		Format: ctx.GlobalString("log-format"),
	}

	if err := lc.Configure(); err != nil {
		return exit(err, 1)
	}

	// Initialize config
	if err := cfg.Parse(ctx.GlobalString("config")); err != nil {
		return exit(err, 1)
	}

	cfg.MergeWithContext(ctx)

	log.Infof("Starting exporter")
	log.Infof("Configured GitLab endpoint : %s", cfg.Gitlab.URL)
	log.Infof("Polling projects every %ds", cfg.ProjectsPollingIntervalSeconds)
	log.Infof("Polling refs every %ds", cfg.RefsPollingIntervalSeconds)
	log.Infof("Polling pipelines every %ds", cfg.PipelinesPollingIntervalSeconds)
	log.Infof("Global rate limit for the GitLab API set to %d req/s", cfg.MaximumGitLabAPIRequestsPerSecond)

	// Configure GitLab client
	httpTransport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: cfg.Gitlab.SkipTLSVerify},
	}
	opts := []gitlab.ClientOptionFunc{
		gitlab.WithHTTPClient(&http.Client{Transport: httpTransport}),
		gitlab.WithBaseURL(cfg.Gitlab.URL),
	}
	gc, err := gitlab.NewClient(cfg.Gitlab.Token, opts...)
	if err != nil {
		return exit(err, 1)
	}
	c := &Client{
		Client:      gc,
		RateLimiter: ratelimit.New(cfg.MaximumGitLabAPIRequestsPerSecond),
	}
	c.UserAgent = userAgent

	// Graceful shutdowns
	onShutdown := make(chan os.Signal, 1)
	signal.Notify(onShutdown, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)

	stopPolling := make(chan bool)
	c.pollProjects(stopPolling)

	// Configure liveness and readiness probes
	health := healthcheck.NewHandler()
	if !cfg.Gitlab.SkipTLSVerify {
		health.AddReadinessCheck("gitlab-reachable", healthcheck.HTTPGetCheck(cfg.Gitlab.HealthURL, 5*time.Second))
	} else {
		log.Warn("TLS verification has been disabled. Readiness checks won't be operated.")
	}

	// Register registry
	registry := newMetricsRegistry(log.StandardLogger(), defaultMetrics...)

	// Expose the registered registry via HTTP
	mux := http.NewServeMux()
	mux.HandleFunc("/health/live", health.LiveEndpoint)
	mux.HandleFunc("/health/ready", health.ReadyEndpoint)
	mux.Handle("/metrics", metricsHandlerFor(registry, cfg.PrometheusOpenmetricsEncoding))

	srv := &http.Server{
		Addr:    ctx.GlobalString("listen-address"),
		Handler: mux,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	log.Infof("Started listening onto %s", ctx.GlobalString("listen-address"))

	<-onShutdown
	stopPolling <- true
	log.Print("Stopped!")

	ctxt, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctxt); err != nil {
		log.Fatalf("Shutdown failed: %+v", err)
	}

	return exit(nil, 0)
}

func metricsHandlerFor(metrics *prometheus.Registry, openMetricsEncoder bool) http.Handler {
	return promhttp.HandlerFor(metrics, promhttp.HandlerOpts{
		Registry:          metrics,
		EnableOpenMetrics: openMetricsEncoder,
	})
}

func exit(err error, exitCode int) *cli.ExitError {
	if err != nil {
		log.Error(err.Error())
		return cli.NewExitError(err.Error(), exitCode)
	}

	return cli.NewExitError("", exitCode)
}
