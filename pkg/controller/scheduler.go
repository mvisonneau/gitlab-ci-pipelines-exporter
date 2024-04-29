package controller

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
	"github.com/vmihailenco/taskq/memqueue/v4"
	"github.com/vmihailenco/taskq/redisq/v4"
	"github.com/vmihailenco/taskq/v4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/config"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/monitor"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/store"
)

// TaskController holds task related clients.
type TaskController struct {
	Factory                  taskq.Factory
	Queue                    taskq.Queue
	TaskMap                  *taskq.TaskMap
	TaskSchedulingMonitoring map[schemas.TaskType]*monitor.TaskSchedulingStatus
}

// NewTaskController initializes and returns a new TaskController object.
func NewTaskController(ctx context.Context, r *redis.Client, maximumJobsQueueSize int) (t TaskController) {
	ctx, span := otel.Tracer(tracerName).Start(ctx, "controller:NewTaskController")
	defer span.End()

	t.TaskMap = &taskq.TaskMap{}

	queueOptions := &taskq.QueueConfig{
		Name:                 "default",
		PauseErrorsThreshold: 3,
		Handler:              t.TaskMap,
		BufferSize:           maximumJobsQueueSize,
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
	if err := t.Queue.Purge(ctx); err != nil {
		log.WithContext(ctx).
			WithError(err).
			Error("purging the pulling queue")
	}

	if r != nil {
		if err := t.Factory.StartConsumers(context.TODO()); err != nil {
			log.WithContext(ctx).
				WithError(err).
				Fatal("starting consuming the task queue")
		}
	}

	t.TaskSchedulingMonitoring = make(map[schemas.TaskType]*monitor.TaskSchedulingStatus)

	return
}

// TaskHandlerPullProject ..
func (c *Controller) TaskHandlerPullProject(ctx context.Context, name string, pull config.ProjectPull) error {
	defer c.unqueueTask(ctx, schemas.TaskTypePullProject, name)

	return c.PullProject(ctx, name, pull)
}

// TaskHandlerPullProjectsFromWildcard ..
func (c *Controller) TaskHandlerPullProjectsFromWildcard(ctx context.Context, id string, w config.Wildcard) error {
	defer c.unqueueTask(ctx, schemas.TaskTypePullProjectsFromWildcard, id)

	return c.PullProjectsFromWildcard(ctx, w)
}

// TaskHandlerPullEnvironmentsFromProject ..
func (c *Controller) TaskHandlerPullEnvironmentsFromProject(ctx context.Context, p schemas.Project) {
	defer c.unqueueTask(ctx, schemas.TaskTypePullEnvironmentsFromProject, string(p.Key()))

	// On errors, we do not want to retry these tasks
	if p.Pull.Environments.Enabled {
		if err := c.PullEnvironmentsFromProject(ctx, p); err != nil {
			log.WithContext(ctx).
				WithFields(log.Fields{
					"project-name": p.Name,
				}).
				WithError(err).
				Warn("pulling environments from project")
		}
	}
}

// TaskHandlerPullEnvironmentMetrics ..
func (c *Controller) TaskHandlerPullEnvironmentMetrics(ctx context.Context, env schemas.Environment) {
	defer c.unqueueTask(ctx, schemas.TaskTypePullEnvironmentMetrics, string(env.Key()))

	// On errors, we do not want to retry these tasks
	if err := c.PullEnvironmentMetrics(ctx, env); err != nil {
		log.WithContext(ctx).
			WithFields(log.Fields{
				"project-name":     env.ProjectName,
				"environment-name": env.Name,
				"environment-id":   env.ID,
			}).
			WithError(err).
			Warn("pulling environment metrics")
	}
}

// TaskHandlerPullRefsFromProject ..
func (c *Controller) TaskHandlerPullRefsFromProject(ctx context.Context, p schemas.Project) {
	defer c.unqueueTask(ctx, schemas.TaskTypePullRefsFromProject, string(p.Key()))

	// On errors, we do not want to retry these tasks
	if err := c.PullRefsFromProject(ctx, p); err != nil {
		log.WithContext(ctx).
			WithFields(log.Fields{
				"project-name": p.Name,
			}).
			WithError(err).
			Warn("pulling refs from project")
	}
}

// TaskHandlerPullRefMetrics ..
func (c *Controller) TaskHandlerPullRefMetrics(ctx context.Context, ref schemas.Ref) {
	defer c.unqueueTask(ctx, schemas.TaskTypePullRefMetrics, string(ref.Key()))

	// On errors, we do not want to retry these tasks
	if err := c.PullRefMetrics(ctx, ref); err != nil {
		log.WithContext(ctx).
			WithFields(log.Fields{
				"project-name": ref.Project.Name,
				"ref":          ref.Name,
			}).
			WithError(err).
			Warn("pulling ref metrics")
	}
}

// TaskHandlerPullProjectsFromWildcards ..
func (c *Controller) TaskHandlerPullProjectsFromWildcards(ctx context.Context) {
	defer c.unqueueTask(ctx, schemas.TaskTypePullProjectsFromWildcards, "_")
	defer c.TaskController.monitorLastTaskScheduling(schemas.TaskTypePullProjectsFromWildcards)

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
	defer c.unqueueTask(ctx, schemas.TaskTypePullEnvironmentsFromProjects, "_")
	defer c.TaskController.monitorLastTaskScheduling(schemas.TaskTypePullEnvironmentsFromProjects)

	projectsCount, err := c.Store.ProjectsCount(ctx)
	if err != nil {
		log.WithContext(ctx).
			WithError(err).
			Error()
	}

	log.WithFields(
		log.Fields{
			"projects-count": projectsCount,
		},
	).Info("scheduling environments from projects pull")

	projects, err := c.Store.Projects(ctx)
	if err != nil {
		log.WithContext(ctx).
			WithError(err).
			Error()
	}

	for _, p := range projects {
		c.ScheduleTask(ctx, schemas.TaskTypePullEnvironmentsFromProject, string(p.Key()), p)
	}
}

// TaskHandlerPullRefsFromProjects ..
func (c *Controller) TaskHandlerPullRefsFromProjects(ctx context.Context) {
	defer c.unqueueTask(ctx, schemas.TaskTypePullRefsFromProjects, "_")
	defer c.TaskController.monitorLastTaskScheduling(schemas.TaskTypePullRefsFromProjects)

	projectsCount, err := c.Store.ProjectsCount(ctx)
	if err != nil {
		log.WithContext(ctx).
			WithError(err).
			Error()
	}

	log.WithFields(
		log.Fields{
			"projects-count": projectsCount,
		},
	).Info("scheduling refs from projects pull")

	projects, err := c.Store.Projects(ctx)
	if err != nil {
		log.WithContext(ctx).
			WithError(err).
			Error()
	}

	for _, p := range projects {
		c.ScheduleTask(ctx, schemas.TaskTypePullRefsFromProject, string(p.Key()), p)
	}
}

// TaskHandlerPullMetrics ..
func (c *Controller) TaskHandlerPullMetrics(ctx context.Context) {
	defer c.unqueueTask(ctx, schemas.TaskTypePullMetrics, "_")
	defer c.TaskController.monitorLastTaskScheduling(schemas.TaskTypePullMetrics)

	refsCount, err := c.Store.RefsCount(ctx)
	if err != nil {
		log.WithContext(ctx).
			WithError(err).
			Error()
	}

	envsCount, err := c.Store.EnvironmentsCount(ctx)
	if err != nil {
		log.WithContext(ctx).
			WithError(err).
			Error()
	}

	log.WithFields(
		log.Fields{
			"environments-count": envsCount,
			"refs-count":         refsCount,
		},
	).Info("scheduling metrics pull")

	// ENVIRONMENTS
	envs, err := c.Store.Environments(ctx)
	if err != nil {
		log.WithContext(ctx).
			WithError(err).
			Error()
	}

	for _, env := range envs {
		c.ScheduleTask(ctx, schemas.TaskTypePullEnvironmentMetrics, string(env.Key()), env)
	}

	// REFS
	refs, err := c.Store.Refs(ctx)
	if err != nil {
		log.WithContext(ctx).
			WithError(err).
			Error()
	}

	for _, ref := range refs {
		c.ScheduleTask(ctx, schemas.TaskTypePullRefMetrics, string(ref.Key()), ref)
	}
}

// TaskHandlerGarbageCollectProjects ..
func (c *Controller) TaskHandlerGarbageCollectProjects(ctx context.Context) error {
	defer c.unqueueTask(ctx, schemas.TaskTypeGarbageCollectProjects, "_")
	defer c.TaskController.monitorLastTaskScheduling(schemas.TaskTypeGarbageCollectProjects)

	return c.GarbageCollectProjects(ctx)
}

// TaskHandlerGarbageCollectEnvironments ..
func (c *Controller) TaskHandlerGarbageCollectEnvironments(ctx context.Context) error {
	defer c.unqueueTask(ctx, schemas.TaskTypeGarbageCollectEnvironments, "_")
	defer c.TaskController.monitorLastTaskScheduling(schemas.TaskTypeGarbageCollectEnvironments)

	return c.GarbageCollectEnvironments(ctx)
}

// TaskHandlerGarbageCollectRefs ..
func (c *Controller) TaskHandlerGarbageCollectRefs(ctx context.Context) error {
	defer c.unqueueTask(ctx, schemas.TaskTypeGarbageCollectRefs, "_")
	defer c.TaskController.monitorLastTaskScheduling(schemas.TaskTypeGarbageCollectRefs)

	return c.GarbageCollectRefs(ctx)
}

// TaskHandlerGarbageCollectMetrics ..
func (c *Controller) TaskHandlerGarbageCollectMetrics(ctx context.Context) error {
	defer c.unqueueTask(ctx, schemas.TaskTypeGarbageCollectMetrics, "_")
	defer c.TaskController.monitorLastTaskScheduling(schemas.TaskTypeGarbageCollectMetrics)

	return c.GarbageCollectMetrics(ctx)
}

// Schedule ..
func (c *Controller) Schedule(ctx context.Context, pull config.Pull, gc config.GarbageCollect) {
	ctx, span := otel.Tracer(tracerName).Start(ctx, "controller:Schedule")
	defer span.End()

	go func() {
		c.GetGitLabMetadata(ctx)
	}()

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

		if c.Redis != nil {
			c.ScheduleRedisSetKeepalive(ctx)
		}
	}
}

// ScheduleRedisSetKeepalive will ensure that whilst the process is running,
// a key is periodically updated within Redis to let other instances know this
// one is alive and processing tasks.
func (c *Controller) ScheduleRedisSetKeepalive(ctx context.Context) {
	ctx, span := otel.Tracer(tracerName).Start(ctx, "controller:ScheduleRedisSetKeepalive")
	defer span.End()

	go func(ctx context.Context) {
		ticker := time.NewTicker(time.Duration(5) * time.Second)

		for {
			select {
			case <-ctx.Done():
				log.Info("stopped redis keepalive")

				return
			case <-ticker.C:
				if _, err := c.Store.(*store.Redis).SetKeepalive(ctx, c.UUID.String(), time.Duration(10)*time.Second); err != nil {
					log.WithContext(ctx).
						WithError(err).
						Fatal("setting keepalive")
				}
			}
		}
	}(ctx)
}

// ScheduleTask ..
func (c *Controller) ScheduleTask(ctx context.Context, tt schemas.TaskType, uniqueID string, args ...interface{}) {
	ctx, span := otel.Tracer(tracerName).Start(ctx, "controller:ScheduleTask")
	defer span.End()

	span.SetAttributes(attribute.String("task_type", string(tt)))
	span.SetAttributes(attribute.String("task_unique_id", uniqueID))

	logFields := log.Fields{
		"task_type":      tt,
		"task_unique_id": uniqueID,
	}
	task := c.TaskController.TaskMap.Get(string(tt))
	msg := task.NewJob(args...)

	qlen, err := c.TaskController.Queue.Len(ctx)
	if err != nil {
		log.WithContext(ctx).
			WithFields(logFields).
			Warn("unable to read task queue length, skipping scheduling of task..")

		return
	}

	if qlen >= c.TaskController.Queue.Options().BufferSize {
		log.WithContext(ctx).
			WithFields(logFields).
			Warn("queue buffer size exhausted, skipping scheduling of task..")

		return
	}

	queued, err := c.Store.QueueTask(ctx, tt, uniqueID, c.UUID.String())
	if err != nil {
		log.WithContext(ctx).
			WithFields(logFields).
			Warn("unable to declare the queueing, skipping scheduling of task..")

		return
	}

	if !queued {
		log.WithFields(logFields).
			Debug("task already queued, skipping scheduling of task..")

		return
	}

	go func(job *taskq.Job) {
		if err := c.TaskController.Queue.AddJob(ctx, job); err != nil {
			log.WithContext(ctx).
				WithError(err).
				Warn("scheduling task")
		}
	}(msg)
}

// ScheduleTaskWithTicker ..
func (c *Controller) ScheduleTaskWithTicker(ctx context.Context, tt schemas.TaskType, intervalSeconds int) {
	ctx, span := otel.Tracer(tracerName).Start(ctx, "controller:ScheduleTaskWithTicker")
	defer span.End()
	span.SetAttributes(attribute.String("task_type", string(tt)))
	span.SetAttributes(attribute.Int("interval_seconds", intervalSeconds))

	if intervalSeconds <= 0 {
		log.WithContext(ctx).
			WithField("task", tt).
			Warn("task scheduling misconfigured, currently disabled")

		return
	}

	log.WithFields(log.Fields{
		"task":             tt,
		"interval_seconds": intervalSeconds,
	}).Debug("task scheduled")

	c.TaskController.monitorNextTaskScheduling(tt, intervalSeconds)

	go func(ctx context.Context) {
		ticker := time.NewTicker(time.Duration(intervalSeconds) * time.Second)

		for {
			select {
			case <-ctx.Done():
				log.WithField("task", tt).Info("scheduling of task stopped")

				return
			case <-ticker.C:
				c.ScheduleTask(ctx, tt, "_")
				c.TaskController.monitorNextTaskScheduling(tt, intervalSeconds)
			}
		}
	}(ctx)
}

func (tc *TaskController) monitorNextTaskScheduling(tt schemas.TaskType, duration int) {
	if _, ok := tc.TaskSchedulingMonitoring[tt]; !ok {
		tc.TaskSchedulingMonitoring[tt] = &monitor.TaskSchedulingStatus{}
	}

	tc.TaskSchedulingMonitoring[tt].Next = time.Now().Add(time.Duration(duration) * time.Second)
}

func (tc *TaskController) monitorLastTaskScheduling(tt schemas.TaskType) {
	if _, ok := tc.TaskSchedulingMonitoring[tt]; !ok {
		tc.TaskSchedulingMonitoring[tt] = &monitor.TaskSchedulingStatus{}
	}

	tc.TaskSchedulingMonitoring[tt].Last = time.Now()
}
