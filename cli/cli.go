package cli

import (
	"io/ioutil"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/cmd"
	"github.com/urfave/cli"
)

// Init : Generates CLI configuration for the application
func Init(version *string) (app *cli.App) {
	app = cli.NewApp()
	app.Name = "gitlab-ci-pipelines-exporter"
	app.Version = *version
	app.Usage = "Export metrics about GitLab CI pipelines statuses"
	app.EnableBashCompletion = true

	app.Flags = []cli.Flag{
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
		cli.StringFlag{
			Name:   "listen-address, l",
			EnvVar: "GCPE_LISTEN_ADDRESS",
			Usage:  "listen-address `address:port`",
			Value:  ":8080",
		},
		cli.StringFlag{
			Name:   "config, c",
			EnvVar: "GCPE_CONFIG",
			Usage:  "config `file`",
			Value:  "~/.gitlab-ci-pipelines-exporter.yml",
		},
	}

	app.Action = cmd.Run

	// Disable error handling from urfave/cli
	cli.ErrWriter = ioutil.Discard

	return
}
