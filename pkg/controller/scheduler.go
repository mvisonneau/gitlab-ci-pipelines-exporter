package controller

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/config"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	log "github.com/sirupsen/logrus"
	"github.com/vmihailenco/taskq/v3"
	"github.com/vmihailenco/taskq/v3/memqueue"
	"github.com/vmihailenco/taskq/v3/redisq"
)

// TaskController holds task related clients
type TaskController struct {
	Factory taskq.Factory
	Queue   taskq.Queue
	TaskMap *taskq.TaskMap
}

// TaskType represents the type of a task
type TaskType string

const (
	// TaskTypePullProjectsFromWildcard ..
	TaskTypePullProjectsFromWildcard TaskType = "PullProjectsFromWildcard"

	// TaskTypePullProjectsFromWildcards ..
	TaskTypePullProjectsFromWildcards TaskType = "PullProjectsFromWildcards"

	// TaskTypePullEnvironmentsFromProject ..
	TaskTypePullEnvironmentsFromProject TaskType = "PullEnvironmentsFromProject"

	// TaskTypePullEnvironmentsFromProjects ..
	TaskTypePullEnvironmentsFromProjects TaskType = "PullEnvironmentsFromProjects"

	// TaskTypePullEnvironmentMetrics ..
	TaskTypePullEnvironmentMetrics TaskType = "PullEnvironmentMetrics"

	// TaskTypePullMetrics ..
	TaskTypePullMetrics TaskType = "PullMetrics"

	// TaskTypePullRefsFromProject ..
	TaskTypePullRefsFromProject TaskType = "PullRefsFromProject"

	// TaskTypePullRefsFromProjects ..
	TaskTypePullRefsFromProjects TaskType = "PullRefsFromProjects"

	// TaskTypePullRefsFromPipelines ..
	TaskTypePullRefsFromPipelines TaskType = "PullRefsFromPipelines"

	// TaskTypePullRefMetrics ..
	TaskTypePullRefMetrics TaskType = "PullRefMetrics"

	// TaskTypeGarbageCollectProjects ..
	TaskTypeGarbageCollectProjects TaskType = "GarbageCollectProjects"

	// TaskTypeGarbageCollectEnvironments ..
	TaskTypeGarbageCollectEnvironments TaskType = "GarbageCollectEnvironments"

	// TaskTypeGarbageCollectRefs ..
	TaskTypeGarbageCollectRefs TaskType = "GarbageCollectRefs"

	// TaskTypeGarbageCollectMetrics ..
	TaskTypeGarbageCollectMetrics TaskType = "GarbageCollectMetrics"
)

// NewTaskController initializes and returns a new TaskController object
func NewTaskController(r *redis.Client) (t TaskController) {
	t.TaskMap = &taskq.TaskMap{}

	queueOptions := &taskq.QueueOptions{
		Name:                 "default",
		PauseErrorsThreshold: 3,
		Handler:              t.TaskMap,

		// Disable system resources checks
		MinSystemResources: taskq.SystemResources{
			Load1PerCPU:          -1,
			MemoryFreeMB:         0,
			MemoryFreePercentage: 0,
		},
	}

	if r != nil {
		t.Factory = redisq.NewFactory()
		queueOptions.Redis = r
	} else {
		t.Factory = memqueue.NewFactory()
	}

	t.Queue = t.Factory.RegisterQueue(queueOptions)

	// Purge the queue when we start
	// I am only partially convinced this will not cause issues in HA fashion
	if err := t.Queue.Purge(); err != nil {
		log.WithField("error", err.Error()).Error("purging the pulling queue")
	}

	if r != nil {
		if err := t.Factory.StartConsumers(context.TODO()); err != nil {
			log.WithError(err).Fatal("starting consuming the task queue")
		}
	}

	return
}

// TaskHandlerPullProjectsFromWildcard ..
func (c *Controller) TaskHandlerPullProjectsFromWildcard(ctx context.Context, w config.Wildcard) error {
	return c.PullProjectsFromWildcard(ctx, w)
}

// TaskHandlerPullEnvironmentsFromProject ..
func (c *Controller) TaskHandlerPullEnvironmentsFromProject(ctx context.Context, p config.Project) {
	// On errors, we do not want to retry these tasks
	if err := c.PullEnvironmentsFromProject(ctx, p); err != nil {
		log.WithFields(log.Fields{
			"project-name": p.Name,
			"error":        err.Error(),
		}).Warn("pulling environments from project")
	}
}

// TaskHandlerPullEnvironmentMetrics ..
func (c *Controller) TaskHandlerPullEnvironmentMetrics(env schemas.Environment) {
	// On errors, we do not want to retry these tasks
	if err := c.PullEnvironmentMetrics(env); err != nil {
		log.WithFields(log.Fields{
			"project-name":     env.ProjectName,
			"environment-name": env.Name,
			"environment-id":   env.ID,
			"error":            err.Error(),
		}).Warn("pulling environment metrics")
	}
}

// TaskHandlerPullRefsFromProject ..
func (c *Controller) TaskHandlerPullRefsFromProject(ctx context.Context, p config.Project) {
	// On errors, we do not want to retry these tasks
	if err := c.PullRefsFromProject(ctx, p); err != nil {
		log.WithFields(log.Fields{
			"project-name": p.Name,
			"error":        err.Error(),
		}).Warn("pulling refs from project")
	}
}

// TaskHandlerPullRefsFromPipelines ..
func (c *Controller) TaskHandlerPullRefsFromPipelines(ctx context.Context, p config.Project) {
	// On errors, we do not want to retry these tasks
	if err := c.PullRefsFromPipelines(ctx, p); err != nil {
		log.WithFields(log.Fields{
			"project-name": p.Name,
			"error":        err.Error(),
		}).Warn("pulling projects refs from pipelines")
	}
}

// TaskHandlerPullRefMetrics ..
func (c *Controller) TaskHandlerPullRefMetrics(ref schemas.Ref) {
	// On errors, we do not want to retry these tasks
	if err := c.PullRefMetrics(ref); err != nil {
		log.WithFields(log.Fields{
			"project-name": ref.ProjectName,
			"ref":          ref.Name,
			"error":        err.Error(),
		}).Warn("pulling ref metrics")
	}
}

// TaskHandlerPullProjectsFromWildcards ..
func (c *Controller) TaskHandlerPullProjectsFromWildcards(ctx context.Context) {
	log.WithFields(
		log.Fields{
			"wildcards-count": len(c.Wildcards),
		},
	).Info("scheduling projects from wildcards pull")

	for _, w := range c.Wildcards {
		c.ScheduleTask(ctx, TaskTypePullProjectsFromWildcard, w)
	}
}

// TaskHandlerPullEnvironmentsFromProjects ..
func (c *Controller) TaskHandlerPullEnvironmentsFromProjects(ctx context.Context) {
	projectsCount, err := c.Store.ProjectsCount()
	if err != nil {
		log.Error(err.Error())
	}

	log.WithFields(
		log.Fields{
			"projects-count": projectsCount,
		},
	).Info("scheduling environments from projects pull")

	projects, err := c.Store.Projects()
	if err != nil {
		log.Error(err)
	}

	for _, p := range projects {
		c.ScheduleTask(ctx, TaskTypePullEnvironmentsFromProject, p)
	}
}

// TaskHandlerPullRefsFromProjects ..
func (c *Controller) TaskHandlerPullRefsFromProjects(ctx context.Context) {
	projectsCount, err := c.Store.ProjectsCount()
	if err != nil {
		log.Error(err.Error())
	}

	log.WithFields(
		log.Fields{
			"projects-count": projectsCount,
		},
	).Info("scheduling refs from projects pull")

	projects, err := c.Store.Projects()
	if err != nil {
		log.Error(err)
	}

	for _, p := range projects {
		c.ScheduleTask(ctx, TaskTypePullRefsFromProject, p)
	}
}

// TaskHandlerPullMetrics ..
func (c *Controller) TaskHandlerPullMetrics(ctx context.Context) {
	refsCount, err := c.Store.RefsCount()
	if err != nil {
		log.Error(err)
	}

	envsCount, err := c.Store.EnvironmentsCount()
	if err != nil {
		log.Error(err)
	}

	log.WithFields(
		log.Fields{
			"environments-count": envsCount,
			"refs-count":         refsCount,
		},
	).Info("scheduling metrics pull")

	// ENVIRONMENTS
	envs, err := c.Store.Environments()
	if err != nil {
		log.Error(err)
	}

	for _, env := range envs {
		c.ScheduleTask(ctx, TaskTypePullEnvironmentMetrics, env)
	}

	// REFS
	refs, err := c.Store.Refs()
	if err != nil {
		log.Error(err)
	}

	for _, ref := range refs {
		c.ScheduleTask(ctx, TaskTypePullRefMetrics, ref)
	}
}

// TaskHandlerGarbageCollectProjects ..
func (c *Controller) TaskHandlerGarbageCollectProjects(ctx context.Context) error {
	return c.GarbageCollectProjects(ctx)
}

// TaskHandlerGarbageCollectEnvironments ..
func (c *Controller) TaskHandlerGarbageCollectEnvironments(ctx context.Context) error {
	return c.GarbageCollectEnvironments(ctx)
}

// TaskHandlerGarbageCollectRefs ..
func (c *Controller) TaskHandlerGarbageCollectRefs(ctx context.Context) error {
	return c.GarbageCollectRefs(ctx)
}

// TaskHandlerGarbageCollectMetrics ..
func (c *Controller) TaskHandlerGarbageCollectMetrics(ctx context.Context) error {
	return c.GarbageCollectMetrics(ctx)
}

// Schedule ..
func (c *Controller) Schedule(ctx context.Context, pull config.Pull, gc config.GarbageCollect) {
	for tt, cfg := range map[TaskType]config.SchedulerConfig{
		TaskTypePullProjectsFromWildcards:    config.SchedulerConfig(pull.ProjectsFromWildcards),
		TaskTypePullEnvironmentsFromProjects: config.SchedulerConfig(pull.EnvironmentsFromProjects),
		TaskTypePullRefsFromProjects:         config.SchedulerConfig(pull.RefsFromProjects),
		TaskTypePullMetrics:                  config.SchedulerConfig(pull.Metrics),
		TaskTypeGarbageCollectProjects:       config.SchedulerConfig(gc.Projects),
		TaskTypeGarbageCollectEnvironments:   config.SchedulerConfig(gc.Environments),
		TaskTypeGarbageCollectRefs:           config.SchedulerConfig(gc.Refs),
		TaskTypeGarbageCollectMetrics:        config.SchedulerConfig(gc.Metrics),
	} {
		if cfg.OnInit {
			c.ScheduleTask(ctx, tt)
		}

		if cfg.Scheduled {
			c.ScheduleTaskWithTicker(ctx, tt, cfg.IntervalSeconds)
		}
	}
}

// ScheduleTask ..
func (c *Controller) ScheduleTask(ctx context.Context, tt TaskType, args ...interface{}) {
	task := c.TaskController.TaskMap.Get(string(tt))
	msg := task.WithArgs(ctx, args...)
	if err := c.TaskController.Queue.Add(msg); err != nil {
		log.WithError(err).Warning("scheduling task")
	}
}

// ScheduleTaskWithTicker ..
func (c *Controller) ScheduleTaskWithTicker(ctx context.Context, tt TaskType, intervalSeconds int) {
	if intervalSeconds <= 0 {
		log.WithField("task", tt).Warn("task scheduling misconfigured, currently disabled")
		return
	}

	log.WithFields(log.Fields{
		"task":             tt,
		"interval_seconds": intervalSeconds,
	}).Debug("task scheduled")

	go func(ctx context.Context) {
		ticker := time.NewTicker(time.Duration(intervalSeconds) * time.Second)
		for {
			select {
			case <-ctx.Done():
				log.WithField("task", tt).Info("scheduling of task stopped")
				return
			case <-ticker.C:
				switch tt {
				default:
					c.ScheduleTask(ctx, tt)
				}
			}
		}
	}(ctx)
}
