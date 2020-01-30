# 🦊 gitlab-ci-pipelines-exporter

[![GoDoc](https://godoc.org/github.com/mvisonneau/gitlab-ci-pipelines-exporter?status.svg)](https://godoc.org/github.com/mvisonneau/gitlab-ci-pipelines-exporter/app)
[![Go Report Card](https://goreportcard.com/badge/github.com/mvisonneau/gitlab-ci-pipelines-exporter)](https://goreportcard.com/report/github.com/mvisonneau/gitlab-ci-pipelines-exporter)
[![Docker Pulls](https://img.shields.io/docker/pulls/mvisonneau/gitlab-ci-pipelines-exporter.svg)](https://hub.docker.com/r/mvisonneau/gitlab-ci-pipelines-exporter/)
[![Build Status](https://cloud.drone.io/api/badges/mvisonneau/gitlab-ci-pipelines-exporter/status.svg)](https://cloud.drone.io/mvisonneau/gitlab-ci-pipelines-exporter)
[![Coverage Status](https://coveralls.io/repos/github/mvisonneau/gitlab-ci-pipelines-exporter/badge.svg?branch=master)](https://coveralls.io/github/mvisonneau/gitlab-ci-pipelines-exporter?branch=master)

`gitlab-ci-pipelines-exporter` is a small binary (~10MB) which allows you to monitor your [GitLab CI pipelines](https://docs.gitlab.com/ee/ci/pipelines.html) with [Prometheus](https://prometheus.io/) or any monitoring solution supporting the [OpenMetrics](https://github.com/OpenObservability/OpenMetrics) format.

## TL:DR

Here is a [Grafana dashboard](https://grafana.com/grafana/dashboards/10620) I have been able to craft, using those metrics:

![grafana_dashboard](/docs/images/grafana_dashboard.png)

If you are solely interested into trying it out, have a look into the [example/](./example) folder which contains documentation to provision test version of the exporter, prometheus and also grafana in ~5min. 

## Getting started

```bash
# Write your config file somewhere on disk
~$ cat <<EOF > $(pwd)/config.yml
# URL and Token with sufficient permissions to access your GitLab's projects
# pipelines informations
gitlab:
  url: https://gitlab.example.com
  token: xrN14n9-ywvAFxxxxxx                         # Gitlab access token. Omit this field to use "GITLAB_TOKEN" from the environment.
  # health_url: https://gitlab.example.com/-/health  # Alternative URL for determining health of GitLab API (readiness probe)
  # skip_tls_verify: false                           # disable TLS verification

# Global rate limit for the GitLab API request/sec
maximum_gitlab_api_requests_per_second: 10

# Custom waiting time between polls for projects, their refs and pipelines (in seconds, optional)
projects_polling_interval_seconds: 1800 # only used for wildcards
refs_polling_interval_seconds: 300
pipelines_polling_interval_seconds: 60
pipelines_max_polling_interval_seconds: 1800 # when no pipeline exists for a given ref, the exporter will exponentially backoff up to this value

# Whether to attempt retrieving refs from pipelines when the exporter starts (default: false)
on_init_fetch_refs_from_pipelines: false
# Maximum number of pipelines to analyze per project to search for refs on init (default: 100)
on_init_fetch_refs_from_pipelines_depth_limit: 100

# Default regexp for parsing the refs (branches and tags) to monitor (optional, default to master)
# default_refs: "^master$"

# The list of the projects you want to monitor
projects:
  - name: foo/project
  - name: bar/project
    refs: "^master|dev$"

# Dynamically fetch projects to monitor using a wildcard
wildcards:
  # Fetch projects belonging to a group and potentially its subgroups
  - owner:
      name: foo
      kind: group
      include_subgroups: true # optional (default: false)
    refs: "^master|1.0$"
    search: 'bar' # optional (defaults to '')

  # Fetch projects belonging to a specific user
  - owner:
      name: bar
      kind: user
    refs: ".*"
    search: 'bar' # optional (defaults to '')

  # Search for projects globally
  - refs: ".*"
    search: 'baz' # optional (defaults to '')
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
INFO[2019-07-30T18:12:24+01:00] Starting exporter
INFO[2019-07-30T18:12:24+01:00] Configured GitLab endpoint : https://gitlab.example.com
INFO[2019-07-30T18:12:24+01:00] Polling projects every 15s
INFO[2019-07-30T18:12:24+01:00] Polling refs every 10s
INFO[2019-07-30T18:12:24+01:00] Polling pipelines every 60s
INFO[2019-07-30T18:12:24+01:00] 2 project(s) configured
INFO[2019-07-30T18:12:24+01:00] Listing all projects using search pattern : 'bar' with owner 'foo' (group)
INFO[2019-07-30T18:12:24+01:00] Found project : foo/bar
INFO[2019-07-30T18:12:24+01:00] Polling refs for project : foo/project
INFO[2019-07-30T18:12:24+01:00] Polling refs for project : bar/project
INFO[2019-07-30T18:12:24+01:00] Polling refs for project : foo/bar
INFO[2019-07-30T18:12:24+01:00] Found ref 'master' for project 'foo/project'
INFO[2019-07-30T18:12:24+01:00] Found ref 'master' for project 'bar/project'
INFO[2019-07-30T18:12:24+01:00] Found ref 'dev' for project 'bar/project'
INFO[2019-07-30T18:12:24+01:00] Found ref 'master' for project 'foo/bar'
INFO[2019-07-30T18:12:24+01:00] Found ref '1.0' for project 'foo/bar'
INFO[2019-07-30T18:12:24+01:00] Polling foo/project:master (1)
INFO[2019-07-30T18:12:24+01:00] Polling bar/project:master (2)
INFO[2019-07-30T18:12:24+01:00] Polling bar/project:dev (2)
INFO[2019-07-30T18:12:24+01:00] Polling foo/bar:master (1)
INFO[2019-07-30T18:12:24+01:00] Polling foo/bar:1.0 (1)
```

And this is an example of the metrics you should expect to retrieve

```bash
~$ curl -s localhost:8080/metrics | grep gitlab_ci_pipeline
# HELP gitlab_ci_pipeline_coverage Coverage of the most recent pipeline
# TYPE gitlab_ci_pipeline_coverage gauge
gitlab_ci_pipeline_coverage{project="foo/project",ref="dev"} 65.4
# HELP gitlab_ci_pipeline_last_run_duration_seconds Duration of last pipeline run
# TYPE gitlab_ci_pipeline_last_run_duration_seconds gauge
gitlab_ci_pipeline_last_run_duration_seconds{project="bar/project",ref="master"} 676
gitlab_ci_pipeline_last_run_duration_seconds{project="foo/project",ref="master"} 33
gitlab_ci_pipeline_last_run_duration_seconds{project="bar/project",ref="dev"} 701
gitlab_ci_pipeline_last_run_duration_seconds{project="foo/bar",ref="master"} 570
gitlab_ci_pipeline_last_run_duration_seconds{project="foo/bar",ref="1.0"} 571
# HELP gitlab_ci_pipeline_last_run_id ID of the most recent pipeline
# TYPE gitlab_ci_pipeline_last_run_id gauge
gitlab_ci_pipeline_last_run_duration_seconds{project="bar/project",ref="master"} 2.2772738e+07
gitlab_ci_pipeline_last_run_duration_seconds{project="foo/project",ref="master"} 3.0094592e+07
gitlab_ci_pipeline_last_run_duration_seconds{project="bar/project",ref="dev"} 4.0059611e+07
gitlab_ci_pipeline_last_run_duration_seconds{project="foo/bar",ref="master"} 4.082622e+07
gitlab_ci_pipeline_last_run_duration_seconds{project="foo/bar",ref="1.0"} 6.8400336e+07
# HELP gitlab_ci_pipeline_last_run_status Status of the most recent pipeline
# TYPE gitlab_ci_pipeline_last_run_status gauge
gitlab_ci_pipeline_last_run_status{project="bar/project",ref="master",status="failed"} 0
gitlab_ci_pipeline_last_run_status{project="bar/project",ref="master",status="running"} 0
gitlab_ci_pipeline_last_run_status{project="bar/project",ref="master",status="success"} 1
gitlab_ci_pipeline_last_run_status{project="foo/project",ref="master",status="failed"} 0
gitlab_ci_pipeline_last_run_status{project="bar/project",ref="dev",status="running"} 0
gitlab_ci_pipeline_last_run_status{project="bar/project",ref="dev",status="success"} 1
gitlab_ci_pipeline_last_run_status{project="foo/bar",ref="master",status="running"} 0
gitlab_ci_pipeline_last_run_status{project="foo/bar",ref="master",status="success"} 1
gitlab_ci_pipeline_last_run_status{project="foo/bar",ref="1.0",status="running"} 0
gitlab_ci_pipeline_last_run_status{project="foo/bar",ref="1.0",status="success"} 1
# HELP gitlab_ci_pipeline_run_count GitLab CI pipeline run count
# TYPE gitlab_ci_pipeline_run_count counter
gitlab_ci_pipeline_run_count{project="bar/project",ref="master"} 0
gitlab_ci_pipeline_run_count{project="foo/project",ref="master"} 0
gitlab_ci_pipeline_run_count{project="bar/project",ref="dev"} 0
gitlab_ci_pipeline_run_count{project="foo/bar",ref="master"} 0
gitlab_ci_pipeline_run_count{project="foo/bar",ref="1.0"} 0
# HELP gitlab_ci_pipeline_time_since_last_run_seconds Elapsed time since most recent GitLab CI pipeline run.
# TYPE gitlab_ci_pipeline_time_since_last_run_seconds gauge
gitlab_ci_pipeline_time_since_last_run_seconds{project="bar/project",ref="master"} 87627
gitlab_ci_pipeline_time_since_last_run_seconds{project="foo/project",ref="master"} 29531
gitlab_ci_pipeline_time_since_last_run_seconds{project="bar/project",ref="dev"} 2950
gitlab_ci_pipeline_time_since_last_run_seconds{project="foo/bar",ref="master"} 2951
gitlab_ci_pipeline_time_since_last_run_seconds{project="foo/bar",ref="1.0"} 2900
```

## Usage

```bash
~$ gitlab-ci-pipelines-exporter --help
NAME:
   gitlab-ci-pipelines-exporter - Export metrics about GitLab CI pipeliens statuses

USAGE:
   gitlab-ci-pipelines-exporter [global options] command [command options] [arguments...]

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --log-level level              log level (debug,info,warn,fatal,panic) (default: "info") [$GCPE_LOG_LEVEL]
   --log-format format            log format (json,text) (default: "text") [$GCPE_LOG_FORMAT]
   --listen-address address:port  listen-address address:port (default: ":8080") [$GCPE_LISTEN_ADDRESS]
   --config file                  config file (default: "~/.gitlab-ci-pipelines-exporter.yml") [$GCPE_CONFIG]
   --help, -h                     show help
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
