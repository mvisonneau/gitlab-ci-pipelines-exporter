package cmd

import (
	"flag"
	"fmt"
	"strings"

	"testing"

	"github.com/urfave/cli"
)

func TestRunWrongLogLevel(t *testing.T) {
	set := flag.NewFlagSet("", 0)
	set.String("log-level", "foo", "")
	set.String("log-format", "json", "")
	fmt.Println(Run(cli.NewContext(nil, set, nil)))
	if err := Run(cli.NewContext(nil, set, nil)); !strings.HasPrefix(err.Error(), "not a valid logrus Level") {
		t.Fatalf("Unexpected error : %v", err)
	}
}

func TestRunWrongLogType(t *testing.T) {
	set := flag.NewFlagSet("", 0)
	set.String("log-level", "debug", "")
	set.String("log-format", "foo", "")
	if err := Run(cli.NewContext(nil, set, nil)); !strings.HasPrefix(err.Error(), "Invalid log format") {
		t.Fatalf("Unexpected error : %v", err)
	}
}

func TestRunInvalidConfigFile(t *testing.T) {
	set := flag.NewFlagSet("", 0)
	set.String("log-level", "debug", "")
	set.String("log-format", "json", "")
	set.String("config", "path_does_not_exist", "")
	if err := Run(cli.NewContext(nil, set, nil)); !strings.HasPrefix(err.Error(), "Couldn't open config file :") {
		t.Fatalf("Unexpected error : %v", err)
	}
}
