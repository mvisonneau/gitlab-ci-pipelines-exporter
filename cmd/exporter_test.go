package cmd

import (
	"flag"

	"testing"

	"github.com/urfave/cli"
)

func TestRunWrongLogLevel(t *testing.T) {
	set := flag.NewFlagSet("", 0)
	set.String("log-level", "foo", "")
	set.String("log-format", "json", "")
	if err := Run(cli.NewContext(nil, set, nil)); err.ExitCode() != 1 {
		t.Fatal("Expected to get a non-zero exit code")
	}
}

func TestRunWrongLogType(t *testing.T) {
	set := flag.NewFlagSet("", 0)
	set.String("log-level", "debug", "")
	set.String("log-format", "foo", "")
	if err := Run(cli.NewContext(nil, set, nil)); err.ExitCode() != 1 {
		t.Fatal("Expected to get a non-zero exit code")
	}
}

func TestRunInvalidConfigFile(t *testing.T) {
	set := flag.NewFlagSet("", 0)
	set.String("log-level", "debug", "")
	set.String("log-format", "json", "")
	set.String("config", "path_does_not_exist", "")
	if err := Run(cli.NewContext(nil, set, nil)); err.ExitCode() != 1 {
		t.Fatal("Expected to get a non-zero exit code")
	}
}
