package main

import (
	"testing"
)

func TestRunCli(t *testing.T) {
	app := runCli()
	if app.Name != "gitlab-ci-pipelines-exporter" {
		t.Fatalf("Expected c.Name to be gitlab-ci-pipelines-exporter, got '%v'", app.Name)
	}
}
