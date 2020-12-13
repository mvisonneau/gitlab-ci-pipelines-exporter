# GitLab CI Pipelines Exporter - Metrics

## Metrics

| Metric name | Description | Labels | Configuration |
|---|---|---|---|
| `gitlab_ci_environment_behind_commits_count` | Number of commits the environment is behind given its last deployment | [project], [environment] | `project_defaults.pull.environments.enabled` |
| `gitlab_ci_environment_behind_duration_seconds` | Duration in seconds the environment is behind the most recent commit given its last deployment | [project], [environment] | `project_defaults.pull.environments.enabled` |
| `gitlab_ci_environment_deployment_count` |Number of deployments for an environment | [project], [environment] | `project_defaults.pull.environments.enabled` |
| `gitlab_ci_environment_deployment_duration_seconds` | Duration in seconds of the most recent deployment of the environment | [project], [environment] | `project_defaults.pull.environments.enabled` |
| `gitlab_ci_environment_deployment_job_id` | ID of the most recent deployment job for an environment | [project], [environment] | `project_defaults.pull.environments.enabled` |
| `gitlab_ci_environment_deployment_status` | Status of the most recent deployment of the environment | [project], [environment], [status] | `project_defaults.pull.environments.enabled` |
| `gitlab_ci_environment_deployment_timestamp` | Creation date of the most recent deployment of the environment | [project], [environment] | `project_defaults.pull.environments.enabled` |
| `gitlab_ci_environment_information` | Information about the environment | [project], [environment], [environment_id], [external_url], [kind], [ref], [latest_commit_short_id], [current_commit_short_id], [available], [author_email] | `project_defaults.pull.environments.enabled` |
| `gitlab_ci_pipeline_coverage` | Coverage of the most recent pipeline | [project], [topics], [ref], [kind], [variables] | *available by default* |
| `gitlab_ci_pipeline_duration_seconds` | Duration in seconds of the most recent pipeline | [project], [topics], [ref], [kind], [variables] | *available by default* |
| `gitlab_ci_pipeline_id` | ID of the most recent pipeline | [project], [topics], [ref], [kind], [variables] | *available by default* |
| `gitlab_ci_pipeline_job_artifact_size_bytes` | Artifact size in bytes (sum of all of them) of the most recent job | [project], [topics], [ref], [runner_description], [kind], [variables], [stage], [job_name] | `project_defaults.pull.pipeline.jobs.enabled` |
| `gitlab_ci_pipeline_job_duration_seconds` | Duration in seconds of the most recent job | [project], [topics], [ref], [runner_description], [kind], [variables], [stage], [job_name] | `project_defaults.pull.pipeline.jobs.enabled` |
| `gitlab_ci_pipeline_job_id` | ID of the most recent job | [project], [topics], [ref], [runner_description], [kind], [variables], [stage], [job_name] | `project_defaults.pull.pipeline.jobs.enabled` |
| `gitlab_ci_pipeline_job_run_count` | Number of executions of a job | [project], [topics], [ref], [runner_description], [kind], [variables], [stage], [job_name] | `project_defaults.pull.pipeline.jobs.enabled` |
| `gitlab_ci_pipeline_job_status` | Status of the most recent job | [project], [topics], [ref], [runner_description], [kind], [variables], [stage], [job_name], [status] | `project_defaults.pull.pipeline.jobs.enabled` |
| `gitlab_ci_pipeline_job_timestamp` | Creation date timestamp of the the most recent job | [project], [topics], [ref], [runner_description], [kind], [variables], [stage], [job_name] | `project_defaults.pull.pipeline.jobs.enabled` |
| `gitlab_ci_pipeline_status` | Status of the most recent pipeline | [project], [topics], [ref], [kind], [variables], [status] | *available by default* |
| `gitlab_ci_pipeline_timestamp` | Timestamp of the last update of the most recent pipeline | [project], [topics], [ref], [kind], [variables] | *available by default* |
| `gitlab_ci_pipeline_run_count` | Number of executions of a pipeline | [project], [topics], [ref], [kind], [variables] | *available by default* |

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

### Variables

User defined variables for the pipelines.
Those are not fetched by default, you need to set `project_defaults.pull.pipeline.variables.enabled` to **true**

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

### Author email

Email of whom created the most recent deployment on the environment

### Status

Status of the pipeline or deployment

### Stage

Stage of the job

### Job name

Name of the job

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

[author_email]: #author-email
[available]: #available
[current_commit_short_id]: #current-commit-short-id
[environment]: #environment
[environment_id]: #environment-id
[external_url]: #external-url
[job_name]: #job-name
[kind]: #ref-kind
[latest_commit_short_id]: #latest-commit-short-id
[project]: #project
[ref]: #ref-name
[stage]: #stage
[status]: #status
[topics]: #topics
[variables]: #variables
[runner_description]: #runner-description