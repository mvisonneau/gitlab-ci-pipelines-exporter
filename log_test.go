package main

import (
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestConfigureLoggingFatalText(t *testing.T) {
	configureLogging("fatal", "text")

	if log.GetLevel() != log.FatalLevel {
		t.Fatalf("Expected log.Level to be 'fatal' but got %s", log.GetLevel())
	}
}

func TestConfigureLoggingDefault(t *testing.T) {
	err := configureLogging("fatal", "default")

	if err == nil {
		t.Fatal("Expected function to return an error, got nil")
	}
}
