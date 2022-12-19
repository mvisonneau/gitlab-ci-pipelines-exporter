package controller

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
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
	// TODO: implement correctly
}
