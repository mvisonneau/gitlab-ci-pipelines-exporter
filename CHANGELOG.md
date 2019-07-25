# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [0ver](https://0ver.org).

## [Unreleased]

### Added

- Added `securityContext` configuration capability to the chart
- Added proper `liveness` and `readiness` checks

### Changed

- Updated default `--listen-port` to **8080**
- Bumped dependencies

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

[Unreleased]: https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/compare/0.2.0...HEAD
[0.2.0]: https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/tree/0.2.0
[0.1.0]: https://github.com/mvisonneau/gitlab-ci-pipelines-exporter/tree/0.1.0
