package cli

import (
	"os"
	"time"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/cmd"
	"github.com/urfave/cli"
)

// Run handles the instanciation of the CLI application
func Run(version string) {
	NewApp(version, time.Now()).Run(os.Args)
}

// NewApp configures the CLI application
func NewApp(version string, start time.Time) (app *cli.App) {
	app = cli.NewApp()
	app.Name = "gitlab-ci-pipelines-exporter"
	app.Version = version
	app.Usage = "Export metrics about GitLab CI pipelines statuses"
	app.EnableBashCompletion = true

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "config, c",
			EnvVar: "GCPE_CONFIG",
			Usage:  "config `file`",
			Value:  "~/.gitlab-ci-pipelines-exporter.yml",
		},
		cli.BoolFlag{
			Name:   "enable-pprof",
			EnvVar: "GCPE_ENABLE_PPROF",
			Usage:  "Enable profiling endpoints at /debug/pprof",
		},
		cli.StringFlag{
			Name:   "gitlab-token",
			EnvVar: "GCPE_GITLAB_TOKEN",
			Usage:  "GitLab access `token`. Can be use to override the gitlab token in config file",
		},
		cli.StringFlag{
			Name:   "listen-address, l",
			EnvVar: "GCPE_LISTEN_ADDRESS",
			Usage:  "listen-address `address:port`",
			Value:  ":8080",
		},
		cli.StringFlag{
			Name:   "log-level",
			EnvVar: "GCPE_LOG_LEVEL",
			Usage:  "log `level` (debug,info,warn,fatal,panic)",
			Value:  "info",
		},
		cli.StringFlag{
			Name:   "log-format",
			EnvVar: "GCPE_LOG_FORMAT",
			Usage:  "log `format` (json,text)",
			Value:  "text",
		},
	}

	app.Action = cmd.ExecWrapper(cmd.Run)

	app.Metadata = map[string]interface{}{
		"startTime": start,
	}

	return
}
