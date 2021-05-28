package schemas

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestMetricKey(t *testing.T) {
	assert.Equal(t, MetricKey("3273426995"), Metric{
		Kind: MetricKindCoverage,
		Labels: prometheus.Labels{
			"foo": "bar",
		},
	}.Key())

	assert.Equal(t, MetricKey("2573719482"), Metric{
		Kind: MetricKindEnvironmentInformation,
		Labels: prometheus.Labels{
			"project":     "foo",
			"environment": "bar",
			"foo":         "bar",
		},
	}.Key())

	assert.Equal(t, MetricKey("2573719482"), Metric{
		Kind: MetricKindEnvironmentInformation,
		Labels: prometheus.Labels{
			"project":     "foo",
			"environment": "bar",
			"bar":         "baz",
		},
	}.Key())

	assert.Equal(t, MetricKey("1258247728"), Metric{
		Kind: MetricKindEnvironmentInformation,
	}.Key())
}
