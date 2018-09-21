# mvisonneau/gitlab-ci-pipelines-exporter

[![GoDoc](https://godoc.org/github.com/mvisonneau/gitlab-ci-pipelines-exporter?status.svg)](https://godoc.org/github.com/mvisonneau/gitlab-ci-pipelines-exporter/app)
[![Go Report Card](https://goreportcard.com/badge/github.com/mvisonneau/gitlab-ci-pipelines-exporter)](https://goreportcard.com/report/github.com/mvisonneau/gitlab-ci-pipelines-exporter)
[![Docker Pulls](https://img.shields.io/docker/pulls/mvisonneau/gitlab-ci-pipelines-exporter.svg)](https://hub.docker.com/r/mvisonneau/gitlab-ci-pipelines-exporter/)
[![Build Status](https://travis-ci.org/mvisonneau/gitlab-ci-pipelines-exporter.svg?branch=master)](https://travis-ci.org/mvisonneau/gitlab-ci-pipelines-exporter)
[![Coverage Status](https://coveralls.io/repos/github/mvisonneau/gitlab-ci-pipelines-exporter/badge.svg?branch=master)](https://coveralls.io/github/mvisonneau/gitlab-ci-pipelines-exporter?branch=master)

`gitlab-ci-pipelines-exporter` is a very small binary that allows you to monitor your [GitLab CI pipelines](https://docs.gitlab.com/ee/ci/pipelines.html) with [Prometheus](https://prometheus.io/) or any monitoring solution supporting the [OpenMetrics](https://github.com/OpenObservability/OpenMetrics) format.

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
