# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [0ver](https://0ver.org) (more or less).

## [Unreleased]

### Changed

- Disabled taskq consumers system resources checks which may leave the exporter in a hanging state
- Embeded taskq logs into logrus
- Fixed a bug preventing the redis URL to be read from the config
- Bumped goreleaser to v0.144.1

## [0.4.0] - 2020-10-09

**BREAKING CHANGES**

1. The [configuration syntax](./docs/configuration_syntax.md) has been restructured quite a bit and some runtime flags have been moved in it as well. Refer to the updated documentation to check what you may need to update.

2. Some metrics have been renamed:

| Original metric name | New metric name|
|---|---|
|*gitlab_ci_pipeline_last_job_run_artifact_size*|`gitlab_ci_pipeline_job_artifact_size_bytes`|
|*gitlab_ci_pipeline_last_job_run_artifact_size*|`gitlab_ci_pipeline_job_artifact_size_bytes`|
|*gitlab_ci_pipeline_last_job_run_duration_seconds*|`gitlab_ci_pipeline_job_duration_seconds`|
|*gitlab_ci_pipeline_last_job_run_id*|`gitlab_ci_pipeline_job_id`|
|*gitlab_ci_pipeline_last_job_run_status*|`gitlab_ci_pipeline_job_status`|
|*gitlab_ci_pipeline_last_run_duration_seconds*|`gitlab_ci_pipeline_duration_seconds`|
|*gitlab_ci_pipeline_last_run_id*|`gitlab_ci_pipeline_id`|
|*gitlab_ci_pipeline_last_run_status*|`gitlab_ci_pipeline_status`|
|*gitlab_ci_pipeline_time_since_last_job_run_seconds*|`gitlab_ci_pipeline_job_timestamp`|
|*gitlab_ci_pipeline_time_since_last_run_seconds*|`gitlab_ci_pipeline_timestamp`|

3. On top of being renamed, the `.*time_since.*` metrics have been also converted to timestamps.
You will need to update your PromQL queries to leverage the new format. eg: `time() - gitlab_ci_pipeline_timestamp`

4. We now output sparse status metrics by default, if you want to revert to the default behaviour you will need to add this
statement to your config file:

```yaml
defaults:
  output_sparse_status_metrics: false
```

### Added

- **HA configuration capabilities using Redis** (optional feature, [example here](examples/ha-setup/README.md))
- **Push based approach leveraging pipelines & jobs webhooks** (optional feature, [example here](examples/webhooks/README.md))
- gosec testing

### Changed

- Upgraded `urfave/cli` to **v2**
- Refactored the codebase to make it compliant with golang standards and more domain-driven
- Included the version of the app in the user agent of GitLab queries
- Rewritten the scheduling of the polling using `vmihailenco/taskq`
- Updated the rate limiter to work globally across several workers
- Fixed an issue preventing the jobs from being updated accordingly when restarted
- Updated the example grafana dashboard with the new metrics naming
- Bumped Grafana and Prometheus versions in the example

### Removed

- `polling_workers` configuration parameter

## [0.3.5] - 2020-09-17

### Changed

- Health endpoints to avoid issues with default configuration
- Bumped go-gitlab to `0.38.1`
- Bumped golang to `1.15`
- Switch default branch to **main**

## [0.3.4] - 2020-07-23

### Added

- New `gitlab_ci_pipeline_last_job_run_id` metric which returns the ID of the most recent job run.

### Changed

- Fixed some issues with the polling of the jobs information which led to innacurate results.
- Bumped all dependencies
  - goreleaser to `0.140.0`
  - go-gitlab to `0.33.0`

## [0.3.3] - 2020-06-09

### Changed

- Fixed a bug where `gitlab_ci_pipeline_time_since_last_run_seconds` and `gitlab_ci_pipeline_time_since_last_job_run_seconds` would not get updated after being fetched for the first time on each pipelines (#106)

## [0.3.2] - 2020-05-27

### Changed

- Fixed a bug where `gitlab_ci_pipeline_last_run_status` would not get updated after being fetched for the first time (#102)
- Fixed a bug on `gitlab_ci_pipeline_run_count`, not being updated when a job in a pipeline gets restarted (linked to #102)
- Bumped all dependencies
  - goreleaser to `0.136.0`
  - go-gitlab to `0.32.0`

## [0.3.1] - 2020-04-30

### Added

- Added `--enable-pprof` flag which provides pprof http endpoint at **/debug/pprof**

### Changed

- Fixed a critical bug introduced with the refactoring of workers in **v0.3.0** where the exporter would hang if there are more project refs to poll than workers available
- Fixed a bug where multiple go routines were accessing a single variable without semaphore
- Renamed `maximum_projects_poller_workers` into `polling_workers`
- Enhanced signals handling using a global context with derivatives throughout go routines

## [0.3.0] - 2020-04-29

### **BREAKING CHANGES**

- `default_refs` & `refs` parameter have been renamed respectively to `default_refs_regexp` and `refs_regexp` to make them more explicit
- In the config, there is a new `defaults` parameter group for any setting which can be overridden at the `project` or `wildcard` level. It includes the following parameters:
  - **fetch_pipeline_job_metrics**
  - **fetch_pipeline_variables**
  - **output_sparse_status_metrics**
  - **pipeline_variables_filter_regex**
  - **refs_regexp**
- Renamed the following parameters (their behaviour remains the same):
  - `projects_polling_interval_seconds` into `wildcards_projects_discover_interval_seconds`
  - `refs_polling_interval_seconds` into `projects_refs_discover_interval_seconds`
  - `pipelines_polling_interval_seconds` into `projects_refs_polling_interval_seconds`

### Added

- `kind` label on all metrics which reflects the type of the ref : branch, tag or merge-request
- project/wildcard parameters `fetch_merge_request_pipelines_refs` and `fetch_merge_request_pipelines_refs_init_limit` to enable the metrics polling of merge requests pipelines
- Configuration for OpenMetrics Encoding in metrics HTTP endpoint. Enabled by default but can be disable using `disable_openmetrics_encoding: true`
- Worker pool for projects polling: set `maximum_projects_poller_workers` with an integer value to control parallelism (defaults to `runtime.GOMAXPROCS(0)`)  
- Augmented `disable_tls_verify` with `disable_health_check` additional parameter to drive the behaviour of checking healthiness of target service 
- Reading pipeline variables if enabled setting `fetch_pipeline_variables` to `true` (defaults to `false`)
- Pipeline variables can be filtered with `pipeline_variables_filter_regex` (defaults to `.*`)
- Configurable ServiceMonitor resource through the helm chart

### Changed

- Projects polling from GitLab API is done in parallel using `maximum_projects_poller_workers` pollers and concurrently fetching refs and projects
- Fixed a bug causing duplicate metrics when status changes with sparse flag enabled
- Updated labels syntax in helm chart to comply with standards
- Updated logging, using more extensively the log.WithFields parameter for an enhanced troubleshooting experience
- Bumped prometheus/client_golang to `1.6`

## [0.2.14] - 2020-04-09

### Added

- Support for pipelines status : `manual`

### Changed

- Bumped **go-gitlab** to `v0.31.0` which includes an exponentional backoff retry mechanism on API errors
- Renamed the `job` label into `job_name`
- Fixed a bug in the helm deployment when using service labels

## [0.2.13] - 2020-03-27

### Added

- **new** `fetch_pipeline_job_metrics` configuration flag (default `false`).
When enabled, various statistics for the jobs from the last pipeline run will be collected.

- **new** `output_sparse_status_metrics` flag (default `false`).
When enabled, only reports the status metric currently matching the last pipeline run.
Reduces reported metric count, at the cost of status values being expired from storage
if not seen in a long time.

### Changed

- Corrected the ordering of variable assigments in the assertion tests functions
- Updated the user agent to `gitlab-ci-pipelines-exporter`
- Bumped goreleaser to 0.129.0

## [0.2.12] - 2020-03-09

### Added

- Now using stretchr/testify for asserting test results
- Capability to filter in/out archived projects

### Changed

- Fix `--gitlab-token` and improve docs/chart
- Bumped to go 1.14
- Bumped goreleaser to 0.128.0

## [0.2.11] - 2020-02-03

### Added

- Added global rate limit capability to avoid hammering GitLab API endpoints
- Added `--gitlab-token` flag. Can be use to specify the gitlab token as flag or env var.

### Changed

- Bumped gitlab & prometheus libaries to their latest versions

## [0.2.10] - 2019-12-20

### Added

- Capability to fetch removed refs by analyzing recent project pipelines
- New label `topics` which gather project topics

### Changed

- Refactored the fetching logic to get faster inits
- Enhanced the logic to prevent fatal failures on connectivity issues
- Bumped go librairies to their latest versions

## [0.2.9] - 2019-11-15

### Added

- New `gitlab_ci_pipeline_coverage` metric that fetches the coverage value of the most recent pipeline [GH-32]

### Changed

- Fixed a bug causing panic on DNS lookup failure
- Enhanced the polling logic to reduce the amount of network calls
- Bumped dependencies versions
- Reduced default verbosity

## [0.2.8] - 2019-10-01

### Added

- Capability to automatically fetch projects from subgroups
- List projects without specifying an user or a group as owner, referring to what is discoverable by the token

### Changed

- Upgraded to go `1.13`

## [0.2.7] - 2019-09-12

### Added

- Graceful shutdowns
- Configurable health URLs for readiness checks
- Disabled readiness checks if SkipTLSVerification is set

### Changed

- Got more flexibility for the helm chart configuration

## [0.2.6] - 2019-09-09

### Added

- Missing pipelines statuses from the API spec
- Tests for config file parsing and some gitlab related functions

### Changed

- Fix nil pointer dereference on pollProjectRef function
- Refactored codebase with `cli`, `cmd` and `logger` packages
- Refactored the config and client structures, exported them
- Switched from yaml.v2 to yaml.v3

## [0.2.5] - 2019-08-27

### Added

- New `gitlab_ci_pipeline_last_run_id` metric
- Added `disable_tls_verify` config parameter for the GitLab client
- Added `-c` and `-l` aliases for `config` and `listen-adress` flags
- Backoff mechanism for pollings refs with no pipelines

### Changed

- Renamed `gitlab_ci_pipeline_status` metric into `gitlab_ci_pipeline_last_run_status`
- Initialize `gitlab_ci_pipeline_run_count` with a value of `0` when the exporter starts

## [0.2.4] - 2019-08-02

### Added

- Added an `example/` folder that allow people to get a fully working test environment in a few minutes using **docker-compose**.

### Changed

- Fixed an issue that prevented from loading all projects and branches/tags when using wildcard definitions #10

## [0.2.3] - 2019-08-01

### Added

- Released packages for `Mac OS X`, `Linux` & `Windows` and updated documentation
- Support for customisable environment variables on the chart

### Changed

- Replaced alpine/musl with a busybox/glibc based container image
- Fixed a bug introduced with the wildcard support preventing mux from starting correctly

### Removed

- Liveness check around goroutines

## [0.2.2] - 2019-07-30

### Added

- Added automatic refresh of available projects when using wildcards
- Added support for wildcard on refs (branches & tags) with automatic refresh of available ones

### Changed

- Replaced cli with `urfave/cli`
- Replaced log with `sirupsen/logrus`

## [0.2.1] - 2019-07-26

### Added

- Added `securityContext` configuration capability to the chart
- Added proper `liveness` and `readiness` checks
- Added support for dynamic discovery of the projects using a wildcard

### Changed

- Updated default `--listen-port` to **8080** so that you can run it without `root` user
- Fixed a bug causing a panic when no pipelines were created on a ref
- Bumped dependencies
- Updated Grafana dashboards

## [0.2.0] - 2019-05-27

### Added

- Automated releases of the binaries

### Changed

- Fixed linting errors
- Actually include the helm chart..
- Switched to go modules
- Upgraded to go 1.12
- Rewrote license in markdown
- Switched CI to drone
- Upgraded Docker release to alpine 3.9
- Bumped prometheus and gitlab SDK to their latest versions

## [0.1.0] - 2018-09-21

### Added

- Working state of the app
- Helm chart for K8S deployments
- New metric : `gitlab_ci_pipeline_last_run_duration_seconds`
- New metric : `gitlab_ci_pipeline_run_count`
- New metric : `gitlab_ci_pipeline_status`
- New metric : `gitlab_ci_pipeline_time_since_last_run_seconds`
- Makefile
- LICENSE
- README

[Unreleased]: https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/compare/0.4.0...HEAD
[0.4.0]: https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/tree/0.4.0
[0.3.5]: https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/tree/0.3.5
[0.3.4]: https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/tree/0.3.4
[0.3.3]: https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/tree/0.3.3
[0.3.2]: https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/tree/0.3.2
[0.3.1]: https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/tree/0.3.1
[0.3.0]: https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/tree/0.3.0
[0.2.14]: https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/tree/0.2.14
[0.2.13]: https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/tree/0.2.13
[0.2.12]: https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/tree/0.2.12
[0.2.11]: https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/tree/0.2.11
[0.2.10]: https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/tree/0.2.10
[0.2.9]: https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/tree/0.2.9
[0.2.8]: https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/tree/0.2.8
[0.2.7]: https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/tree/0.2.7
[0.2.6]: https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/tree/0.2.6
[0.2.5]: https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/tree/0.2.5
[0.2.4]: https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/tree/0.2.4
[0.2.3]: https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/tree/0.2.3
[0.2.2]: https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/tree/0.2.2
[0.2.1]: https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/tree/0.2.1
[0.2.0]: https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/tree/0.2.0
[0.1.0]: https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/tree/0.1.0
