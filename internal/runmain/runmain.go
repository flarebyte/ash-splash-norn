// purpose: Shares CLI runtime bootstrap logic so multiple binaries can execute the same command surface.
// responsibilities: Instantiate cli.Runner with build metadata and execute argv with consistent IO streams.
// architecture notes: This indirection intentionally avoids duplicated main wiring across compatibility entrypoints.

package runmain

import (
	"os"

	"github.com/flarebyte/ash-splash-norn/internal/cli"
)

func Run(version, commit, date string) int {
	r := cli.Runner{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Build: cli.BuildInfo{
			Version: version,
			Commit:  commit,
			Date:    date,
		},
	}
	return r.Run(os.Args[1:])
}
