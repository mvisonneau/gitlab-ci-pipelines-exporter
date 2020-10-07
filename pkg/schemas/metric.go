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

	// MetricKindJobLastRunArtifactSize ..
	MetricKindJobLastRunArtifactSize

	// MetricKindJobLastRunDuration ..
	MetricKindJobLastRunDuration

	// MetricKindJobLastRunID ..
	MetricKindJobLastRunID

	// MetricKindJobLastRunStatus ..
	MetricKindJobLastRunStatus

	// MetricKindJobRunCount ..
	MetricKindJobRunCount

	// MetricKindJobTimeSinceLastRun ..
	MetricKindJobTimeSinceLastRun

	// MetricKindLastRunDuration ..
	MetricKindLastRunDuration

	// MetricKindLastRunID ..
	MetricKindLastRunID

	// MetricKindLastRunStatus ..
	MetricKindLastRunStatus

	// MetricKindRunCount ..
	MetricKindRunCount

	// MetricKindTimeSinceLastRun ..
	MetricKindTimeSinceLastRun
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
	return MetricKey(strconv.Itoa(int(crc32.ChecksumIEEE([]byte(strconv.Itoa(int(m.Kind)) + fmt.Sprintf("%v", m.Labels))))))
}
