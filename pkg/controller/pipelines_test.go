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
			"duration":300,"queued_duration":60,"status":"running","coverage":"30.2","source":"schedule"}`)
		})

	mux.HandleFunc("/api/v4/projects/foo/pipelines/1/variables",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			fmt.Fprint(w, `[{"key":"foo","value":"bar"}]`)
		})

	mux.HandleFunc("/api/v4/projects/foo/pipelines/1/test_report",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			fmt.Fprint(w, `{"total_time": 5, "total_count": 1, "success_count": 1, "failed_count": 0, "skipped_count": 0, "error_count": 0, "test_suites": [{"name": "Secure", "total_time": 5, "total_count": 1, "success_count": 1, "failed_count": 0, "skipped_count": 0, "error_count": 0, "test_cases": [{"status": "success", "name": "Security Reports can create an auto-remediation MR", "classname": "vulnerability_management_spec", "execution_time": 5, "system_output": null, "stack_trace": null}]}]}`)
		})

	// Metrics pull shall succeed
	p := schemas.NewProject("foo")
	p.Pull.Pipeline.Variables.Enabled = true
	p.Pull.Pipeline.TestReports.Enabled = true
	p.Pull.Pipeline.TestReports.TestCases.Enabled = true

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
		"source":    "schedule",
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

func TestPullRefTestReportMetrics(t *testing.T) {
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
			"duration":300,"queued_duration":60,"status":"success","coverage":"30.2","source":"schedule"}`)
		})

	mux.HandleFunc("/api/v4/projects/foo/pipelines/1/variables",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			fmt.Fprint(w, `[{"key":"foo","value":"bar"}]`)
		})

	mux.HandleFunc("/api/v4/projects/foo/pipelines/1/test_report",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			fmt.Fprint(w, `{"total_time": 5, "total_count": 1, "success_count": 1, "failed_count": 0, "skipped_count": 0, "error_count": 0, "test_suites": [{"name": "Secure", "total_time": 5, "total_count": 1, "success_count": 1, "failed_count": 0, "skipped_count": 0, "error_count": 0, "test_cases": [{"status": "success", "name": "Security Reports can create an auto-remediation MR", "classname": "vulnerability_management_spec", "execution_time": 5, "system_output": null, "stack_trace": null}]}]}`)
		})

	// Metrics pull shall succeed
	p := schemas.NewProject("foo")
	p.Pull.Pipeline.Variables.Enabled = true
	p.Pull.Pipeline.TestReports.Enabled = true
	p.Pull.Pipeline.TestReports.TestCases.Enabled = true

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
		"source":    "schedule",
	}

	trTotalTime := schemas.Metric{
		Kind:   schemas.MetricKindTestReportTotalTime,
		Labels: labels,
		Value:  5,
	}
	assert.Equal(t, trTotalTime, metrics[trTotalTime.Key()])

	trTotalCount := schemas.Metric{
		Kind:   schemas.MetricKindTestReportTotalCount,
		Labels: labels,
		Value:  1,
	}
	assert.Equal(t, trTotalCount, metrics[trTotalCount.Key()])

	trSuccessCount := schemas.Metric{
		Kind:   schemas.MetricKindTestReportSuccessCount,
		Labels: labels,
		Value:  1,
	}
	assert.Equal(t, trSuccessCount, metrics[trSuccessCount.Key()])

	trFailedCount := schemas.Metric{
		Kind:   schemas.MetricKindTestReportFailedCount,
		Labels: labels,
		Value:  0,
	}
	assert.Equal(t, trFailedCount, metrics[trFailedCount.Key()])

	trSkippedCount := schemas.Metric{
		Kind:   schemas.MetricKindTestReportSkippedCount,
		Labels: labels,
		Value:  0,
	}
	assert.Equal(t, trSkippedCount, metrics[trSkippedCount.Key()])

	trErrorCount := schemas.Metric{
		Kind:   schemas.MetricKindTestReportErrorCount,
		Labels: labels,
		Value:  0,
	}
	assert.Equal(t, trErrorCount, metrics[trErrorCount.Key()])

	labels["test_suite_name"] = "Secure"

	tsTotalTime := schemas.Metric{
		Kind:   schemas.MetricKindTestSuiteTotalTime,
		Labels: labels,
		Value:  5,
	}
	assert.Equal(t, tsTotalTime, metrics[tsTotalTime.Key()])

	tsTotalCount := schemas.Metric{
		Kind:   schemas.MetricKindTestSuiteTotalCount,
		Labels: labels,
		Value:  1,
	}
	assert.Equal(t, tsTotalCount, metrics[tsTotalCount.Key()])

	tsSuccessCount := schemas.Metric{
		Kind:   schemas.MetricKindTestSuiteSuccessCount,
		Labels: labels,
		Value:  1,
	}
	assert.Equal(t, tsSuccessCount, metrics[tsSuccessCount.Key()])

	tsFailedCount := schemas.Metric{
		Kind:   schemas.MetricKindTestSuiteFailedCount,
		Labels: labels,
		Value:  0,
	}
	assert.Equal(t, tsFailedCount, metrics[tsFailedCount.Key()])

	tsSkippedCount := schemas.Metric{
		Kind:   schemas.MetricKindTestSuiteSkippedCount,
		Labels: labels,
		Value:  0,
	}
	assert.Equal(t, tsSkippedCount, metrics[tsSkippedCount.Key()])

	tsErrorCount := schemas.Metric{
		Kind:   schemas.MetricKindTestSuiteErrorCount,
		Labels: labels,
		Value:  0,
	}
	assert.Equal(t, tsErrorCount, metrics[tsErrorCount.Key()])

	labels["test_case_name"] = "Security Reports can create an auto-remediation MR"
	labels["test_case_classname"] = "vulnerability_management_spec"

	tcExecutionTime := schemas.Metric{
		Kind:   schemas.MetricKindTestCaseExecutionTime,
		Labels: labels,
		Value:  5,
	}
	assert.Equal(t, tcExecutionTime, metrics[tcExecutionTime.Key()])

	labels["status"] = "success"
	tcStatus := schemas.Metric{
		Kind:   schemas.MetricKindTestCaseStatus,
		Labels: labels,
		Value:  1,
	}
	assert.Equal(t, tcStatus, metrics[tcStatus.Key()])
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
			fmt.Fprint(w, `{"id":1,"updated_at":"2016-08-11T11:28:34.085Z","duration":300,"status":"running","coverage":"30.2","source":"schedule"}`)
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
