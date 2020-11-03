package storage

import (
	"testing"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	"github.com/openlyinc/pointy"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestLocalProjectFunctions(t *testing.T) {
	p := schemas.Project{
		Name: "foo/bar",
		ProjectParameters: schemas.ProjectParameters{
			OutputSparseStatusMetricsValue: pointy.Bool(false),
		},
	}

	l := NewLocalStorage()
	l.SetProject(p)

	// Set project
	projects, err := l.Projects()
	assert.NoError(t, err)
	assert.Contains(t, projects, p.Key())
	assert.Equal(t, p, projects[p.Key()])

	// Project exists
	exists, err := l.ProjectExists(p.Key())
	assert.NoError(t, err)
	assert.True(t, exists)

	// GetProject should succeed
	newProject := schemas.Project{
		Name: "foo/bar",
	}
	assert.NoError(t, l.GetProject(&newProject))
	assert.Equal(t, p, newProject)

	// Count
	count, err := l.ProjectsCount()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Delete project
	l.DelProject(p.Key())
	projects, err = l.Projects()
	assert.NoError(t, err)
	assert.NotContains(t, projects, p.Key())

	exists, err = l.ProjectExists(p.Key())
	assert.NoError(t, err)
	assert.False(t, exists)

	// GetProject should not update the var this time
	newProject = schemas.Project{
		Name: "foo/bar",
	}
	assert.NoError(t, l.GetProject(&newProject))
	assert.NotEqual(t, p, newProject)
}

func TestLocalEnvironmentFunctions(t *testing.T) {
	environment := schemas.Environment{
		ProjectName: "foo",
		ID:          1,
	}

	l := NewLocalStorage()
	l.SetEnvironment(environment)

	// Set project
	environments, err := l.Environments()
	assert.NoError(t, err)
	assert.Contains(t, environments, environment.Key())
	assert.Equal(t, environment, environments[environment.Key()])

	// Environment exists
	exists, err := l.EnvironmentExists(environment.Key())
	assert.NoError(t, err)
	assert.True(t, exists)

	// GetEnvironment should succeed
	newEnvironment := schemas.Environment{
		ProjectName: "foo",
		ID:          1,
	}
	assert.NoError(t, l.GetEnvironment(&newEnvironment))
	assert.Equal(t, environment, newEnvironment)

	// Count
	count, err := l.EnvironmentsCount()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Delete Environment
	l.DelEnvironment(environment.Key())
	environments, err = l.Environments()
	assert.NoError(t, err)
	assert.NotContains(t, environments, environment.Key())

	exists, err = l.EnvironmentExists(environment.Key())
	assert.NoError(t, err)
	assert.False(t, exists)

	// GetEnvironment should not update the var this time
	newEnvironment = schemas.Environment{
		ProjectName: "foo",
		ID:          1,
		ExternalURL: "foo",
	}
	assert.NoError(t, l.GetEnvironment(&newEnvironment))
	assert.NotEqual(t, environment, newEnvironment)
}

func TestLocalMetricFunctions(t *testing.T) {
	m := schemas.Metric{
		Kind: schemas.MetricKindCoverage,
		Labels: prometheus.Labels{
			"foo": "bar",
		},
		Value: 5,
	}

	l := NewLocalStorage()
	l.SetMetric(m)

	// Set metric
	metrics, err := l.Metrics()
	assert.NoError(t, err)
	assert.Contains(t, metrics, m.Key())
	assert.Equal(t, m, metrics[m.Key()])

	// Metric exists
	exists, err := l.MetricExists(m.Key())
	assert.NoError(t, err)
	assert.True(t, exists)

	// GetMetric should succeed
	newMetric := schemas.Metric{
		Kind: schemas.MetricKindCoverage,
		Labels: prometheus.Labels{
			"foo": "bar",
		},
	}
	assert.NoError(t, l.GetMetric(&newMetric))
	assert.Equal(t, m, newMetric)

	// Count
	count, err := l.MetricsCount()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Delete Metric
	l.DelMetric(m.Key())
	metrics, err = l.Metrics()
	assert.NoError(t, err)
	assert.NotContains(t, metrics, m.Key())

	exists, err = l.MetricExists(m.Key())
	assert.NoError(t, err)
	assert.False(t, exists)

	// GetMetric should not update the var this time
	newMetric = schemas.Metric{
		Kind: schemas.MetricKindCoverage,
		Labels: prometheus.Labels{
			"foo": "bar",
		},
	}
	assert.NoError(t, l.GetMetric(&newMetric))
	assert.NotEqual(t, m, newMetric)
}
