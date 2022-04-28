# Example monitoring of gitlab-ci-pipelines-exporter with Jaeger

## Requirements

- **~5 min of your time**
- A personal access token on [gitlab.com](https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html) (or your own instance) with `read_api` scope
- [git](https://git-scm.com/) & [docker-compose](https://docs.docker.com/compose/)

## 🚀

```bash
# Clone this repository
~$ git clone https://github.com/mvisonneau/gitlab-ci-pipelines-exporter.git
~$ cd gitlab-ci-pipelines-exporter/examples/opentelemetry

# Provide your personal GitLab API access token (needs read_api permissions)
~$ sed -i 's/<your_token>/xXF_xxjV_xxyzxzz/' gitlab-ci-pipelines-exporter.yml

# Start gitlab-ci-pipelines-exporter, prometheus and grafana containers !
~$ docker-compose up -d
Creating network "opentelemetry_default" with driver "bridge"
Creating opentelemetry_jaeger_1 ... done
Creating opentelemetry_redis_1  ... done
Creating opentelemetry_otel-collector_1 ... done
Creating opentelemetry_gitlab-ci-pipelines-exporter_1 ... done
Creating opentelemetry_prometheus_1                   ... done
Creating opentelemetry_grafana_1                      ... done
```

You should now have a stack completely configured and accessible at these locations:

- `gitlab-ci-pipelines-exporter`: [http://localhost:8080/metrics](http://localhost:8080/metrics)
- `jaeger`: [http://localhost:16686](http://localhost:16686)
- `prometheus`: [http://localhost:9090](http://localhost:9090)
- `grafana`: [http://localhost:3000](http://localhost:3000) (if you want/need to login, creds are _admin/admin_)

## Use and troubleshoot

### Validate that containers are running

```bash
~$ docker-compose ps   
                    Name                                  Command               State                                     Ports                                  
-----------------------------------------------------------------------------------------------------------------------------------------------------------------
opentelemetry_gitlab-ci-pipelines-exporter_1   /usr/local/bin/gitlab-ci-p ...   Up      0.0.0.0:8080->8080/tcp                                                   
opentelemetry_grafana_1                        /run.sh                          Up      0.0.0.0:3000->3000/tcp                                                   
opentelemetry_jaeger_1                         /go/bin/all-in-one-linux         Up      14250/tcp, 14268/tcp, 0.0.0.0:16686->16686/tcp, 5775/udp, 5778/tcp,      
                                                                                        6831/udp, 6832/udp                                                       
opentelemetry_otel-collector_1                 /otelcontribcol --config=/ ...   Up      0.0.0.0:4317->4317/tcp, 55679/tcp, 55680/tcp                             
opentelemetry_prometheus_1                     /bin/prometheus --config.f ...   Up      0.0.0.0:9090->9090/tcp                                                   
opentelemetry_redis_1                          /opt/bitnami/scripts/redis ...   Up      0.0.0.0:6379->6379/tcp 
```

## Cleanup

```bash
~$ docker-compose down
```
