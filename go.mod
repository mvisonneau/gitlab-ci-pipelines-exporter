module github.com/mvisonneau/gitlab-ci-pipelines-exporter

go 1.18

require (
	github.com/alecthomas/chroma v0.10.0
	github.com/alicebob/miniredis/v2 v2.20.0
	github.com/charmbracelet/bubbles v0.10.3
	github.com/charmbracelet/bubbletea v0.20.0
	github.com/charmbracelet/lipgloss v0.5.0
	github.com/creasty/defaults v1.6.0
	github.com/go-playground/validator/v10 v10.10.1
	github.com/go-redis/redis/extra/redisotel/v8 v8.11.5
	github.com/go-redis/redis/v8 v8.11.5
	github.com/go-redis/redis_rate/v9 v9.1.2
	github.com/google/uuid v1.3.0
	github.com/heptiolabs/healthcheck v0.0.0-20211123025425-613501dd5deb
	github.com/imdario/mergo v0.3.12
	github.com/mvisonneau/go-helpers v0.0.1
	github.com/openlyinc/pointy v1.1.2
	github.com/paulbellamy/ratecounter v0.2.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.12.1
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.1
	github.com/uptrace/opentelemetry-go-extra/otellogrus v0.1.12
	github.com/urfave/cli/v2 v2.5.0
	github.com/vmihailenco/msgpack/v5 v5.3.5
	github.com/vmihailenco/taskq/v3 v3.2.8
	github.com/xanzy/go-gitlab v0.64.0
	github.com/xeonx/timeago v1.0.0-rc4
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.31.0
	go.opentelemetry.io/otel v1.7.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.7.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.7.0
	go.opentelemetry.io/otel/sdk v1.7.0
	go.opentelemetry.io/otel/trace v1.7.0
	go.uber.org/ratelimit v0.2.0
	google.golang.org/grpc v1.46.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

require (
	github.com/alicebob/gopher-json v0.0.0-20200520072559-a9ecdc9d1d3a // indirect
	github.com/andres-erbsen/clock v0.0.0-20160526145045-9e14626cd129 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bsm/redislock v0.7.2 // indirect
	github.com/capnm/sysinfo v0.0.0-20130621111458-5909a53897f3 // indirect
	github.com/cenkalti/backoff/v4 v4.1.3 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/charmbracelet/harmonica v0.2.0 // indirect
	github.com/containerd/console v1.0.3 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dlclark/regexp2 v1.4.0 // indirect
	github.com/felixge/httpsnoop v1.0.2 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-playground/locales v0.14.0 // indirect
	github.com/go-playground/universal-translator v0.18.0 // indirect
	github.com/go-redis/redis/extra/rediscmd/v8 v8.11.5 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.7.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.1 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/klauspost/compress v1.15.2 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/muesli/ansi v0.0.0-20211031195517-c9f0611b6c70 // indirect
	github.com/muesli/reflow v0.3.0 // indirect
	github.com/muesli/termenv v0.11.1-0.20220212125758-44cd13922739 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.34.0 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/uptrace/opentelemetry-go-extra/otelutil v0.1.12 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/yuin/gopher-lua v0.0.0-20210529063254-f4c35e4016d9 // indirect
	go.opentelemetry.io/otel/exporters/otlp/internal/retry v1.7.0 // indirect
	go.opentelemetry.io/otel/metric v0.28.0 // indirect
	go.opentelemetry.io/proto/otlp v0.16.0 // indirect
	golang.org/x/crypto v0.0.0-20220411220226-7b82a4e95df4 // indirect
	golang.org/x/net v0.0.0-20220425223048-2871e0cb64e4 // indirect
	golang.org/x/oauth2 v0.0.0-20220411215720-9780585627b5 // indirect
	golang.org/x/sys v0.0.0-20220422013727-9388b58f7150 // indirect
	golang.org/x/term v0.0.0-20220411215600-e5f449aeb171 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/time v0.0.0-20220411224347-583f2d630306 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20211118181313-81c1377c94b1 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/DATA-DOG/go-sqlmock.v1 v1.3.0 // indirect
)

replace github.com/vmihailenco/taskq/v3 => github.com/mvisonneau/taskq/v3 v3.2.4-0.20220317203606-c903565cde04
