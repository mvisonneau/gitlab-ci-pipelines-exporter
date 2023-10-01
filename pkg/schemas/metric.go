package schemas

import (
	"fmt"
	"hash/crc32"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// MetricKindCoverage refers to the coerage of a job/pipeline.
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

	// MetricKindJobQueuedDurationSeconds ..
	MetricKindJobQueuedDurationSeconds

	// MetricKindJobRunCount ..
	MetricKindJobRunCount

	// MetricKindJobStatus ..
	MetricKindJobStatus

	// MetricKindJobTimestamp ..
	MetricKindJobTimestamp

	// MetricKindQueuedDurationSeconds ..
	MetricKindQueuedDurationSeconds

	// MetricKindRunCount ..
	MetricKindRunCount

	// MetricKindStatus ..
	MetricKindStatus

	// MetricKindTimestamp ..
	MetricKindTimestamp

	// MetricKindTestReportTotalTime ..
	MetricKindTestReportTotalTime

	// MetricKindTestReportTotalCount ..
	MetricKindTestReportTotalCount

	// MetricKindTestReportSuccessCount ..
	MetricKindTestReportSuccessCount

	// MetricKindTestReportFailedCount ..
	MetricKindTestReportFailedCount

	// MetricKindTestReportSkippedCount ..
	MetricKindTestReportSkippedCount

	// MetricKindTestReportErrorCount ..
	MetricKindTestReportErrorCount

	// MetricKindTestSuiteTotalTime ..
	MetricKindTestSuiteTotalTime

	// MetricKindTestSuiteTotalCount ..
	MetricKindTestSuiteTotalCount

	// MetricKindTestSuiteSuccessCount ..
	MetricKindTestSuiteSuccessCount

	// MetricKindTestSuiteFailedCount ..
	MetricKindTestSuiteFailedCount

	// MetricKindTestSuiteSkippedCount ..
	MetricKindTestSuiteSkippedCount

	// MetricKindTestSuiteErrorCount ..
	MetricKindTestSuiteErrorCount

	// MetricKindTestCaseExecutionTime ..
	MetricKindTestCaseExecutionTime

	// MetricKindTestCaseStatus ..
	MetricKindTestCaseStatus
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
	case MetricKindCoverage, MetricKindDurationSeconds, MetricKindID, MetricKindQueuedDurationSeconds, MetricKindRunCount, MetricKindStatus, MetricKindTimestamp, MetricKindTestReportTotalCount, MetricKindTestReportErrorCount, MetricKindTestReportFailedCount, MetricKindTestReportSkippedCount, MetricKindTestReportSuccessCount, MetricKindTestReportTotalTime:
		key += fmt.Sprintf("%v", []string{
			m.Labels["project"],
			m.Labels["kind"],
			m.Labels["ref"],
			m.Labels["source"],
		})

	case MetricKindJobArtifactSizeBytes, MetricKindJobDurationSeconds, MetricKindJobID, MetricKindJobQueuedDurationSeconds, MetricKindJobRunCount, MetricKindJobStatus, MetricKindJobTimestamp:
		key += fmt.Sprintf("%v", []string{
			m.Labels["project"],
			m.Labels["kind"],
			m.Labels["ref"],
			m.Labels["stage"],
			m.Labels["tag_list"],
			m.Labels["job_name"],
			m.Labels["failure_reason"],
		})

	case MetricKindEnvironmentBehindCommitsCount, MetricKindEnvironmentBehindDurationSeconds, MetricKindEnvironmentDeploymentCount, MetricKindEnvironmentDeploymentDurationSeconds, MetricKindEnvironmentDeploymentJobID, MetricKindEnvironmentDeploymentStatus, MetricKindEnvironmentDeploymentTimestamp, MetricKindEnvironmentInformation:
		key += fmt.Sprintf("%v", []string{
			m.Labels["project"],
			m.Labels["environment"],
		})

	case MetricKindTestSuiteErrorCount, MetricKindTestSuiteFailedCount, MetricKindTestSuiteSkippedCount, MetricKindTestSuiteSuccessCount, MetricKindTestSuiteTotalCount, MetricKindTestSuiteTotalTime:
		key += fmt.Sprintf("%v", []string{
			m.Labels["project"],
			m.Labels["kind"],
			m.Labels["ref"],
			m.Labels["test_suite_name"],
		})

	case MetricKindTestCaseExecutionTime, MetricKindTestCaseStatus:
		key += fmt.Sprintf("%v", []string{
			m.Labels["project"],
			m.Labels["kind"],
			m.Labels["ref"],
			m.Labels["test_suite_name"],
			m.Labels["test_case_name"],
			m.Labels["test_case_classname"],
		})
	}

	// If the metric is a "status" one, add the status label
	switch m.Kind {
	case MetricKindJobStatus, MetricKindEnvironmentDeploymentStatus, MetricKindStatus, MetricKindTestCaseStatus:
		key += m.Labels["status"]
	}

	return MetricKey(strconv.Itoa(int(crc32.ChecksumIEEE([]byte(key)))))
}
