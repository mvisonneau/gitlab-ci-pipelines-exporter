package store

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/config"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	log "github.com/sirupsen/logrus"
)

// Store ..
type Store interface {
	SetProject(config.Project) error
	DelProject(config.ProjectKey) error
	GetProject(*config.Project) error
	ProjectExists(config.ProjectKey) (bool, error)
	Projects() (config.Projects, error)
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

// NewLocalStore ..
func NewLocalStore() Store {
	return &Local{
		projects:     make(config.Projects),
		environments: make(schemas.Environments),
		refs:         make(schemas.Refs),
		metrics:      make(schemas.Metrics),
	}
}

// NewRedisStore ..
func NewRedisStore(client *redis.Client) Store {
	return &Redis{
		Client: client,
		ctx:    context.TODO(),
	}
}

// New creates a new store and populates it with
// provided []config.Project
func New(
	r *redis.Client,
	projects []config.Project,
) (s Store) {
	if r != nil {
		s = NewRedisStore(r)
	} else {
		s = NewLocalStore()
	}

	// Load all the configured projects in the store
	for _, p := range projects {
		exists, err := s.ProjectExists(p.Key())
		if err != nil {
			log.WithFields(log.Fields{
				"project-name": p.Name,
				"error":        err.Error(),
			}).Error("reading project from the store")
		}

		if !exists {
			if err = s.SetProject(p); err != nil {
				log.WithFields(log.Fields{
					"project-name": p.Name,
					"error":        err.Error(),
				}).Error("writing project in the store")
			}
		}
	}

	return
}
