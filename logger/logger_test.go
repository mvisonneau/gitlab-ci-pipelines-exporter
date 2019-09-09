package logger

import (
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestLoggerConfigureFatalText(t *testing.T) {
	c := &Config{Level: "fatal", Format: "text"}
	c.Configure()

	if log.GetLevel() != log.FatalLevel {
		t.Fatalf("Expected log.Level to be 'fatal' but got %s", log.GetLevel())
	}
}

func TestLoggerConfigureDefault(t *testing.T) {
	c := &Config{Level: "fatal", Format: "foo"}
	if err := c.Configure(); err == nil {
		t.Fatal("Expected function to return an error, got nil")
	}
}

func TestLoggerConfigureJson(t *testing.T) {
	c := &Config{Level: "debug", Format: "json"}
	if err := c.Configure(); err != nil {
		t.Fatalf("Function is not expected to return an error, got '%s'", err.Error())
	}
}

func TestLoggerConfigureInvalidLogFormat(t *testing.T) {
	c := &Config{Level: "foo", Format: "text"}
	if err := c.Configure(); err == nil {
		t.Fatal("Expected function to return an error, got nil")
	}
}
