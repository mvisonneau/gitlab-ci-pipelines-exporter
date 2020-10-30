package storage

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	"github.com/openlyinc/pointy"
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
		ProjectParameters: schemas.ProjectParameters{
			OutputSparseStatusMetricsValue: pointy.Bool(false),
		},
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

	// GetProject should succeed
	newProject := schemas.Project{
		Name: "foo/bar",
	}
	assert.NoError(t, r.GetProject(&newProject))
	assert.Equal(t, p, newProject)

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

	// GetProject should not update the var this time
	newProject = schemas.Project{
		Name: "foo/bar",
	}
	assert.NoError(t, r.GetProject(&newProject))
	assert.NotEqual(t, p, newProject)
}

func TestRedisRefFunctions(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	r := NewRedisStorage(redis.NewClient(&redis.Options{Addr: s.Addr()}))

	ref := schemas.Ref{
		Project: schemas.Project{
			Name: "foo/bar",
		},
		Ref:    "sweet",
		Topics: "salty",
	}

	// Set project
	r.SetRef(ref)
	projectsRefs, err := r.Refs()
	assert.NoError(t, err)
	assert.Contains(t, projectsRefs, ref.Key())
	assert.Equal(t, ref, projectsRefs[ref.Key()])

	// Ref exists
	exists, err := r.RefExists(ref.Key())
	assert.NoError(t, err)
	assert.True(t, exists)

	// GetRef should succeed
	newRef := schemas.Ref{
		Project: schemas.Project{
			Name: "foo/bar",
		},
		Ref: "sweet",
	}
	assert.NoError(t, r.GetRef(&newRef))
	assert.Equal(t, ref, newRef)

	// Count
	count, err := r.RefsCount()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Delete Ref
	r.DelRef(ref.Key())
	projectsRefs, err = r.Refs()
	assert.NoError(t, err)
	assert.NotContains(t, projectsRefs, ref.Key())

	exists, err = r.RefExists(ref.Key())
	assert.NoError(t, err)
	assert.False(t, exists)

	// GetRef should not update the var this time
	newRef = schemas.Ref{
		Project: schemas.Project{
			Name: "foo/bar",
		},
		Ref: "sweet",
	}
	assert.NoError(t, r.GetRef(&newRef))
	assert.NotEqual(t, ref, newRef)
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
		Value: 5,
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

	// GetMetric should succeed
	newMetric := schemas.Metric{
		Kind: schemas.MetricKindCoverage,
		Labels: prometheus.Labels{
			"foo": "bar",
		},
	}
	assert.NoError(t, r.GetMetric(&newMetric))
	assert.Equal(t, m, newMetric)

	// Count
	count, err := r.MetricsCount()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Delete Metric
	r.DelMetric(m.Key())
	metrics, err = r.Metrics()
	assert.NoError(t, err)
	assert.NotContains(t, metrics, m.Key())

	exists, err = r.MetricExists(m.Key())
	assert.NoError(t, err)
	assert.False(t, exists)

	// GetMetric should not update the var this time
	newMetric = schemas.Metric{
		Kind: schemas.MetricKindCoverage,
		Labels: prometheus.Labels{
			"foo": "bar",
		},
	}
	assert.NoError(t, r.GetMetric(&newMetric))
	assert.NotEqual(t, m, newMetric)
}
