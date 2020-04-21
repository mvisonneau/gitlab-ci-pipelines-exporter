package cmd

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
)

var (
	coverage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_coverage",
			Help: "Coverage of the most recent pipeline",
		},
		[]string{"project", "topics", "ref"},
	)

	lastRunDuration = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_last_run_duration_seconds",
			Help: "Duration of last pipeline run",
		},
		[]string{"project", "topics", "ref"},
	)

	lastRunJobDuration = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_last_job_run_duration_seconds",
			Help: "Duration of last job run",
		},
		[]string{"project", "topics", "ref", "stage", "job_name"},
	)

	lastRunJobStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_last_job_run_status",
			Help: "Status of the most recent job",
		},
		[]string{"project", "topics", "ref", "stage", "job_name", "status"},
	)

	lastRunJobArtifactSize = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_last_job_run_artifact_size",
			Help: "Filesize of the most recent job artifacts",
		},
		[]string{"project", "topics", "ref", "stage", "job_name"},
	)

	timeSinceLastJobRun = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_time_since_last_job_run_seconds",
			Help: "Elapsed time since most recent GitLab CI job run.",
		},
		[]string{"project", "topics", "ref", "stage", "job_name"},
	)

	jobRunCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gitlab_ci_pipeline_job_run_count",
			Help: "GitLab CI pipeline job run count",
		},
		[]string{"project", "topics", "ref", "stage", "job_name"},
	)

	lastRunID = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_last_run_id",
			Help: "ID of the most recent pipeline",
		},
		[]string{"project", "topics", "ref"},
	)

	lastRunStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_last_run_status",
			Help: "Status of the most recent pipeline",
		},
		[]string{"project", "topics", "ref", "status"},
	)

	runCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gitlab_ci_pipeline_run_count",
			Help: "GitLab CI pipeline run count",
		},
		[]string{"project", "topics", "ref"},
	)

	timeSinceLastRun = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_time_since_last_run_seconds",
			Help: "Elapsed time since most recent GitLab CI pipeline run.",
		},
		[]string{"project", "topics", "ref"},
	)
)

var (
	registry       = prometheus.NewRegistry()
	defaultMetrics = []prometheus.Collector{coverage, lastRunDuration, lastRunID, lastRunStatus, runCount, timeSinceLastJobRun, lastRunDuration, lastRunJobStatus, jobRunCount, timeSinceLastJobRun, lastRunJobArtifactSize}
)

func registerMetricOn(registry *prometheus.Registry, log *log.Logger, metrics ...prometheus.Collector) {
	for _, m := range metrics {
		if err := registry.Register(m); err != nil {
			log.Fatalf("could not add provided metric '%v' to the Prometheus registry: %v", m, err)
		}
	}
}

func emitStatusMetric(metric *prometheus.GaugeVec, labels []string, statuses []string, status string, sparseMetrics bool) {
	// Moved into separate function to reduce cyclomatic complexity
	// List of available statuses from the API spec
	// ref: https://docs.gitlab.com/ee/api/jobs.html#list-pipeline-jobs
	for _, s := range statuses {
		args := append(labels, s)
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

func variableLabelledCounter(vars []gitlab.PipelineVariable) prometheus.Counter {
	var labels []string
	for _, v := range vars {
		labels = append(labels, v.Key)
	}
	counter := prometheus.NewCounterVec(prometheus.CounterOpts{Name: "test_vars"}, labels).WithLabelValues(labels...)
	counter.Add(1.0)
	return counter
}
