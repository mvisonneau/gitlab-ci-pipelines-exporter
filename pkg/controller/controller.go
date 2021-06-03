package controller

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/vmihailenco/taskq/v3"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/config"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/gitlab"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/ratelimit"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/store"
)

// Controller holds the necessary clients to run the app and handle requests
type Controller struct {
	Config         config.Config
	Redis          *redis.Client
	Gitlab         *gitlab.Client
	Store          store.Store
	TaskController TaskController
}

// New creates a new controller
func New(ctx context.Context, cfg config.Config, version string) (c Controller, err error) {
	c.Config = cfg

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
	for n, h := range map[TaskType]interface{}{
		TaskTypeGarbageCollectEnvironments:   c.TaskHandlerGarbageCollectEnvironments,
		TaskTypeGarbageCollectMetrics:        c.TaskHandlerGarbageCollectMetrics,
		TaskTypeGarbageCollectProjects:       c.TaskHandlerGarbageCollectProjects,
		TaskTypeGarbageCollectRefs:           c.TaskHandlerGarbageCollectRefs,
		TaskTypePullEnvironmentMetrics:       c.TaskHandlerPullEnvironmentMetrics,
		TaskTypePullEnvironmentsFromProject:  c.TaskHandlerPullEnvironmentsFromProject,
		TaskTypePullEnvironmentsFromProjects: c.TaskHandlerPullEnvironmentsFromProjects,
		TaskTypePullMetrics:                  c.TaskHandlerPullMetrics,
		TaskTypePullProjectsFromWildcard:     c.TaskHandlerPullProjectsFromWildcard,
		TaskTypePullProjectsFromWildcards:    c.TaskHandlerPullProjectsFromWildcards,
		TaskTypePullRefMetrics:               c.TaskHandlerPullRefMetrics,
		TaskTypePullRefsFromProject:          c.TaskHandlerPullRefsFromProject,
		TaskTypePullRefsFromProjects:         c.TaskHandlerPullRefsFromProjects,
	} {
		_, _ = c.TaskController.TaskMap.Register(&taskq.TaskOptions{
			Name:    string(n),
			Handler: h,
		})
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
	if len(url) > 0 {
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
	}
	return
}
