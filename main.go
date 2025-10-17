package main

import (
	"github.com/codila125/foedus/cli"
	"os"
)

func main() {
	defer os.Exit(0)
	cli := cli.CommandLine{}
	cli.Run()
}
