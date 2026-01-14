package cmd

import (
	"context"

	"github.com/urfave/cli/v3"

	monitorUI "github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/monitor/ui"
)

// Monitor ..
func Monitor(_ context.Context, cmd *cli.Command) (int, error) {
	cfg, err := parseGlobalFlags(cmd)
	if err != nil {
		return 1, err
	}

	monitorUI.Start(
		appVersion,
		cfg.InternalMonitoringListenerAddress,
	)

	return 0, nil
}
