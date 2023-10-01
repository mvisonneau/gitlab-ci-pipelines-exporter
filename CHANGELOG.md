# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [0ver](https://0ver.org) (more or less).

## [Unreleased]

## [v0.5.5] - 2023-05-22

### Added

- new metrics:
  - `gitlab_ci_pipeline_test_report_total_time` -> Duration in seconds of all the tests in the most recently finished pipeline
  - `gitlab_ci_pipeline_test_report_total_count` -> Number of total tests in the most recently finished pipeline
  - `gitlab_ci_pipeline_test_report_success_count` -> Number of successful tests in the most recently finished pipeline
  - `gitlab_ci_pipeline_test_report_failed_count` -> Number of failed tests in the most recently finished pipeline
  - `gitlab_ci_pipeline_test_report_skipped_count` -> Number of skipped tests in the most recently finished pipeline
  - `gitlab_ci_pipeline_test_report_error_count` -> Number of errored tests in the most recently finished pipeline
  - `gitlab_ci_pipeline_test_suite_total_time` -> Duration in seconds for the test suite
  - `gitlab_ci_pipeline_test_suite_total_count` -> Number of total tests for the test suite
  - `gitlab_ci_pipeline_test_suite_success_count` -> Number of successful tests for the test suite
  - `gitlab_ci_pipeline_test_suite_failed_count` -> Number of failed tests for the test suite
  - `gitlab_ci_pipeline_test_suite_skipped_count` -> Number of skipped tests for the test suite
  - `gitlab_ci_pipeline_test_suite_error_count` -> Duration in errored tests for the test suite
- new configuration parameter: `gitlab.burstable_requests_per_second`, introducing a burstable amount of API RPS
- new configuration parameter: `gitlab.maximum_jobs_queue_size`, controlling the queue buffer size
- new label for pipelines and jobs: `source` to indicate the reason the pipeline started

### Changed

- Upgraded golang to **v1.20**
- Upgraded most dependencies to their latest versions
- Reduced the amount of data being pulled from the project list API calls

## [v0.5.4] - 2022-08-25

### Added

- Kickstarted tracing support through `opentelemetry` implementation for most of the network calls
- Now passing a `context.Context` to most functional calls
- Aggregated already used linters and added new ones through the implementation of `golangci`
- Release `.apk` packages for Alpine linux
- Added man pages and autocompletion scripts (bash & zsh) to `.apk`, `.deb`, `.rpm` & `homebrew` packages
- Release "fat" binaries (arm64 + amd64 combined) for MacOS under `_all` suffix

### Changed

- Fixed a config issue preventing the arm deb/rpm packages to be released correctly
- Upgraded golang to **v1.19**
- Upgraded most dependencies to their lastest versions
- Fixed child pipeline jobs not found whilst looking up through bridges (#345)
- `gitlab_ci_pipeline_job_queued_duration_seconds` & `gitlab_ci_pipeline_queued_duration_seconds` will now be leveraging the value returned through the GitLab API instead of computing it with (startedAt - createdAt)
- Refactored the RPC layer used for CLI monitoring with gRPC

## [v0.5.3] - 2022-02-11

### Added

- `linux/arm/v6` & `linux/arm/v7` binary & container image releases
- `quay.io` container image releases
- New internal metrics about exporter's health:
  - `gcpe_gitlab_api_requests_remaining` -  GitLab API requests remaining in the API Limit
  - `gcpe_gitlab_api_requests_limit` - GitLab API requests available in the API Limit

### Changed

- Fixed an issue when running in cluster mode where tasks could hang when the exporter restarted.
- Fixed a bug in some cases where pagination headers are not returned from GitLab's API
- Upgraded most dependencies to their latest versions
- fixed json syntax error in webhook error body
- dashboards: fixed owner multiselect filters
- fixed a bug on `gitlab_ci_pipeline_run_count` being incremented multiple times

## [v0.5.2] - 2021-08-25
### Changed

- Updated default metrics garbage collection intervals from 5 to 10m
- monitor: correctly sanitize the config output
- fixed an issue introduced in v0.5.1 causing the exporter to fail when the monitoring listener address was not defined
- fixed a bug which could cause nil pointer deferences on readiness checks
- Updated golang to `1.17`

## [v0.5.1] - 2021-07-19

### Added

- New monitoring CLI to retrieve information about the exporter
- New internal metrics about exporter's health:
  - `gcpe_currently_queued_tasks_count` - Number of tasks in the queue
  - `gcpe_environments_count` - Number of GitLab environments being exported
  - `gcpe_executed_tasks_count` - Number of tasks executed
  - `gcpe_gitlab_api_requests_count` - GitLab API requests count
  - `gcpe_metrics_count` - Number of GitLab pipelines metrics being exported
  - `gcpe_projects_count` - Number of GitLab projects being exported
  - `gcpe_refs_count` - Number of GitLab refs being exported

### Changed

- fixed a goroutines/memory leak introduced in `0.5.0` which can occur when attempting to process more tasks than the ratelimit permits
- fixed a bug causing the refreshing of tag based jobs to not match any API results, generating lots of unnecessary requests
- webhooks: added more details when processing wildcards
- examples/quickstart: updated prometheus and grafana versions
- updated most libraries to their most recent versions

## [v0.5.0] - 2021-06-02

**BREAKING CHANGES**

- The configuration syntax has evolved, you can refer to the [updated version of the doc](docs/configuration_syntax.md)
  - `pull.maximum_gitlab_api_requests_per_second` has moved to `gitlab.maximum_requests_per_second`
  - `project_defaults.pull.refs.*` has been redone in a hopefully more comprehensible fashion
    - `project_defaults.pull.refs.(branches|tags|merge_requests).*` parameters can now be used to have a finer granularity
      on the management of per-ref-kind settings
    - `project_defaults.pull.refs.from.(pipelines|merge_requests)` is not natively managed as part of the per-ref-kind specific parameters
  - `project_defaults.pull.environments.name_regexp` has moved to `project_defaults.pull.environments.regexp`
  - `project_defaults.pull.environments.tags_regexp` was removed to avoid confusion

- Logging configuration now has to be done as part of the config file instead of CLI flags: 

```yaml
log:
  level: info
  format: text
```

- By default, when exporting metrics for `environments`, stopped ones will not be considered any more.

### Added

- New metric `gitlab_ci_pipeline_queued_duration_seconds`
- New metric `gitlab_ci_pipeline_job_queued_duration_seconds`
- (en|dis)able pulling `branches` / `tags` / `merge_requests` on a global or per-project basis
- Choose to export metrics for only the most 'n' recently updated refs pipelines
- Choose to export metrics for refs with pipelines 'updated in the last x seconds'
- Choose to export metrics for deleted `branches` or `tags`
- Choose to export metrics for available `environments` only

### Changed

- Bumped all dependencies
- Enhanced the function signatures for the ref objects management 
- Fixed a bug causing the jobs metrics pull to fail on ref.Kind=RefKindMergeRequest
- Fixed a bug causing the environments to be garbage collected on every run
- Fixed the error handling when comparing 2 refs which resulted into nil pointer dereferences
- Fixed the pulling of merge-request based pipelines
- Fixed unit tests on windows
- Replaced custom config mangement implementation with `creasty/defaults` and `go-playground/validator`
- Rewrote the non-OOP approach used so far for the controller

## [v0.4.9] - 2021-05-05

### Changed

- Fixed a bug preventing to set `output_sparse_status_metrics` globally or at the wildcard level
- Updated all dependencies to their latest versions
- Reduced the default GitLab API request rate limit from 10 to 1rps

## [v0.4.8] - 2021-03-08

### Added

- Snapcraft releases
- darwin/arm64 releases
- `username` label from the **gitlab_ci_environment_information** flag

### Removed

- `author_email` label from the **gitlab_ci_environment_information** flag (replaced by `username`)

### Changed

- Scoped down the projects fetched from wildcard onto the one starting with the owner's name to make it clearer for endusers
- Upgraded to go 1.16
- Bumped all dependencies to their latest versions

## [v0.4.7] - 2021-01-28

### Added

- GPG sign released artifacts checksums
- Support for performing requests through a forward proxy using standard env variables
- New parameters to enable/disable/aggregate the export of the runner description which executed the job

### Changed

- Fixed a bug on child/downstream pipelines pull when the trigger has not been fired yet
- Made the default config local file location not hidden as it makes very little sense
- Updated examples/webhook with ngrok+docker-compose instead of hashicorp waypoint
- Bumped goreleaser to 0.155.0 and leverage docker buildx
- Enhanced logging of the settings at startup to get more visibility on the interpreted parameters
- Fixed a bug preventing the webhooks from working properly when pull is completely disabled in conjunction of wildcards
- Updated all dependencies

## [v0.4.6] - 2020-12-16

### Added

- When configured to export job metrics, it will now pull child/downstream pipelines jobs related ones as well
- New `runner_description` label for all the `job` related metrics.
- Release GitHub container registry based images: [ghcr.io/mvisonneau/gitlab-ci-pipelines-exporter](https://github.com/users/mvisonneau/packages/container/package/gitlab-ci-pipelines-exporter)
- Release `arm64v8` based container images as part of docker manifests in both **docker.io** and **ghcr.io**

### Changed

- Ensure consistency of the exported metrics by distinguishing immutable from mutable labels used as metric key
- Updated all dependencies
- Migrated CI from Drone to GitHub actions

## [v0.4.5] - 2020-11-27

### Added

- Implemented a `max_age_seconds` parameter to determine whether to pull a "stale ref" or not

### Changed

- When fetching refs from pipelines, capped the maximum length to 100 in order to prevent the API call from failing
- Garbage collect merge requests refs in order to keep the depth value as a maximum
- Prevent potential nil pointers when exporting environments information
- Updated all dependencies

## [v0.4.4] - 2020-11-20

### Added

- Automatically refresh _pkg.go.dev_ on new releases

### Changed

- Do not delete environments metrics when ref is a tag which may not be configured to be monitored for pipelines/jobs (#182)
- Fixed a bug making `latest_commit_short_id` labels reflect incorrect information on environment metrics when tags are used as refs
- Enhanced logging for jobs pulling function
- Bumped goreleaser to 0.147.1
- Bumped all dependencies

## [v0.4.3] - 2020-11-04

### Added

- Export `environments/deployments` related metrics
- New `environments/deployments` and `jobs` grafana dashboards
- Documented the list of exported metrics
- Released **.deb** and **.rpm** packages
- More complete garbage collector capabilities
- Newly supported statuses for pipelines and jobs: `created`, `waiting_for_resource`, `preparing`, `scheduled`
- GitLab links for pipelines, jobs, environments and deployments in the dashboards

### Changed

- Prefix new releases with `^v` to make pkg.go.dev happy
- Bumped all dependencies
- Fixed race conditions during tests
- Always return coverage metric
- Enhanced the scheduling of the pull functions on init
- Improved webhook parsing functions performance
- Fixed a bug preventing the `gitlab_ci_pipeline_run_count` from being initialized correctly at startup
- Fixed the `gitlab_ci_pipeline_job_run_count` and `gitlab_ci_pipeline_run_count` metrics incrementing algorithm
- Improved the `pipelines` grafana dashboard
- Fixed a bug which could lead to an overwrite of the refs and environments at scale, inducing unecessary GitLab API calls and discrepancy for some metrics
- Optimized the storage layer implementation
- Ensure group wildcards only returns projects belonging directly to the group

## [0.4.2] - 2020-10-22

**BREAKING CHANGES**

- Moved helm chart definition to https://github.com/mvisonneau/helm-charts/tree/main/charts/gitlab-ci-pipelines-exporter

### Added

- More unit tests!

### Changed

- Fixed a bug preventing the webhooks implementation to correctly update the pertinent metrics (also creating pseudo duplicates)
- Fixed some missing columns and not ideal default sorting on the example grafana dashboard
- pkg/storage/local: added rw mutexes to prevent some read/write race condition issues from happening
- Bumped go-redis/redis/v8 to `v8.3.2`
- Bumped goreleaser to `v0.145.0`
- Bumped prometheus/client_golang to `v1.8.0`
- Bumped xanzy/go-gitlab to `v0.38.2`
- pkg/storage/local: added per variables mutexes

## [0.4.1] - 2020-10-14

### Added

- **Garbage collector for projects, refs and metrics**

By default, on regular basis, the exporter will now attempt to remove unconfigured/wanted projects, refs and associated metrics

### Changed

- Disabled taskq consumers system resources checks which may leave the exporter in a hanging state
- Fixed a bug preventing the redis URL to be read from the config
- Project discovery from wildcards will now ignore projects with disabled jobs/pipelines feature
- Embeded taskq logs into logrus
- Prevent taskq consumer from being paused on errors
- Changed the way we handle project ref pipelines not being found, log in debug factory instead of erroring in a crashloop
- Bumped goreleaser to `v0.144.1`
- Improved logging
- Increased amount of items fetched per API call to 100 (maximum value)

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

[Unreleased]: https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/compare/v0.5.4...HEAD
[v0.5.4]: https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/tree/v0.5.4
[v0.5.3]: https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/tree/v0.5.3
[v0.5.2]: https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/tree/v0.5.2
[v0.5.1]: https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/tree/v0.5.1
[v0.5.0]: https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/tree/v0.5.0
[v0.4.9]: https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/tree/v0.4.9
[v0.4.8]: https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/tree/v0.4.8
[v0.4.7]: https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/tree/v0.4.7
[v0.4.6]: https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/tree/v0.4.6
[v0.4.5]: https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/tree/v0.4.5
[v0.4.4]: https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/tree/v0.4.4
[v0.4.3]: https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/tree/v0.4.3
[0.4.2]: https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/tree/0.4.2
[0.4.1]: https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/tree/0.4.1
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
