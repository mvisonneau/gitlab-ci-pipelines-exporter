package controller

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/config"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
)

// TestTaskHandlerPullRefMetricsRequeuesWhenDirty ensures that if a pipeline/job
// webhook for a ref arrives while a pull for that same ref is already in flight,
// it gets coalesced into a rescheduled pull instead of being silently dropped --
// which is what previously left "running" pipelines stuck forever (the ref's
// status metric would never be updated again once the in-flight pull completed).
func TestTaskHandlerPullRefMetricsRequeuesWhenDirty(t *testing.T) {
	ctx, c, mux, srv := newTestController(config.Config{})
	defer srv.Close()

	mux.HandleFunc("/api/v4/projects/foo/pipelines",
		func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprint(w, `[{"id":1}]`)
		})

	mux.HandleFunc("/api/v4/projects/foo/pipelines/1",
		func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprint(w, `{"id":1,"created_at":"2016-08-11T11:27:00.085Z",
			"started_at":"2016-08-11T11:28:00.085Z","duration":300,"queued_duration":60,
			"status":"running","source":"schedule"}`)
		})

	ref := schemas.NewRef(schemas.NewProject("foo"), schemas.RefKindBranch, "bar")

	// Simulate the race: ScheduleTask locks the ref for a first pull, then a
	// second webhook comes in for the same ref while that pull is still running.
	ok, err := c.Store.QueueTask(ctx, schemas.TaskTypePullRefMetrics, string(ref.Key()), c.UUID.String())
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = c.Store.QueueTask(ctx, schemas.TaskTypePullRefMetrics, string(ref.Key()), c.UUID.String())
	assert.NoError(t, err)
	assert.False(t, ok, "second call should find the task already queued/running")

	// Run the (first) pull to completion.
	c.TaskHandlerPullRefMetrics(ctx, ref)

	// The handler must have detected the dirty flag left by the second call and
	// rescheduled a pull, which re-acquires the lock synchronously before the
	// task is actually re-run.
	ok, err = c.Store.QueueTask(ctx, schemas.TaskTypePullRefMetrics, string(ref.Key()), c.UUID.String())
	assert.NoError(t, err)
	assert.False(t, ok, "expected the ref pull to have been rescheduled and thus still locked")
}

// TestTaskHandlerPullEnvironmentMetricsRequeuesWhenDirty mirrors
// TestTaskHandlerPullRefMetricsRequeuesWhenDirty for deployment webhooks.
func TestTaskHandlerPullEnvironmentMetricsRequeuesWhenDirty(t *testing.T) {
	ctx, c, mux, srv := newTestController(config.Config{})
	defer srv.Close()

	mux.HandleFunc("/api/v4/projects/foo/environments/1",
		func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprint(w, `{"id":1,"name":"prod","external_url":"https://foo.example.com","state":"available"}`)
		})

	env := schemas.Environment{
		ProjectName: "foo",
		Name:        "prod",
		ID:          1,
	}

	ok, err := c.Store.QueueTask(ctx, schemas.TaskTypePullEnvironmentMetrics, string(env.Key()), c.UUID.String())
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = c.Store.QueueTask(ctx, schemas.TaskTypePullEnvironmentMetrics, string(env.Key()), c.UUID.String())
	assert.NoError(t, err)
	assert.False(t, ok, "second call should find the task already queued/running")

	c.TaskHandlerPullEnvironmentMetrics(ctx, env)

	ok, err = c.Store.QueueTask(ctx, schemas.TaskTypePullEnvironmentMetrics, string(env.Key()), c.UUID.String())
	assert.NoError(t, err)
	assert.False(t, ok, "expected the environment pull to have been rescheduled and thus still locked")
}
