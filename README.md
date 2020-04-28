# 🦊 gitlab-ci-pipelines-exporter

[![GoDoc](https://godoc.org/github.com/mvisonneau/gitlab-ci-pipelines-exporter?status.svg)](https://godoc.org/github.com/mvisonneau/gitlab-ci-pipelines-exporter/app)
[![Go Report Card](https://goreportcard.com/badge/github.com/mvisonneau/gitlab-ci-pipelines-exporter)](https://goreportcard.com/report/github.com/mvisonneau/gitlab-ci-pipelines-exporter)
[![Docker Pulls](https://img.shields.io/docker/pulls/mvisonneau/gitlab-ci-pipelines-exporter.svg)](https://hub.docker.com/r/mvisonneau/gitlab-ci-pipelines-exporter/)
[![Build Status](https://cloud.drone.io/api/badges/mvisonneau/gitlab-ci-pipelines-exporter/status.svg)](https://cloud.drone.io/mvisonneau/gitlab-ci-pipelines-exporter)
[![Coverage Status](https://coveralls.io/repos/github/mvisonneau/gitlab-ci-pipelines-exporter/badge.svg?branch=master)](https://coveralls.io/github/mvisonneau/gitlab-ci-pipelines-exporter?branch=master)

`gitlab-ci-pipelines-exporter` allows you to monitor your [GitLab CI pipelines](https://docs.gitlab.com/ee/ci/pipelines.html) with [Prometheus](https://prometheus.io/) or any monitoring solution supporting the [OpenMetrics](https://github.com/OpenObservability/OpenMetrics) format.

## TL:DR

Here is a [Grafana dashboard](https://grafana.com/grafana/dashboards/10620) I have been able to craft, using those metrics:

![grafana_dashboard](/docs/images/grafana_dashboard.png)

If you are solely interested into trying it out, have a look into the [example/](./example) folder which contains documentation to provision test version of the exporter, prometheus and also grafana in `~5min`.

## Getting started

```bash
# Write your config file somewhere on disk
~$ cat <<EOF > $(pwd)/config.yml
# URL and Token with sufficient permissions to access your GitLab's projects
# pipelines informations
gitlab:
  # URL of your GitLab instance (defaults to https://gitlab.com)
  url: https://gitlab.example.com

  # Gitlab access token. You can omit this field when --gitlab-token or $GCPE_GITLAB_TOKEN  are set
  token: xrN14n9-ywvAFxxxxxx

  # Alternative URL for determining health of GitLab API (readiness probe)
  # health_url: https://gitlab.example.com/-/health
  
  # disable verification of readiness for target GitLab instance calling `health_url`
  # disable_health_check: false

  # disable TLS validation for target GitLab instance (handy when self-hosting)
  # disable_tls_verify: false

# Global rate limit for the GitLab API request/sec
maximum_gitlab_api_requests_per_second: 10

# Interval in seconds to discover projects from wildcards (optional, default: 1800)
wildcards_projects_discover_interval_seconds: 1800

# Interval in seconds to discover refs from projects (optional, default: 300)
projects_refs_discover_interval_seconds: 300

# Interval in seconds to poll metrics from discovered project refs (optional, default: 30)
projects_refs_polling_interval_seconds: 30

# Sets the parallelism for polling projects from the API (default to available CPUs: runtime.GOMAXPROCS(0))
# maximum_projects_poller_workers: 1

# Disable OpenMetrics content encoding in prometheus HTTP handler (default: false)
# see: https://godoc.org/github.com/prometheus/client_golang/prometheus/promhttp#HandlerOpts
# disable_openmetrics_encoding: true

# Whether to attempt retrieving refs from pipelines when the exporter starts (default: false)
# on_init_fetch_refs_from_pipelines: false

# Maximum number of pipelines to analyze per project to search for refs on init (default: 100)
# on_init_fetch_refs_from_pipelines_depth_limit: 100

# Default settings which can be overridden at the project or wildcard level
defaults:
  # Whether to attempt retrieving job level metrics from pipelines. Increases the number of output metrics significantly! (default: false)
  # fetch_pipeline_job_metrics: false

  # Fetch pipeline variables in a separate metric (default: false)
  # fetch_pipeline_variables: false

  # Whether to output sparse job and pipeline status metrics. When enabled, only the status label matching the last run of a pipeline or jb will be submitted (default: false)
  # output_sparse_status_metrics: false

  # Filter pipelines variables to include (default: ".*", all variables)
  # pipeline_variables_filter_regex: ".*"

  # Filter refs (branches/tags) to include (default: "^master$" -- master branch)
  # refs_regexp: "^master$"

# The list of the projects you want to monitor
projects:
  - name: foo/project
  - name: bar/project
    refs_regexp: "^master|dev$"
    fetch_pipeline_job_metrics: true # optional, overrides global setting of the same name
    output_sparse_status_metrics: true # optional, overrides global setting of the same name

# Dynamically fetch projects to monitor using a wildcard
wildcards:
  # Fetch projects belonging to a group and potentially its subgroups
  - owner:
      name: foo
      kind: group
      include_subgroups: true # optional (default: false)
    refs_regexp: "^master|1.0$"
    search: 'bar' # optional (defaults to '')
    archived: true # optional (default: false)
    fetch_pipeline_job_metrics: true # optional, overrides global setting of the same name
    output_sparse_status_metrics: true # optional, overrides global setting of the same name

  # Fetch projects belonging to a specific user
  - owner:
      name: bar
      kind: user
    refs_regexp: ".*"
    search: 'bar' # optional (defaults to '')
    archived: true # optional (default: false)

  # Search for projects globally
  - refs_regexp: ".*"
    search: 'baz' # optional (defaults to '')
    archived: true # optional (default: false)
EOF

# If you have docker installed, it is as easy as :
~$ docker run -d \
   --name gitlab-ci-pipelines-exporter \
   -v $(pwd)/config.yml:/etc/config.yml \
   -p 8080:8080 \
   mvisonneau/gitlab-ci-pipelines-exporter:latest \
   --config /etc/config.yml

# Otherwise for Mac OS X
~$ brew install mvisonneau/tap/gitlab-ci-pipelines-exporter
~$ gitlab-ci-pipelines-exporter --config /etc/config.yml

# Linux
~$ export GCPE_VERSION=$(curl -s "https://api.github.com/repos/mvisonneau/s5/gitlab-ci-pipelines-exporter/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
~$ wget https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/releases/download/${GCPE_VERSION}/gitlab-ci-pipelines-exporter_${GCPE_VERSION}_linux_amd64.deb
~$ tar zxvf gitlab-ci-pipelines-exporter_${GCPE_VERSION}_linux_amd64.deb -C /usr/local/bin
~$ gitlab-ci-pipelines-exporter --config /etc/config.yml

# Windows
~$ scoop bucket add https://github.com/mvisonneau/scoops
~$ scoop install gitlab-ci-pipelines-exporter
~$ gitlab-ci-pipelines-exporter --config <path_to_config_file>
```

You should then be able to see the following logs

```bash
~$ docker logs -f gitlab-ci-pipelines-exporter
INFO[0000] starting exporter                             gitlab-endpoint="https://gitlab.com" on-init-fetch-refs-from-pipelines=true polling-pipelines-every=60s polling-projects-every=15s polling-refs-every=10s rate-limit=10rps
INFO[0000] configured wildcards                          count=1
INFO[0000] found new project                             project-name=foo/project wildcard-archived=false wildcard-owner-include-subgroups=false wildcard-owner-kind=group wildcard-owner-name=foo wildcard-search=
INFO[0000] found new project                             project-name=foo/bar wildcard-archived=false wildcard-owner-include-subgroups=false wildcard-owner-kind=group wildcard-owner-name=foo wildcard-search=
INFO[0000] configured projects                           count=3
INFO[0000] started, now serving requests                 listen-address=":8080"
INFO[0000] found project refs                            project-path-with-namespace=foo/project project-ref=master
INFO[0000] found project refs                            project-path-with-namespace=bar/project project-ref=master
INFO[0000] found project refs                            project-path-with-namespace=foo/bar project-ref=master
```

And this is an example of the metrics you should expect to retrieve

```bash
~$ curl -s localhost:8080/metrics | grep gitlab_ci_pipeline
# HELP gitlab_ci_pipeline_last_run_duration_seconds Duration of last pipeline run
# TYPE gitlab_ci_pipeline_last_run_duration_seconds gauge
gitlab_ci_pipeline_last_run_duration_seconds{project="foo/project",ref="dev",topics="",variables=""} 81
gitlab_ci_pipeline_last_run_duration_seconds{project="foo/project",ref="master",topics="",variables=""} 420
gitlab_ci_pipeline_last_run_duration_seconds{project="bar/project",ref="master",topics="",variables=""} 334
gitlab_ci_pipeline_last_run_duration_seconds{project="foo/bar",ref="master",topics="",variables="FOO:BAR"} 55
# HELP gitlab_ci_pipeline_last_run_id ID of the most recent pipeline
# TYPE gitlab_ci_pipeline_last_run_id gauge
gitlab_ci_pipeline_last_run_id{project="foo/project",ref="dev",topics="",variables=""} 4.0059611e+07
gitlab_ci_pipeline_last_run_id{project="foo/project",ref="master",topics="",variables=""} 1.25351545e+08
gitlab_ci_pipeline_last_run_id{project="bar/project",ref="master",topics="",variables=""} 1.33308085e+08
gitlab_ci_pipeline_last_run_id{project="foo/bar",ref="master",topics="",variables="FOO:BAR"} 1.40420947e+08
# HELP gitlab_ci_pipeline_last_run_status Status of the most recent pipeline
# TYPE gitlab_ci_pipeline_last_run_status gauge
gitlab_ci_pipeline_last_run_status{project="foo/project",ref="dev",status="canceled",topics="",variables=""} 0
gitlab_ci_pipeline_last_run_status{project="foo/project",ref="dev",status="failed",topics="",variables=""} 1
gitlab_ci_pipeline_last_run_status{project="foo/project",ref="dev",status="manual",topics="",variables=""} 0
gitlab_ci_pipeline_last_run_status{project="foo/project",ref="dev",status="pending",topics="",variables=""} 0
gitlab_ci_pipeline_last_run_status{project="foo/project",ref="dev",status="running",topics="",variables=""} 0
gitlab_ci_pipeline_last_run_status{project="foo/project",ref="dev",status="skipped",topics="",variables=""} 0
gitlab_ci_pipeline_last_run_status{project="foo/project",ref="dev",status="success",topics="",variables=""} 0
gitlab_ci_pipeline_last_run_status{project="foo/project",ref="master",status="canceled",topics="",variables=""} 0
gitlab_ci_pipeline_last_run_status{project="foo/project",ref="master",status="failed",topics="",variables=""} 0
gitlab_ci_pipeline_last_run_status{project="foo/project",ref="master",status="manual",topics="",variables=""} 0
gitlab_ci_pipeline_last_run_status{project="foo/project",ref="master",status="pending",topics="",variables=""} 0
gitlab_ci_pipeline_last_run_status{project="foo/project",ref="master",status="running",topics="",variables=""} 0
gitlab_ci_pipeline_last_run_status{project="foo/project",ref="master",status="skipped",topics="",variables=""} 0
gitlab_ci_pipeline_last_run_status{project="foo/project",ref="master",status="success",topics="",variables=""} 1
gitlab_ci_pipeline_last_run_status{project="bar/project",ref="master",status="canceled",topics="",variables=""} 0
gitlab_ci_pipeline_last_run_status{project="bar/project",ref="master",status="failed",topics="",variables=""} 0
gitlab_ci_pipeline_last_run_status{project="bar/project",ref="master",status="manual",topics="",variables=""} 0
gitlab_ci_pipeline_last_run_status{project="bar/project",ref="master",status="pending",topics="",variables=""} 0
gitlab_ci_pipeline_last_run_status{project="bar/project",ref="master",status="running",topics="",variables=""} 0
gitlab_ci_pipeline_last_run_status{project="bar/project",ref="master",status="skipped",topics="",variables=""} 0
gitlab_ci_pipeline_last_run_status{project="bar/project",ref="master",status="success",topics="",variables=""} 1
gitlab_ci_pipeline_last_run_status{project="foo/bar",ref="master",status="canceled",topics="",variables="FOO:BAR"} 0
gitlab_ci_pipeline_last_run_status{project="foo/bar",ref="master",status="failed",topics="",variables="FOO:BAR"} 0
gitlab_ci_pipeline_last_run_status{project="foo/bar",ref="master",status="manual",topics="",variables="FOO:BAR"} 0
gitlab_ci_pipeline_last_run_status{project="foo/bar",ref="master",status="pending",topics="",variables="FOO:BAR"} 0
gitlab_ci_pipeline_last_run_status{project="foo/bar",ref="master",status="running",topics="",variables="FOO:BAR"} 0
gitlab_ci_pipeline_last_run_status{project="foo/bar",ref="master",status="skipped",topics="",variables="FOO:BAR"} 0
gitlab_ci_pipeline_last_run_status{project="foo/bar",ref="master",status="success",topics="",variables="FOO:BAR"} 1
# HELP gitlab_ci_pipeline_run_count GitLab CI pipeline run count
# TYPE gitlab_ci_pipeline_run_count counter
gitlab_ci_pipeline_run_count{project="foo/project",ref="dev",topics="",variables=""} 1
gitlab_ci_pipeline_run_count{project="foo/project",ref="master",topics="",variables=""} 2
gitlab_ci_pipeline_run_count{project="bar/project",ref="master",topics="",variables=""} 1
gitlab_ci_pipeline_run_count{project="foo/bar",ref="master",topics="",variables="FOO:BAR"} 2
# HELP gitlab_ci_pipeline_time_since_last_run_seconds Elapsed time since most recent GitLab CI pipeline run.
# TYPE gitlab_ci_pipeline_time_since_last_run_seconds gauge
gitlab_ci_pipeline_time_since_last_run_seconds{project="foo/project",ref="dev",topics="",variables=""} 4.3368877e+07
gitlab_ci_pipeline_time_since_last_run_seconds{project="foo/project",ref="master",topics="",variables=""} 4.151883e+06
gitlab_ci_pipeline_time_since_last_run_seconds{project="bar/project",ref="master",topics="",variables=""} 1.907042e+06
gitlab_ci_pipeline_time_since_last_run_seconds{project="foo/bar",ref="master",topics="",variables="FOO:BAR"} 65456
```

If `fetch_pipeline_job_metrics` is enabled, expect additional metrics:

```bash
~$ curl -s http://localhost:8080/metrics | grep job
# HELP gitlab_ci_pipeline_job_run_count GitLab CI pipeline job run count
# TYPE gitlab_ci_pipeline_job_run_count counter
gitlab_ci_pipeline_job_run_count{job="build",project="bar/project",ref="master",stage="build",topics=""} 1
gitlab_ci_pipeline_job_run_count{job="test",project="bar/project",ref="master",stage="build",topics=""} 1
# HELP gitlab_ci_pipeline_last_job_run_artifact_size Filesize of the most recent job artifacts
# TYPE gitlab_ci_pipeline_last_job_run_artifact_size gauge
gitlab_ci_pipeline_last_job_run_artifact_size{job="build",project="bar/project",ref="master",stage="build",topics=""} 1.3793677e+07
gitlab_ci_pipeline_last_job_run_artifact_size{job="test",project="bar/project",ref="master",stage="build",topics=""} 257737
# HELP gitlab_ci_pipeline_last_job_run_duration_seconds Duration of last job run
# TYPE gitlab_ci_pipeline_last_job_run_duration_seconds gauge
gitlab_ci_pipeline_last_job_run_duration_seconds{job="build",project="bar/project",ref="master",stage="build",topics=""} 826.064469
gitlab_ci_pipeline_last_job_run_duration_seconds{job="test",project="bar/project",ref="master",stage="test",topics=""} 519.873374
# HELP gitlab_ci_pipeline_last_job_run_status Status of the most recent job
# TYPE gitlab_ci_pipeline_last_job_run_status gauge
gitlab_ci_pipeline_last_job_run_status{job="build",project="bar/project",ref="master",stage="build",status="failed",topics=""} 0
gitlab_ci_pipeline_last_job_run_status{job="build",project="bar/project",ref="master",stage="build",status="running",topics=""} 0
gitlab_ci_pipeline_last_job_run_status{job="build",project="bar/project",ref="master",stage="build",status="success",topics=""} 1
gitlab_ci_pipeline_last_job_run_status{job="test",project="bar/project",ref="master",stage="test",status="failed",topics=""} 1
gitlab_ci_pipeline_last_job_run_status{job="test",project="bar/project",ref="master",stage="test",status="running",topics=""} 0
gitlab_ci_pipeline_last_job_run_status{job="test",project="bar/project",ref="master",stage="test",status="success",topics=""} 0 
# HELP gitlab_ci_pipeline_time_since_last_job_run_seconds Elapsed time since most recent GitLab CI job run.
# TYPE gitlab_ci_pipeline_time_since_last_job_run_seconds gauge
gitlab_ci_pipeline_time_since_last_job_run_seconds{job="build",project="bar/project",ref="master",stage="build",topics=""} 52422
gitlab_ci_pipeline_time_since_last_job_run_seconds{job="test",project="bar/project",ref="master",stage="test",topics=""} 1.0260727e+07
```

If the number of metrics generated by fetching jobs becomes a problem, you can enable `output_sparse_status_metrics` on a global, per-project or per-wildcard basis. When enabled, only labels matching the previous pipeline or job status will be submitted (with value `1`) rather than all label combinations submitted but with `0` value where the status does not match the previous run, for example:

```bash
# output_sparse_status_metrics: false
gitlab_ci_pipeline_last_job_run_status{job="test",project="bar/project",ref="master",stage="test",status="canceled",topics=""} 0
gitlab_ci_pipeline_last_job_run_status{job="test",project="bar/project",ref="master",stage="test",status="failed",topics=""} 0
gitlab_ci_pipeline_last_job_run_status{job="test",project="bar/project",ref="master",stage="test",status="manual",topics=""} 0
gitlab_ci_pipeline_last_job_run_status{job="test",project="bar/project",ref="master",stage="test",status="pending",topics=""} 0
gitlab_ci_pipeline_last_job_run_status{job="test",project="bar/project",ref="master",stage="test",status="running",topics=""} 0
gitlab_ci_pipeline_last_job_run_status{job="test",project="bar/project",ref="master",stage="test",status="skipped",topics=""} 0
gitlab_ci_pipeline_last_job_run_status{job="test",project="bar/project",ref="master",stage="test",status="success",topics=""} 1

# output_sparse_status_metrics: true
gitlab_ci_pipeline_last_job_run_status{job="test",project="bar/project",ref="master",stage="test",status="success",topics=""} 1
```

## Usage

```bash
~$ gitlab-ci-pipelines-exporter --help
NAME:
   gitlab-ci-pipelines-exporter - Export metrics about GitLab CI pipelines statuses

USAGE:
   gitlab-ci-pipelines-exporter [global options] command [command options] [arguments...]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --log-level level                               log level (debug,info,warn,fatal,panic) (default: "info") [$GCPE_LOG_LEVEL]
   --log-format format                             log format (json,text) (default: "text") [$GCPE_LOG_FORMAT]
   --listen-address address:port, -l address:port  listen-address address:port (default: ":8080") [$GCPE_LISTEN_ADDRESS]
   --config file, -c file                          config file (default: "~/.gitlab-ci-pipelines-exporter.yml") [$GCPE_CONFIG]
   --gitlab-token token                            GitLab access token. Can be use to override the gitlab token in config file [$GCPE_GITLAB_TOKEN]
   --help, -h                                      show help
   --version, -v                  print the version
```

## HELM

If you want to make it run on [kubernetes](https://kubernetes.io/), there is a [helm chart](https://docs.helm.sh/) for that!

```bash
~$ git clone git@github.com:mvisonneau/gitlab-ci-pipelines-exporter.git
~$ cat <<EOF > values.yml
config:
  gitlab:
    url: https://gitlab.example.com
    token: xrN14n9-ywvAFxxxxxx
  projects:
    - name: foo/project
    - name: bar/project
  wildcards:
    - owner:
        name: foo
        kind: group
EOF
~$ helm upgrade -i gitlab-ci-pipelines-exporter ./chart -f values.yml
```

## Develop / Test

If you use docker, you can easily get started using :

```bash
~$ make dev-env
# You should then be able to use go commands to work onto the project, eg:
~docker$ make fmt
~docker$ gitlab-ci-pipelines-exporter
```

## Contribute

Contributions are more than welcome! Feel free to submit a [PR](https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/pulls).
