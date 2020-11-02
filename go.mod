module github.com/mvisonneau/gitlab-ci-pipelines-exporter

go 1.15

require (
	github.com/alicebob/miniredis v2.5.0+incompatible
	github.com/alicebob/miniredis/v2 v2.14.1
	github.com/go-redis/redis/v8 v8.3.3
	github.com/go-redis/redis_rate/v9 v9.0.2
	github.com/gomodule/redigo v1.8.2 // indirect
	github.com/heptiolabs/healthcheck v0.0.0-20180807145615-6ff867650f40
	github.com/mvisonneau/go-helpers v0.0.0-20201013090751-e69b7251ab02
	github.com/openlyinc/pointy v1.1.2
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.8.0
	github.com/sirupsen/logrus v1.7.0
	github.com/stretchr/testify v1.6.1
	github.com/urfave/cli/v2 v2.3.0
	github.com/vmihailenco/msgpack/v5 v5.0.0-beta.9
	github.com/vmihailenco/taskq/v3 v3.2.0
	github.com/xanzy/go-gitlab v0.39.0
	go.uber.org/ratelimit v0.1.0
	gopkg.in/DATA-DOG/go-sqlmock.v1 v1.3.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
)

replace (
	github.com/vmihailenco/taskq/v3 v3.2.0 => github.com/mvisonneau/taskq/v3 v3.1.2-0.20201102181824-e475be0420f0
	github.com/xanzy/go-gitlab v0.39.0 => github.com/mvisonneau/go-gitlab v0.20.2-0.20201031120209-a4b33b12e52f
)
