package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/internal/cmd"
)

// Run handles the instanciation of the CLI application.
func Run(version string, args []string) {
	err := NewApp(version, time.Now()).Run(args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// NewApp configures the CLI application.
func NewApp(version string, start time.Time) (app *cli.App) {
	app = cli.NewApp()
	app.Name = "gitlab-ci-pipelines-exporter"
	app.Version = version
	app.Usage = "Export metrics about GitLab CI pipelines statuses"
	app.EnableBashCompletion = true

	app.Flags = cli.FlagsByName{
		&cli.StringFlag{
			Name:    "internal-monitoring-listener-address",
			Aliases: []string{"m"},
			EnvVars: []string{"GCPE_INTERNAL_MONITORING_LISTENER_ADDRESS"},
			Usage:   "internal monitoring listener address",
		},
	}

	app.Commands = cli.CommandsByName{
		{
			Name:   "run",
			Usage:  "start the exporter",
			Action: cmd.ExecWrapper(cmd.Run),
			Flags: cli.FlagsByName{
				&cli.StringFlag{
					Name:    "config",
					Aliases: []string{"c"},
					EnvVars: []string{"GCPE_CONFIG"},
					Usage:   "config `file`",
					Value:   "./gitlab-ci-pipelines-exporter.yml",
				},
				&cli.StringFlag{
					Name:    "redis-url",
					EnvVars: []string{"GCPE_REDIS_URL"},
					Usage:   "redis `url` for an HA setup (format: redis[s]://[:password@]host[:port][/db-number][?option=value]) (overrides config file parameter)",
				},
				&cli.StringFlag{
					Name:    "gitlab-token",
					EnvVars: []string{"GCPE_GITLAB_TOKEN"},
					Usage:   "GitLab API access `token` (overrides config file parameter)",
				},
				&cli.StringFlag{
					Name:    "webhook-secret-token",
					EnvVars: []string{"GCPE_WEBHOOK_SECRET_TOKEN"},
					Usage:   "`token` used to authenticate legitimate requests (overrides config file parameter)",
				},
				&cli.StringFlag{
					Name:    "gitlab-health-url",
					EnvVars: []string{"GCPE_GITLAB_HEALTH_URL"},
					Usage:   "GitLab health URL (overrides config file parameter)",
				},
			},
		},
		{
			Name:   "validate",
			Usage:  "validate the configuration file",
			Action: cmd.ExecWrapper(cmd.Validate),
			Flags: cli.FlagsByName{
				&cli.StringFlag{
					Name:    "config",
					Aliases: []string{"c"},
					EnvVars: []string{"GCPE_CONFIG"},
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
	}

	app.Metadata = map[string]interface{}{
		"startTime": start,
	}

	return
}
