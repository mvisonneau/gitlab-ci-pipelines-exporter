package exporter

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	"github.com/openlyinc/pointy"
	"github.com/stretchr/testify/assert"
	"github.com/xanzy/go-gitlab"
)

func TestPullProjectRefMetricsSucceed(t *testing.T) {
	resetGlobalValues()
	mux, server := configureMockedGitlabClient()
	defer server.Close()
	configureStore()

	mux.HandleFunc("/api/v4/projects/1/pipelines",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `[{"id":1}]`)
		})

	mux.HandleFunc("/api/v4/projects/1/pipelines/1",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `{"id":1,"updated_at":"2016-08-11T11:28:34.085Z","duration":300,"status":"running","coverage":"30.2"}`)
		})

	mux.HandleFunc(fmt.Sprintf("/api/v4/projects/1/pipelines/1/variables"),
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			fmt.Fprint(w, `[{"key":"foo","value":"bar"}]`)
		})

	pr := schemas.ProjectRef{
		Kind:              schemas.ProjectRefKindBranch,
		ID:                1,
		PathWithNamespace: "foo/bar",
		Ref:               "baz",
	}
	pr.Pull.Pipeline.Variables.EnabledValue = pointy.Bool(true)

	// Metrics pull shall succeed
	assert.NoError(t, pullProjectRefMetrics(pr))

	// Check if all the metrics exist
	metrics, _ := store.Metrics()
	labels := map[string]string{
		"project":   "foo/bar",
		"topics":    "",
		"ref":       "baz",
		"kind":      string(schemas.ProjectRefKindBranch),
		"variables": "foo:bar",
	}

	runCount := schemas.Metric{
		Kind:   schemas.MetricKindRunCount,
		Labels: labels,
		Value:  1,
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

func TestPullProjectRefMetricsMergeRequestPipeline(t *testing.T) {
	pr := schemas.ProjectRef{
		Kind: schemas.ProjectRefKindMergeRequest,
		MostRecentPipeline: &gitlab.Pipeline{
			Status: "success",
		},
	}

	assert.NoError(t, pullProjectRefMetrics(pr))
}
