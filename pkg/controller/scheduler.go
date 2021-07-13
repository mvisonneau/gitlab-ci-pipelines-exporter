package controller

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/config"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	log "github.com/sirupsen/logrus"
	"github.com/vmihailenco/taskq/v3"
	"github.com/vmihailenco/taskq/v3/memqueue"
	"github.com/vmihailenco/taskq/v3/redisq"
)

const bufferSize = 1000

// TaskController holds task related clients
type TaskController struct {
	Factory taskq.Factory
	Queue   taskq.Queue
	TaskMap *taskq.TaskMap
}

// NewTaskController initializes and returns a new TaskController object
func NewTaskController(r *redis.Client) (t TaskController) {
	t.TaskMap = &taskq.TaskMap{}

	queueOptions := &taskq.QueueOptions{
		Name:                 "default",
		PauseErrorsThreshold: 3,
		Handler:              t.TaskMap,
		BufferSize:           bufferSize,

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
func (c *Controller) TaskHandlerPullProjectsFromWildcard(ctx context.Context, id string, w config.Wildcard) error {
	defer c.unqueueTask(schemas.TaskTypePullProjectsFromWildcard, id)

	return c.PullProjectsFromWildcard(ctx, w)
}

// TaskHandlerPullEnvironmentsFromProject ..
func (c *Controller) TaskHandlerPullEnvironmentsFromProject(ctx context.Context, p schemas.Project) {
	defer c.unqueueTask(schemas.TaskTypePullEnvironmentsFromProject, string(p.Key()))

	// On errors, we do not want to retry these tasks
	if p.Pull.Environments.Enabled {
		if err := c.PullEnvironmentsFromProject(ctx, p); err != nil {
			log.WithFields(log.Fields{
				"project-name": p.Name,
				"error":        err.Error(),
			}).Warn("pulling environments from project")
		}
	}
}

// TaskHandlerPullEnvironmentMetrics ..
func (c *Controller) TaskHandlerPullEnvironmentMetrics(env schemas.Environment) {
	defer c.unqueueTask(schemas.TaskTypePullEnvironmentMetrics, string(env.Key()))

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
func (c *Controller) TaskHandlerPullRefsFromProject(ctx context.Context, p schemas.Project) {
	defer c.unqueueTask(schemas.TaskTypePullRefsFromProject, string(p.Key()))

	// On errors, we do not want to retry these tasks
	if err := c.PullRefsFromProject(ctx, p); err != nil {
		log.WithFields(log.Fields{
			"project-name": p.Name,
			"error":        err.Error(),
		}).Warn("pulling refs from project")
	}
}

// TaskHandlerPullRefMetrics ..
func (c *Controller) TaskHandlerPullRefMetrics(ref schemas.Ref) {
	defer c.unqueueTask(schemas.TaskTypePullRefMetrics, string(ref.Key()))

	// On errors, we do not want to retry these tasks
	if err := c.PullRefMetrics(ref); err != nil {
		log.WithFields(log.Fields{
			"project-name": ref.Project.Name,
			"ref":          ref.Name,
			"error":        err.Error(),
		}).Warn("pulling ref metrics")
	}
}

// TaskHandlerPullProjectsFromWildcards ..
func (c *Controller) TaskHandlerPullProjectsFromWildcards(ctx context.Context) {
	defer c.unqueueTask(schemas.TaskTypePullProjectsFromWildcards, "_")

	log.WithFields(
		log.Fields{
			"wildcards-count": len(c.Config.Wildcards),
		},
	).Info("scheduling projects from wildcards pull")

	for id, w := range c.Config.Wildcards {
		c.ScheduleTask(ctx, schemas.TaskTypePullProjectsFromWildcard, strconv.Itoa(id), strconv.Itoa(id), w)
	}
}

// TaskHandlerPullEnvironmentsFromProjects ..
func (c *Controller) TaskHandlerPullEnvironmentsFromProjects(ctx context.Context) {
	defer c.unqueueTask(schemas.TaskTypePullEnvironmentsFromProjects, "_")

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
		c.ScheduleTask(ctx, schemas.TaskTypePullEnvironmentsFromProject, string(p.Key()), p)
	}
}

// TaskHandlerPullRefsFromProjects ..
func (c *Controller) TaskHandlerPullRefsFromProjects(ctx context.Context) {
	defer c.unqueueTask(schemas.TaskTypePullRefsFromProjects, "_")

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
		c.ScheduleTask(ctx, schemas.TaskTypePullRefsFromProject, string(p.Key()), p)
	}
}

// TaskHandlerPullMetrics ..
func (c *Controller) TaskHandlerPullMetrics(ctx context.Context) {
	defer c.unqueueTask(schemas.TaskTypePullMetrics, "_")

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
		c.ScheduleTask(ctx, schemas.TaskTypePullEnvironmentMetrics, string(env.Key()), env)
	}

	// REFS
	refs, err := c.Store.Refs()
	if err != nil {
		log.Error(err)
	}

	for _, ref := range refs {
		c.ScheduleTask(ctx, schemas.TaskTypePullRefMetrics, string(ref.Key()), ref)
	}
}

// TaskHandlerGarbageCollectProjects ..
func (c *Controller) TaskHandlerGarbageCollectProjects(ctx context.Context) error {
	defer c.unqueueTask(schemas.TaskTypeGarbageCollectProjects, "_")
	return c.GarbageCollectProjects(ctx)
}

// TaskHandlerGarbageCollectEnvironments ..
func (c *Controller) TaskHandlerGarbageCollectEnvironments(ctx context.Context) error {
	defer c.unqueueTask(schemas.TaskTypeGarbageCollectEnvironments, "_")
	return c.GarbageCollectEnvironments(ctx)
}

// TaskHandlerGarbageCollectRefs ..
func (c *Controller) TaskHandlerGarbageCollectRefs(ctx context.Context) error {
	defer c.unqueueTask(schemas.TaskTypeGarbageCollectRefs, "_")
	return c.GarbageCollectRefs(ctx)
}

// TaskHandlerGarbageCollectMetrics ..
func (c *Controller) TaskHandlerGarbageCollectMetrics(ctx context.Context) error {
	defer c.unqueueTask(schemas.TaskTypeGarbageCollectMetrics, "_")
	return c.GarbageCollectMetrics(ctx)
}

// Schedule ..
func (c *Controller) Schedule(ctx context.Context, pull config.Pull, gc config.GarbageCollect) {
	for tt, cfg := range map[schemas.TaskType]config.SchedulerConfig{
		schemas.TaskTypePullProjectsFromWildcards:    config.SchedulerConfig(pull.ProjectsFromWildcards),
		schemas.TaskTypePullEnvironmentsFromProjects: config.SchedulerConfig(pull.EnvironmentsFromProjects),
		schemas.TaskTypePullRefsFromProjects:         config.SchedulerConfig(pull.RefsFromProjects),
		schemas.TaskTypePullMetrics:                  config.SchedulerConfig(pull.Metrics),
		schemas.TaskTypeGarbageCollectProjects:       config.SchedulerConfig(gc.Projects),
		schemas.TaskTypeGarbageCollectEnvironments:   config.SchedulerConfig(gc.Environments),
		schemas.TaskTypeGarbageCollectRefs:           config.SchedulerConfig(gc.Refs),
		schemas.TaskTypeGarbageCollectMetrics:        config.SchedulerConfig(gc.Metrics),
	} {
		if cfg.OnInit {
			c.ScheduleTask(ctx, tt, "_")
		}

		if cfg.Scheduled {
			c.ScheduleTaskWithTicker(ctx, tt, cfg.IntervalSeconds)
		}
	}
}

// ScheduleTask ..
func (c *Controller) ScheduleTask(ctx context.Context, tt schemas.TaskType, uniqueID string, args ...interface{}) {
	logFields := log.Fields{
		"task_type":      tt,
		"task_unique_id": uniqueID,
	}
	task := c.TaskController.TaskMap.Get(string(tt))
	msg := task.WithArgs(ctx, args...)

	qlen, err := c.TaskController.Queue.Len()
	if err != nil {
		log.WithFields(logFields).Warn("unable to read task queue length, skipping scheduling of task..")
		return
	}

	if qlen >= c.TaskController.Queue.Options().BufferSize {
		log.WithFields(logFields).Warn("queue buffer size exhausted, skipping scheduling of task..")
		return
	}

	queued, err := c.Store.QueueTask(tt, uniqueID)
	if err != nil {
		log.WithFields(logFields).Warn("unable to declare the queueing, skipping scheduling of task..")
		return
	}

	if !queued {
		log.WithFields(logFields).Debug("task already queued, skipping scheduling of task..")
		return
	}

	go func(msg *taskq.Message) {
		if err := c.TaskController.Queue.Add(msg); err != nil {
			log.WithError(err).Warning("scheduling task")
		}
	}(msg)
}

// ScheduleTaskWithTicker ..
func (c *Controller) ScheduleTaskWithTicker(ctx context.Context, tt schemas.TaskType, intervalSeconds int) {
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
					c.ScheduleTask(ctx, tt, "_")
				}
			}
		}
	}(ctx)
}
