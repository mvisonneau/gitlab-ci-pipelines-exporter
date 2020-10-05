package schemas

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// MetricKindCoverage refers to the coerage of a job/pipeline
	MetricKindCoverage MetricKind = iota

	// MetricKindJobRunCount ..
	MetricKindJobRunCount

	// MetricKindLastJobRunID ..
	MetricKindLastJobRunID

	// MetricKindLastRunDuration ..
	MetricKindLastRunDuration

	// MetricKindLastRunID ..
	MetricKindLastRunID

	// MetricKindLastRunJobArtifactSize ..
	MetricKindLastRunJobArtifactSize

	// MetricKindLastRunJobDuration ..
	MetricKindLastRunJobDuration

	// MetricKindLastRunJobStatus ..
	MetricKindLastRunJobStatus

	// MetricKindLastRunStatus ..
	MetricKindLastRunStatus

	// MetricKindRunCount ..
	MetricKindRunCount

	// MetricKindTimeSinceLastJobRun ..
	MetricKindTimeSinceLastJobRun

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
	sum := sha1.Sum([]byte(strconv.Itoa(int(m.Kind)) + fmt.Sprintf("%v", m.Labels)))
	return MetricKey(base64.URLEncoding.EncodeToString(sum[:]))
}
