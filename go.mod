module github.com/mvisonneau/gitlab-ci-pipelines-exporter

go 1.16

require (
	github.com/alicebob/miniredis v2.5.0+incompatible
	github.com/alicebob/miniredis/v2 v2.14.3
	github.com/go-redis/redis/v8 v8.8.2
	github.com/go-redis/redis_rate/v9 v9.1.1
	github.com/gomodule/redigo v1.8.4 // indirect
	github.com/heptiolabs/healthcheck v0.0.0-20180807145615-6ff867650f40
	github.com/mvisonneau/go-helpers v0.0.1
	github.com/openlyinc/pointy v1.1.2
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.10.0
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/urfave/cli/v2 v2.3.0
	github.com/vmihailenco/msgpack/v5 v5.3.1
	github.com/vmihailenco/taskq/v3 v3.2.3
	github.com/xanzy/go-gitlab v0.48.0
	go.uber.org/ratelimit v0.2.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

replace github.com/vmihailenco/taskq/v3 => github.com/mvisonneau/taskq/v3 v3.2.4-0.20201127170227-fddacd1811f5
