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

# Disable OpenMetrics content encoding in
# prometheus HTTP handler (optional, default: false)
# see: https://godoc.org/github.com/prometheus/client_golang/prometheus/promhttp#HandlerOpts
disable_openmetrics_encoding: false

pull:
  # Global rate limit for the GitLab API request/sec
  # (optional, default: 10)
  maximum_gitlab_api_requests_per_second: 10

  metrics:
    # Whether or not to trigger a pull of the metrics when the
    # exporter starts (optional, default: true)
    on_init: true

    # Whether or not to attempt refreshing the metrics
    # on a regular basis (optional, default: true)
    scheduled: true

    # Interval in seconds to pull metrics from
    # discovered project refs (optional, default: 30)
    interval_seconds: 30

  projects_from_wildcards:
    # Whether to trigger a discovery or not when the
    # exporter starts (optional, default: true)
    on_init: true

    # Whether to attempt retrieving new projects from wildcards
    # on a regular basis (optional, default: true)
    scheduled: true

    # Interval in seconds to discover projects
    # from wildcards (optional, default: 1800)
    interval_seconds: 1800

  project_refs_from_branches_tags_and_mrs:
    # Whether to trigger a discovery of project refs from
    # branches, tags and merge requests when the
    # exporter starts (optional, default: true)
    # nb: merge requests refs discovery needs to be
    # additionally enabled on a per project basis
    on_init: true

    # Whether to attempt retrieving project refs from branches,
    # tags & merge requests on a regular basis (optional, default: true)
    scheduled: true

    # Interval in seconds to discover refs
    # from projects branches and tags (optional, default: 300)
    interval_seconds: 300

# Default settings which can be overridden at the project
# or wildcard level (optional)
project_defaults:
  # Whether to output sparse job and pipeline status metrics.
  # When enabled, only the status label matching the last run
  # of a pipeline or job will be submitted (optional, default: true)
  output_sparse_status_metrics: true

  pull:
    refs:
      # Filter refs (branches/tags only) to include
      # (optional, default: "^main|master$" -- main or master branch)
      regexp: "^main|master$"

      from:
        pipelines:
          # Whether to trigger a discovery of the projects refs
          # from the most recent project pipelines when the
          # project is configured/discovered (optional, default: false)
          # This flag is useful if you want/need to obtain pipelines
          # metrics of deleted refs
          enabled: false

          # Maximum number of pipelines to analyze per project
          # to search for refs on init (optional, default: 100)
          depth: 100

        merge_requests:
          # Fetch merge request pipelines refs (optional, default: false)
          enabled: false

          # Maximum number for merge requests pipelines to
          # attempt fetch on each project ref discovery (optional, default: 1)
          depth: 1

    pipeline:
      jobs:
        # Whether to attempt retrieving job level metrics from pipelines.
        # Increases the number of outputed metrics significantly!
        # (optional, default: false)
        enabled: false

      variables:
        # Fetch pipeline variables in a separate metric (optional, default: false)
        enabled: false

        # Filter pipelines variables to include
        # (optional, default: ".*", all variables)
        regexp: ".*"

# The list of the projects you want to monitor (optional)
projects:
  - # Name of the project (actually path with namespace) to fetch
    # (required)
    name: foo/bar

    # Here are all the project parameters which can be overriden (optional)
    pull:
      refs:
        # Filter refs (branches/tags only) to include
        # (optional, default: "^main|master$" -- main or master branch)
        regexp: "^main|master$"

        from:
          pipelines:
            # Whether to trigger a discovery of the projects refs
            # from the most recent project pipelines when the
            # project is configured/discovered (optional, default: false)
            # This flag is useful if you want/need to obtain pipelines
            # metrics of deleted refs
            enabled: false

            # Maximum number of pipelines to analyze per project
            # to search for refs on init (optional, default: 100)
            depth: 100

          merge_requests:
            # Fetch merge request pipelines refs (optional, default: false)
            enabled: false

            # Maximum number for merge requests pipelines to
            # attempt fetch on each project ref discovery (optional, default: 1)
            depth: 1

      pipeline:
        jobs:
          # Whether to attempt retrieving job level metrics from pipelines.
          # Increases the number of outputed metrics significantly!
          # (optional, default: false)
          enabled: false

        variables:
          # Fetch pipeline variables in a separate metric (optional, default: false)
          enabled: false

          # Filter pipelines variables to include
          # (optional, default: ".*", all variables)
          regexp: ".*"

# Dynamically fetch projects to monitor using a wildcard (optional)
wildcards:
  - # Define the owner of the projects we want to look for (optional)
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

    # Here are all the project parameters which can be overriden (optional)
    pull:
      refs:
        # Filter refs (branches/tags only) to include
        # (optional, default: "^main|master$" -- main or master branch)
        regexp: "^main|master$"

        from:
          pipelines:
            # Whether to trigger a discovery of the projects refs
            # from the most recent project pipelines when the
            # project is configured/discovered (optional, default: false)
            # This flag is useful if you want/need to obtain pipelines
            # metrics of deleted refs
            enabled: false

            # Maximum number of pipelines to analyze per project
            # to search for refs on init (optional, default: 100)
            depth: 100

          merge_requests:
            # Fetch merge request pipelines refs (optional, default: false)
            enabled: false

            # Maximum number for merge requests pipelines to
            # attempt fetch on each project ref discovery (optional, default: 1)
            depth: 1

      pipeline:
        jobs:
          # Whether to attempt retrieving job level metrics from pipelines.
          # Increases the number of outputed metrics significantly!
          # (optional, default: false)
          enabled: false

        variables:
          # Fetch pipeline variables in a separate metric (optional, default: false)
          enabled: false

          # Filter pipelines variables to include
          # (optional, default: ".*", all variables)
          regexp: ".*"
```

## Pull all projects accessible by the provided token

If you want to pull all your GitLab projects (accessible by the token), you can use the following wildcard:

```yaml
wildcards:
  - {}
```

The exporter will then search for all accessible projects and start pulling their metrics.
