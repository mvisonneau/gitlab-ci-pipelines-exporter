module github.com/mvisonneau/gitlab-ci-pipelines-exporter

go 1.16

require (
	github.com/alicebob/miniredis/v2 v2.14.3
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/creasty/defaults v1.5.1
	github.com/go-playground/validator/v10 v10.6.1
	github.com/go-redis/redis/v8 v8.10.0
	github.com/go-redis/redis_rate/v9 v9.1.1
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/uuid v1.2.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.0 // indirect
	github.com/heptiolabs/healthcheck v0.0.0-20180807145615-6ff867650f40
	github.com/imdario/mergo v0.3.12
	github.com/klauspost/compress v1.12.3 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/mvisonneau/go-helpers v0.0.1
	github.com/openlyinc/pointy v1.1.2
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/urfave/cli/v2 v2.3.0
	github.com/vmihailenco/msgpack/v5 v5.3.4
	github.com/vmihailenco/taskq/v3 v3.2.4
	github.com/xanzy/go-gitlab v0.50.0
	go.uber.org/ratelimit v0.2.0
	golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a // indirect
	golang.org/x/net v0.0.0-20210525063256-abc453219eb5 // indirect
	golang.org/x/oauth2 v0.0.0-20210514164344-f6687ab2804c // indirect
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba // indirect
	google.golang.org/appengine v1.6.7 // indirect
	gopkg.in/DATA-DOG/go-sqlmock.v1 v1.3.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

replace github.com/vmihailenco/taskq/v3 => github.com/mvisonneau/taskq/v3 v3.2.4-0.20201127170227-fddacd1811f5
