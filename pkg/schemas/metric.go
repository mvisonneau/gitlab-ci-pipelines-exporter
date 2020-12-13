package schemas

import (
	"fmt"
	"hash/crc32"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// MetricKindCoverage refers to the coerage of a job/pipeline
	MetricKindCoverage MetricKind = iota

	// MetricKindDurationSeconds ..
	MetricKindDurationSeconds

	// MetricKindEnvironmentBehindCommitsCount ..
	MetricKindEnvironmentBehindCommitsCount

	// MetricKindEnvironmentBehindDurationSeconds ..
	MetricKindEnvironmentBehindDurationSeconds

	// MetricKindEnvironmentDeploymentCount ..
	MetricKindEnvironmentDeploymentCount

	// MetricKindEnvironmentDeploymentDurationSeconds ..
	MetricKindEnvironmentDeploymentDurationSeconds

	// MetricKindEnvironmentDeploymentJobID ..
	MetricKindEnvironmentDeploymentJobID

	// MetricKindEnvironmentDeploymentStatus ..
	MetricKindEnvironmentDeploymentStatus

	// MetricKindEnvironmentDeploymentTimestamp ..
	MetricKindEnvironmentDeploymentTimestamp

	// MetricKindEnvironmentInformation ..
	MetricKindEnvironmentInformation

	// MetricKindID ..
	MetricKindID

	// MetricKindJobArtifactSizeBytes ..
	MetricKindJobArtifactSizeBytes

	// MetricKindJobDurationSeconds ..
	MetricKindJobDurationSeconds

	// MetricKindJobID ..
	MetricKindJobID

	// MetricKindJobRunCount ..
	MetricKindJobRunCount

	// MetricKindJobStatus ..
	MetricKindJobStatus

	// MetricKindJobTimestamp ..
	MetricKindJobTimestamp

	// MetricKindStatus ..
	MetricKindStatus

	// MetricKindRunCount ..
	MetricKindRunCount

	// MetricKindTimestamp ..
	MetricKindTimestamp
)

// MetricKind ..
type MetricKind int32

// Metric ..
type Metric struct {
	Kind   MetricKind
	Labels prometheus.Labels
	Value  float64
}

// MetricKey ..
type MetricKey string

// Metrics ..
type Metrics map[MetricKey]Metric

// Key ..
func (m Metric) Key() MetricKey {
	key := strconv.Itoa(int(m.Kind))

	switch m.Kind {
	case MetricKindCoverage, MetricKindDurationSeconds, MetricKindID, MetricKindStatus, MetricKindRunCount, MetricKindTimestamp:
		key += fmt.Sprintf("%v", []string{
			m.Labels["project"],
			m.Labels["kind"],
			m.Labels["ref"],
		})

	case MetricKindJobArtifactSizeBytes, MetricKindJobDurationSeconds, MetricKindJobID, MetricKindJobRunCount, MetricKindJobStatus, MetricKindJobTimestamp:
		key += fmt.Sprintf("%v", []string{
			m.Labels["project"],
			m.Labels["kind"],
			m.Labels["ref"],
			m.Labels["stage"],
			m.Labels["job_name"],
		})

	case MetricKindEnvironmentBehindCommitsCount, MetricKindEnvironmentBehindDurationSeconds, MetricKindEnvironmentDeploymentCount, MetricKindEnvironmentDeploymentDurationSeconds, MetricKindEnvironmentDeploymentJobID, MetricKindEnvironmentDeploymentStatus, MetricKindEnvironmentDeploymentTimestamp, MetricKindEnvironmentInformation:
		key += fmt.Sprintf("%v", []string{
			m.Labels["project"],
			m.Labels["environment"],
		})
	}

	// If the metric is a "status" one, add the status label
	switch m.Kind {
	case MetricKindJobStatus, MetricKindEnvironmentDeploymentStatus, MetricKindStatus:
		key += m.Labels["status"]
	}

	return MetricKey(strconv.Itoa(int(crc32.ChecksumIEEE([]byte(key)))))
}
