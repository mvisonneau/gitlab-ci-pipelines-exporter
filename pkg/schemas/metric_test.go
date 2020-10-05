package schemas

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestMetricKey(t *testing.T) {
	m := Metric{
		Kind: MetricKindCoverage,
		Labels: prometheus.Labels{
			"foo": "bar",
		},
	}

	assert.Equal(t, MetricKey("SBbFyHaOYDqfbN0hhFI13xHQ6MY="), m.Key())
}
