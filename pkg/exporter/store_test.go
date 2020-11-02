package exporter

import (
	"testing"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestMetricLogFields(t *testing.T) {
	m := schemas.Metric{
		Kind: schemas.MetricKindDurationSeconds,
		Labels: prometheus.Labels{
			"foo": "bar",
		},
	}
	expected := log.Fields{
		"metric-kind":   schemas.MetricKindDurationSeconds,
		"metric-labels": prometheus.Labels{"foo": "bar"},
	}
	assert.Equal(t, expected, metricLogFields(m))
}

func TestStoreGetSetDelMetric(_ *testing.T) {
	resetGlobalValues()

	storeGetMetric(&schemas.Metric{})
	storeSetMetric(schemas.Metric{})
	storeDelMetric(schemas.Metric{})
}
