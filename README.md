# mvisonneau/gitlab-ci-pipelines-exporter

[![GoDoc](https://godoc.org/github.com/mvisonneau/gitlab-ci-pipelines-exporter?status.svg)](https://godoc.org/github.com/mvisonneau/gitlab-ci-pipelines-exporter/app)
[![Go Report Card](https://goreportcard.com/badge/github.com/mvisonneau/gitlab-ci-pipelines-exporter)](https://goreportcard.com/report/github.com/mvisonneau/gitlab-ci-pipelines-exporter)
[![Docker Pulls](https://img.shields.io/docker/pulls/mvisonneau/gitlab-ci-pipelines-exporter.svg)](https://hub.docker.com/r/mvisonneau/gitlab-ci-pipelines-exporter/)
[![Build Status](https://travis-ci.org/mvisonneau/gitlab-ci-pipelines-exporter.svg?branch=master)](https://travis-ci.org/mvisonneau/gitlab-ci-pipelines-exporter)
[![Coverage Status](https://coveralls.io/repos/github/mvisonneau/gitlab-ci-pipelines-exporter/badge.svg?branch=master)](https://coveralls.io/github/mvisonneau/gitlab-ci-pipelines-exporter?branch=master)

`gitlab-ci-pipelines-exporter` is a very small binary that allows you to monitor your [GitLab CI pipelines](https://docs.gitlab.com/ee/ci/pipelines.html) with [Prometheus](https://prometheus.io/) or any monitoring solution supporting the [OpenMetrics](https://github.com/OpenObservability/OpenMetrics) format.

## TL:DR

```bash
# Write your config file somewhere on disk
~$ cat <<EOF > $(pwd)/config.yml
# URL and Token with sufficient permissions to access your GitLab's projects
# pipelines informations
gitlab:
  url: https://gitlab.example.com
  token: xrN14n9-ywvAFxxxxxx

# Waiting time between polls for each projects
polling_interval_seconds: 30

# The list of the projects you want to monitor
projects:
  - name: foo/project
    ref: master
  - name: bar/project
    ref: master
EOF

# If you have docker installed, it is as easy as :
~$ docker run -d \
   --name gitlab-ci-pipelines-exporter \
   -v $(pwd)/config.yml:/etc/config.yml \
   -p 8080:80 \
   mvisonneau/gitlab-ci-pipelines-exporter:latest \
   -config /etc/config.yml
```

You should then be able to see the following logs

```
~$ docker logs -f gitlab-ci-pipelines-exporter
2018/09/21 12:44:05 -> Starting exporter
2018/09/21 12:44:05 -> Polling https://gitlab.example.com every 30s
2018/09/21 12:44:05 -> 2 project(s) configured
2018/09/21 12:44:05 --> Polling ID: 1 | foo/project:master
2018/09/21 12:44:05 --> Polling ID: 2 | bar/project:master
```

And this is an example of the metrics could expect to retrieve

```
~$ curl -s localhost:8080/metrics | grep gitlab_ci_pipeline
# HELP gitlab_ci_pipeline_last_run_duration_seconds Duration of last pipeline run
# TYPE gitlab_ci_pipeline_last_run_duration_seconds gauge
gitlab_ci_pipeline_last_run_duration_seconds{project="bar/project",ref="master"} 676
gitlab_ci_pipeline_last_run_duration_seconds{project="foo/project",ref="master"} 33
# HELP gitlab_ci_pipeline_run_count GitLab CI pipeline run count
# TYPE gitlab_ci_pipeline_run_count counter
gitlab_ci_pipeline_run_count{project="bar/project",ref="master"} 0
gitlab_ci_pipeline_run_count{project="foo/project",ref="master"} 0
# HELP gitlab_ci_pipeline_status GitLab CI pipeline current status
# TYPE gitlab_ci_pipeline_status gauge
gitlab_ci_pipeline_status{project="bar/project",ref="master",status="failed"} 0
gitlab_ci_pipeline_status{project="bar/project",ref="master",status="running"} 0
gitlab_ci_pipeline_status{project="bar/project",ref="master",status="success"} 1
gitlab_ci_pipeline_status{project="foo/project",ref="master",status="failed"} 0
gitlab_ci_pipeline_status{project="foo/project",ref="master",status="running"} 0
gitlab_ci_pipeline_status{project="foo/project",ref="master",status="success"} 1
# HELP gitlab_ci_pipeline_time_since_last_run_seconds Elapsed time since most recent GitLab CI pipeline run.
# TYPE gitlab_ci_pipeline_time_since_last_run_seconds gauge
gitlab_ci_pipeline_time_since_last_run_seconds{project="bar/project",ref="master"} 87627
gitlab_ci_pipeline_time_since_last_run_seconds{project="foo/project",ref="master"} 29531
```

## Usage

```
~$ gitlab-ci-pipelines-exporter -h
Usage of .gitlab-ci-pipelines-exporter:
  -config string
    	Config file path (default "~/.gitlab-ci-pipelines-exporter.yml")
  -listen-address string
    	Listening address (default ":80")
```

## HELM

If you want to make it run on [kubernetes](https://kubernetes.io/), there is a [helm chart](https://docs.helm.sh/) for that!

```
~$ git clone git@github.com:mvisonneau/gitlab-ci-pipelines-exporter.git
~$ cd gitlab-ci-pipelines-exporter/charts
~$ cat <<EOF > values.yml
gitlab:
  url: https://gitlab.example.com
  token: xrN14n9-ywvAFxxxxxx
polling_interval_seconds: 30
projects:
  - name: foo/project
    ref: master
  - name: bar/project
    ref: master
EOF
~$ helm package gitlab-ci-pipelines-exporter
~$ helm upgrade -i gitlab-ci-pipelines-exporter ./gitlab-ci-pipelines-exporter-0.0.0.tgz -f values.yml
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
