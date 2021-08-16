module github.com/mvisonneau/gitlab-ci-pipelines-exporter

go 1.16

require (
	github.com/alecthomas/chroma v0.9.2
	github.com/alicebob/miniredis/v2 v2.15.1
	github.com/bsm/redislock v0.7.1 // indirect
	github.com/charmbracelet/bubbles v0.8.0
	github.com/charmbracelet/bubbletea v0.14.1
	github.com/charmbracelet/lipgloss v0.3.0
	github.com/containerd/console v1.0.2 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/creasty/defaults v1.5.1
	github.com/go-playground/validator/v10 v10.9.0
	github.com/go-redis/redis/v8 v8.11.2
	github.com/go-redis/redis_rate/v9 v9.1.1
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.0 // indirect
	github.com/heptiolabs/healthcheck v0.0.0-20180807145615-6ff867650f40
	github.com/imdario/mergo v0.3.12
	github.com/klauspost/compress v1.13.1 // indirect
	github.com/muesli/reflow v0.3.0 // indirect
	github.com/muesli/termenv v0.9.0
	github.com/mvisonneau/go-helpers v0.0.1
	github.com/openlyinc/pointy v1.1.2
	github.com/paulbellamy/ratecounter v0.2.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/prometheus/common v0.29.0 // indirect
	github.com/prometheus/procfs v0.7.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/urfave/cli/v2 v2.3.0
	github.com/vmihailenco/msgpack/v5 v5.3.4
	github.com/vmihailenco/taskq/v3 v3.2.4
	github.com/xanzy/go-gitlab v0.50.3
	github.com/xeonx/timeago v1.0.0-rc4
	go.uber.org/ratelimit v0.2.0
	golang.org/x/net v0.0.0-20210716203947-853a461950ff // indirect
	golang.org/x/oauth2 v0.0.0-20210628180205-a41e5a781914 // indirect
	golang.org/x/term v0.0.0-20210615171337-6886f2dfbf5b // indirect
	golang.org/x/time v0.0.0-20210611083556-38a9dc6acbc6 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/DATA-DOG/go-sqlmock.v1 v1.3.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

replace (
	github.com/vmihailenco/taskq/v3 => github.com/mvisonneau/taskq/v3 v3.2.4-0.20210712150957-0533f0c347b7
	github.com/xanzy/go-gitlab v0.50.1 => github.com/mvisonneau/go-gitlab v0.20.2-0.20210713152017-e61123733123
)
