package main

import (
	"os"
)

var version = "<devel>"

func main() {
	runCli().Run(os.Args)
}
