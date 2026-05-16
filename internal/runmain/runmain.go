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
