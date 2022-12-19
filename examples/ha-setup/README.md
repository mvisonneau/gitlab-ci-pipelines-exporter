# Example usage of gitlab-ci-pipelines-exporter in an highly available fashion

This is a more advanced setup for users in the need of enhanced reliability for this exporter.

## Requirements

There are the same as for the [quickstart example](../quickstart/README.md):

- A personal access token on [gitlab.com](https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html) (or your own instance) with `read_repository` scope
- [git](https://git-scm.com/) & [docker-compose](https://docs.docker.com/compose/)

The [docker-compose.yml](./docker-compose.yml) is configured to spinup the necessary middleware which is a redis instance.

## 🚀

```bash
# Clone this repository
~$ git clone https://github.com/mvisonneau/gitlab-ci-pipelines-exporter.git
~$ cd gitlab-ci-pipelines-exporter/examples/ha-setup

# Provide your personal GitLab API access token (needs read_api permissions)
~$ sed -i 's/<your_token>/xXF_xxjV_xxyzxzz/' gitlab-ci-pipelines-exporter.yml

# Start redis and gitlab-ci-pipelines-exporter containers
~$ docker-compose up -d
Creating network "ha-setup_default" with driver "bridge"
Creating ha-setup_redis_1 ... done
Creating ha-setup_gitlab-ci-pipelines-exporter-1_1 ... done
Creating ha-setup_gitlab-ci-pipelines-exporter-3_1 ... done
Creating ha-setup_gitlab-ci-pipelines-exporter-2_1 ... done
```

## Use and troubleshoot

### Validate that containers are running

```bash
~$ docker ps
CONTAINER ID        IMAGE                                            COMMAND                  CREATED             STATUS              PORTS                    NAMES
e7f951e18af2        mvisonneau/gitlab-ci-pipelines-exporter:latest   "/usr/local/bin/gitl…"   5 seconds ago       Up 4 seconds        0.0.0.0:8082->8080/tcp   ha-setup_gitlab-ci-pipelines-exporter-2_1
a5c379c7203c        mvisonneau/gitlab-ci-pipelines-exporter:latest   "/usr/local/bin/gitl…"   5 seconds ago       Up 4 seconds        0.0.0.0:8081->8080/tcp   ha-setup_gitlab-ci-pipelines-exporter-1_1
ddcd8e257973        mvisonneau/gitlab-ci-pipelines-exporter:latest   "/usr/local/bin/gitl…"   5 seconds ago       Up 4 seconds        0.0.0.0:8083->8080/tcp   ha-setup_gitlab-ci-pipelines-exporter-3_1
6107feacefb7        bitnami/redis:6.0.8                              "/opt/bitnami/script…"   6 seconds ago       Up 5 seconds        0.0.0.0:6379->6379/tcp   ha-setup_redis_1
```

### Check logs from all containers

```bash
~$ docker-compose logs -f | grep -v redis
gitlab-ci-pipelines-exporter-3_1  | time="2020-10-05T19:57:05Z" level=debug msg="listing project pipelines" project-id=11915984 project-ref=master
gitlab-ci-pipelines-exporter-2_1  | time="2020-10-05T19:57:06Z" level=info msg="discovered new project ref" project-id=250833 project-path-with-namespace=gitlab-org/gitlab-runner project-ref=master project-ref-kind=branch

gitlab-ci-pipelines-exporter-3_1  | time="2020-10-05T19:57:22Z" level=info msg="scheduling projects refs most recent pipelines pulling" total=2
gitlab-ci-pipelines-exporter-1_1  | time="2020-10-05T19:57:22Z" level=debug msg="listing project pipelines" project-id=11915984 project-ref=master
gitlab-ci-pipelines-exporter-3_1  | time="2020-10-05T19:57:22Z" level=info msg="scheduling projects refs most recent jobs pulling" total=2
gitlab-ci-pipelines-exporter-2_1  | time="2020-10-05T19:57:22Z" level=info msg="scheduling projects refs most recent pipelines pulling" total=2
gitlab-ci-pipelines-exporter-2_1  | time="2020-10-05T19:57:22Z" level=info msg="scheduling projects refs most recent jobs pulling" total=2
gitlab-ci-pipelines-exporter-1_1  | time="2020-10-05T19:57:22Z" level=info msg="scheduling projects refs most recent pipelines pulling" total=2
gitlab-ci-pipelines-exporter-1_1  | time="2020-10-05T19:57:22Z" level=info msg="scheduling projects refs most recent jobs pulling" total=2
gitlab-ci-pipelines-exporter-2_1  | time="2020-10-05T19:57:22Z" level=debug msg="listing project pipelines" project-id=250833 project-ref=master
gitlab-ci-pipelines-exporter-1_1  | time="2020-10-05T19:57:22Z" level=debug msg="listing project pipelines" project-id=250833 project-ref=master
gitlab-ci-pipelines-exporter-2_1  | time="2020-10-05T19:57:22Z" level=debug msg="listing project pipelines" project-id=250833 project-ref=master
gitlab-ci-pipelines-exporter-1_1  | time="2020-10-05T19:57:23Z" level=debug msg="listing project pipelines" project-id=11915984 project-ref=master
gitlab-ci-pipelines-exporter-2_1  | time="2020-10-05T19:57:23Z" level=debug msg="listing project pipelines" project-id=11915984 project-ref=master
gitlab-ci-pipelines-exporter-2_1  | time="2020-10-05T19:57:24Z" level=debug msg="listing project pipelines" project-id=250833 project-ref=master
```

### Check we can fetch metrics from any exporter container

```bash
~$ curl -s http://localhost:8081/metrics | grep project
gitlab_ci_pipeline_coverage{kind="branch",project="gitlab-org/gitlab-runner",ref="master",topics="",variables=""} 75.6
gitlab_ci_pipeline_last_run_duration_seconds{kind="branch",project="gitlab-org/charts/auto-deploy-app",ref="master",topics="",variables=""} 198
gitlab_ci_pipeline_last_run_duration_seconds{kind="branch",project="gitlab-org/gitlab-runner",ref="master",topics="",variables=""} 2831
gitlab_ci_pipeline_last_run_id{kind="branch",project="gitlab-org/charts/auto-deploy-app",ref="master",topics="",variables=""} 1.64436288e+08
gitlab_ci_pipeline_last_run_id{kind="branch",project="gitlab-org/gitlab-runner",ref="master",topics="",variables=""} 1.9807024e+08
gitlab_ci_pipeline_last_run_status{kind="branch",project="gitlab-org/charts/auto-deploy-app",ref="master",status="canceled",topics="",variables=""} 0
gitlab_ci_pipeline_last_run_status{kind="branch",project="gitlab-org/charts/auto-deploy-app",ref="master",status="failed",topics="",variables=""} 0
gitlab_ci_pipeline_last_run_status{kind="branch",project="gitlab-org/charts/auto-deploy-app",ref="master",status="manual",topics="",variables=""} 0
gitlab_ci_pipeline_last_run_status{kind="branch",project="gitlab-org/charts/auto-deploy-app",ref="master",status="pending",topics="",variables=""} 0
gitlab_ci_pipeline_last_run_status{kind="branch",project="gitlab-org/charts/auto-deploy-app",ref="master",status="running",topics="",variables=""} 0
gitlab_ci_pipeline_last_run_status{kind="branch",project="gitlab-org/charts/auto-deploy-app",ref="master",status="skipped",topics="",variables=""} 0
gitlab_ci_pipeline_last_run_status{kind="branch",project="gitlab-org/charts/auto-deploy-app",ref="master",status="success",topics="",variables=""} 1
gitlab_ci_pipeline_last_run_status{kind="branch",project="gitlab-org/gitlab-runner",ref="master",status="canceled",topics="",variables=""} 0
gitlab_ci_pipeline_last_run_status{kind="branch",project="gitlab-org/gitlab-runner",ref="master",status="failed",topics="",variables=""} 1
gitlab_ci_pipeline_last_run_status{kind="branch",project="gitlab-org/gitlab-runner",ref="master",status="manual",topics="",variables=""} 0
gitlab_ci_pipeline_last_run_status{kind="branch",project="gitlab-org/gitlab-runner",ref="master",status="pending",topics="",variables=""} 0
gitlab_ci_pipeline_last_run_status{kind="branch",project="gitlab-org/gitlab-runner",ref="master",status="running",topics="",variables=""} 0
gitlab_ci_pipeline_last_run_status{kind="branch",project="gitlab-org/gitlab-runner",ref="master",status="skipped",topics="",variables=""} 0
gitlab_ci_pipeline_last_run_status{kind="branch",project="gitlab-org/gitlab-runner",ref="master",status="success",topics="",variables=""} 0
gitlab_ci_pipeline_time_since_last_run_seconds{kind="branch",project="gitlab-org/charts/auto-deploy-app",ref="master",topics="",variables=""} 7.706591e+06
gitlab_ci_pipeline_time_since_last_run_seconds{kind="branch",project="gitlab-org/gitlab-runner",ref="master",topics="",variables=""} 42310
```

You can validate that you get the same results from any container :

```bash
~$ curl -s http://localhost:8081/metrics | grep project
35
~$ curl -s http://localhost:8082/metrics | grep project
35
~$ curl -s http://localhost:8083/metrics | grep project
35
```

## Cleanup

```bash
~$ docker-compose down
```
