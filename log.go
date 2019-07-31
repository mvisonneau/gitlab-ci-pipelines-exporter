package main

import (
	"errors"
	"os"

	log "github.com/sirupsen/logrus"
)

func configureLogging(level string, format string) error {
	parsedLevel, _ := log.ParseLevel(level)
	log.SetLevel(parsedLevel)

	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(formatter)

	switch format {
	case "text":
	case "json":
		log.SetFormatter(&log.JSONFormatter{})
	default:
		return errors.New("Invalid log format")
	}

	log.SetOutput(os.Stdout)

	return nil
}
