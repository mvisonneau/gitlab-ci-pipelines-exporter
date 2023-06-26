package store

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
)

func TestLocalProjectFunctions(t *testing.T) {
	p := schemas.NewProject("foo/bar")
	p.OutputSparseStatusMetrics = false

	l := NewLocalStore()
	assert.NoError(t, l.SetProject(testCtx, p))

	// Set project
	projects, err := l.Projects(testCtx)
	assert.NoError(t, err)
	assert.Contains(t, projects, p.Key())
	assert.Equal(t, p, projects[p.Key()])

	// Project exists
	exists, err := l.ProjectExists(testCtx, p.Key())
	assert.NoError(t, err)
	assert.True(t, exists)

	// GetProject should succeed
	newProject := schemas.NewProject("foo/bar")
	assert.NoError(t, l.GetProject(testCtx, &newProject))
	assert.Equal(t, p, newProject)

	// Count
	count, err := l.ProjectsCount(testCtx)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Delete project
	assert.NoError(t, l.DelProject(testCtx, p.Key()))
	projects, err = l.Projects(testCtx)
	assert.NoError(t, err)
	assert.NotContains(t, projects, p.Key())

	exists, err = l.ProjectExists(testCtx, p.Key())
	assert.NoError(t, err)
	assert.False(t, exists)

	// GetProject should not update the var this time
	newProject = schemas.NewProject("foo/bar")
	assert.NoError(t, l.GetProject(testCtx, &newProject))
	assert.NotEqual(t, p, newProject)
}

func TestLocalEnvironmentFunctions(t *testing.T) {
	environment := schemas.Environment{
		ProjectName: "foo",
		ID:          1,
	}

	l := NewLocalStore()
	assert.NoError(t, l.SetEnvironment(testCtx, environment))

	// Set project
	environments, err := l.Environments(testCtx)
	assert.NoError(t, err)
	assert.Contains(t, environments, environment.Key())
	assert.Equal(t, environment, environments[environment.Key()])

	// Environment exists
	exists, err := l.EnvironmentExists(testCtx, environment.Key())
	assert.NoError(t, err)
	assert.True(t, exists)

	// GetEnvironment should succeed
	newEnvironment := schemas.Environment{
		ProjectName: "foo",
		ID:          1,
	}
	assert.NoError(t, l.GetEnvironment(testCtx, &newEnvironment))
	assert.Equal(t, environment, newEnvironment)

	// Count
	count, err := l.EnvironmentsCount(testCtx)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Delete Environment
	assert.NoError(t, l.DelEnvironment(testCtx, environment.Key()))
	environments, err = l.Environments(testCtx)
	assert.NoError(t, err)
	assert.NotContains(t, environments, environment.Key())

	exists, err = l.EnvironmentExists(testCtx, environment.Key())
	assert.NoError(t, err)
	assert.False(t, exists)

	// GetEnvironment should not update the var this time
	newEnvironment = schemas.Environment{
		ProjectName: "foo",
		ID:          1,
		ExternalURL: "foo",
	}
	assert.NoError(t, l.GetEnvironment(testCtx, &newEnvironment))
	assert.NotEqual(t, environment, newEnvironment)
}

func TestLocalRefFunctions(t *testing.T) {
	p := schemas.NewProject("foo/bar")
	p.Topics = "salty"
	ref := schemas.NewRef(
		p,
		schemas.RefKindBranch,
		"sweet",
	)

	// Set project
	l := NewLocalStore()
	assert.NoError(t, l.SetRef(testCtx, ref))

	projectsRefs, err := l.Refs(testCtx)
	assert.NoError(t, err)
	assert.Contains(t, projectsRefs, ref.Key())
	assert.Equal(t, ref, projectsRefs[ref.Key()])

	// Ref exists
	exists, err := l.RefExists(testCtx, ref.Key())
	assert.NoError(t, err)
	assert.True(t, exists)

	// GetRef should succeed
	newRef := schemas.Ref{
		Project: schemas.NewProject("foo/bar"),
		Kind:    schemas.RefKindBranch,
		Name:    "sweet",
	}
	assert.NoError(t, l.GetRef(testCtx, &newRef))
	assert.Equal(t, ref, newRef)

	// Count
	count, err := l.RefsCount(testCtx)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Delete Ref
	assert.NoError(t, l.DelRef(testCtx, ref.Key()))
	projectsRefs, err = l.Refs(testCtx)
	assert.NoError(t, err)
	assert.NotContains(t, projectsRefs, ref.Key())

	exists, err = l.RefExists(testCtx, ref.Key())
	assert.NoError(t, err)
	assert.False(t, exists)

	// GetRef should not update the var this time
	newRef = schemas.Ref{
		Kind:    schemas.RefKindBranch,
		Project: schemas.NewProject("foo/bar"),
		Name:    "sweet",
	}
	assert.NoError(t, l.GetRef(testCtx, &newRef))
	assert.NotEqual(t, ref, newRef)
}

func TestLocalMetricFunctions(t *testing.T) {
	m := schemas.Metric{
		Kind: schemas.MetricKindCoverage,
		Labels: prometheus.Labels{
			"foo": "bar",
		},
		Value: 5,
	}

	l := NewLocalStore()
	assert.NoError(t, l.SetMetric(testCtx, m))

	// Set metric
	metrics, err := l.Metrics(testCtx)
	assert.NoError(t, err)
	assert.Contains(t, metrics, m.Key())
	assert.Equal(t, m, metrics[m.Key()])

	// Metric exists
	exists, err := l.MetricExists(testCtx, m.Key())
	assert.NoError(t, err)
	assert.True(t, exists)

	// GetMetric should succeed
	newMetric := schemas.Metric{
		Kind: schemas.MetricKindCoverage,
		Labels: prometheus.Labels{
			"foo": "bar",
		},
	}
	assert.NoError(t, l.GetMetric(testCtx, &newMetric))
	assert.Equal(t, m, newMetric)

	// Count
	count, err := l.MetricsCount(testCtx)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Delete Metric
	l.DelMetric(testCtx, m.Key())
	metrics, err = l.Metrics(testCtx)
	assert.NoError(t, err)
	assert.NotContains(t, metrics, m.Key())

	exists, err = l.MetricExists(testCtx, m.Key())
	assert.NoError(t, err)
	assert.False(t, exists)

	// GetMetric should not update the var this time
	newMetric = schemas.Metric{
		Kind: schemas.MetricKindCoverage,
		Labels: prometheus.Labels{
			"foo": "bar",
		},
	}
	assert.NoError(t, l.GetMetric(testCtx, &newMetric))
	assert.NotEqual(t, m, newMetric)
}

func TestLocalQueueTask(t *testing.T) {
	l := NewLocalStore()
	ok, err := l.QueueTask(testCtx, schemas.TaskTypePullMetrics, "foo", "")
	assert.True(t, ok)
	assert.NoError(t, err)

	ok, err = l.QueueTask(testCtx, schemas.TaskTypePullMetrics, "foo", "")
	assert.False(t, ok)
	assert.NoError(t, err)

	_, _ = l.QueueTask(testCtx, schemas.TaskTypePullMetrics, "bar", "")
	ok, err = l.QueueTask(testCtx, schemas.TaskTypePullMetrics, "bar", "")
	assert.False(t, ok)
	assert.NoError(t, err)
}

func TestLocalUnqueueTask(t *testing.T) {
	l := NewLocalStore()
	_, _ = l.QueueTask(testCtx, schemas.TaskTypePullMetrics, "foo", "")
	assert.Equal(t, uint64(0), l.(*Local).executedTasksCount)
	assert.NoError(t, l.UnqueueTask(testCtx, schemas.TaskTypePullMetrics, "foo"))
	assert.Equal(t, uint64(1), l.(*Local).executedTasksCount)
}

func TestLocalCurrentlyQueuedTasksCount(t *testing.T) {
	l := NewLocalStore()
	_, _ = l.QueueTask(testCtx, schemas.TaskTypePullMetrics, "foo", "")
	_, _ = l.QueueTask(testCtx, schemas.TaskTypePullMetrics, "bar", "")
	_, _ = l.QueueTask(testCtx, schemas.TaskTypePullMetrics, "baz", "")

	count, _ := l.CurrentlyQueuedTasksCount(testCtx)
	assert.Equal(t, uint64(3), count)
	assert.NoError(t, l.UnqueueTask(testCtx, schemas.TaskTypePullMetrics, "foo"))
	count, _ = l.CurrentlyQueuedTasksCount(testCtx)
	assert.Equal(t, uint64(2), count)
}

func TestLocalExecutedTasksCount(t *testing.T) {
	l := NewLocalStore()
	_, _ = l.QueueTask(testCtx, schemas.TaskTypePullMetrics, "foo", "")
	_, _ = l.QueueTask(testCtx, schemas.TaskTypePullMetrics, "bar", "")
	_ = l.UnqueueTask(testCtx, schemas.TaskTypePullMetrics, "foo")
	_ = l.UnqueueTask(testCtx, schemas.TaskTypePullMetrics, "foo")

	count, _ := l.ExecutedTasksCount(testCtx)
	assert.Equal(t, uint64(1), count)
}
