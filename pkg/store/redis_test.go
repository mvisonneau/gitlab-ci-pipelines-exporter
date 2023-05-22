package store

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
)

func newTestRedisStore(t *testing.T) (mr *miniredis.Miniredis, r Store) {
	mr, err := miniredis.Run()
	if err != nil {
		panic(err)
	}

	t.Cleanup(func() {
		mr.Close()
	})

	return mr, NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})).(*Redis)
}

func TestRedisProjectFunctions(t *testing.T) {
	_, r := newTestRedisStore(t)

	p := schemas.NewProject("foo/bar")
	p.OutputSparseStatusMetrics = false

	// Set project
	r.SetProject(testCtx, p)
	projects, err := r.Projects(testCtx)
	assert.NoError(t, err)
	assert.Contains(t, projects, p.Key())
	assert.Equal(t, p, projects[p.Key()])

	// Project exists
	exists, err := r.ProjectExists(testCtx, p.Key())
	assert.NoError(t, err)
	assert.True(t, exists)

	// GetProject should succeed
	newProject := schemas.NewProject("foo/bar")
	assert.NoError(t, r.GetProject(testCtx, &newProject))
	assert.Equal(t, p, newProject)

	// Count
	count, err := r.ProjectsCount(testCtx)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Delete project
	r.DelProject(testCtx, p.Key())
	projects, err = r.Projects(testCtx)
	assert.NoError(t, err)
	assert.NotContains(t, projects, p.Key())

	exists, err = r.ProjectExists(testCtx, p.Key())
	assert.NoError(t, err)
	assert.False(t, exists)

	// GetProject should not update the var this time
	newProject = schemas.NewProject("foo/bar")
	assert.NoError(t, r.GetProject(testCtx, &newProject))
	assert.NotEqual(t, p, newProject)
}

func TestRedisEnvironmentFunctions(t *testing.T) {
	_, r := newTestRedisStore(t)

	environment := schemas.Environment{
		ProjectName: "foo",
		ID:          1,
		ExternalURL: "bar",
	}

	// Set project
	r.SetEnvironment(testCtx, environment)
	environments, err := r.Environments(testCtx)
	assert.NoError(t, err)
	assert.Contains(t, environments, environment.Key())
	assert.Equal(t, environment.ProjectName, environments[environment.Key()].ProjectName)
	assert.Equal(t, environment.ID, environments[environment.Key()].ID)

	// Environment exists
	exists, err := r.EnvironmentExists(testCtx, environment.Key())
	assert.NoError(t, err)
	assert.True(t, exists)

	// GetEnvironment should succeed
	newEnvironment := schemas.Environment{
		ProjectName: "foo",
		ID:          1,
	}
	assert.NoError(t, r.GetEnvironment(testCtx, &newEnvironment))
	assert.Equal(t, environment.ExternalURL, newEnvironment.ExternalURL)

	// Count
	count, err := r.EnvironmentsCount(testCtx)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Delete Environment
	r.DelEnvironment(testCtx, environment.Key())
	environments, err = r.Environments(testCtx)
	assert.NoError(t, err)
	assert.NotContains(t, environments, environment.Key())

	exists, err = r.EnvironmentExists(testCtx, environment.Key())
	assert.NoError(t, err)
	assert.False(t, exists)

	// GetEnvironment should not update the var this time
	newEnvironment = schemas.Environment{
		ProjectName: "foo",
		ID:          1,
	}
	assert.NoError(t, r.GetEnvironment(testCtx, &newEnvironment))
	assert.NotEqual(t, environment, newEnvironment)
}

func TestRedisRefFunctions(t *testing.T) {
	_, r := newTestRedisStore(t)

	p := schemas.NewProject("foo/bar")
	p.Topics = "salty"
	ref := schemas.NewRef(
		p,
		schemas.RefKindBranch,
		"sweet",
	)

	// Set project
	r.SetRef(testCtx, ref)
	projectsRefs, err := r.Refs(testCtx)
	assert.NoError(t, err)
	assert.Contains(t, projectsRefs, ref.Key())
	assert.Equal(t, ref, projectsRefs[ref.Key()])

	// Ref exists
	exists, err := r.RefExists(testCtx, ref.Key())
	assert.NoError(t, err)
	assert.True(t, exists)

	// GetRef should succeed
	newRef := schemas.Ref{
		Project: schemas.NewProject("foo/bar"),
		Kind:    schemas.RefKindBranch,
		Name:    "sweet",
	}
	assert.NoError(t, r.GetRef(testCtx, &newRef))
	assert.Equal(t, ref, newRef)

	// Count
	count, err := r.RefsCount(testCtx)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Delete Ref
	r.DelRef(testCtx, ref.Key())
	projectsRefs, err = r.Refs(testCtx)
	assert.NoError(t, err)
	assert.NotContains(t, projectsRefs, ref.Key())

	exists, err = r.RefExists(testCtx, ref.Key())
	assert.NoError(t, err)
	assert.False(t, exists)

	// GetRef should not update the var this time
	newRef = schemas.Ref{
		Kind:    schemas.RefKindBranch,
		Project: schemas.NewProject("foo/bar"),
		Name:    "sweet",
	}
	assert.NoError(t, r.GetRef(testCtx, &newRef))
	assert.NotEqual(t, ref, newRef)
}

func TestRedisMetricFunctions(t *testing.T) {
	_, r := newTestRedisStore(t)

	m := schemas.Metric{
		Kind: schemas.MetricKindCoverage,
		Labels: prometheus.Labels{
			"foo": "bar",
		},
		Value: 5,
	}

	// Set metric
	r.SetMetric(testCtx, m)
	metrics, err := r.Metrics(testCtx)
	assert.NoError(t, err)
	assert.Contains(t, metrics, m.Key())
	assert.Equal(t, m, metrics[m.Key()])

	// Metric exists
	exists, err := r.MetricExists(testCtx, m.Key())
	assert.NoError(t, err)
	assert.True(t, exists)

	// GetMetric should succeed
	newMetric := schemas.Metric{
		Kind: schemas.MetricKindCoverage,
		Labels: prometheus.Labels{
			"foo": "bar",
		},
	}
	assert.NoError(t, r.GetMetric(testCtx, &newMetric))
	assert.Equal(t, m, newMetric)

	// Count
	count, err := r.MetricsCount(testCtx)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Delete Metric
	r.DelMetric(testCtx, m.Key())
	metrics, err = r.Metrics(testCtx)
	assert.NoError(t, err)
	assert.NotContains(t, metrics, m.Key())

	exists, err = r.MetricExists(testCtx, m.Key())
	assert.NoError(t, err)
	assert.False(t, exists)

	// GetMetric should not update the var this time
	newMetric = schemas.Metric{
		Kind: schemas.MetricKindCoverage,
		Labels: prometheus.Labels{
			"foo": "bar",
		},
	}
	assert.NoError(t, r.GetMetric(testCtx, &newMetric))
	assert.NotEqual(t, m, newMetric)
}

func TestRedisKeepalive(t *testing.T) {
	mr, r := newTestRedisStore(t)

	uuidString := uuid.New().String()
	resp, err := r.(*Redis).SetKeepalive(testCtx, uuidString, time.Second)
	assert.True(t, resp)
	assert.NoError(t, err)

	resp, err = r.(*Redis).KeepaliveExists(testCtx, uuidString)
	assert.True(t, resp)
	assert.NoError(t, err)

	mr.FastForward(2 * time.Second)

	resp, err = r.(*Redis).KeepaliveExists(testCtx, uuidString)
	assert.False(t, resp)
	assert.NoError(t, err)
}

func TestGetRedisQueueKey(t *testing.T) {
	assert.Equal(t, "task:GarbageCollectEnvironments:foo", getRedisQueueKey(schemas.TaskTypeGarbageCollectEnvironments, "foo"))
}

func TestRedisQueueTask(t *testing.T) {
	mr, r := newTestRedisStore(t)

	r.(*Redis).SetKeepalive(testCtx, "controller1", time.Second)

	ok, err := r.QueueTask(testCtx, schemas.TaskTypePullMetrics, "foo", "controller1")
	assert.True(t, ok)
	assert.NoError(t, err)

	// The keepalive of controller1 not being expired, we should not requeue the task
	ok, err = r.QueueTask(testCtx, schemas.TaskTypePullMetrics, "foo", "controller2")
	assert.False(t, ok)
	assert.NoError(t, err)

	// The keepalive of controller1 being expired, we should requeue the task
	mr.FastForward(2 * time.Second)

	ok, err = r.QueueTask(testCtx, schemas.TaskTypePullMetrics, "foo", "controller2")
	assert.True(t, ok)
	assert.NoError(t, err)
}

func TestRedisUnqueueTask(t *testing.T) {
	_, r := newTestRedisStore(t)

	r.QueueTask(testCtx, schemas.TaskTypePullMetrics, "foo", "")
	count, _ := r.ExecutedTasksCount(testCtx)
	assert.Equal(t, uint64(0), count)

	assert.NoError(t, r.UnqueueTask(testCtx, schemas.TaskTypePullMetrics, "foo"))
	count, _ = r.ExecutedTasksCount(testCtx)
	assert.Equal(t, uint64(1), count)
}

func TestRedisCurrentlyQueuedTasksCount(t *testing.T) {
	_, r := newTestRedisStore(t)

	r.QueueTask(testCtx, schemas.TaskTypePullMetrics, "foo", "")
	r.QueueTask(testCtx, schemas.TaskTypePullMetrics, "bar", "")
	r.QueueTask(testCtx, schemas.TaskTypePullMetrics, "baz", "")

	count, _ := r.CurrentlyQueuedTasksCount(testCtx)
	assert.Equal(t, uint64(3), count)
	r.UnqueueTask(testCtx, schemas.TaskTypePullMetrics, "foo")
	count, _ = r.CurrentlyQueuedTasksCount(testCtx)
	assert.Equal(t, uint64(2), count)
}

func TestRedisExecutedTasksCount(t *testing.T) {
	_, r := newTestRedisStore(t)

	r.QueueTask(testCtx, schemas.TaskTypePullMetrics, "foo", "")
	r.QueueTask(testCtx, schemas.TaskTypePullMetrics, "bar", "")
	r.UnqueueTask(testCtx, schemas.TaskTypePullMetrics, "foo")
	r.UnqueueTask(testCtx, schemas.TaskTypePullMetrics, "foo")

	count, _ := r.ExecutedTasksCount(testCtx)
	assert.Equal(t, uint64(1), count)
}
