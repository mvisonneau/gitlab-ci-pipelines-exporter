package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func Validate(cliCtx *cli.Context) (int, error) {
	log.Debug("Validating configuration..")

	if _, err := configure(cliCtx); err != nil {
		log.WithError(err).Error("Failed to configure")

		return 1, err
	}

	log.Debug("Configuration is valid")

	return 0, nil
}
