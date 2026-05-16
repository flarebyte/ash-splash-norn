package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
)

var ErrUsage = errors.New("usage")

type Runner struct {
	Stdout io.Writer
	Stderr io.Writer
	Build  BuildInfo
}

type BuildInfo struct {
	Version string
	Commit  string
	Date    string
}

func (r Runner) Run(args []string) int {
	if len(args) == 0 {
		r.printRootHelp()
		return 1
	}
	cmd := args[0]
	rest := args[1:]
	var err error
	switch cmd {
	case "validate":
		err = r.runValidate(rest)
	case "lint":
		err = r.runLint(rest)
	case "dry-run-preview":
		err = r.runDryRunPreview(rest)
	case "version":
		err = r.runVersion(rest)
	case "help", "--help", "-h":
		r.printRootHelp()
		return 0
	default:
		fmt.Fprintf(r.Stderr, "unknown command: %s\n\n", cmd)
		r.printRootHelp()
		return 1
	}
	if err == nil {
		return 0
	}
	if errors.Is(err, ErrUsage) {
		return 2
	}
	fmt.Fprintf(r.Stderr, "%v\n", err)
	return 1
}

func newFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	return fs
}

func (r Runner) printRootHelp() {
	fmt.Fprintln(r.Stdout, "Usage: flyb <command> [flags]")
	fmt.Fprintln(r.Stdout, "")
	fmt.Fprintln(r.Stdout, "Commands:")
	fmt.Fprintln(r.Stdout, "  validate         Validate registry/config inputs")
	fmt.Fprintln(r.Stdout, "  lint             Run aggregate lint checks (scaffold)")
	fmt.Fprintln(r.Stdout, "  dry-run-preview  Preview planned artifacts")
	fmt.Fprintln(r.Stdout, "  version          Print build metadata")
}
