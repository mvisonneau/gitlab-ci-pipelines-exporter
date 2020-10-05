package main

import (
	"os"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/internal/cli"
)

var version = ""

func main() {
	cli.Run(version, os.Args)
}
