package main

import (
	"os"

	"github.com/flarebyte/ash-splash-norn/internal/cli"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	r := cli.Runner{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Build: cli.BuildInfo{
			Version: version,
			Commit:  commit,
			Date:    date,
		},
	}
	os.Exit(r.Run(os.Args[1:]))
}
