# Example usage of gitlab-ci-pipelines-exporter leveraging pipelines and jobs webhooks

This is a more advanced setup for users looking to reduce the amount of requests being made onto your GitLab API endpoint.

## Requirements

- A personal access token on [gitlab.com](https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html) (or your own instance) with `read_repository` scope
- [git](https://git-scm.com/) & [docker-compose](https://docs.docker.com/compose/)
- GitLab [webhook configuration](https://docs.gitlab.com/ee/user/project/integrations/webhooks.html) privileges/capabilities on group(s) or project(s) you want to monitor.
- For this use case, we will need network connectivity from the GitLab rails processes towards the exporter's HTTP endpoint. If you are attempting this from your laptop, you will need something like [ngrok](https://ngrok.com/) or equivalent to be able to do so.

/!\ This implementation is for test/example only, I would not recommended to leverage ngrok endpoints for production purposes.

## 🚀

### Configure and start the exporter

```bash
# Clone this repository
~$ git clone https://github.com/mvisonneau/gitlab-ci-pipelines-exporter.git
~$ cd gitlab-ci-pipelines-exporter/examples/webhooks

# Provide your personal GitLab API access token (needs read_api permissions)
~$ sed -i 's/<your_token>/xXF_xxjV_xxyzxzz' gitlab-ci-pipelines-exporter.yml

# Configure a secret token for your webhooks authz
~$ export SECRET_TOKEN=$(openssl rand -base64 32)
~$ sed -i "s/<strong_arbitrary_secret_token>/${SECRET_TOKEN}" gitlab-ci-pipelines-exporter.yml

# Configure a project on which you are authorized to configure webhooks
~$ sed -i 's;<your_project_path_with_namespace>;my_group/my_project' gitlab-ci-pipelines-exporter.yml

# Start the exporter!
~$ docker-compose up -d

# Start ngrok
~$ ngrok http 8080
ngrok by @inconshreveable                                                                                                                                                                                                                                                                                                                                (Ctrl+C to quit)

Session Status                online
Version                       2.3.35
Region                        Europe (eu)
Web Interface                 http://127.0.0.1:4040
Forwarding                    http://0ba537eaa697.eu.ngrok.io -> http://localhost:8080
Forwarding                    https://0ba537eaa697.eu.ngrok.io -> http://localhost:8080

Connections                   ttl     opn     rt1     rt5     p50     p90
                              0       0       0.00    0.00    0.00    0.00
```

## Attempt to reach your exporter http endpoint from the public address

After a few seconds, you should be able to query the URL you got from `waypoint up`

```bash
~$ curl -i https://0ba537eaa697.eu.ngrok.io/health/ready
HTTP/1.1 200 OK
Date: Thu, 15 Oct 2020 22:18:27 GMT
Content-Type: application/json; charset=utf-8
Content-Length: 3
Connection: keep-alive

{}
```

gitlab.com should also be able to reach https://0ba537eaa697.eu.ngrok.io/webhook now! 🎉

### Configure GitLab group(s) or project(s)

I will showcase how to do it from the web UI but this can also be achieved using GitLab's API

```bash
# Retrieve your secret token configured for the exporter
~$ echo $SECRET_TOKEN
UYqDp5DvHLrtCnkfHA8aBPEkyKfgHjTGAWZRUD4olZU=
```

Go onto the project's configuration page and configure a new webhook using:

- **URL**: `https://0ba537eaa697.eu.ngrok.io/webhook`
- **Secret Token**: `UYqDp5DvHLrtCnkfHA8aBPEkyKfgHjTGAWZRUD4olZU=`
- Untick `Push events` and tick `Pipeline events`
- If you are running on GitLab >= 13.5 and want to export environments/deployments metrics, you can also tick `Deployment events`
- Hit the `Add webhook` button

![webhook_configuration](../../docs/images/webhook_configuration.png)

You can then trigger a manual test:

![webhook_trigger_test](../../docs/images/webhook_trigger_test.png)

If the last pipeline which ran on your project is on a ref that is configured to be exported, you will see the following logs:

```bash
~$ docker-compose logs -f
[..]
time="2021-01-28T09:51:28Z" level=debug msg="webhook request" ip-address="[::1]:56118" user-agent=GitLab/13.9.0-pre
time="2021-01-28T09:51:28Z" level=info msg="received a pipeline webhook from GitLab for a ref, triggering metrics pull" project-name=foo/bar ref=main ref-kind=branch
[..]
```

If you query the `/metrics` endpoint of the exporter you should be able to see associated metrics:

```shell
gitlab_ci_pipeline_coverage{kind="branch",project="foo/bar",ref="main",topics="",variables=""} 0
gitlab_ci_pipeline_duration_seconds{kind="branch",project="foo/bar",ref="main",topics="",variables=""} 494
gitlab_ci_pipeline_queued_duration_seconds{kind="branch",project="foo/bar",ref="main",topics="",variables=""} 60
gitlab_ci_pipeline_id{kind="branch",project="foo/bar",ref="main",topics="",variables=""} 1.00308162e+08
gitlab_ci_pipeline_run_count{kind="branch",project="foo/bar",ref="main",topics="",variables=""} 0
gitlab_ci_pipeline_status{kind="branch",project="foo/bar",ref="main",status="canceled",topics="",variables=""} 0
gitlab_ci_pipeline_status{kind="branch",project="foo/bar",ref="main",status="created",topics="",variables=""} 0
gitlab_ci_pipeline_status{kind="branch",project="foo/bar",ref="main",status="failed",topics="",variables=""} 0
gitlab_ci_pipeline_status{kind="branch",project="foo/bar",ref="main",status="manual",topics="",variables=""} 0
gitlab_ci_pipeline_status{kind="branch",project="foo/bar",ref="main",status="pending",topics="",variables=""} 0
gitlab_ci_pipeline_status{kind="branch",project="foo/bar",ref="main",status="preparing",topics="",variables=""} 0
gitlab_ci_pipeline_status{kind="branch",project="foo/bar",ref="main",status="running",topics="",variables=""} 0
gitlab_ci_pipeline_status{kind="branch",project="foo/bar",ref="main",status="scheduled",topics="",variables=""} 0
gitlab_ci_pipeline_status{kind="branch",project="foo/bar",ref="main",status="skipped",topics="",variables=""} 0
gitlab_ci_pipeline_status{kind="branch",project="foo/bar",ref="main",status="success",topics="",variables=""} 1
gitlab_ci_pipeline_status{kind="branch",project="foo/bar",ref="main",status="waiting_for_resource",topics="",variables=""} 0
gitlab_ci_pipeline_timestamp{kind="branch",project="foo/bar",ref="main",topics="",variables=""} 1.611826041e+09
```

In this configuration, the exporter does not even attempt to fetch the metrics on init, this means that on start, you won't be getting any metrics populated/exported until
a webhook has been triggered). This can be mitigated by whether using it in a ["ha-fashion" with a Redis backend](../ha-setup) or by enabling the following in the exporter config:

```yaml
pull:
  projects_from_wildcards:
    on_init: true

  environments_from_projects:
    on_init: true

  refs_from_projects:
    on_init: true

  metrics:
    on_init: true
```

You can also use it in an hybrid fashion pull/push with greater pull intervals if you want to ensure you have a consistent/convergent state over time.

## Cleanup

```bash
~$ docker-compose down
~$ <stop ngrok>
```

Also remove the configured test webhook on your GitLab project.
