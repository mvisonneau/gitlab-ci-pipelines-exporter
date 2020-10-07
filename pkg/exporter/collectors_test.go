package exporter

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestNewCollectorFunctions(t *testing.T) {
	for _, f := range [](func() prometheus.Collector){
		NewCollectorCoverage,
		NewCollectorJobLastRunID,
		NewCollectorLastRunDuration,
		NewCollectorLastRunID,
		NewCollectorJobLastRunArtifactSize,
		NewCollectorJobLastRunDuration,
		NewCollectorJobLastRunStatus,
		NewCollectorLastRunStatus,
		NewCollectorJobTimeSinceLastRun,
		NewCollectorTimeSinceLastRun,
	} {
		c := f()
		assert.NotNil(t, c)
		assert.IsType(t, &prometheus.GaugeVec{}, c)
	}

	for _, f := range [](func() prometheus.Collector){
		NewCollectorJobRunCount,
		NewCollectorRunCount,
	} {
		c := f()
		assert.NotNil(t, c)
		assert.IsType(t, &prometheus.CounterVec{}, c)
	}
}
