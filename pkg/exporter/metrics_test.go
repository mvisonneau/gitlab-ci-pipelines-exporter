package exporter

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	assert.NotNil(t, r.Registry)
	assert.NotNil(t, r.Collectors)
}

// introduce a test to check the /metrics endpoint body
func TestMetricsHandler(t *testing.T) {
	configureStore()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	MetricsHandler(w, r)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	// TODO: Find a way to see if expected metrics are present
}

func TestRegistryGetCollector(t *testing.T) {
	r := NewRegistry()
	assert.Equal(t, r.Collectors[schemas.MetricKindCoverage], r.GetCollector(schemas.MetricKindCoverage))
	assert.Nil(t, r.GetCollector(150))
}

func TestExportMetrics(t *testing.T) {
	r := NewRegistry()
	configureStore()

	store.SetMetric(schemas.Metric{
		Kind: schemas.MetricKindCoverage,
		Labels: prometheus.Labels{
			"project":   "foo",
			"topics":    "alpha",
			"ref":       "bar",
			"kind":      "branch",
			"variables": "beta",
		},
		Value: float64(107.7),
	})

	store.SetMetric(schemas.Metric{
		Kind: schemas.MetricKindRunCount,
		Labels: prometheus.Labels{
			"project":   "foo",
			"topics":    "alpha",
			"ref":       "bar",
			"kind":      "branch",
			"variables": "beta",
		},
		Value: float64(10),
	})

	assert.NoError(t, r.ExportMetrics())
	// TODO: Assert that we have the correct metrics being rendered by the exporter
}

// var testOkFetchFn = func(interface{}, int, ...gitlab.RequestOptionFunc) ([]*gitlab.PipelineVariable, *gitlab.Response, error) {
// 	return []*gitlab.PipelineVariable{{Key: "test", Value: "testval"}, {Key: "test-2", Value: "aaaa", VariableType: "env_var"}}, nil, nil
// }
// var testErrFetchFn = func(interface{}, int, ...gitlab.RequestOptionFunc) ([]*gitlab.PipelineVariable, *gitlab.Response, error) {
// 	return nil, nil, fmt.Errorf("some error")
// }

// TODO: Rework those functions
// func TestEmitVariablesMetric(t *testing.T) {
// 	client := &Client{
// 		RateLimiter: ratelimit.New(2),
// 	}
// 	rx, err := regexp.Compile(`.*`)
// 	details := &ProjectDetails{&schemas.Project{Name: "foo"}, "foo/bar", "tag", "master", 0}
// 	if assert.Nil(t, err) {
// 		assert.NoError(t,
// 			emitPipelineVariablesMetric(client,
// 				variableLabelledCounter("gitlab_ci_pipeline_run_count_with_variable", []string{"project", "topics", "ref", "pipeline_variables"}),
// 				details, 0,
// 				testOkFetchFn,
// 				rx))
// 		assert.Error(t,
// 			emitPipelineVariablesMetric(client,
// 				variableLabelledCounter("gitlab_ci_pipeline_run_count_with_variable", []string{"project", "topics", "ref", "pipeline_variables"}),
// 				details,
// 				0,
// 				testErrFetchFn,
// 				rx))
// 	}
// }

// func TestEmitFilteredVariablesMetric(t *testing.T) {
// 	client := &Client{
// 		RateLimiter: ratelimit.New(2),
// 	}
// 	counter := variableLabelledCounter("gitlab_ci_pipeline_run_count_with_variable", []string{"project", "topics", "ref", "pipeline_variables"})
// 	rx, err := regexp.Compile(`^test$`)

// 	details := &projectDetails{&schemas.Project{Name: "foo"}, "foo/bar", "tag", "master", 0}

// 	if assert.Nil(t, err) {
// 		assert.NoError(t,
// 			emitPipelineVariablesMetric(client, counter, details, 0, testOkFetchFn, rx))

// 		g, err := counter.GetMetricWithLabelValues("test", "tag", "master", "foo/bar")
// 		assert.NoError(t, err)
// 		assert.NotNil(t, g.Desc())
// 		assert.Contains(t, g.Desc().String(), "pipeline_variables")

// 		g2, err := counter.GetMetricWith(prometheus.Labels{"pipeline_variables": "test-2"})
// 		assert.Error(t, err)
// 		assert.Nil(t, g2)
// 	}
// }
