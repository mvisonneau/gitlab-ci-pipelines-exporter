module github.com/mvisonneau/gitlab-ci-pipelines-exporter

go 1.14

require (
	github.com/docker/distribution v2.7.1+incompatible
	github.com/gorilla/mux v1.7.4 // indirect
	github.com/heptiolabs/healthcheck v0.0.0-20180807145615-6ff867650f40
	github.com/mvisonneau/go-helpers v0.0.0-20200224131125-cb5cc4e6def9
	github.com/openlyinc/pointy v1.1.2
	github.com/prometheus/client_golang v1.6.0
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.5.1
	github.com/urfave/cli v1.22.4
	github.com/xanzy/go-gitlab v0.32.0
	go.uber.org/ratelimit v0.1.0
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c
)
