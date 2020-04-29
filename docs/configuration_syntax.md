# GitLab CI Pipelines Exporter - Configuration syntax

```yaml
# URL and Token with sufficient permissions to access
# your GitLab's projects pipelines informations (optional)
gitlab:
  # URL of your GitLab instance (optional, defaults to https://gitlab.com)
  url: https://gitlab.com

  # Token to use to authenticate against the GitLab API
  # it requires api and read_repository permissions
  # (required but can also be configured using --gitlab-token
  # or the $GCPE_GITLAB_TOKEN environment variable)
  token: xrN14n9-ywvAFxxxxxx

  # Alternative URL for determining health of
  # GitLab API for the readiness probe (optional)
  health_url: https://gitlab.example.com/-/health
  
  # disable verification of readiness for target
  # GitLab instance calling `health_url` (optional)
  disable_health_check: false

  # disable TLS validation for target
  # GitLab instance (handy when self-hosting) (optional)
  disable_tls_verify: false

# Global rate limit for the GitLab API request/sec
# (optional, default: 10)
maximum_gitlab_api_requests_per_second: 10

# Interval in seconds to discover projects
# from wildcards (optional, default: 1800)
wildcards_projects_discover_interval_seconds: 1800

# Interval in seconds to discover refs
# from projects (optional, default: 300)
projects_refs_discover_interval_seconds: 300

# Interval in seconds to poll metrics from
# discovered project refs (optional, default: 30)
projects_refs_polling_interval_seconds: 30

# Sets the parallelism for polling projects from
# the API (optional, default to available CPUs: runtime.GOMAXPROCS(0))
maximum_projects_poller_workers: 1

# Disable OpenMetrics content encoding in
# prometheus HTTP handler (optional, default: false)
# see: https://godoc.org/github.com/prometheus/client_golang/prometheus/promhttp#HandlerOpts
disable_openmetrics_encoding: false

# Whether to attempt retrieving refs from pipelines
# when the exporter starts (optional, default: false)
on_init_fetch_refs_from_pipelines: false

# Maximum number of pipelines to analyze per project
# to search for refs on init (optional, default: 100)
on_init_fetch_refs_from_pipelines_depth_limit: 100

# Default settings which can be overridden at the project
# or wildcard level (optional)
defaults:
  # Whether to attempt retrieving job level metrics from pipelines.
  # Increases the number of output metrics significantly!
  # (optional, default: false)
  fetch_pipeline_job_metrics: false

  # Fetch pipeline variables in a separate metric (optional, default: false)
  fetch_pipeline_variables: false

  # Whether to output sparse job and pipeline status metrics.
  # When enabled, only the status label matching the last run
  # of a pipeline or jb will be submitted (optional, default: false)
  output_sparse_status_metrics: false

  # Filter pipelines variables to include
  # (optional, default: ".*", all variables)
  pipeline_variables_filter_regexp: ".*"

  # Filter refs (branches/tags) to include
  # (optional, default: "^master$" -- master branch)
  refs_regexp: "^master$"

  # Fetch merge request pipelines refs (optional, default: false)
  fetch_merge_request_pipelines_refs: false

  # Maximum number for merge requests pipelines to
  # attempt fetch on each ref discovery (optional, default: 1)
  fetch_merge_request_pipelines_refs_init_limit: 1

# The list of the projects you want to monitor (optional)
projects:
  - # Name of the project (actually path with namespace) to fetch
    # (required)
    name: foo/bar

    # Whether to attempt retrieving job level metrics from pipelines.
    # Increases the number of output metrics significantly!
    # (optional, default: false)
    fetch_pipeline_job_metrics: false

    # Fetch pipeline variables in a separate metric (optional, default: false)
    fetch_pipeline_variables: false

    # Whether to output sparse job and pipeline status metrics.
    # When enabled, only the status label matching the last run
    # of a pipeline or jb will be submitted (optional, default: false)
    output_sparse_status_metrics: false

    # Filter pipelines variables to include
    # (optional, default: ".*", all variables)
    pipeline_variables_filter_regexp: ".*"

    # Filter refs (branches/tags) to include
    # (optional, default: "^master$" -- master branch)
    refs_regexp: "^master$"

    # Fetch merge request pipelines refs (optional, default: false)
    fetch_merge_request_pipelines_refs: false

    # Maximum number for merge requests pipelines to
    # attempt fetch on each ref discovery (optional, default: 1)
    fetch_merge_request_pipelines_refs_init_limit: 1


# Dynamically fetch projects to monitor using a wildcard (optional)
wildcards:
  - # Define the owner of the projects we want to look for
    # (required)
    owner:
      # Name of the owner (required)
      name: foo

      # Owner kind: can be either 'group' or 'user' (required)
      kind: group

      # if owner kind is 'group', whether to include subgroups
      # or not (optional, default: false)
      include_subgroups: false

    # Search expression to filter out projects
    # (optional, default: '' -- no filter/all projects)
    search: ''

    # Including archived projects or not
    # (optional, default: false)
    archived: false

    # Here are all the default parameters which can be overriden

    # Whether to attempt retrieving job level metrics from pipelines.
    # Increases the number of output metrics significantly!
    # (optional, default: false)
    fetch_pipeline_job_metrics: false

    # Fetch pipeline variables in a separate metric (optional, default: false)
    fetch_pipeline_variables: false

    # Whether to output sparse job and pipeline status metrics.
    # When enabled, only the status label matching the last run
    # of a pipeline or jb will be submitted (optional, default: false)
    output_sparse_status_metrics: false

    # Filter pipelines variables to include
    # (optional, default: ".*", all variables)
    pipeline_variables_filter_regexp: ".*"

    # Filter refs (branches/tags) to include
    # (optional, default: "^master$" -- master branch)
    refs_regexp: "^master$"

    # Fetch merge request pipelines refs (optional, default: false)
    fetch_merge_request_pipelines_refs: false

    # Maximum number for merge requests pipelines to
    # attempt fetch on each ref discovery (optional, default: 1)
    fetch_merge_request_pipelines_refs_init_limit: 1
```
