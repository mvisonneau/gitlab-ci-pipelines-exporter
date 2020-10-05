package exporter

import "github.com/prometheus/client_golang/prometheus"

var (
	defaultLabels = []string{"project", "topics", "ref", "kind", "variables"}
	jobLabels     = []string{"stage", "job_name"}
	statusLabels  = []string{"status"}
	statusesList  = [...]string{"running", "pending", "success", "failed", "canceled", "skipped", "manual"}
)

// NewCollectorCoverage returns a new collector for the gitlab_ci_pipeline_coverage metric
func NewCollectorCoverage() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_coverage",
			Help: "Coverage of the most recent pipeline",
		},
		defaultLabels,
	)
}

// NewCollectorJobRunCount returns a new collector for the gitlab_ci_pipeline_job_run_count metric
func NewCollectorJobRunCount() prometheus.Collector {
	return prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gitlab_ci_pipeline_job_run_count",
			Help: "GitLab CI pipeline job run count",
		},
		append(defaultLabels, jobLabels...),
	)
}

// NewCollectorLastJobRunID returns a new collector for the gitlab_ci_pipeline_last_job_run_id metric
func NewCollectorLastJobRunID() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_last_job_run_id",
			Help: "ID of the most recent job",
		},
		append(defaultLabels, jobLabels...),
	)
}

// NewCollectorLastRunDuration returns a new collector for the gitlab_ci_pipeline_last_run_duration_seconds metric
func NewCollectorLastRunDuration() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_last_run_duration_seconds",
			Help: "Duration of last pipeline run",
		},
		defaultLabels,
	)
}

// NewCollectorLastRunID returns a new collector for the gitlab_ci_pipeline_last_run_id metric
func NewCollectorLastRunID() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_last_run_id",
			Help: "ID of the most recent pipeline",
		},
		defaultLabels,
	)
}

// NewCollectorLastRunJobArtifactSize returns a new collector for the gitlab_ci_pipeline_last_job_run_artifact_size metric
func NewCollectorLastRunJobArtifactSize() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_last_job_run_artifact_size",
			Help: "Filesize of the most recent job artifacts",
		},
		append(defaultLabels, jobLabels...),
	)
}

// NewCollectorLastRunJobDuration returns a new collector for the gitlab_ci_pipeline_last_job_run_duration_seconds metric
func NewCollectorLastRunJobDuration() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_last_job_run_duration_seconds",
			Help: "Duration of last job run",
		},
		append(defaultLabels, jobLabels...),
	)
}

// NewCollectorLastRunJobStatus returns a new collector for the gitlab_ci_pipeline_last_job_run_status metric
func NewCollectorLastRunJobStatus() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_last_job_run_status",
			Help: "Status of the most recent job",
		},
		append(defaultLabels, append(jobLabels, statusLabels...)...),
	)
}

// NewCollectorLastRunStatus returns a new collector for the gitlab_ci_pipeline_last_run_status metric
func NewCollectorLastRunStatus() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_last_run_status",
			Help: "Status of the most recent pipeline",
		},
		append(defaultLabels, "status"),
	)
}

// NewCollectorRunCount returns a new collector for the gitlab_ci_pipeline_run_count metric
func NewCollectorRunCount() prometheus.Collector {
	return prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gitlab_ci_pipeline_run_count",
			Help: "GitLab CI pipeline run count",
		},
		defaultLabels,
	)
}

// NewCollectorTimeSinceLastJobRun returns a new collector for the gitlab_ci_pipeline_time_since_last_job_run_seconds metric
func NewCollectorTimeSinceLastJobRun() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_time_since_last_job_run_seconds",
			Help: "Elapsed time since most recent GitLab CI job run.",
		},
		append(defaultLabels, jobLabels...),
	)
}

// NewCollectorTimeSinceLastRun returns a new collector for the gitlab_ci_pipeline_time_since_last_run_seconds metric
func NewCollectorTimeSinceLastRun() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_time_since_last_run_seconds",
			Help: "Elapsed time since most recent GitLab CI pipeline run.",
		},
		defaultLabels,
	)
}
