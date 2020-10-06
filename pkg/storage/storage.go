package storage

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
)

// Storage ..
type Storage interface {
	SetProject(schemas.Project) error
	DelProject(schemas.ProjectKey) error
	GetProject(*schemas.Project) error
	ProjectExists(schemas.ProjectKey) (bool, error)
	Projects() (schemas.Projects, error)
	ProjectsCount() (int64, error)

	SetProjectRef(schemas.ProjectRef) error
	DelProjectRef(schemas.ProjectRefKey) error
	GetProjectRef(*schemas.ProjectRef) error
	ProjectRefExists(schemas.ProjectRefKey) (bool, error)
	ProjectsRefs() (schemas.ProjectsRefs, error)
	ProjectsRefsCount() (int64, error)

	SetMetric(schemas.Metric) error
	DelMetric(schemas.MetricKey) error
	GetMetric(*schemas.Metric) error
	MetricExists(schemas.MetricKey) (bool, error)
	Metrics() (schemas.Metrics, error)
	MetricsCount() (int64, error)
}

// NewLocalStorage ..
func NewLocalStorage() Storage {
	return &Local{
		projects:     make(schemas.Projects),
		projectsRefs: make(schemas.ProjectsRefs),
		metrics:      make(schemas.Metrics),
	}
}

// NewRedisStorage ..
func NewRedisStorage(client *redis.Client) Storage {
	return &Redis{
		Client: client,
		ctx:    context.TODO(),
	}
}
