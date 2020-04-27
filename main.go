package main

import (
	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/cli"
)

var version = ""

func main() {
	cli.Run(version)
}
