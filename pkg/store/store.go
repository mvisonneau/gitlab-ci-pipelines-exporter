package store

import (
	"context"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/config"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
)

// Store ..
type Store interface {
	SetProject(context.Context, schemas.Project) error
	DelProject(context.Context, schemas.ProjectKey) error
	GetProject(context.Context, *schemas.Project) error
	ProjectExists(context.Context, schemas.ProjectKey) (bool, error)
	Projects(context.Context) (schemas.Projects, error)
	ProjectsCount(context.Context) (int64, error)
	SetEnvironment(context.Context, schemas.Environment) error
	DelEnvironment(context.Context, schemas.EnvironmentKey) error
	GetEnvironment(context.Context, *schemas.Environment) error
	EnvironmentExists(context.Context, schemas.EnvironmentKey) (bool, error)
	Environments(context.Context) (schemas.Environments, error)
	EnvironmentsCount(context.Context) (int64, error)
	SetRef(context.Context, schemas.Ref) error
	DelRef(context.Context, schemas.RefKey) error
	GetRef(context.Context, *schemas.Ref) error
	RefExists(context.Context, schemas.RefKey) (bool, error)
	Refs(context.Context) (schemas.Refs, error)
	RefsCount(context.Context) (int64, error)
	SetMetric(context.Context, schemas.Metric) error
	DelMetric(context.Context, schemas.MetricKey) error
	GetMetric(context.Context, *schemas.Metric) error
	MetricExists(context.Context, schemas.MetricKey) (bool, error)
	Metrics(context.Context) (schemas.Metrics, error)
	MetricsCount(context.Context) (int64, error)

	// Helpers to keep track of currently queued tasks and avoid scheduling them
	// twice at the risk of ending up with loads of dangling goroutines being locked
	QueueTask(context.Context, schemas.TaskType, string, string) (bool, error)
	UnqueueTask(context.Context, schemas.TaskType, string) error
	CurrentlyQueuedTasksCount(context.Context) (uint64, error)
	ExecutedTasksCount(context.Context) (uint64, error)
}

// NewLocalStore ..
func NewLocalStore() Store {
	return &Local{
		projects:     make(schemas.Projects),
		environments: make(schemas.Environments),
		refs:         make(schemas.Refs),
		metrics:      make(schemas.Metrics),
	}
}

// NewRedisStore ..
func NewRedisStore(client *redis.Client) Store {
	return &Redis{
		Client: client,
	}
}

// New creates a new store and populates it with
// provided []schemas.Project.
func New(
	ctx context.Context,
	r *redis.Client,
	projects config.Projects,
) (s Store) {
	ctx, span := otel.Tracer("gitlab-ci-pipelines-exporter").Start(ctx, "store:New")
	defer span.End()

	if r != nil {
		s = NewRedisStore(r)
	} else {
		s = NewLocalStore()
	}

	// Load all the configured projects in the store
	for _, p := range projects {
		sp := schemas.Project{Project: p}

		exists, err := s.ProjectExists(ctx, sp.Key())
		if err != nil {
			log.WithContext(ctx).
				WithFields(log.Fields{
					"project-name": p.Name,
				}).
				WithError(err).
				Error("reading project from the store")
		}

		if !exists {
			if err = s.SetProject(ctx, sp); err != nil {
				log.WithContext(ctx).
					WithFields(log.Fields{
						"project-name": p.Name,
					}).
					WithError(err).
					Error("writing project in the store")
			}
		}
	}

	return
}
