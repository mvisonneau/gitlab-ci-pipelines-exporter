# Example usage of gitlab-ci-pipelines-exporter with Prometheus & Grafana

## Requirements

- **~5 min of your time**
- A personal access token on [gitlab.com](https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html) (or your own instance) with `read_repository` scope
- [git](https://git-scm.com/) & [docker-compose](https://docs.docker.com/compose/)

## 🚀

```bash
# Clone this repository
~$ git clone https://github.com/mvisonneau/gitlab-ci-pipelines-exporter.git
~$ cd gitlab-ci-pipelines-exporter/examples/quickstart

# Provide your personal GitLab API access token (needs read_api permissions)
~$ sed -i 's/<your_token>/xXF_xxjV_xxyzxzz' gitlab-ci-pipelines-exporter/config.yml

# Start gitlab-ci-pipelines-exporter, prometheus and grafana containers !
~$ docker-compose up -d
Creating network "quickstart_default" with driver "bridge"
Creating quickstart_gitlab-ci-pipelines-exporter_1 ... done
Creating quickstart_prometheus_1                   ... done
Creating quickstart_grafana_1                      ... done
```

You should now have a stack completely configured and accessible at these locations:

- `gitlab-ci-pipelines-exporter`: [http://localhost:8080/metrics](http://localhost:8080/metrics)
- `prometheus`: [http://localhost:9090](http://localhost:9090)
- `grafana`: [http://localhost:3000/d/gitlab_ci_pipeline_last_run_statuses](localhost:3000/d/gitlab_ci_pipeline_last_run_statuses)

## Use and troubleshoot

### Validate that containers are running

```bash
~$ docker ps
CONTAINER ID        IMAGE                                            COMMAND                  CREATED             STATUS              PORTS                    NAMES
c9aedfdefe41        grafana/grafana:latest                          "/run.sh"                6 seconds ago       Up 4 seconds        0.0.0.0:3000->3000/tcp   quickstart_grafana_1
b3500bff6038        prom/prometheus:latest                          "/bin/prometheus --c…"   7 seconds ago       Up 5 seconds        0.0.0.0:9090->9090/tcp   quickstart_prometheus_1
930b76005b13        mvisonneau/gitlab-ci-pipelines-exporter:latest  "/usr/local/bin/gitl…"   8 seconds ago       Up 6 seconds        0.0.0.0:8080->8080/tcp   quickstart_gitlab-ci-pipelines-exporter_1
```

### Check logs from the gitlab-ci-pipelines-exporter container

```bash
~$ docker logs -f quickstart_gitlab-ci-pipelines-exporter_1
time="2020-04-28T23:09:01Z" level=info msg="starting exporter" discover-projects-refs-interval=300s discover-wildcard-projects-interval=1800s gitlab-endpoint="https://gitlab.com" on-init-fetch-refs-from-pipelines=false pulling-projects-refs-interval=30s rate-limit=10rps
time="2020-04-28T23:09:01Z" level=info msg="started, now serving requests" listen-address=":8080"
time="2020-04-28T23:09:01Z" level=info msg="discover wildcards" count=0
time="2020-04-28T23:09:14Z" level=info msg="discovered new project ref" project-id=250833 project-path-with-namespace=gitlab-org/gitlab-runner project-ref=master project-ref-kind=branch
time="2020-04-28T23:09:15Z" level=info msg="discovered new project ref" project-id=11915984 project-path-with-namespace=gitlab-org/charts/auto-deploy-app project-ref=master project-ref-kind=branch
time="2020-04-28T23:09:15Z" level=info msg="pulling metrics from projects refs" count=2
```

### Check we can fetch metrics from the exporter container

```bash
~$ curl -s http://localhost:8080/metrics | grep project
gitlab_ci_pipeline_last_run_duration_seconds{project="gitlab-org/charts/auto-deploy-app",ref="master"} 36
gitlab_ci_pipeline_last_run_duration_seconds{project="gitlab-org/gitlab-runner",ref="master"} 3875
gitlab_ci_pipeline_status{project="gitlab-org/charts/auto-deploy-app",ref="master",status="failed"} 0
gitlab_ci_pipeline_status{project="gitlab-org/charts/auto-deploy-app",ref="master",status="running"} 0
gitlab_ci_pipeline_status{project="gitlab-org/charts/auto-deploy-app",ref="master",status="success"} 1
gitlab_ci_pipeline_status{project="gitlab-org/gitlab-runner",ref="master",status="failed"} 0
gitlab_ci_pipeline_status{project="gitlab-org/gitlab-runner",ref="master",status="running"} 0
gitlab_ci_pipeline_status{project="gitlab-org/gitlab-runner",ref="master",status="success"} 1
gitlab_ci_pipeline_time_since_last_run_seconds{project="gitlab-org/charts/auto-deploy-app",ref="master"} 1.251363e+06
gitlab_ci_pipeline_time_since_last_run_seconds{project="gitlab-org/gitlab-runner",ref="master"} 91799
```

### Checkout prometheus targets and available metrics

You can open this URL in your browser and should see the exporter is being configured and pulled correctly:

[http://localhost:9090/targets](http://localhost:9090/targets)

![prometheus_targets](/docs/images/prometheus_targets.png)

You should then be able to see that following metrics being displayed:

[http://localhost:9090/graph?g0.range_input=1h&g0.expr=gitlab_ci_pipeline_status&g0.tab=1](http://localhost:9090/graph?g0.range_input=1h&g0.expr=gitlab_ci_pipeline_status&g0.tab=1)

![prometheus_metrics](/docs/images/prometheus_metrics.png)

### Checkout the grafana example dashboard

An example dashboard should be available at this address:

[http://localhost:3000/d/gitlab_ci_pipeline_statuses/gitlab-ci-pipelines-statuses](http://localhost:3000/d/gitlab_ci_pipeline_statuses/gitlab-ci-pipelines-statuses)

![grafana_dashboard_example](/docs/images/grafana_dashboard_example.png)

## Perform configuration changes

I believe it would be more interesting for you to be monitoring your own projects. To perform configuration changes, there are 2 simple steps:

```bash
# Edit the configuration file for the exporter
~$ vi ./gitlab-ci-pipelines-exporter/config.yml

# Restart the exporter container
~$ docker-compose restart gitlab-ci-pipelines-exporter
```

## Cleanup

```bash
~$ docker-compose down
```
