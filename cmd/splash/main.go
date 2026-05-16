package main

import (
	"os"

	"github.com/flarebyte/ash-splash-norn/internal/runmain"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	os.Exit(runmain.Run(version, commit, date))
}
