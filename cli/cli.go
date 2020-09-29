package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/cmd"
	cli "github.com/urfave/cli/v2"
)

// Run handles the instanciation of the CLI application
func Run(version string, args []string) {
	err := NewApp(version, time.Now()).Run(args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// NewApp configures the CLI application
func NewApp(version string, start time.Time) (app *cli.App) {
	app = cli.NewApp()
	app.Name = "gitlab-ci-pipelines-exporter"
	app.Version = version
	app.Usage = "Export metrics about GitLab CI pipelines statuses"
	app.EnableBashCompletion = true

	app.Flags = cli.FlagsByName{
		&cli.StringFlag{
			Name:    "config",
			Aliases: []string{"c"},
			EnvVars: []string{"GCPE_CONFIG"},
			Usage:   "config `file`",
			Value:   "~/.gitlab-ci-pipelines-exporter.yml",
		},
		&cli.BoolFlag{
			Name:    "enable-pprof",
			EnvVars: []string{"GCPE_ENABLE_PPROF"},
			Usage:   "Enable profiling endpoints at /debug/pprof",
		},
		&cli.StringFlag{
			Name:    "gitlab-token",
			EnvVars: []string{"GCPE_GITLAB_TOKEN"},
			Usage:   "GitLab access `token`. Can be use to override the gitlab token in config file",
		},
		&cli.StringFlag{
			Name:    "listen-address",
			Aliases: []string{"l"},
			EnvVars: []string{"GCPE_LISTEN_ADDRESS"},
			Usage:   "listen-address `address:port`",
			Value:   ":8080",
		},
		&cli.StringFlag{
			Name:    "log-level",
			EnvVars: []string{"GCPE_LOG_LEVEL"},
			Usage:   "log `level` (debug,info,warn,fatal,panic)",
			Value:   "info",
		},
		&cli.StringFlag{
			Name:    "log-format",
			EnvVars: []string{"GCPE_LOG_FORMAT"},
			Usage:   "log `format` (json,text)",
			Value:   "text",
		},
	}

	app.Action = cmd.ExecWrapper(cmd.Run)

	app.Metadata = map[string]interface{}{
		"startTime": start,
	}

	return
}
