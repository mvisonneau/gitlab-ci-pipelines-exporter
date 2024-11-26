# GitLab CI Pipelines Exporter - Metrics

## Metrics

| Metric name | Description | Labels | Configuration |
|---|---|---|---|
| `gcpe_currently_queued_tasks_count` | Number of tasks in the queue || *available by default* |
| `gcpe_environments_count` | Number of GitLab environments being exported || *available by default* |
| `gcpe_executed_tasks_count` | Number of tasks executed || *available by default* |
| `gcpe_gitlab_api_requests_count` | GitLab API requests count || *available by default* |
| `gcpe_gitlab_api_requests_remaining` | GitLab API requests remaining in the API Limit || *available by default* |
| `gcpe_gitlab_api_requests_limit` | GitLab API requests available in the API Limit || *available by default* |
| `gcpe_metrics_count` | Number of GitLab pipelines metrics being exported || *available by default* |
| `gcpe_projects_count` | Number of GitLab projects being exported || *available by default* |
| `gcpe_refs_count` | Number of GitLab refs being exported || *available by default* |
| `gitlab_ci_environment_behind_commits_count` | Number of commits the environment is behind given its last deployment | [project], [environment] | `project_defaults.pull.environments.enabled` |
| `gitlab_ci_environment_behind_duration_seconds` | Duration in seconds the environment is behind the most recent commit given its last deployment | [project], [environment] | `project_defaults.pull.environments.enabled` |
| `gitlab_ci_environment_deployment_count` |Number of deployments for an environment | [project], [environment] | `project_defaults.pull.environments.enabled` |
| `gitlab_ci_environment_deployment_duration_seconds` | Duration in seconds of the most recent deployment of the environment | [project], [environment] | `project_defaults.pull.environments.enabled` |
| `gitlab_ci_environment_deployment_job_id` | ID of the most recent deployment job for an environment | [project], [environment] | `project_defaults.pull.environments.enabled` |
| `gitlab_ci_environment_deployment_status` | Status of the most recent deployment of the environment | [project], [environment], [status] | `project_defaults.pull.environments.enabled` |
| `gitlab_ci_environment_deployment_timestamp` | Creation date of the most recent deployment of the environment | [project], [environment] | `project_defaults.pull.environments.enabled` |
| `gitlab_ci_environment_information` | Information about the environment | [project], [environment], [environment_id], [external_url], [kind], [ref], [latest_commit_short_id], [current_commit_short_id], [available], [username] | `project_defaults.pull.environments.enabled` |
| `gitlab_ci_pipeline_coverage` | Coverage of the most recent pipeline | [project], [topics], [ref], [kind], [source], [variables] | *available by default* |
| `gitlab_ci_pipeline_duration_seconds` | Duration in seconds of the most recent pipeline | [project], [topics], [ref], [kind], [source], [variables] | *available by default* |
| `gitlab_ci_pipeline_id` | ID of the most recent pipeline | [project], [topics], [ref], [kind], [source], [variables] | *available by default* |
| `gitlab_ci_pipeline_job_artifact_size_bytes` | Artifact size in bytes (sum of all of them) of the most recent job | [project], [topics], [ref], [runner_description], [kind], [source], [variables], [stage], [job_name], [tag_list], [failure_reason] | `project_defaults.pull.pipeline.jobs.enabled` |
| `gitlab_ci_pipeline_job_duration_seconds` | Duration in seconds of the most recent job | [project], [topics], [ref], [runner_description], [kind], [source], [variables], [stage], [job_name], [tag_list], [failure_reason] | `project_defaults.pull.pipeline.jobs.enabled` |
| `gitlab_ci_pipeline_job_id` | ID of the most recent job | [project], [topics], [ref], [runner_description], [kind], [source], [variables], [stage], [job_name], [tag_list], [failure_reason] | `project_defaults.pull.pipeline.jobs.enabled` |
| `gitlab_ci_pipeline_job_queued_duration_seconds` | Duration in seconds the most recent job has been queued before starting | [project], [topics], [ref], [runner_description], [kind], [source], [variables], [stage], [job_name], [tag_list], [failure_reason] | `project_defaults.pull.pipeline.jobs.enabled` |
| `gitlab_ci_pipeline_job_run_count` | Number of executions of a job | [project], [topics], [ref], [runner_description], [kind], [source], [variables], [stage], [job_name], [tag_list], [failure_reason] | `project_defaults.pull.pipeline.jobs.enabled` |
| `gitlab_ci_pipeline_job_status` | Status of the most recent job | [project], [topics], [ref], [runner_description], [kind], [source], [variables], [stage], [job_name], [tag_list], [status], [failure_reason] | `project_defaults.pull.pipeline.jobs.enabled` |
| `gitlab_ci_pipeline_job_timestamp` | Creation date timestamp of the the most recent job | [project], [topics], [ref], [runner_description], [kind], [source], [variables], [stage], [job_name], [tag_list], [failure_reason] | `project_defaults.pull.pipeline.jobs.enabled` |
| `gitlab_ci_pipeline_queued_duration_seconds` | Duration in seconds the most recent pipeline has been queued before starting | [project], [topics], [ref], [kind], [source], [variables] | *available by default* |
| `gitlab_ci_pipeline_run_count` | Number of executions of a pipeline | [project], [topics], [ref], [kind], [source], [variables] | *available by default* |
| `gitlab_ci_pipeline_status` | Status of the most recent pipeline | [project], [topics], [ref], [kind], [source], [variables], [status] | *available by default* |
| `gitlab_ci_pipeline_timestamp` | Timestamp of the last update of the most recent pipeline | [project], [topics], [ref], [kind], [source], [variables] | *available by default* |
| `gitlab_ci_pipeline_test_report_total_time` | Duration in seconds of all the tests in the most recently finished pipeline | [project], [topics], [ref], [kind], [source], [variables] | `project_defaults.pull.pipeline.test_reports.enabled` |
| `gitlab_ci_pipeline_test_report_total_count` | Number of total tests in the most recently finished pipeline | [project], [topics], [ref], [kind], [source], [variables] | `project_defaults.pull.pipeline.test_reports.enabled` |
| `gitlab_ci_pipeline_test_report_success_count` | Number of successful tests in the most recently finished pipeline | [project], [topics], [ref], [kind], [source], [variables] | `project_defaults.pull.pipeline.test_reports.enabled` |
| `gitlab_ci_pipeline_test_report_failed_count` | Number of failed tests in the most recently finished pipeline | [project], [topics], [ref], [kind], [source], [variables] | `project_defaults.pull.pipeline.test_reports.enabled` |
| `gitlab_ci_pipeline_test_report_skipped_count` | Number of skipped tests in the most recently finished pipeline | [project], [topics], [ref], [kind], [source], [variables] | `project_defaults.pull.pipeline.test_reports.enabled` |
| `gitlab_ci_pipeline_test_report_error_count` | Number of errored tests in the most recently finished pipeline | [project], [topics], [ref], [kind], [source], [variables] | `project_defaults.pull.pipeline.test_reports.enabled` |
| `gitlab_ci_pipeline_test_suite_total_time` | Duration in seconds for the test suite | [project], [topics], [ref], [kind], [source], [variables], [test_suite_name] | `project_defaults.pull.pipeline.test_reports.enabled` |
| `gitlab_ci_pipeline_test_suite_total_count` | Number of total tests for the test suite | [project], [topics], [ref], [kind], [source], [variables], [test_suite_name] | `project_defaults.pull.pipeline.test_reports.enabled` |
| `gitlab_ci_pipeline_test_suite_success_count` | Number of successful tests for the test suite | [project], [topics], [ref], [kind], [source], [variables], [test_suite_name] | `project_defaults.pull.pipeline.test_reports.enabled` |
| `gitlab_ci_pipeline_test_suite_failed_count` | Number of failed tests for the test suite | [project], [topics], [ref], [kind], [source], [variables], [test_suite_name] | `project_defaults.pull.pipeline.test_reports.enabled` |
| `gitlab_ci_pipeline_test_suite_skipped_count` | Number of skipped tests for the test suite | [project], [topics], [ref], [kind], [source], [variables], [test_suite_name] | `project_defaults.pull.pipeline.test_reports.enabled` |
| `gitlab_ci_pipeline_test_suite_error_count` | Duration in errored tests for the test suite | [project], [topics], [ref], [kind], [source], [variables], [test_suite_name] | `project_defaults.pull.pipeline.test_reports.enabled` |
| `gitlab_ci_pipeline_test_case_execution_time` | Duration in seconds for the test case | [project], [topics], [ref], [kind], [source], [variables], [test_suite_name], [test_case_name], [test_case_classname] | `project_defaults.pull.pipeline.test_reports.test_cases.enabled` |
| `gitlab_ci_pipeline_test_case_status` | Status of the most recent test case | [project], [topics], [ref], [kind], [source], [variables], [test_suite_name], [test_case_name], [test_case_classname], [status] | `project_defaults.pull.pipeline.test_reports.test_cases.enabled` |

## Labels

### Project

Path with namespace of the project

### Topics

Topics configured on the project

### Ref Name

Name of the ref (branch, tag or merge request) used by the pipeline

### Runner Description

Description of the runner on which the most recent job ran

### Ref Kind

Type of the ref used by the pipeline. Can be either **branch**, **tag** or **merge_request**

### Source

The reason the pipeline exists.

### Variables

User defined variables for the pipelines.
Those are not fetched by default, you need to set `project_defaults.pull.pipeline.variables.enabled` to **true**

### Test Suite Name

Name of the test suite.
This is not fetched by default, you need to set `project_default.pull.pipeline.test_reports.enabled` to **true**

### Test Case Name

Name of the test case.
This is not fetched by default, you need to set `project_default.pull.pipeline.test_reports.test_cases.enabled` to **true**

### Test Case ClassName

Name of the test case classname.
This is not fetched by default, you need to set `project_default.pull.pipeline.test_reports.test_cases.enabled` to **true**

### Environment

Name of the environment

### Available

Whether the environment is available or not

### External URL

External URL of the environment

### Latest commit short ID

Most recent commit short ID on the ref which was last used to deploy to the environment

### Current commit short ID

Currently deployed commit short ID on the environment

### Username

GitLab username of the person which triggered the most recent deployment of the environment

### Status

Status of the pipeline, deployment or test case

### Stage

Stage of the job

### Job name

Name of the job

### Tag list

Tag list of the job

### Environment ID

ID of the environment

### Sparse status metrics

If the amount of status metrics generated by fetching jobs becomes a problem, you can enable `output_sparse_status_metrics` on a global, per-project or per-wildcard basis. When enabled, only labels matching the previous pipeline or job status will be submitted (with value `1`) rather than all label combinations submitted but with `0` value where the status does not match the previous run, for example:

```bash
# output_sparse_status_metrics: false
gitlab_ci_pipeline_last_job_run_status{job="test",project="bar/project",ref="main",stage="test",status="canceled",topics=""} 0
gitlab_ci_pipeline_last_job_run_status{job="test",project="bar/project",ref="main",stage="test",status="failed",topics=""} 0
gitlab_ci_pipeline_last_job_run_status{job="test",project="bar/project",ref="main",stage="test",status="manual",topics=""} 0
gitlab_ci_pipeline_last_job_run_status{job="test",project="bar/project",ref="main",stage="test",status="pending",topics=""} 0
gitlab_ci_pipeline_last_job_run_status{job="test",project="bar/project",ref="main",stage="test",status="running",topics=""} 0
gitlab_ci_pipeline_last_job_run_status{job="test",project="bar/project",ref="main",stage="test",status="skipped",topics=""} 0
gitlab_ci_pipeline_last_job_run_status{job="test",project="bar/project",ref="main",stage="test",status="success",topics=""} 1

# output_sparse_status_metrics: true
gitlab_ci_pipeline_last_job_run_status{job="test",project="bar/project",ref="main",stage="test",status="success",topics=""} 1
```

This flag affect every `_status$` metrics:

- `gitlab_ci_pipeline_environment_deployment_status`
- `gitlab_ci_pipeline_job_status`
- `gitlab_ci_pipeline_status`
- `gitlab_ci_pipeline_test_case_status`

[available]: #available
[current_commit_short_id]: #current-commit-short-id
[environment]: #environment
[environment_id]: #environment-id
[external_url]: #external-url
[job_name]: #job-name
[tag_list]: #tag-list
[kind]: #ref-kind
[latest_commit_short_id]: #latest-commit-short-id
[project]: #project
[ref]: #ref-name
[runner_description]: #runner-description
[stage]: #stage
[status]: #status
[topics]: #topics
[username]: #username
[source]: #source
[variables]: #variables
[test_suite_name]: #test-suite-name
[test_case_name]: #test-case-name
[test_case_classname]: #test-case-classname
