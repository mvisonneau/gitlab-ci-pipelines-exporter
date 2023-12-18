package controller

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
	"github.com/vmihailenco/taskq/v4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"google.golang.org/grpc"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/config"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/gitlab"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/ratelimit"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/schemas"
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/store"
)

const tracerName = "gitlab-ci-pipelines-exporter"

// Controller holds the necessary clients to run the app and handle requests.
type Controller struct {
	Config         config.Config
	Redis          *redis.Client
	Gitlab         *gitlab.Client
	Store          store.Store
	TaskController TaskController

	// UUID is used to identify this controller/process amongst others when
	// the exporter is running in cluster mode, leveraging Redis.
	UUID uuid.UUID
}

// New creates a new controller.
func New(ctx context.Context, cfg config.Config, version string) (c Controller, err error) {
	c.Config = cfg
	c.UUID = uuid.New()

	if err = configureTracing(ctx, cfg.OpenTelemetry.GRPCEndpoint); err != nil {
		return
	}

	if err = c.configureRedis(ctx, cfg.Redis.URL); err != nil {
		return
	}

	c.TaskController = NewTaskController(ctx, c.Redis, cfg.Gitlab.MaximumJobsQueueSize)
	c.registerTasks()

	c.Store = store.New(ctx, c.Redis, c.Config.Projects)

	if err = c.configureGitlab(cfg.Gitlab, version); err != nil {
		return
	}

	// Start the scheduler
	c.Schedule(ctx, cfg.Pull, cfg.GarbageCollect)

	return
}

func (c *Controller) registerTasks() {
	for n, h := range map[schemas.TaskType]interface{}{
		schemas.TaskTypeGarbageCollectEnvironments:   c.TaskHandlerGarbageCollectEnvironments,
		schemas.TaskTypeGarbageCollectMetrics:        c.TaskHandlerGarbageCollectMetrics,
		schemas.TaskTypeGarbageCollectProjects:       c.TaskHandlerGarbageCollectProjects,
		schemas.TaskTypeGarbageCollectRefs:           c.TaskHandlerGarbageCollectRefs,
		schemas.TaskTypePullEnvironmentMetrics:       c.TaskHandlerPullEnvironmentMetrics,
		schemas.TaskTypePullEnvironmentsFromProject:  c.TaskHandlerPullEnvironmentsFromProject,
		schemas.TaskTypePullEnvironmentsFromProjects: c.TaskHandlerPullEnvironmentsFromProjects,
		schemas.TaskTypePullMetrics:                  c.TaskHandlerPullMetrics,
		schemas.TaskTypePullProject:                  c.TaskHandlerPullProject,
		schemas.TaskTypePullProjectsFromWildcard:     c.TaskHandlerPullProjectsFromWildcard,
		schemas.TaskTypePullProjectsFromWildcards:    c.TaskHandlerPullProjectsFromWildcards,
		schemas.TaskTypePullRefMetrics:               c.TaskHandlerPullRefMetrics,
		schemas.TaskTypePullRefsFromProject:          c.TaskHandlerPullRefsFromProject,
		schemas.TaskTypePullRefsFromProjects:         c.TaskHandlerPullRefsFromProjects,
	} {
		_, _ = c.TaskController.TaskMap.Register(string(n), &taskq.TaskConfig{
			Handler:    h,
			RetryLimit: 1,
		})
	}
}

func (c *Controller) unqueueTask(ctx context.Context, tt schemas.TaskType, uniqueID string) {
	if err := c.Store.UnqueueTask(ctx, tt, uniqueID); err != nil {
		log.WithContext(ctx).
			WithFields(log.Fields{
				"task_type":      tt,
				"task_unique_id": uniqueID,
			}).
			WithError(err).
			Warn("unqueuing task")
	}
}

func configureTracing(ctx context.Context, grpcEndpoint string) error {
	if len(grpcEndpoint) == 0 {
		log.Debug("opentelemetry.grpc_endpoint is not configured, skipping open telemetry support")

		return nil
	}

	log.WithFields(log.Fields{
		"opentelemetry_grpc_endpoint": grpcEndpoint,
	}).Info("opentelemetry gRPC endpoint provided, initializing connection..")

	traceClient := otlptracegrpc.NewClient(
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(grpcEndpoint),
		otlptracegrpc.WithDialOption(grpc.WithBlock()))

	traceExp, err := otlptrace.New(ctx, traceClient)
	if err != nil {
		return err
	}

	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String("gitlab-ci-pipelines-exporter"),
		),
	)
	if err != nil {
		return err
	}

	bsp := sdktrace.NewBatchSpanProcessor(traceExp)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	otel.SetTracerProvider(tracerProvider)

	return nil
}

func (c *Controller) configureGitlab(cfg config.Gitlab, version string) (err error) {
	var rl ratelimit.Limiter

	if c.Redis != nil {
		rl = ratelimit.NewRedisLimiter(c.Redis, cfg.MaximumRequestsPerSecond)
	} else {
		rl = ratelimit.NewLocalLimiter(cfg.MaximumRequestsPerSecond, cfg.BurstableRequestsPerSecond)
	}

	c.Gitlab, err = gitlab.NewClient(gitlab.ClientConfig{
		URL:              cfg.URL,
		Token:            cfg.Token,
		DisableTLSVerify: !cfg.EnableTLSVerify,
		UserAgentVersion: version,
		RateLimiter:      rl,
		ReadinessURL:     cfg.HealthURL,
	})

	return
}

func (c *Controller) configureRedis(ctx context.Context, url string) (err error) {
	ctx, span := otel.Tracer(tracerName).Start(ctx, "controller:configureRedis")
	defer span.End()

	if len(url) <= 0 {
		log.Debug("redis url is not configured, skipping configuration & using local driver")

		return
	}

	log.Info("redis url configured, initializing connection..")

	var opt *redis.Options

	if opt, err = redis.ParseURL(url); err != nil {
		return
	}

	c.Redis = redis.NewClient(opt)

	if err = redisotel.InstrumentTracing(c.Redis); err != nil {
		return
	}

	if _, err := c.Redis.Ping(ctx).Result(); err != nil {
		return errors.Wrap(err, "connecting to redis")
	}

	log.Info("connected to redis")

	return
}
