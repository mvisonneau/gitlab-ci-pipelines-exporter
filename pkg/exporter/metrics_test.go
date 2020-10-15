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
	resetGlobalValues()

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
	resetGlobalValues()

	r := NewRegistry()

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
