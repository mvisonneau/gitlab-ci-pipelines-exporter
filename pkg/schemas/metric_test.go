package schemas

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestMetricKey(t *testing.T) {
	assert.Equal(t, MetricKey("2152003002"), Metric{
		Kind: MetricKindCoverage,
		Labels: prometheus.Labels{
			"foo": "bar",
		},
	}.Key())

	assert.Equal(t, MetricKey("3948534280"), Metric{
		Kind: MetricKindEnvironmentInformation,
		Labels: prometheus.Labels{
			"project":     "foo",
			"environment": "bar",
			"foo":         "bar",
		},
	}.Key())

	assert.Equal(t, MetricKey("3948534280"), Metric{
		Kind: MetricKindEnvironmentInformation,
		Labels: prometheus.Labels{
			"project":     "foo",
			"environment": "bar",
			"bar":         "baz",
		},
	}.Key())

	assert.Equal(t, MetricKey("4100769768"), Metric{
		Kind: MetricKindEnvironmentInformation,
	}.Key())
}
