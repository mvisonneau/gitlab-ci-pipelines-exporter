# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [0ver](https://0ver.org).

## [Unreleased]

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
- Added `skip_tls_verify` config parameter for the GitLab client
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

[Unreleased]: https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/compare/0.2.9...HEAD
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
