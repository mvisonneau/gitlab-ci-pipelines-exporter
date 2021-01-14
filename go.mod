module github.com/mvisonneau/gitlab-ci-pipelines-exporter

go 1.15

require (
	github.com/alicebob/miniredis v2.5.0+incompatible
	github.com/alicebob/miniredis/v2 v2.14.1
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/go-redis/redis/v8 v8.4.8
	github.com/go-redis/redis_rate/v9 v9.1.0
	github.com/gomodule/redigo v1.8.3 // indirect
	github.com/google/uuid v1.1.4 // indirect
	github.com/hashicorp/go-retryablehttp v0.6.8 // indirect
	github.com/heptiolabs/healthcheck v0.0.0-20180807145615-6ff867650f40
	github.com/klauspost/compress v1.11.6 // indirect
	github.com/mvisonneau/go-helpers v0.0.1
	github.com/openlyinc/pointy v1.1.2
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.9.0
	github.com/prometheus/procfs v0.3.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sirupsen/logrus v1.7.0
	github.com/stretchr/testify v1.6.1
	github.com/urfave/cli/v2 v2.3.0
	github.com/vmihailenco/msgpack/v5 v5.1.4
	github.com/vmihailenco/taskq/v3 v3.2.3
	github.com/xanzy/go-gitlab v0.42.0
	go.opentelemetry.io/otel v0.16.0 // indirect
	go.uber.org/ratelimit v0.1.0
	golang.org/x/net v0.0.0-20201224014010-6772e930b67b // indirect
	golang.org/x/oauth2 v0.0.0-20210113205817-d3ed898aa8a3 // indirect
	golang.org/x/sys v0.0.0-20210113181707-4bcb84eeeb78 // indirect
	golang.org/x/time v0.0.0-20201208040808-7e3f01d25324 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	gopkg.in/DATA-DOG/go-sqlmock.v1 v1.3.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

replace github.com/vmihailenco/taskq/v3 => github.com/mvisonneau/taskq/v3 v3.2.4-0.20201127170227-fddacd1811f5
