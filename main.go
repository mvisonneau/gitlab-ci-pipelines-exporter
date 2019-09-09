package main

import (
	"log"
	"os"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/cli"
)

var version = ""

func main() {
	if err := cli.Init(&version).Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
