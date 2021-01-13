# GitLab CI Pipelines Exporter - Configuration syntax

```yaml
# Exporter HTTP servers configuration
server:
  # [address:port] to make the process listen
  # upon (optional, default: :8080)
  listen_address: :8080
  
  # Enable profiling pages
  # at /debug/pprof (optional, default: false)
  enable_pprof: false
  
  metrics:
    # Enable /metrics endpoint (optional, default: true)
    enabled: true

    # Enable OpenMetrics content encoding in
    # prometheus HTTP handler (optional, default: false)
    # see: https://godoc.org/github.com/prometheus/client_golang/prometheus/promhttp#HandlerOpts
    enable_openmetrics_encoding: true

  webhook:
    # Enable /webhook endpoint to
    # support GitLab requests (optional, default: false)
    enabled: false

    # Secret token to authenticate legitimate webhook
    # requests coming from the GitLab server
    # (required if enabled but can also be configured using
    # the --webhook-secret-token flag or $GCPE_WEBHOOK_SECRET_TOKEN
    # environment variable)
    secret_token: 063f51ec-09a4-11eb-adc1-0242ac120002

# Redis configuration, optional and solely useful for an HA setup.
# By default the data is held in memory of the exporter
redis:
  # URL used to connect onto the redis endpoint
  # format: redis[s]://[:password@]host[:port][/db-number][?option=value])
  # (required to use the feature but can also be configured using
  # the --redis-url flag or $GCPE_REDIS_URL
  # environment variable)
  url: redis://foo:bar@redis.example.net:6379

# URL and Token with sufficient permissions to access
# your GitLab's projects pipelines informations (optional)
gitlab:
  # URL of your GitLab instance (optional, default: https://gitlab.com)
  url: https://gitlab.com

  # Token to use to authenticate against the GitLab API
  # it requires api and read_repository permissions
  # (required but can also be configured using the --gitlab-token
  # flag or the $GCPE_GITLAB_TOKEN environment variable)
  token: xrN14n9-ywvAFxxxxxx

  # Alternative URL for determining health of
  # GitLab API for the readiness probe (optional, default: https://gitlab.com)
  # it can also be defined using the --gitlab-health-url flag or $GCPE_GITLAB_HEALTH_URL
  # environment variable
  health_url: https://gitlab.example.com/-/health
  
  # Enable verification of readiness for target
  # GitLab instance calling `health_url` (optional, default: true)
  enable_health_check: true

  # Enable TLS validation for target
  # GitLab instance (handy when self-hosting) (optional, default: true)
  enable_tls_verify: true

pull:
  # Global rate limit for the GitLab API request/sec
  # (optional, default: 10)
  maximum_gitlab_api_requests_per_second: 10

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

  environments_from_projects:
    # Whether to trigger a discovery of project environments when
    # exporter starts (optional, default: true)
    on_init: true

    # Whether to attempt retrieving project environments
    # on a regular basis (optional, default: true)
    scheduled: true

    # Interval in seconds to discover project environments
    # (optional, default: 300)
    interval_seconds: 300

  refs_from_projects:
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

garbage_collect:
  projects:
    # Whether or not to trigger a garbage collection of the
    # projects when the exporter starts (optional, default: false)
    on_init: false

    # Whether or not to attempt garbage collecting the projects
    # on a regular basis (optional, default: true)
    scheduled: true

    # Interval in seconds to garbage collect projects
    # (optional, default: 14400)
    interval_seconds: 14400

  environments:
    # Whether or not to trigger a garbage collection of the
    # environments when the exporter starts (optional, default: false)
    on_init: false

    # Whether or not to attempt garbage collecting the environments
    # on a regular basis (optional, default: true)
    scheduled: true

    # Interval in seconds to garbage collect environments
    # (optional, default: 14400)
    interval_seconds: 14400

  refs:
    # Whether or not to trigger a garbage collection of the
    # projects refs when the exporter starts (optional, default: false)
    on_init: false

    # Whether or not to attempt garbage collecting the projects refs
    # on a regular basis (optional, default: true)
    scheduled: true

    # Interval in seconds to garbage collect projects refs
    # from projects branches and tags (optional, default: 1800)
    interval_seconds: 1800

  metrics:
    # Whether or not to trigger a garbage collection of the
    # metrics when the exporter starts (optional, default: false)
    on_init: false

    # Whether or not to attempt garbage collecting the metrics
    # on a regular basis (optional, default: true)
    scheduled: true

    # Interval in seconds to garbage collect metrics
    # (optional, default: 300)
    interval_seconds: 300

# Default settings which can be overridden at the project
# or wildcard level (optional)
project_defaults:
  # Whether to output sparse job and pipeline status metrics.
  # When enabled, only the status label matching the last run
  # of a pipeline or job will be submitted (optional, default: true)
  output_sparse_status_metrics: true

  pull:
    environments:
      # Whether or not to pull project environments & their deployments
      # (optional, default: false)
      enabled: false

      # Filter out by name environments to include
      # (optional, default: ".*")
      name_regexp: ".*"

      # When deployments are based upon tags, you can
      # choose to filter out the ones which you are
      # using to deploy your environment (optional, default: ".*")
      tags_regexp: ".*"

    refs:
      # Filter refs (branches/tags only) to include
      # (optional, default: "^main|master$" -- main or master branch)
      regexp: "^main|master$"

      # If the age of the most recent commit for the ref is greater than
      # this value, the ref won't get exported (optional, default: 0 (disabled))
      # nb: when used in conjuction of pull.from.(pipelines|merge_requests).enabled = true, the creation date
      # of the pipeline is taken in account, not the age of the commit
      max_age_seconds: 0

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

        from_child_pipelines:
          # Collect jobs from subsequent child/downstream pipelines
          # (optional, default: true)
          enabled: true

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
      environments:
        # Whether or not to pull project environments & their deployments
        # (optional, default: false)
        enabled: false

        # Filter out by name environments to include
        # (optional, default: ".*")
        name_regexp: ".*"

        # When deployments are based upon tags, you can
        # choose to filter out the ones which you are
        # using to deploy your environment (optional, default: ".*")
        tags_regexp: ".*"
  
      refs:
        # Filter refs (branches/tags only) to include
        # (optional, default: "^main|master$" -- main or master branch)
        regexp: "^main|master$"

        # If the age of the most recent commit for the ref is greater than
        # this value, the ref won't get exported (optional, default: 0 (disabled))
        # nb: when used in conjuction of pull.from.(pipelines|merge_requests).enabled = true, the creation date
        # of the pipeline is taken in account, not the age of the commit
        max_age_seconds: 0

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

          from_child_pipelines:
            # Collect jobs from subsequent child/downstream pipelines
            # (optional, default: true)
            enabled: true

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
      environments:
        # Whether or not to pull project environments & their deployments
        # (optional, default: false)
        enabled: false

        # Filter out by name environments to include
        # (optional, default: ".*")
        name_regexp: ".*"

        # When deployments are based upon tags, you can
        # choose to filter out the ones which you are
        # using to deploy your environment (optional, default: ".*")
        tags_regexp: ".*"

      refs:
        # Filter refs (branches/tags only) to include
        # (optional, default: "^main|master$" -- main or master branch)
        regexp: "^main|master$"

        # If the age of the most recent commit for the ref is greater than
        # this value, the ref won't get exported (optional, default: 0 (disabled))
        # nb: when used in conjuction of pull.from.(pipelines|merge_requests).enabled = true, the creation date
        # of the pipeline is taken in account, not the age of the commit
        max_age_seconds: 0

        from:
          pipelines:
            # Whether to trigger a discovery of the projects refs
            # from the most recent project pipelines when the
            # project is configured/discovered (optional, default: false)
            # This flag is useful if you want/need to obtain pipelines
            # metrics of deleted refs
            enabled: false

            # Maximum number of pipelines to analyze per project
            # to search for refs on init (optional, default: 100, min: 1, max: 100)
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

          from_child_pipelines:
            # Collect jobs from subsequent child/downstream pipelines
            # (optional, default: true)
            enabled: true

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

## Using a forward proxy to reach GitLab's endpoints

You can refer to the documentation of the [net/http package regarding ProxyFromEnvironment](https://godoc.org/net/http#ProxyFromEnvironment)

```bash
# eg:
export HTTP_PROXY=http://10.x.x.x:3128
```
