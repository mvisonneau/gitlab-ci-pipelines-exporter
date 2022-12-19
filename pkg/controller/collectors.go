package controller

import "github.com/prometheus/client_golang/prometheus"

var (
	defaultLabels                = []string{"project", "topics", "kind", "ref", "variables"}
	jobLabels                    = []string{"stage", "job_name", "runner_description"}
	statusLabels                 = []string{"status"}
	environmentLabels            = []string{"project", "environment"}
	environmentInformationLabels = []string{"environment_id", "external_url", "kind", "ref", "latest_commit_short_id", "current_commit_short_id", "available", "username"}
	statusesList                 = [...]string{"created", "waiting_for_resource", "preparing", "pending", "running", "success", "failed", "canceled", "skipped", "manual", "scheduled"}
)

// NewInternalCollectorCurrentlyQueuedTasksCount returns a new collector for the gcpe_currently_queued_tasks_count metric.
func NewInternalCollectorCurrentlyQueuedTasksCount() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gcpe_currently_queued_tasks_count",
			Help: "Number of tasks in the queue",
		},
		[]string{},
	)
}

// NewInternalCollectorEnvironmentsCount returns a new collector for the gcpe_environments_count metric.
func NewInternalCollectorEnvironmentsCount() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gcpe_environments_count",
			Help: "Number of GitLab environments being exported",
		},
		[]string{},
	)
}

// NewInternalCollectorExecutedTasksCount returns a new collector for the gcpe_executed_tasks_count metric.
func NewInternalCollectorExecutedTasksCount() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gcpe_executed_tasks_count",
			Help: "Number of tasks executed",
		},
		[]string{},
	)
}

// NewInternalCollectorGitLabAPIRequestsCount returns a new collector for the gcpe_gitlab_api_requests_count metric.
func NewInternalCollectorGitLabAPIRequestsCount() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gcpe_gitlab_api_requests_count",
			Help: "GitLab API requests count",
		},
		[]string{},
	)
}

// NewInternalCollectorGitLabAPIRequestsRemaining returns a new collector for the gcpe_gitlab_api_requests_remaining metric.
func NewInternalCollectorGitLabAPIRequestsRemaining() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gcpe_gitlab_api_requests_remaining",
			Help: "GitLab API requests remaining in the api limit",
		},
		[]string{},
	)
}

// NewInternalCollectorGitLabAPIRequestsLimit returns a new collector for the gcpe_gitlab_api_requests_limit metric.
func NewInternalCollectorGitLabAPIRequestsLimit() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gcpe_gitlab_api_requests_limit",
			Help: "GitLab API requests available in the api limit",
		},
		[]string{},
	)
}

// NewInternalCollectorMetricsCount returns a new collector for the gcpe_metrics_count metric.
func NewInternalCollectorMetricsCount() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gcpe_metrics_count",
			Help: "Number of GitLab pipelines metrics being exported",
		},
		[]string{},
	)
}

// NewInternalCollectorProjectsCount returns a new collector for the gcpe_projects_count metric.
func NewInternalCollectorProjectsCount() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gcpe_projects_count",
			Help: "Number of GitLab projects being exported",
		},
		[]string{},
	)
}

// NewInternalCollectorRefsCount returns a new collector for the gcpe_refs_count metric.
func NewInternalCollectorRefsCount() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gcpe_refs_count",
			Help: "Number of GitLab refs being exported",
		},
		[]string{},
	)
}

// NewCollectorCoverage returns a new collector for the gitlab_ci_pipeline_coverage metric.
func NewCollectorCoverage() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_coverage",
			Help: "Coverage of the most recent pipeline",
		},
		defaultLabels,
	)
}

// NewCollectorDurationSeconds returns a new collector for the gitlab_ci_pipeline_duration_seconds metric.
func NewCollectorDurationSeconds() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_duration_seconds",
			Help: "Duration in seconds of the most recent pipeline",
		},
		defaultLabels,
	)
}

// NewCollectorQueuedDurationSeconds returns a new collector for the gitlab_ci_pipeline_queued_duration_seconds metric.
func NewCollectorQueuedDurationSeconds() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_queued_duration_seconds",
			Help: "Duration in seconds the most recent pipeline has been queued before starting",
		},
		defaultLabels,
	)
}

// NewCollectorEnvironmentBehindCommitsCount returns a new collector for the gitlab_ci_environment_behind_commits_count metric.
func NewCollectorEnvironmentBehindCommitsCount() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_environment_behind_commits_count",
			Help: "Number of commits the environment is behind given its last deployment",
		},
		environmentLabels,
	)
}

// NewCollectorEnvironmentBehindDurationSeconds returns a new collector for the gitlab_ci_environment_behind_duration_seconds metric.
func NewCollectorEnvironmentBehindDurationSeconds() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_environment_behind_duration_seconds",
			Help: "Duration in seconds the environment is behind the most recent commit given its last deployment",
		},
		environmentLabels,
	)
}

// NewCollectorEnvironmentDeploymentCount returns a new collector for the gitlab_ci_environment_deployment_count metric.
func NewCollectorEnvironmentDeploymentCount() prometheus.Collector {
	return prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gitlab_ci_environment_deployment_count",
			Help: "Number of deployments for an environment",
		},
		environmentLabels,
	)
}

// NewCollectorEnvironmentDeploymentDurationSeconds returns a new collector for the gitlab_ci_environment_deployment_duration_seconds metric.
func NewCollectorEnvironmentDeploymentDurationSeconds() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_environment_deployment_duration_seconds",
			Help: "Duration in seconds of the most recent deployment of the environment",
		},
		environmentLabels,
	)
}

// NewCollectorEnvironmentDeploymentJobID returns a new collector for the gitlab_ci_environment_deployment_id metric.
func NewCollectorEnvironmentDeploymentJobID() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_environment_deployment_job_id",
			Help: "ID of the most recent deployment job of the environment",
		},
		environmentLabels,
	)
}

// NewCollectorEnvironmentDeploymentStatus returns a new collector for the gitlab_ci_environment_deployment_status metric.
func NewCollectorEnvironmentDeploymentStatus() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_environment_deployment_status",
			Help: "Status of the most recent deployment of the environment",
		},
		append(environmentLabels, "status"),
	)
}

// NewCollectorEnvironmentDeploymentTimestamp returns a new collector for the gitlab_ci_environment_deployment_timestamp metric.
func NewCollectorEnvironmentDeploymentTimestamp() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_environment_deployment_timestamp",
			Help: "Creation date of the most recent deployment of the environment",
		},
		environmentLabels,
	)
}

// NewCollectorEnvironmentInformation returns a new collector for the gitlab_ci_environment_information metric.
func NewCollectorEnvironmentInformation() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_environment_information",
			Help: "Information about the environment",
		},
		append(environmentLabels, environmentInformationLabels...),
	)
}

// NewCollectorID returns a new collector for the gitlab_ci_pipeline_id metric.
func NewCollectorID() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_id",
			Help: "ID of the most recent pipeline",
		},
		defaultLabels,
	)
}

// NewCollectorJobArtifactSizeBytes returns a new collector for the gitlab_ci_pipeline_job_artifact_size_bytes metric.
func NewCollectorJobArtifactSizeBytes() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_job_artifact_size_bytes",
			Help: "Artifact size in bytes (sum of all of them) of the most recent job",
		},
		append(defaultLabels, jobLabels...),
	)
}

// NewCollectorJobDurationSeconds returns a new collector for the gitlab_ci_pipeline_job_duration_seconds metric.
func NewCollectorJobDurationSeconds() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_job_duration_seconds",
			Help: "Duration in seconds of the most recent job",
		},
		append(defaultLabels, jobLabels...),
	)
}

// NewCollectorJobID returns a new collector for the gitlab_ci_pipeline_job_id metric.
func NewCollectorJobID() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_job_id",
			Help: "ID of the most recent job",
		},
		append(defaultLabels, jobLabels...),
	)
}

// NewCollectorJobQueuedDurationSeconds returns a new collector for the gitlab_ci_pipeline_job_queued_duration_seconds metric.
func NewCollectorJobQueuedDurationSeconds() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_job_queued_duration_seconds",
			Help: "Duration in seconds the most recent job has been queued before starting",
		},
		append(defaultLabels, jobLabels...),
	)
}

// NewCollectorJobRunCount returns a new collector for the gitlab_ci_pipeline_job_run_count metric.
func NewCollectorJobRunCount() prometheus.Collector {
	return prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gitlab_ci_pipeline_job_run_count",
			Help: "Number of executions of a job",
		},
		append(defaultLabels, jobLabels...),
	)
}

// NewCollectorJobStatus returns a new collector for the gitlab_ci_pipeline_job_status metric.
func NewCollectorJobStatus() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_job_status",
			Help: "Status of the most recent job",
		},
		append(defaultLabels, append(jobLabels, statusLabels...)...),
	)
}

// NewCollectorJobTimestamp returns a new collector for the gitlab_ci_pipeline_job_timestamp metric.
func NewCollectorJobTimestamp() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_job_timestamp",
			Help: "Creation date timestamp of the most recent job",
		},
		append(defaultLabels, jobLabels...),
	)
}

// NewCollectorStatus returns a new collector for the gitlab_ci_pipeline_status metric.
func NewCollectorStatus() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_status",
			Help: "Status of the most recent pipeline",
		},
		append(defaultLabels, "status"),
	)
}

// NewCollectorTimestamp returns a new collector for the gitlab_ci_pipeline_timestamp metric.
func NewCollectorTimestamp() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gitlab_ci_pipeline_timestamp",
			Help: "Timestamp of the last update of the most recent pipeline",
		},
		defaultLabels,
	)
}

// NewCollectorRunCount returns a new collector for the gitlab_ci_pipeline_run_count metric.
func NewCollectorRunCount() prometheus.Collector {
	return prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gitlab_ci_pipeline_run_count",
			Help: "Number of executions of a pipeline",
		},
		defaultLabels,
	)
}
