package main

import (
	"os"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/cli"
)

var version = ""

func main() {
	cli.Init(&version).Run(os.Args)
}
