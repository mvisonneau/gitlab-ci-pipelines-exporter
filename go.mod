module github.com/mvisonneau/gitlab-ci-pipelines-exporter

go 1.14

require (
	github.com/google/go-cmp v0.4.0
	github.com/heptiolabs/healthcheck v0.0.0-20180807145615-6ff867650f40
	github.com/jpillora/backoff v1.0.0
	github.com/mvisonneau/go-helpers v0.0.0-20200224131125-cb5cc4e6def9
	github.com/prometheus/client_golang v1.5.1
	github.com/sirupsen/logrus v1.5.0
	github.com/stretchr/testify v1.5.1
	github.com/urfave/cli v1.22.3
	github.com/xanzy/go-gitlab v0.29.0
	go.uber.org/atomic v1.6.0 // indirect
	go.uber.org/ratelimit v0.1.0
	gopkg.in/DATA-DOG/go-sqlmock.v1 v1.3.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200121175148-a6ecf24a6d71
)
