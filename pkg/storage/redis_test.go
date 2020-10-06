package storage

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestRedisProjectFunctions(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	r := NewRedisStorage(redis.NewClient(&redis.Options{Addr: s.Addr()}))

	p := schemas.Project{
		Name: "foo/bar",
	}

	// Set project
	r.SetProject(p)
	projects, err := r.Projects()
	assert.NoError(t, err)
	assert.Contains(t, projects, p.Key())
	assert.Equal(t, p, projects[p.Key()])

	// Project exists
	exists, err := r.ProjectExists(p.Key())
	assert.NoError(t, err)
	assert.True(t, exists)

	// Count
	count, err := r.ProjectsCount()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Delete project
	r.DelProject(p.Key())
	projects, err = r.Projects()
	assert.NoError(t, err)
	assert.NotContains(t, projects, p.Key())

	exists, err = r.ProjectExists(p.Key())
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestRedisProjectRefFunctions(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	r := NewRedisStorage(redis.NewClient(&redis.Options{Addr: s.Addr()}))

	pr := schemas.ProjectRef{
		Project: schemas.Project{
			Name: "foo/bar",
		},
		Ref: "sweet",
	}

	// Set project
	r.SetProjectRef(pr)
	projectsRefs, err := r.ProjectsRefs()
	assert.NoError(t, err)
	assert.Contains(t, projectsRefs, pr.Key())
	assert.Equal(t, pr, projectsRefs[pr.Key()])

	// ProjectRef exists
	exists, err := r.ProjectRefExists(pr.Key())
	assert.NoError(t, err)
	assert.True(t, exists)

	// Count
	count, err := r.ProjectsRefsCount()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Delete ProjectRef
	r.DelProjectRef(pr.Key())
	projectsRefs, err = r.ProjectsRefs()
	assert.NoError(t, err)
	assert.NotContains(t, projectsRefs, pr.Key())

	exists, err = r.ProjectRefExists(pr.Key())
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestRedisMetricFunctions(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	r := NewRedisStorage(redis.NewClient(&redis.Options{Addr: s.Addr()}))

	m := schemas.Metric{
		Kind: schemas.MetricKindCoverage,
		Labels: prometheus.Labels{
			"foo": "bar",
		},
		Value: 1,
	}

	// Set metric
	r.SetMetric(m)
	metrics, err := r.Metrics()
	assert.NoError(t, err)
	assert.Contains(t, metrics, m.Key())
	assert.Equal(t, m, metrics[m.Key()])

	// Metric exists
	exists, err := r.MetricExists(m.Key())
	assert.NoError(t, err)
	assert.True(t, exists)

	// Count
	count, err := r.MetricsCount()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Pull value
	m.Value = 0
	err = r.PullMetricValue(&m)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), m.Value)

	// Delete Metric
	r.DelMetric(m.Key())
	metrics, err = r.Metrics()
	assert.NoError(t, err)
	assert.NotContains(t, metrics, m.Key())

	exists, err = r.MetricExists(m.Key())
	assert.NoError(t, err)
	assert.False(t, exists)

	// Pull value
	m.Value = 10
	err = r.PullMetricValue(&m)
	assert.NoError(t, err)
	assert.Equal(t, float64(10), m.Value)
}
