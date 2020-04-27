package cmd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/xanzy/go-gitlab"
	"go.uber.org/ratelimit"
)

// introduce a test to check the /metrics endpoint body
func TestMetricsRegistryContainsMetricsWhenSet(t *testing.T) {
	// a custom additional metric added to the registry
	some := "test_something"
	aCounter := prometheus.NewCounter(prometheus.CounterOpts{Name: some})
	reg := prometheus.NewRegistry()
	registerMetricOn(reg, aCounter)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	metricsHandlerFor(reg, false).ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	assert.Contains(t, w.Body.String(), some)
}

// introduce a test to check the /metrics endpoint body
func TestMetricsRegistryFailsWhenDouble(t *testing.T) {
	// a custom additional metric added to the registry
	some := "test_something"
	aCounter := prometheus.NewCounter(prometheus.CounterOpts{Name: some})
	reg := prometheus.NewRegistry()
	registerMetricOn(reg, aCounter)
	panicFn := func() {
		registerMetricOn(reg, aCounter)
	}
	assert.Panics(t, panicFn)
}

var testOkFetchFn = func(interface{}, int, ...gitlab.RequestOptionFunc) ([]*gitlab.PipelineVariable, *gitlab.Response, error) {
	return []*gitlab.PipelineVariable{{Key: "test", Value: "testval"}, {Key: "test-2", Value: "aaaa", VariableType: "env_var"}}, nil, nil
}
var testErrFetchFn = func(interface{}, int, ...gitlab.RequestOptionFunc) ([]*gitlab.PipelineVariable, *gitlab.Response, error) {
	return nil, nil, fmt.Errorf("some error")
}

func TestEmitVariablesMetric(t *testing.T) {
	client := &Client{
		RateLimiter: ratelimit.New(2),
	}
	rx, err := regexp.Compile(variablesCatchallRegex)
	details := &projectDetails{"test-project", "tag", "master", 0}
	if assert.Nil(t, err) {
		assert.NoError(t,
			emitPipelineVariablesMetric(client,
				variableLabelledCounter("gitlab_ci_pipeline_run_count_with_variable", []string{"project", "topics", "ref", "pipeline_variables"}),
				details, 0,
				testOkFetchFn,
				rx))
		assert.Error(t,
			emitPipelineVariablesMetric(client,
				variableLabelledCounter("gitlab_ci_pipeline_run_count_with_variable", []string{"project", "topics", "ref", "pipeline_variables"}),
				details,
				0,
				testErrFetchFn,
				rx))
	}
}

func TestEmitFilteredVariablesMetric(t *testing.T) {
	client := &Client{
		RateLimiter: ratelimit.New(2),
	}
	counter := variableLabelledCounter("gitlab_ci_pipeline_run_count_with_variable", []string{"project", "topics", "ref", "pipeline_variables"})
	rx, err := regexp.Compile(`^test$`)

	details := &projectDetails{"test-project", "tag", "master", 0}

	if assert.Nil(t, err) {
		assert.NoError(t,
			emitPipelineVariablesMetric(client, counter, details, 0, testOkFetchFn, rx))

		g, err := counter.GetMetricWithLabelValues("test", "tag", "test-project", "master")
		assert.NoError(t, err)
		assert.NotNil(t, g.Desc())

		g2, err := counter.GetMetricWith(prometheus.Labels{"pipeline_variables": "test-2"})
		assert.Error(t, err)
		assert.Nil(t, g2)
	}
}
