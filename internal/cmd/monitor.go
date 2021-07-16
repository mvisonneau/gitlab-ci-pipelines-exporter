package cmd

import (
	monitorUI "github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/monitor/ui"
	"github.com/urfave/cli/v2"
)

// Monitor ..
func Monitor(ctx *cli.Context) (int, error) {
	monitorUI.Start(ctx.App.Version)
	return 0, nil
}
