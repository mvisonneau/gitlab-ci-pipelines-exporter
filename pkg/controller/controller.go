package controller

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/config"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/gitlab"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/ratelimit"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/store"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/vmihailenco/taskq/v3"
)

// Controller holds the necessary clients to run the app and handle requests
type Controller struct {
	Config         config.Config
	Redis          *redis.Client
	Gitlab         *gitlab.Client
	Store          store.Store
	TaskController TaskController

	// UUID is used to identify this controller/process amongst others when
	// the exporter is running in cluster mode, leveraging Redis.
	UUID uuid.UUID
}

// New creates a new controller
func New(ctx context.Context, cfg config.Config, version string) (c Controller, err error) {
	c.Config = cfg
	c.UUID = uuid.New()

	if err = c.configureRedis(cfg.Redis.URL); err != nil {
		return
	}

	c.TaskController = NewTaskController(c.Redis)
	c.registerTasks()

	c.Store = store.New(c.Redis, c.Config.Projects)

	if err = c.configureGitlab(cfg.Gitlab, version); err != nil {
		return
	}

	// Start the scheduler
	c.Schedule(ctx, cfg.Pull, cfg.GarbageCollect)

	return
}

func (c *Controller) registerTasks() {
	for n, h := range map[schemas.TaskType]interface{}{
		schemas.TaskTypeGarbageCollectEnvironments:   c.TaskHandlerGarbageCollectEnvironments,
		schemas.TaskTypeGarbageCollectMetrics:        c.TaskHandlerGarbageCollectMetrics,
		schemas.TaskTypeGarbageCollectProjects:       c.TaskHandlerGarbageCollectProjects,
		schemas.TaskTypeGarbageCollectRefs:           c.TaskHandlerGarbageCollectRefs,
		schemas.TaskTypePullEnvironmentMetrics:       c.TaskHandlerPullEnvironmentMetrics,
		schemas.TaskTypePullEnvironmentsFromProject:  c.TaskHandlerPullEnvironmentsFromProject,
		schemas.TaskTypePullEnvironmentsFromProjects: c.TaskHandlerPullEnvironmentsFromProjects,
		schemas.TaskTypePullMetrics:                  c.TaskHandlerPullMetrics,
		schemas.TaskTypePullProjectsFromWildcard:     c.TaskHandlerPullProjectsFromWildcard,
		schemas.TaskTypePullProjectsFromWildcards:    c.TaskHandlerPullProjectsFromWildcards,
		schemas.TaskTypePullRefMetrics:               c.TaskHandlerPullRefMetrics,
		schemas.TaskTypePullRefsFromProject:          c.TaskHandlerPullRefsFromProject,
		schemas.TaskTypePullRefsFromProjects:         c.TaskHandlerPullRefsFromProjects,
	} {
		_, _ = c.TaskController.TaskMap.Register(&taskq.TaskOptions{
			Name:       string(n),
			Handler:    h,
			RetryLimit: 1,
		})
	}
}

func (c *Controller) unqueueTask(tt schemas.TaskType, uniqueID string) {
	if err := c.Store.UnqueueTask(tt, uniqueID); err != nil {
		log.WithFields(log.Fields{
			"task_type":      tt,
			"task_unique_id": uniqueID,
		}).WithError(err).Warn("unqueuing task")
	}
}

func (c *Controller) configureGitlab(cfg config.Gitlab, version string) (err error) {
	var rl ratelimit.Limiter

	if c.Redis != nil {
		rl = ratelimit.NewRedisLimiter(context.Background(), c.Redis, cfg.MaximumRequestsPerSecond)
	} else {
		rl = ratelimit.NewLocalLimiter(cfg.MaximumRequestsPerSecond)
	}

	c.Gitlab, err = gitlab.NewClient(gitlab.ClientConfig{
		URL:              cfg.URL,
		Token:            cfg.Token,
		DisableTLSVerify: !cfg.EnableTLSVerify,
		UserAgentVersion: version,
		RateLimiter:      rl,
		ReadinessURL:     cfg.HealthURL,
	})
	return
}

func (c *Controller) configureRedis(url string) (err error) {
	if len(url) <= 0 {
		log.Debug("redis url is not configured, skipping configuration & using local driver")
		return
	}

	log.Info("redis url configured, initializing connection..")

	var opt *redis.Options
	if opt, err = redis.ParseURL(url); err != nil {
		return
	}

	c.Redis = redis.NewClient(opt)

	if _, err := c.Redis.Ping(context.Background()).Result(); err != nil {
		return errors.Wrap(err, "connecting to redis")
	}

	log.Info("connected to redis")

	return
}
