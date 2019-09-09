package logger

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

// Config : Type that handles logging config
type Config struct {
	Level  string
	Format string
}

// Configure the logger
func (c *Config) Configure() error {
	parsedLevel, err := log.ParseLevel(c.Level)
	if err != nil {
		return err
	}
	log.SetLevel(parsedLevel)

	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(formatter)

	switch c.Format {
	case "text":
	case "json":
		log.SetFormatter(&log.JSONFormatter{})
	default:
		return fmt.Errorf("Invalid log format '%s'", c.Format)
	}

	log.SetOutput(os.Stdout)

	return nil
}
