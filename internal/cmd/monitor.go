package cmd

import (
	"github.com/urfave/cli/v2"

	monitorUI "github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/monitor/ui"
)

// Monitor ..
func Monitor(ctx *cli.Context) (int, error) {
	cfg, err := parseGlobalFlags(ctx)
	if err != nil {
		return 1, err
	}

	monitorUI.Start(
		ctx.App.Version,
		cfg.InternalMonitoringListenerAddress,
	)

	return 0, nil
}
