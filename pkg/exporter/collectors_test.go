package exporter

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestNewCollectorFunctions(t *testing.T) {
	for _, f := range [](func() prometheus.Collector){
		NewCollectorCoverage,
		NewCollectorDurationSeconds,
		NewCollectorEnvironmentBehindCommitsCount,
		NewCollectorEnvironmentBehindDurationSeconds,
		NewCollectorEnvironmentDeploymentDurationSeconds,
		NewCollectorEnvironmentDeploymentJobID,
		NewCollectorEnvironmentDeploymentStatus,
		NewCollectorEnvironmentDeploymentTimestamp,
		NewCollectorEnvironmentInformation,
		NewCollectorID,
		NewCollectorJobArtifactSizeBytes,
		NewCollectorJobDurationSeconds,
		NewCollectorJobID,
		NewCollectorJobStatus,
		NewCollectorJobTimestamp,
		NewCollectorStatus,
		NewCollectorTimestamp,
	} {
		c := f()
		assert.NotNil(t, c)
		assert.IsType(t, &prometheus.GaugeVec{}, c)
	}

	for _, f := range [](func() prometheus.Collector){
		NewCollectorJobRunCount,
		NewCollectorRunCount,
		NewCollectorEnvironmentDeploymentCount,
	} {
		c := f()
		assert.NotNil(t, c)
		assert.IsType(t, &prometheus.CounterVec{}, c)
	}
}
