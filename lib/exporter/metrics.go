package exporter

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Registry wraps a pointer of prometheus.Registry
type Registry struct {
	*prometheus.Registry
}

var (
	defaultLabels = []string{"project", "topics", "ref", "variables"}
	jobLabels     = []string{"stage", "job_name"}
	statusLabels  = []string{"status"}

	coverage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_coverage",
			Help: "Coverage of the most recent pipeline",
		},
		defaultLabels,
	)

	lastRunDuration = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_last_run_duration_seconds",
			Help: "Duration of last pipeline run",
		},
		defaultLabels,
	)

	lastRunJobDuration = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_last_job_run_duration_seconds",
			Help: "Duration of last job run",
		},
		append(defaultLabels, jobLabels...),
	)

	lastRunJobStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_last_job_run_status",
			Help: "Status of the most recent job",
		},
		append(defaultLabels, append(jobLabels, statusLabels...)...),
	)

	lastRunJobArtifactSize = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_last_job_run_artifact_size",
			Help: "Filesize of the most recent job artifacts",
		},
		append(defaultLabels, jobLabels...),
	)

	timeSinceLastJobRun = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_time_since_last_job_run_seconds",
			Help: "Elapsed time since most recent GitLab CI job run.",
		},
		append(defaultLabels, jobLabels...),
	)

	jobRunCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gitlab_ci_pipeline_job_run_count",
			Help: "GitLab CI pipeline job run count",
		},
		append(defaultLabels, jobLabels...),
	)

	lastRunID = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_last_run_id",
			Help: "ID of the most recent pipeline",
		},
		defaultLabels,
	)

	lastRunStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_last_run_status",
			Help: "Status of the most recent pipeline",
		},
		append(defaultLabels, "status"),
	)

	runCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gitlab_ci_pipeline_run_count",
			Help: "GitLab CI pipeline run count",
		},
		defaultLabels,
	)

	timeSinceLastRun = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_time_since_last_run_seconds",
			Help: "Elapsed time since most recent GitLab CI pipeline run.",
		},
		defaultLabels,
	)

	defaultMetrics = []prometheus.Collector{
		coverage,
		lastRunDuration,
		lastRunJobDuration,
		lastRunJobStatus,
		lastRunJobArtifactSize,
		timeSinceLastJobRun,
		jobRunCount,
		lastRunID,
		lastRunStatus,
		runCount,
		timeSinceLastRun,
	}
)

// NewRegistry initialize a new registry
func NewRegistry() *Registry {
	return &Registry{prometheus.NewRegistry()}
}

// RegisterDefaultMetrics add all our metrics to the registry
func (r *Registry) RegisterDefaultMetrics() error {
	for _, m := range defaultMetrics {
		if err := r.Register(m); err != nil {
			return fmt.Errorf("could not add provided metric '%v' to the Prometheus registry: %v", m, err)
		}
	}
	return nil
}

// MetricsHandler returns an http handler containing with the desired configuration
func (r *Registry) MetricsHandler(disableOpenMetricsEncoder bool) http.Handler {
	return promhttp.HandlerFor(r, promhttp.HandlerOpts{
		Registry:          r,
		EnableOpenMetrics: !disableOpenMetricsEncoder,
	})
}

func emitStatusMetric(metric *prometheus.GaugeVec, labelValues []string, statuses []string, status string, sparseMetrics bool) {
	// Moved into separate function to reduce cyclomatic complexity
	// List of available statuses from the API spec
	// ref: https://docs.gitlab.com/ee/api/jobs.html#list-pipeline-jobs
	for _, s := range statuses {
		args := append(labelValues, s)
		if s == status {
			metric.WithLabelValues(args...).Set(1)
		} else {
			if sparseMetrics {
				metric.DeleteLabelValues(args...)
			} else {
				metric.WithLabelValues(args...).Set(0)
			}
		}
	}
}

func variableLabelledCounter(metricName string, labels []string) *prometheus.GaugeVec {
	return prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: metricName}, labels)
}
