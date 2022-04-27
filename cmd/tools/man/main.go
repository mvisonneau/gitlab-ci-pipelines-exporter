package main

import (
	"fmt"
	"time"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/internal/cli"
)

var version = "devel"

func main() {
	fmt.Println(cli.NewApp(version, time.Now()).ToMan())
}
