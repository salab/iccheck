package main

import "github.com/salab/iccheck/cmd"

var (
	version  = "dev"
	revision = "local"
)

func main() {
	cmd.Execute(version, revision)
}
