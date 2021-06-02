package cmd

import (
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/controller"
	"github.com/urfave/cli/v2"
)

// Run launches the exporter
func Run(ctx *cli.Context) (int, error) {
	if err := configure(ctx); err != nil {
		return 1, err
	}

	controller.Run()
	return 0, nil
}
