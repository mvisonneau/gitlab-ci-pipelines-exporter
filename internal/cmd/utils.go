package cmd

import (
	"context"
	"fmt"
	stdlibLog "log"
	"net/url"
	"os"
	"time"

	"github.com/go-logr/stdr"
	log "github.com/sirupsen/logrus"
	"github.com/uptrace/opentelemetry-go-extra/otellogrus"
	"github.com/urfave/cli/v3"
	"github.com/vmihailenco/taskq/v4"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/config"
	"github.com/mvisonneau/go-helpers/logger"
)

var (
	start      time.Time
	appVersion string
)

// Init sets runtime metadata needed by command handlers (timings, version).
func Init(version string, startTime time.Time) {
	appVersion = version
	start = startTime
}

func configure(cmd *cli.Command) (cfg config.Config, err error) {
	assertStringVariableDefined(cmd, "config")

	cfg, err = config.ParseFile(cmd.String("config"))
	if err != nil {
		return
	}

	cfg.Global, err = parseGlobalFlags(cmd)
	if err != nil {
		return
	}

	configCliOverrides(cmd, &cfg)

	if err = cfg.Validate(); err != nil {
		return
	}

	// Configure logger
	if err = logger.Configure(logger.Config{
		Level:  cfg.Log.Level,
		Format: cfg.Log.Format,
	}); err != nil {
		return
	}

	log.AddHook(otellogrus.NewHook(otellogrus.WithLevels(
		log.PanicLevel,
		log.FatalLevel,
		log.ErrorLevel,
		log.WarnLevel,
	)))

	// This hack is to embed taskq logs with logrus
	taskq.SetLogger(stdr.New(stdlibLog.New(log.StandardLogger().WriterLevel(log.WarnLevel), "taskq", 0)))

	log.WithFields(
		log.Fields{
			"gitlab-endpoint":   cfg.Gitlab.URL,
			"gitlab-rate-limit": fmt.Sprintf("%drps", cfg.Gitlab.MaximumRequestsPerSecond),
		},
	).Info("configured")

	log.WithFields(config.SchedulerConfig(cfg.Pull.ProjectsFromWildcards).Log()).Info("pull projects from wildcards")
	log.WithFields(config.SchedulerConfig(cfg.Pull.EnvironmentsFromProjects).Log()).Info("pull environments from projects")
	log.WithFields(config.SchedulerConfig(cfg.Pull.RefsFromProjects).Log()).Info("pull refs from projects")
	log.WithFields(config.SchedulerConfig(cfg.Pull.Metrics).Log()).Info("pull metrics")

	log.WithFields(config.SchedulerConfig(cfg.GarbageCollect.Projects).Log()).Info("garbage collect projects")
	log.WithFields(config.SchedulerConfig(cfg.GarbageCollect.Environments).Log()).Info("garbage collect environments")
	log.WithFields(config.SchedulerConfig(cfg.GarbageCollect.Refs).Log()).Info("garbage collect refs")
	log.WithFields(config.SchedulerConfig(cfg.GarbageCollect.Metrics).Log()).Info("garbage collect metrics")

	return
}

func parseGlobalFlags(cmd *cli.Command) (cfg config.Global, err error) {
	if listenerAddr := cmd.String("internal-monitoring-listener-address"); listenerAddr != "" {
		cfg.InternalMonitoringListenerAddress, err = url.Parse(listenerAddr)
	}

	return
}

func exit(exitCode int, err error) cli.ExitCoder {
	defer func() {
		log.WithFields(
			log.Fields{
				"execution-time": time.Since(start),
			},
		).Debug("exited..")
	}()

	if err != nil {
		log.WithError(err).Error()
	}

	return cli.Exit("", exitCode)
}

// ExecWrapper gracefully logs and exits our `run` functions.
func ExecWrapper(f func(ctx context.Context, cmd *cli.Command) (int, error)) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		return exit(f(ctx, cmd))
	}
}

func configCliOverrides(cmd *cli.Command, cfg *config.Config) {
	if cmd.String("gitlab-token") != "" {
		cfg.Gitlab.Token = cmd.String("gitlab-token")
	}

	if cfg.Server.Webhook.Enabled {
		if cmd.String("webhook-secret-token") != "" {
			cfg.Server.Webhook.SecretToken = cmd.String("webhook-secret-token")
		}
	}

	if cmd.String("redis-url") != "" {
		cfg.Redis.URL = cmd.String("redis-url")
	}

	if healthURL := cmd.String("gitlab-health-url"); healthURL != "" {
		cfg.Gitlab.HealthURL = healthURL
		cfg.Gitlab.EnableHealthCheck = true
	}
}

func assertStringVariableDefined(cmd *cli.Command, k string) {
	if len(cmd.String(k)) == 0 {
		log.Errorf("'--%s' must be set!", k)
		os.Exit(2)
	}
}
