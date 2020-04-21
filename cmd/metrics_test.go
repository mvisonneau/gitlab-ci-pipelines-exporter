package cmd

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/xanzy/go-gitlab"
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

func TestAMetricCanBeAddedLabelDynamically(t *testing.T) {
	// a custom additional metric added to the registry
	counter := prometheus.NewCounterVec(prometheus.CounterOpts{Name: "test_something"}, []string{"first", "second"})
	prometheus.MustRegister(counter)

	curriedCounter, err := counter.CurryWith(prometheus.Labels{"first": "0", "second": "something"})
	if assert.Nil(t, err) {
		assert.Contains(t, curriedCounter.WithLabelValues().Desc().String(), "something")
	}
}

func TestAVariableConstMetricIsUpdated(t *testing.T) {
	someVars := []gitlab.PipelineVariable{{Key: "test", Value: "testval"}, {Key: "test-2", Value: "aaaa", VariableType: "env_var"}}

	counter := variableLabelledCounter(someVars)
	assert.Contains(t, counter.Desc().String(), "test")
	assert.Contains(t, counter.Desc().String(), "test-2")
	assert.NotContains(t, counter.Desc().String(), "testval")

}
