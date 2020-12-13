package exporter

import "github.com/prometheus/client_golang/prometheus"

var (
	defaultLabels                = []string{"project", "topics", "kind", "ref", "variables"}
	jobLabels                    = []string{"stage", "job_name", "runner_description"}
	statusLabels                 = []string{"status"}
	environmentLabels            = []string{"project", "environment"}
	environmentInformationLabels = []string{"environment_id", "external_url", "kind", "ref", "latest_commit_short_id", "current_commit_short_id", "available", "author_email"}
	statusesList                 = [...]string{"created", "waiting_for_resource", "preparing", "pending", "running", "success", "failed", "canceled", "skipped", "manual", "scheduled"}
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

// NewCollectorEnvironmentBehindCommitsCount returns a new collector for the gitlab_ci_environment_behind_commits_count metric
func NewCollectorEnvironmentBehindCommitsCount() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_environment_behind_commits_count",
			Help: "Number of commits the environment is behind given its last deployment",
		},
		environmentLabels,
	)
}

// NewCollectorEnvironmentBehindDurationSeconds returns a new collector for the gitlab_ci_environment_behind_duration_seconds metric
func NewCollectorEnvironmentBehindDurationSeconds() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_environment_behind_duration_seconds",
			Help: "Duration in seconds the environment is behind the most recent commit given its last deployment",
		},
		environmentLabels,
	)
}

// NewCollectorEnvironmentDeploymentCount returns a new collector for the gitlab_ci_environment_deployment_count metric
func NewCollectorEnvironmentDeploymentCount() prometheus.Collector {
	return prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gitlab_ci_environment_deployment_count",
			Help: "Number of deployments for an environment",
		},
		environmentLabels,
	)
}

// NewCollectorEnvironmentDeploymentDurationSeconds returns a new collector for the gitlab_ci_environment_deployment_duration_seconds metric
func NewCollectorEnvironmentDeploymentDurationSeconds() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_environment_deployment_duration_seconds",
			Help: "Duration in seconds of the most recent deployment of the environment",
		},
		environmentLabels,
	)
}

// NewCollectorEnvironmentDeploymentJobID returns a new collector for the gitlab_ci_environment_deployment_id metric
func NewCollectorEnvironmentDeploymentJobID() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_environment_deployment_job_id",
			Help: "ID of the most recent deployment job of the environment",
		},
		environmentLabels,
	)
}

// NewCollectorEnvironmentDeploymentStatus returns a new collector for the gitlab_ci_environment_deployment_status metric
func NewCollectorEnvironmentDeploymentStatus() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_environment_deployment_status",
			Help: "Status of the most recent deployment of the environment",
		},
		append(environmentLabels, "status"),
	)
}

// NewCollectorEnvironmentDeploymentTimestamp returns a new collector for the gitlab_ci_environment_deployment_timestamp metric
func NewCollectorEnvironmentDeploymentTimestamp() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_environment_deployment_timestamp",
			Help: "Creation date of the most recent deployment of the environment",
		},
		environmentLabels,
	)
}

// NewCollectorEnvironmentInformation returns a new collector for the gitlab_ci_environment_information metric
func NewCollectorEnvironmentInformation() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_environment_information",
			Help: "Information about the environment",
		},
		append(environmentLabels, environmentInformationLabels...),
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
