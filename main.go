package main

import (
	"github.com/salab/iccheck/cmd"
	"github.com/salab/iccheck/pkg/utils/cli"
)

var (
	version  = "dev"
	revision = "local"
)

func main() {
	cli.SetVersion(version, revision)
	cmd.Execute()
}
