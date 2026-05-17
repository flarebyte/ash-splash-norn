// purpose: Provides a compatibility binary entrypoint expected by external build tooling.
// responsibilities: Expose build metadata variables and delegate process execution to shared runtime bootstrap.
// architecture notes: Compatibility main intentionally mirrors splash behavior while preserving legacy build path contracts.

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
