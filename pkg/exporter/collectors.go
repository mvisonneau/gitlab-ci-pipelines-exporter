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

// NewCollectorDurationSeconds returns a new collector for the gitlab_ci_pipeline_duration_seconds metric
func NewCollectorDurationSeconds() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_duration_seconds",
			Help: "Duration in seconds of the most recent pipeline",
		},
		defaultLabels,
	)
}

// NewCollectorID returns a new collector for the gitlab_ci_pipeline_id metric
func NewCollectorID() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_id",
			Help: "ID of the most recent pipeline",
		},
		defaultLabels,
	)
}

// NewCollectorJobArtifactSizeBytes returns a new collector for the gitlab_ci_pipeline_job_artifact_size_bytes metric
func NewCollectorJobArtifactSizeBytes() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_job_artifact_size_bytes",
			Help: "Artifact size in bytes (sum of all of them) of the most recent job",
		},
		append(defaultLabels, jobLabels...),
	)
}

// NewCollectorJobDurationSeconds returns a new collector for the gitlab_ci_pipeline_job_duration_seconds metric
func NewCollectorJobDurationSeconds() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_job_duration_seconds",
			Help: "Duration in seconds of the most recent job",
		},
		append(defaultLabels, jobLabels...),
	)
}

// NewCollectorJobID returns a new collector for the gitlab_ci_pipeline_job_id metric
func NewCollectorJobID() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_job_id",
			Help: "ID of the most recent job",
		},
		append(defaultLabels, jobLabels...),
	)
}

// NewCollectorJobRunCount returns a new collector for the gitlab_ci_pipeline_job_run_count metric
func NewCollectorJobRunCount() prometheus.Collector {
	return prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gitlab_ci_pipeline_job_run_count",
			Help: "Number of executions of a job",
		},
		append(defaultLabels, jobLabels...),
	)
}

// NewCollectorJobStatus returns a new collector for the gitlab_ci_pipeline_job_status metric
func NewCollectorJobStatus() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_job_status",
			Help: "Status of the most recent job",
		},
		append(defaultLabels, append(jobLabels, statusLabels...)...),
	)
}

// NewCollectorJobTimestamp returns a new collector for the gitlab_ci_pipeline_job_timestamp metric
func NewCollectorJobTimestamp() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_job_timestamp",
			Help: "Creation date timestamp of the the most recent job",
		},
		append(defaultLabels, jobLabels...),
	)
}

// NewCollectorStatus returns a new collector for the gitlab_ci_pipeline_status metric
func NewCollectorStatus() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_status",
			Help: "Status of the most recent pipeline",
		},
		append(defaultLabels, "status"),
	)
}

// NewCollectorTimestamp returns a new collector for the gitlab_ci_pipeline_timestamp metric
func NewCollectorTimestamp() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_timestamp",
			Help: "Timestamp of the last update of the most recent pipeline",
		},
		defaultLabels,
	)
}

// NewCollectorRunCount returns a new collector for the gitlab_ci_pipeline_run_count metric
func NewCollectorRunCount() prometheus.Collector {
	return prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gitlab_ci_pipeline_run_count",
			Help: "Number of executions of a pipeline",
		},
		defaultLabels,
	)
}
