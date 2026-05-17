// purpose: Provides the primary splash CLI binary entrypoint for end users.
// responsibilities: Expose build metadata variables and delegate process execution to shared runtime bootstrap.
// architecture notes: Main stays thin by design so binary identity can vary without forking command behavior.

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
