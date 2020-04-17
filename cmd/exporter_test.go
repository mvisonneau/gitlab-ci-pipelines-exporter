package cmd

import (
	"flag"
	"fmt"
	"strings"

	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestRunWrongLogLevel(t *testing.T) {
	set := flag.NewFlagSet("", 0)
	set.String("log-level", "foo", "")
	set.String("log-format", "json", "")
	fmt.Println(Run(cli.NewContext(nil, set, nil)))
	err := Run(cli.NewContext(nil, set, nil))
	assert.Equal(t, true, strings.HasPrefix(err.Error(), "not a valid logrus Level"))
}

func TestRunWrongLogType(t *testing.T) {
	set := flag.NewFlagSet("", 0)
	set.String("log-level", "debug", "")
	set.String("log-format", "foo", "")
	err := Run(cli.NewContext(nil, set, nil))
	assert.Equal(t, true, strings.HasPrefix(err.Error(), "Invalid log format"))
}

func TestRunInvalidConfigFile(t *testing.T) {
	set := flag.NewFlagSet("", 0)
	set.String("log-level", "debug", "")
	set.String("log-format", "json", "")
	set.String("config", "path_does_not_exist", "")
	err := Run(cli.NewContext(nil, set, nil))
	assert.Equal(t, true, strings.HasPrefix(err.Error(), "couldn't open config file :"))
}
