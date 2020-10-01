module github.com/mvisonneau/gitlab-ci-pipelines-exporter

go 1.15

require (
	github.com/go-redis/redis/v8 v8.2.3
	github.com/go-redis/redis_rate/v9 v9.0.2
	github.com/heptiolabs/healthcheck v0.0.0-20180807145615-6ff867650f40
	github.com/mvisonneau/go-helpers v0.0.0-20200224131125-cb5cc4e6def9
	github.com/openlyinc/pointy v1.1.2
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.7.1
	github.com/sirupsen/logrus v1.7.0
	github.com/stretchr/testify v1.6.1
	github.com/urfave/cli/v2 v2.2.0
	github.com/xanzy/go-gitlab v0.38.1
	go.uber.org/ratelimit v0.1.0
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
)
