# 🦊 gitlab-ci-pipelines-exporter

[![PkgGoDev](https://pkg.go.dev/badge/github.com/mvisonneau/gitlab-ci-pipelines-exporter)](https://pkg.go.dev/mod/github.com/mvisonneau/gitlab-ci-pipelines-exporter)
[![Go Report Card](https://goreportcard.com/badge/github.com/mvisonneau/gitlab-ci-pipelines-exporter)](https://goreportcard.com/report/github.com/mvisonneau/gitlab-ci-pipelines-exporter)
[![Docker Pulls](https://img.shields.io/docker/pulls/mvisonneau/gitlab-ci-pipelines-exporter.svg)](https://hub.docker.com/r/mvisonneau/gitlab-ci-pipelines-exporter/)
[![test](https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/actions/workflows/test.yml/badge.svg)](https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/actions/workflows/test.yml)
[![Coverage Status](https://coveralls.io/repos/github/mvisonneau/gitlab-ci-pipelines-exporter/badge.svg?branch=main)](https://coveralls.io/github/mvisonneau/gitlab-ci-pipelines-exporter?branch=main)
[![release](https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/actions/workflows/release.yml/badge.svg)](https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/actions/workflows/release.yml)
[![gitlab-ci-pipelines-exporter](https://snapcraft.io/gitlab-ci-pipelines-exporter/badge.svg)](https://snapcraft.io/gitlab-ci-pipelines-exporter)

`gitlab-ci-pipelines-exporter` allows you to monitor your [GitLab CI pipelines](https://docs.gitlab.com/ee/ci/pipelines/) with [Prometheus](https://prometheus.io/) or any monitoring solution supporting the [OpenMetrics](https://github.com/OpenObservability/OpenMetrics) format.

You can find more information [on GitLab docs](https://docs.gitlab.com/ee/ci/pipelines/pipeline_efficiency.html#pipeline-monitoring) about how it takes part improving your pipeline efficiency.

## TL:DR

Here are some [Grafana dashboards](https://grafana.com/grafana/dashboards/10620) I was able to craft using those metrics. Otherwise, the detailed list of **exported metrics** [is maintained here](docs/metrics.md).

### Pipelines

![grafana_dashboard_pipelines](/docs/images/grafana_dashboard_pipelines.jpg)

_grafana.com dashboard_ [#10620](https://grafana.com/grafana/dashboards/10620)

### Jobs

![grafana_dashboard_jobs](/docs/images/grafana_dashboard_jobs.jpg)

_grafana.com dashboard_ [#13328](https://grafana.com/grafana/dashboards/13328)

### Environments / Deployments

![grafana_dashboard_environments](/docs/images/grafana_dashboard_environments.jpg)

_grafana.com dashboard_ [#13329](https://grafana.com/grafana/dashboards/13329)

If you want to quickly try them out with your own data, have a look into the [examples/quickstart](./examples/quickstart/README.md) folder which contains documentation to provision test version of the exporter, prometheus and also grafana in **~5min** using `docker-compose`

## Install

### Go

```bash
~$ go run github.com/mvisonneau/gitlab-ci-pipelines-exporter/cmd/gitlab-ci-pipelines-exporter@latest
```

### Snapcraft

```bash
~$ snap install gitlab-ci-pipelines-exporter
```

### Homebrew

```bash
~$ brew install mvisonneau/tap/gitlab-ci-pipelines-exporter
```

### Docker

```bash
~$ docker run -it --rm docker.io/mvisonneau/gitlab-ci-pipelines-exporter
~$ docker run -it --rm ghcr.io/mvisonneau/gitlab-ci-pipelines-exporter
~$ docker run -it --rm quay.io/mvisonneau/gitlab-ci-pipelines-exporter
```

### Scoop

```bash
~$ scoop bucket add https://github.com/mvisonneau/scoops
~$ scoop install gitlab-ci-pipelines-exporter
```

### NixOS

```
~$ nix-env -iA nixos.prometheus-gitlab-ci-pipelines-exporter
```

### Binaries, DEB and RPM packages

Have a look onto the [latest release page](https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/releases/latest) to pick your flavor and version. Here is an helper to fetch the most recent one:

```bash
~$ export GCPE_VERSION=$(curl -s "https://api.github.com/repos/mvisonneau/gitlab-ci-pipelines-exporter/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
```

```bash
# Binary (eg: linux/amd64)
~$ wget https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/releases/download/${GCPE_VERSION}/gitlab-ci-pipelines-exporter_${GCPE_VERSION}_linux_amd64.tar.gz
~$ tar zxvf gitlab-ci-pipelines-exporter_${GCPE_VERSION}_linux_amd64.tar.gz -C /usr/local/bin

# DEB package (eg: linux/386)
~$ wget https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/releases/download/${GCPE_VERSION}/gitlab-ci-pipelines-exporter_${GCPE_VERSION}_linux_386.deb
~$ dpkg -i gitlab-ci-pipelines-exporter_${GCPE_VERSION}_linux_386.deb

# RPM package (eg: linux/arm64)
~$ wget https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/releases/download/${GCPE_VERSION}/gitlab-ci-pipelines-exporter_${GCPE_VERSION}_linux_arm64.rpm
~$ rpm -ivh gitlab-ci-pipelines-exporter_${GCPE_VERSION}_linux_arm64.rpm
```

### HELM

If you want to make it run on [kubernetes](https://kubernetes.io/), there is a [helm chart](https://github.com/mvisonneau/helm-charts/tree/main/charts/gitlab-ci-pipelines-exporter) available for this purpose.

You can check the chart's [values.yml](https://github.com/mvisonneau/helm-charts/blob/main/charts/gitlab-ci-pipelines-exporter/values.yaml) for complete configuration options.

```bash
# Add the helm repository to your local client
~$ helm repo add mvisonneau https://charts.visonneau.fr

# Configure a minimal configuration for the exporter
~$ cat <<EOF > values.yml
config:
  gitlab:
    url: https://gitlab.example.com
    # You can also configure the token using --gitlab-token
    # or the $GCPE_GITLAB_TOKEN environment variable
    token: xrN14n9-ywvAFxxxxxx
  projects:
    - name: foo/project
EOF

# Release the chart on your Kubernetes cluster
~$ helm upgrade -i gitlab-ci-pipelines-exporter mvisonneau/gitlab-ci-pipelines-exporter -f values.yml
```

## Configuration syntax

The **configuration syntax** [is maintained here](docs/configuration_syntax.md).

## Quickstart

```bash
# Write a minimal config file somewhere on disk
~$ cat <<EOF > $(pwd)/config.yml
gitlab:
  url: https://gitlab.example.com
  # You can also configure the token using --gitlab-token
  # or the $GCPE_GITLAB_TOKEN environment variable
  token: <your_token>
projects:
  - name: foo/project
  - name: bar/project
wildcards:
  - owner:
      name: foo
      kind: group
EOF

# If you have installed the binary
~$ gitlab-ci-pipelines-exporter --config /etc/config.yml

# Otherwise if you have docker available, it is as easy as :
~$ docker run -it --rm \
   --name gitlab-ci-pipelines-exporter \
   -v $(pwd)/config.yml:/etc/config.yml \
   -p 8080:8080 \
   mvisonneau/gitlab-ci-pipelines-exporter:latest \
   run --config /etc/config.yml
```

You should then be able to see the following logs

```bash
INFO[0000] starting exporter                             gitlab-endpoint="https://gitlab.com" on-init-fetch-refs-from-pipelines=true pulling-pipelines-every=60s pulling-projects-every=15s pulling-refs-every=10s pulling-workers=2 rate-limit=10rps
INFO[0000] configured wildcards                          count=1
INFO[0000] found new project                             project-name=foo/project wildcard-archived=false wildcard-owner-include-subgroups=false wildcard-owner-kind=group wildcard-owner-name=foo wildcard-search=
INFO[0000] found new project                             project-name=foo/bar wildcard-archived=false wildcard-owner-include-subgroups=false wildcard-owner-kind=group wildcard-owner-name=foo wildcard-search=
INFO[0000] configured projects                           count=3
INFO[0000] started, now serving requests                 listen-address=":8080"
INFO[0000] found project refs                            project-path-with-namespace=foo/project project-ref=main
INFO[0000] found project refs                            project-path-with-namespace=bar/project project-ref=main
INFO[0000] found project refs                            project-path-with-namespace=foo/bar project-ref=main
```

And this is an example of the metrics you should expect to retrieve, the **detailed list of exported metrics** [is maintained here](docs/metrics.md).

```bash
~$ curl -s localhost:8080/metrics | grep gitlab_ci
# HELP gitlab_ci_pipeline_last_run_duration_seconds Duration of last pipeline run
# TYPE gitlab_ci_pipeline_last_run_duration_seconds gauge
gitlab_ci_pipeline_last_run_duration_seconds{project="foo/project",ref="dev",topics="",variables=""} 81
gitlab_ci_pipeline_last_run_duration_seconds{project="foo/project",ref="main",topics="",variables=""} 420
gitlab_ci_pipeline_last_run_duration_seconds{project="bar/project",ref="main",topics="",variables=""} 334
gitlab_ci_pipeline_last_run_duration_seconds{project="foo/bar",ref="main",topics="",variables="FOO:BAR"} 55
# HELP gitlab_ci_pipeline_last_run_id ID of the most recent pipeline
# TYPE gitlab_ci_pipeline_last_run_id gauge
gitlab_ci_pipeline_last_run_id{project="foo/project",ref="dev",topics="",variables=""} 4.0059611e+07
gitlab_ci_pipeline_last_run_id{project="foo/project",ref="main",topics="",variables=""} 1.25351545e+08
gitlab_ci_pipeline_last_run_id{project="bar/project",ref="main",topics="",variables=""} 1.33308085e+08
gitlab_ci_pipeline_last_run_id{project="foo/bar",ref="main",topics="",variables="FOO:BAR"} 1.40420947e+08
# HELP gitlab_ci_pipeline_last_run_status Status of the most recent pipeline
# TYPE gitlab_ci_pipeline_last_run_status gauge
gitlab_ci_pipeline_last_run_status{project="foo/project",ref="dev",status="canceled",topics="",variables=""} 0
gitlab_ci_pipeline_last_run_status{project="foo/project",ref="dev",status="failed",topics="",variables=""} 1
gitlab_ci_pipeline_last_run_status{project="foo/project",ref="dev",status="manual",topics="",variables=""} 0
gitlab_ci_pipeline_last_run_status{project="foo/project",ref="dev",status="pending",topics="",variables=""} 0
gitlab_ci_pipeline_last_run_status{project="foo/project",ref="dev",status="running",topics="",variables=""} 0
gitlab_ci_pipeline_last_run_status{project="foo/project",ref="dev",status="skipped",topics="",variables=""} 0
gitlab_ci_pipeline_last_run_status{project="foo/project",ref="dev",status="success",topics="",variables=""} 0
gitlab_ci_pipeline_last_run_status{project="foo/project",ref="main",status="canceled",topics="",variables=""} 0
gitlab_ci_pipeline_last_run_status{project="foo/project",ref="main",status="failed",topics="",variables=""} 0
gitlab_ci_pipeline_last_run_status{project="foo/project",ref="main",status="manual",topics="",variables=""} 0
gitlab_ci_pipeline_last_run_status{project="foo/project",ref="main",status="pending",topics="",variables=""} 0
gitlab_ci_pipeline_last_run_status{project="foo/project",ref="main",status="running",topics="",variables=""} 0
gitlab_ci_pipeline_last_run_status{project="foo/project",ref="main",status="skipped",topics="",variables=""} 0
gitlab_ci_pipeline_last_run_status{project="foo/project",ref="main",status="success",topics="",variables=""} 1
gitlab_ci_pipeline_last_run_status{project="bar/project",ref="main",status="canceled",topics="",variables=""} 0
gitlab_ci_pipeline_last_run_status{project="bar/project",ref="main",status="failed",topics="",variables=""} 0
gitlab_ci_pipeline_last_run_status{project="bar/project",ref="main",status="manual",topics="",variables=""} 0
gitlab_ci_pipeline_last_run_status{project="bar/project",ref="main",status="pending",topics="",variables=""} 0
gitlab_ci_pipeline_last_run_status{project="bar/project",ref="main",status="running",topics="",variables=""} 0
gitlab_ci_pipeline_last_run_status{project="bar/project",ref="main",status="skipped",topics="",variables=""} 0
gitlab_ci_pipeline_last_run_status{project="bar/project",ref="main",status="success",topics="",variables=""} 1
gitlab_ci_pipeline_last_run_status{project="foo/bar",ref="main",status="canceled",topics="",variables="FOO:BAR"} 0
gitlab_ci_pipeline_last_run_status{project="foo/bar",ref="main",status="failed",topics="",variables="FOO:BAR"} 0
gitlab_ci_pipeline_last_run_status{project="foo/bar",ref="main",status="manual",topics="",variables="FOO:BAR"} 0
gitlab_ci_pipeline_last_run_status{project="foo/bar",ref="main",status="pending",topics="",variables="FOO:BAR"} 0
gitlab_ci_pipeline_last_run_status{project="foo/bar",ref="main",status="running",topics="",variables="FOO:BAR"} 0
gitlab_ci_pipeline_last_run_status{project="foo/bar",ref="main",status="skipped",topics="",variables="FOO:BAR"} 0
gitlab_ci_pipeline_last_run_status{project="foo/bar",ref="main",status="success",topics="",variables="FOO:BAR"} 1
# HELP gitlab_ci_pipeline_run_count GitLab CI pipeline run count
# TYPE gitlab_ci_pipeline_run_count counter
gitlab_ci_pipeline_run_count{project="foo/project",ref="dev",topics="",variables=""} 1
gitlab_ci_pipeline_run_count{project="foo/project",ref="main",topics="",variables=""} 2
gitlab_ci_pipeline_run_count{project="bar/project",ref="main",topics="",variables=""} 1
gitlab_ci_pipeline_run_count{project="foo/bar",ref="main",topics="",variables="FOO:BAR"} 2
# HELP gitlab_ci_pipeline_time_since_last_run_seconds Elapsed time since most recent GitLab CI pipeline run.
# TYPE gitlab_ci_pipeline_time_since_last_run_seconds gauge
gitlab_ci_pipeline_time_since_last_run_seconds{project="foo/project",ref="dev",topics="",variables=""} 4.3368877e+07
gitlab_ci_pipeline_time_since_last_run_seconds{project="foo/project",ref="main",topics="",variables=""} 4.151883e+06
gitlab_ci_pipeline_time_since_last_run_seconds{project="bar/project",ref="main",topics="",variables=""} 1.907042e+06
gitlab_ci_pipeline_time_since_last_run_seconds{project="foo/bar",ref="main",topics="",variables="FOO:BAR"} 65456
```

## HA implementation

It supports running multiple instances of the exporter in an HA fashion leveraging redis as storage middleware. You simply need to set a redis URL in the `config.yml` or using the `--redis-url` flag or `$GCPE_REDIS_URL` env variable. A quick example using docker-compose is also available here: [examples/ha-setup](examples/ha-setup/README.md)

### How it works

- Pulling of all of the GitLab resources (projects, refs, pipelines, jobs, etc..) is spread evenly across all the running instances
- Rate limit is global across the workers. eg: 3 workers at a 10 rps limit will result in a ~3.3rps limit/worker
- Exported metrics are fetched from the shared storage layer on each call to ensure data integrity/consistency of the requests across the instances

## Push based implementation (leveraging GitLab webhooks)

The exporter supports receiving project pipeline events through GitLab webhooks on the `/webhook` path. This feature is not enabled by default and requires the following parameters to be set in the `config.yml`:

```yaml
server:
   webhook:
      enabled: true
      secret_token: <a_secret_token>
```

A complete example is available here: [examples/webhooks](examples/webhooks/README.md). You can also refer to the [configuration syntax](docs/configuration_syntax.md) for me information.

## Usage

```bash
~$ gitlab-ci-pipelines-exporter --help
NAME:
   gitlab-ci-pipelines-exporter - Export metrics about GitLab CI pipelines statuses

USAGE:
   gitlab-ci-pipelines-exporter [global options] command [command options] [arguments...]

COMMANDS:
   run       start the exporter
   validate  validate the configuration file
   monitor   display information about the currently running exporter
   help, h   Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --internal-monitoring-listener-address value, -m value  internal monitoring listener address [$GCPE_INTERNAL_MONITORING_LISTENER_ADDRESS]
   --help, -h                                              show help
   --version, -v                                           print the version

```

### run

```bash
~$ gitlab-ci-pipelines-exporter run --help
NAME:
   gitlab-ci-pipelines-exporter run - start the exporter

USAGE:
   gitlab-ci-pipelines-exporter run [command options] [arguments...]

OPTIONS:
   --config file, -c file        config file (default: "./gitlab-ci-pipelines-exporter.yml") [$GCPE_CONFIG]
   --redis-url url               redis url for an HA setup (format: redis[s]://[:password@]host[:port][/db-number][?option=value]) (overrides config file parameter) [$GCPE_REDIS_URL]
   --gitlab-token token          GitLab API access token (overrides config file parameter) [$GCPE_GITLAB_TOKEN]
   --webhook-secret-token token  token used to authenticate legitimate requests (overrides config file parameter) [$GCPE_WEBHOOK_SECRET_TOKEN]
   --help, -h                    show help (default: false)
```

### validate

```bash
~$ gitlab-ci-pipelines-exporter validate --help
NAME:
   gitlab-ci-pipelines-exporter validate - validate the configuration file

USAGE:
   gitlab-ci-pipelines-exporter validate [command options] [arguments...]

OPTIONS:
   --config file, -c file  config file (default: "./gitlab-ci-pipelines-exporter.yml") [$GCPE_CONFIG]
   --help, -h              show help
```

### monitor

```bash
~$ gitlab-ci-pipelines-exporter monitor --help
NAME:
   gitlab-ci-pipelines-exporter monitor - display information about the currently running exporter

USAGE:
   gitlab-ci-pipelines-exporter monitor [command options] [arguments...]

OPTIONS:
   --help, -h  show help (default: false)
```

## Monitor / Troubleshoot

![monitor_cli_example](/docs/images/monitor_cli_example.gif)

If you need to dig into your exporter's internal, you can leverage the internal CLI monitoring endpoint. This will get you insights about the following:
- Live telemetry regarding:
  - GitLab API requests
  - Tasks buffer usage
  - Projects count and schedules
  - Environments count and schedules
  - Refs count and schedules
  - Metrics count and schedules
- **Parsed configuration details**

To use it, you have to start your exporter with the following flag `--internal-monitoring-listener-address`, `-m` or the `GCPE_INTERNAL_MONITORING_LISTENER_ADDRESS` env variable.

You can whether use a TCP or UNIX socket eg:

```
~$ gitlab-ci-pipelines-exporter -m 'unix://gcpe-monitor.sock' run
~$ gitlab-ci-pipelines-exporter -m 'tcp://127.0.0.1:9000' run
```

To use the monitor CLI, you need to be able to access the monitoring socket and reuse the same flag:

```
export GCPE_INTERNAL_MONITORING_LISTENER_ADDRESS='unix://gcpe-monitor.sock'
~$ gitlab-ci-pipelines-exporter run &
~$ gitlab-ci-pipelines-exporter monitor
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
