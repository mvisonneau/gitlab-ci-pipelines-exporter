package controller

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/config"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
)

func TestPullRefMetricsSucceed(t *testing.T) {
	ctx, c, mux, srv := newTestController(config.Config{})
	defer srv.Close()

	mux.HandleFunc("/api/v4/projects/foo/pipelines",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "bar", r.URL.Query().Get("ref"))
			fmt.Fprint(w, `[{"id":1}]`)
		})

	mux.HandleFunc("/api/v4/projects/foo/pipelines/1",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `{"id":1,"created_at":"2016-08-11T11:27:00.085Z", "started_at":"2016-08-11T11:28:00.085Z",
			"duration":300,"queued_duration":60,"status":"running","coverage":"30.2"}`)
		})

	mux.HandleFunc("/api/v4/projects/foo/pipelines/1/variables",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			fmt.Fprint(w, `[{"key":"foo","value":"bar"}]`)
		})

	// Metrics pull shall succeed
	p := schemas.NewProject("foo")
	p.Pull.Pipeline.Variables.Enabled = true

	assert.NoError(t, c.PullRefMetrics(
		ctx,
		schemas.NewRef(
			p,
			schemas.RefKindBranch,
			"bar",
		)))

	// Check if all the metrics exist
	metrics, _ := c.Store.Metrics(ctx)
	labels := map[string]string{
		"kind":      string(schemas.RefKindBranch),
		"project":   "foo",
		"ref":       "bar",
		"topics":    "",
		"variables": "foo:bar",
	}

	runCount := schemas.Metric{
		Kind:   schemas.MetricKindRunCount,
		Labels: labels,
		Value:  0,
	}
	assert.Equal(t, runCount, metrics[runCount.Key()])

	coverage := schemas.Metric{
		Kind:   schemas.MetricKindCoverage,
		Labels: labels,
		Value:  30.2,
	}
	assert.Equal(t, coverage, metrics[coverage.Key()])

	runID := schemas.Metric{
		Kind:   schemas.MetricKindID,
		Labels: labels,
		Value:  1,
	}
	assert.Equal(t, runID, metrics[runID.Key()])

	queued := schemas.Metric{
		Kind:   schemas.MetricKindQueuedDurationSeconds,
		Labels: labels,
		Value:  60,
	}
	assert.Equal(t, queued, metrics[queued.Key()])

	labels["status"] = "running"
	status := schemas.Metric{
		Kind:   schemas.MetricKindStatus,
		Labels: labels,
		Value:  1,
	}
	assert.Equal(t, status, metrics[status.Key()])
}

func TestPullRefMetricsMergeRequestPipeline(t *testing.T) {
	ctx, c, mux, srv := newTestController(config.Config{})
	defer srv.Close()

	mux.HandleFunc("/api/v4/projects/foo/pipelines",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "refs/merge-requests/1234/head", r.URL.Query().Get("ref"))
			fmt.Fprint(w, `[{"id":1}]`)
		})

	mux.HandleFunc("/api/v4/projects/foo/pipelines/1",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `{"id":1,"updated_at":"2016-08-11T11:28:34.085Z","duration":300,"status":"running","coverage":"30.2"}`)
		})

	mux.HandleFunc(fmt.Sprintf("/api/v4/projects/foo/pipelines/1/variables"),
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			fmt.Fprint(w, `[{"key":"foo","value":"bar"}]`)
		})

	// Metrics pull shall succeed
	p := schemas.NewProject("foo")
	p.Pull.Pipeline.Variables.Enabled = true

	assert.NoError(t, c.PullRefMetrics(
		ctx,
		schemas.NewRef(
			p,
			schemas.RefKindMergeRequest,
			"1234",
		)))
}
