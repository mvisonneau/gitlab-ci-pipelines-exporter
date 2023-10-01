package schemas

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestMetricKey(t *testing.T) {
	assert.Equal(t, MetricKey("3797596385"), Metric{
		Kind: MetricKindCoverage,
		Labels: prometheus.Labels{
			"foo": "bar",
		},
	}.Key())

	assert.Equal(t, MetricKey("77312310"), Metric{
		Kind: MetricKindEnvironmentInformation,
		Labels: prometheus.Labels{
			"project":     "foo",
			"environment": "bar",
			"foo":         "bar",
		},
	}.Key())

	assert.Equal(t, MetricKey("1288741005"), Metric{
		Kind: MetricKindEnvironmentInformation,
	}.Key())
}
