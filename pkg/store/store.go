package store

import (
	"context"

	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/config"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
)

// Store ..
type Store interface {
	SetProject(ctx context.Context, p schemas.Project) error
	DelProject(ctx context.Context, pk schemas.ProjectKey) error
	GetProject(ctx context.Context, p *schemas.Project) error
	ProjectExists(ctx context.Context, pk schemas.ProjectKey) (bool, error)
	Projects(ctx context.Context) (schemas.Projects, error)
	ProjectsCount(ctx context.Context) (int64, error)
	SetEnvironment(ctx context.Context, e schemas.Environment) error
	DelEnvironment(ctx context.Context, ek schemas.EnvironmentKey) error
	GetEnvironment(ctx context.Context, e *schemas.Environment) error
	EnvironmentExists(ctx context.Context, ek schemas.EnvironmentKey) (bool, error)
	Environments(ctx context.Context) (schemas.Environments, error)
	EnvironmentsCount(ctx context.Context) (int64, error)
	SetRef(ctx context.Context, r schemas.Ref) error
	DelRef(ctx context.Context, rk schemas.RefKey) error
	GetRef(ctx context.Context, r *schemas.Ref) error
	RefExists(ctx context.Context, rk schemas.RefKey) (bool, error)
	Refs(ctx context.Context) (schemas.Refs, error)
	RefsCount(ctx context.Context) (int64, error)
	SetMetric(ctx context.Context, m schemas.Metric) error
	DelMetric(ctx context.Context, mk schemas.MetricKey) error
	GetMetric(ctx context.Context, m *schemas.Metric) error
	MetricExists(ctx context.Context, mk schemas.MetricKey) (bool, error)
	Metrics(ctx context.Context) (schemas.Metrics, error)
	MetricsCount(ctx context.Context) (int64, error)

	// Helpers to keep track of currently queued tasks and avoid scheduling them
	// twice at the risk of ending up with loads of dangling goroutines being locked
	QueueTask(ctx context.Context, tt schemas.TaskType, taskUUID, processUUID string) (bool, error)
	UnqueueTask(ctx context.Context, tt schemas.TaskType, taskUUID string) error
	CurrentlyQueuedTasksCount(ctx context.Context) (uint64, error)
	ExecutedTasksCount(ctx context.Context) (uint64, error)
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
