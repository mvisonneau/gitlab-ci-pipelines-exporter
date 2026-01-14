package cmd

import (
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"
)

func Validate(_ context.Context, cliCmd *cli.Command) (int, error) {
	log.Debug("Validating configuration..")

	if _, err := configure(cliCmd); err != nil {
		log.WithError(err).Error("Failed to configure")

		return 1, err
	}

	log.Debug("Configuration is valid")

	return 0, nil
}
