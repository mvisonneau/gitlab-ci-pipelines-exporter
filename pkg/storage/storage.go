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

	SetEnvironment(schemas.Environment) error
	DelEnvironment(schemas.EnvironmentKey) error
	GetEnvironment(*schemas.Environment) error
	EnvironmentExists(schemas.EnvironmentKey) (bool, error)
	Environments() (schemas.Environments, error)
	EnvironmentsCount() (int64, error)

	SetRef(schemas.Ref) error
	DelRef(schemas.RefKey) error
	GetRef(*schemas.Ref) error
	RefExists(schemas.RefKey) (bool, error)
	Refs() (schemas.Refs, error)
	RefsCount() (int64, error)

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
		environments: make(schemas.Environments),
		refs:         make(schemas.Refs),
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
