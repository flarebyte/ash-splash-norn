package cli

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"
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
	root := r.newRootCommand()
	root.SetArgs(args)
	if err := root.Execute(); err != nil {
		if errors.Is(err, ErrUsage) {
			return 2
		}
		_, _ = fmt.Fprintf(r.Stderr, "%v\n", err)
		return 1
	}
	return 0
}

func (r Runner) newRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:           "splash",
		Short:         "Distributed config CLI",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.SetOut(r.Stdout)
	root.SetErr(r.Stderr)

	root.AddCommand(r.newValidateCommand())
	root.AddCommand(r.newLintCommand())
	root.AddCommand(r.newListCommand())
	root.AddCommand(r.newExplainKeyCommand())
	root.AddCommand(r.newDryRunPreviewCommand())
	root.AddCommand(r.newGenerateCommand())
	root.AddCommand(r.newVersionCommand())

	return root
}
