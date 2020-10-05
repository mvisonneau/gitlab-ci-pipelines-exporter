package storage

import (
	"testing"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestLocalProjectFunctions(t *testing.T) {
	p := schemas.Project{
		Name: "foo/bar",
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
}

func TestLocalProjectRefFunctions(t *testing.T) {
	pr := schemas.ProjectRef{
		Project: schemas.Project{
			Name: "foo/bar",
		},
		Ref: "sweet",
	}

	l := NewLocalStorage()
	l.SetProjectRef(pr)

	// Set project
	projectsRefs, err := l.ProjectsRefs()
	assert.NoError(t, err)
	assert.Contains(t, projectsRefs, pr.Key())
	assert.Equal(t, pr, projectsRefs[pr.Key()])

	// ProjectRef exists
	exists, err := l.ProjectRefExists(pr.Key())
	assert.NoError(t, err)
	assert.True(t, exists)

	// Count
	count, err := l.ProjectsRefsCount()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Delete ProjectRef
	l.DelProjectRef(pr.Key())
	projectsRefs, err = l.ProjectsRefs()
	assert.NoError(t, err)
	assert.NotContains(t, projectsRefs, pr.Key())

	exists, err = l.ProjectRefExists(pr.Key())
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestLocalMetricFunctions(t *testing.T) {
	m := schemas.Metric{
		Kind: schemas.MetricKindCoverage,
		Labels: prometheus.Labels{
			"foo": "bar",
		},
		Value: 1,
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

	// Count
	count, err := l.MetricsCount()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Pull value
	l.PullMetricValue(&m)

	// Delete Metric
	l.DelMetric(m.Key())
	metrics, err = l.Metrics()
	assert.NoError(t, err)
	assert.NotContains(t, metrics, m.Key())

	exists, err = l.MetricExists(m.Key())
	assert.NoError(t, err)
	assert.False(t, exists)
}
