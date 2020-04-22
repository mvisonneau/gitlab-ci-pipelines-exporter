package cmd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
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
	registerMetricOn(reg, nil, aCounter)

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
	registerMetricOn(reg, nil, aCounter)
	registerMetricOn(reg, nil, aCounter)
	defer func(){
		if r := recover(); r!=nil {
			t.Log("failed correctly")
		}
	}()
}


func TestAMetricCanBeAddedLabelDynamically(t *testing.T) {
	// a custom additional metric added to the registry
	counter := prometheus.NewCounterVec(prometheus.CounterOpts{Name: "test_something"}, []string{"first", "second"})
	prometheus.MustRegister(counter)

	curriedCounter, err := counter.CurryWith(prometheus.Labels{"first": "0", "second": "something"})
	if assert.Nil(t, err) {
		assert.Contains(t, curriedCounter.WithLabelValues().Desc().String(), "something")
	}
}

func TestEmitVariablesMetric(t *testing.T) {
	var testOkFetchFn = func(interface{}, int, ...gitlab.RequestOptionFunc) ([]*gitlab.PipelineVariable, *gitlab.Response, error) {
		return []*gitlab.PipelineVariable{{Key: "test", Value: "testval"}, {Key: "test-2", Value: "aaaa", VariableType: "env_var"}}, nil, nil
	}
	var testErrFetchFn = func(interface{}, int, ...gitlab.RequestOptionFunc) ([]*gitlab.PipelineVariable, *gitlab.Response, error) {
		return nil, nil, fmt.Errorf("some error")
	}
	client := &Client{
		RateLimiter: ratelimit.New(2),
	}

	assert.NoError(t,
		emitPipelineVariablesMetric(client, variableLabelledCounter("gitlab_ci_pipeline_run_count_with_variable", []string{"project", "ref", "pipeline_variables"}),
			"test-project", "master", 0, 0, testOkFetchFn))
	assert.Error(t,
		emitPipelineVariablesMetric(client, variableLabelledCounter("gitlab_ci_pipeline_run_count_with_variable", []string{"project", "ref", "pipeline_variables"}),
			"test-project", "master", 0, 0, testErrFetchFn))
}
