package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/internal/cmd"
)

// Run handles the instantiation of the CLI application.
func Run(version string, args []string) {
	err := NewApp(version, time.Now()).Run(context.Background(), args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// NewApp configures the CLI application.
func NewApp(version string, start time.Time) (app *cli.Command) {
	cmd.Init(version, start)

	app = &cli.Command{
		Name:                  "gitlab-ci-pipelines-exporter",
		Version:               version,
		Usage:                 "Export metrics about GitLab CI pipelines statuses",
		EnableShellCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "internal-monitoring-listener-address",
				Aliases: []string{"m"},
				Sources: cli.EnvVars("GCPE_INTERNAL_MONITORING_LISTENER_ADDRESS"),
				Usage:   "internal monitoring listener address",
			},
		},
		Commands: []*cli.Command{
			{
				Name:   "run",
				Usage:  "start the exporter",
				Action: cmd.ExecWrapper(cmd.Run),
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "config",
						Aliases: []string{"c"},
						Sources: cli.EnvVars("GCPE_CONFIG"),
						Usage:   "config `file`",
						Value:   "./gitlab-ci-pipelines-exporter.yml",
					},
					&cli.StringFlag{
						Name:    "redis-url",
						Sources: cli.EnvVars("GCPE_REDIS_URL"),
						Usage:   "redis `url` for an HA setup (format: redis[s]://[:password@]host[:port][/db-number][?option=value]) (overrides config file parameter)",
					},
					&cli.StringFlag{
						Name:    "gitlab-token",
						Sources: cli.EnvVars("GCPE_GITLAB_TOKEN"),
						Usage:   "GitLab API access `token` (overrides config file parameter)",
					},
					&cli.StringFlag{
						Name:    "webhook-secret-token",
						Sources: cli.EnvVars("GCPE_WEBHOOK_SECRET_TOKEN"),
						Usage:   "`token` used to authenticate legitimate requests (overrides config file parameter)",
					},
					&cli.StringFlag{
						Name:    "gitlab-health-url",
						Sources: cli.EnvVars("GCPE_GITLAB_HEALTH_URL"),
						Usage:   "GitLab health URL (overrides config file parameter)",
					},
				},
			},
			{
				Name:   "validate",
				Usage:  "validate the configuration file",
				Action: cmd.ExecWrapper(cmd.Validate),
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "config",
						Aliases: []string{"c"},
						Sources: cli.EnvVars("GCPE_CONFIG"),
						Usage:   "config `file`",
						Value:   "./gitlab-ci-pipelines-exporter.yml",
					},
				},
			},
			{
				Name:   "monitor",
				Usage:  "display information about the currently running exporter",
				Action: cmd.ExecWrapper(cmd.Monitor),
			},
		},
	}

	return
}
