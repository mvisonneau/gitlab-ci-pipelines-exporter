package exporter

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	"github.com/stretchr/testify/assert"
)

func TestPullRefMetricsSucceed(t *testing.T) {
	resetGlobalValues()
	mux, server := configureMockedGitlabClient()
	defer server.Close()

	mux.HandleFunc("/api/v4/projects/foo/pipelines",
		func(w http.ResponseWriter, r *http.Request) {
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
	assert.NoError(t, pullRefMetrics(schemas.Ref{
		Kind:                         schemas.RefKindBranch,
		ProjectName:                  "foo",
		Name:                         "bar",
		PullPipelineVariablesEnabled: true,
	}))

	// Check if all the metrics exist
	metrics, _ := store.Metrics()
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

	labels["status"] = "running"
	status := schemas.Metric{
		Kind:   schemas.MetricKindStatus,
		Labels: labels,
		Value:  1,
	}
	assert.Equal(t, status, metrics[status.Key()])
}

func TestPullRefMetricsMergeRequestPipeline(t *testing.T) {
	resetGlobalValues()
	ref := schemas.Ref{
		Kind: schemas.RefKindMergeRequest,
		LatestPipeline: schemas.Pipeline{
			ID:     1,
			Status: "success",
		},
	}

	assert.NoError(t, pullRefMetrics(ref))
}
