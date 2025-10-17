package main

import (
	"github.com/codila125/foedus-blockchain/cli"
	"os"
)

func main() {
	defer os.Exit(0)
	cli := cli.CommandLine{}
	cli.Run()
}
