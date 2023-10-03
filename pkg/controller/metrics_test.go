package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/config"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
)

func TestNewRegistry(t *testing.T) {
	r := NewRegistry(context.Background())
	assert.NotNil(t, r.Registry)
	assert.NotNil(t, r.Collectors)
}

// introduce a test to check the /metrics endpoint body.
func TestMetricsHandler(t *testing.T) {
	_, c, _, srv := newTestController(config.Config{})
	srv.Close()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	c.MetricsHandler(w, r)

	// TODO: Find a way to see if expected metrics are present
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
}

func TestRegistryGetCollector(t *testing.T) {
	r := NewRegistry(context.Background())
	assert.Equal(t, r.Collectors[schemas.MetricKindCoverage], r.GetCollector(schemas.MetricKindCoverage))
	assert.Nil(t, r.GetCollector(150))
}

func TestExportMetrics(_ *testing.T) {
	r := NewRegistry(context.Background())

	m1 := schemas.Metric{
		Kind: schemas.MetricKindCoverage,
		Labels: prometheus.Labels{
			"project":   "foo",
			"topics":    "alpha",
			"ref":       "bar",
			"kind":      "branch",
			"source":    "schedule",
			"variables": "beta",
		},
		Value: float64(107.7),
	}

	m2 := schemas.Metric{
		Kind: schemas.MetricKindRunCount,
		Labels: prometheus.Labels{
			"project":   "foo",
			"topics":    "alpha",
			"ref":       "bar",
			"kind":      "branch",
			"source":    "schedule",
			"variables": "beta",
		},
		Value: float64(10),
	}

	metrics := schemas.Metrics{
		m1.Key(): m1,
		m2.Key(): m2,
	}

	// TODO: Assert that we have the correct metrics being rendered by the exporter
	r.ExportMetrics(metrics)
}
