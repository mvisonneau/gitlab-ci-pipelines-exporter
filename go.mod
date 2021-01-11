module github.com/mvisonneau/gitlab-ci-pipelines-exporter

go 1.15

require (
	github.com/alicebob/miniredis v2.5.0+incompatible
	github.com/alicebob/miniredis/v2 v2.14.1
	github.com/go-redis/redis/v8 v8.4.8
	github.com/go-redis/redis_rate/v9 v9.1.0
	github.com/gomodule/redigo v1.8.3 // indirect
	github.com/hashicorp/go-retryablehttp v0.6.8 // indirect
	github.com/heptiolabs/healthcheck v0.0.0-20180807145615-6ff867650f40
	github.com/mvisonneau/go-helpers v0.0.1
	github.com/openlyinc/pointy v1.1.2
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.9.0
	github.com/sirupsen/logrus v1.7.0
	github.com/stretchr/testify v1.6.1
	github.com/urfave/cli/v2 v2.3.0
	github.com/vmihailenco/msgpack/v5 v5.1.4
	github.com/vmihailenco/taskq/v3 v3.2.3
	github.com/xanzy/go-gitlab v0.40.3-0.20201224125403-b95aa2de920d
	go.uber.org/ratelimit v0.1.0
	golang.org/x/oauth2 v0.0.0-20200902213428-5d25da1a8d43 // indirect
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	google.golang.org/appengine v1.6.7 // indirect
	gopkg.in/DATA-DOG/go-sqlmock.v1 v1.3.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
)

replace github.com/vmihailenco/taskq/v3 => github.com/mvisonneau/taskq/v3 v3.2.4-0.20201127170227-fddacd1811f5
