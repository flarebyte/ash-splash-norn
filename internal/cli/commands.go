package cli

import (
	"encoding/json"
	"fmt"

	"github.com/flarebyte/ash-splash-norn/internal/app"
	"github.com/flarebyte/ash-splash-norn/internal/diag"
	"github.com/flarebyte/ash-splash-norn/internal/engine"
	"github.com/spf13/cobra"
)

func addInputFlags(cmd *cobra.Command, in *app.Inputs, format *string) {
	cmd.Flags().StringVar(&in.RegistryPath, "registry", "", "registry cue path")
	cmd.Flags().StringVar(&in.RegistrySchemaPath, "registry-schema", "", "registry schema cue path")
	cmd.Flags().StringVar(&in.ConfigPath, "config", "", "config cue path")
	cmd.Flags().StringVar(&in.ConfigSchemaPath, "config-schema", "", "config schema cue path")
	cmd.Flags().StringVar(format, "format", "text", "output format: text|json")
}

func (r Runner) newValidateCommand() *cobra.Command {
	var in app.Inputs
	format := "text"
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate registry/config inputs",
		RunE: func(cmd *cobra.Command, args []string) error {
			entries := engine.ValidateInputs(in)
			if len(entries) > 0 {
				_ = diag.Write(entries, format, r.Stderr)
				return fmt.Errorf("validate failed")
			}
			if format == "json" {
				_, _ = fmt.Fprintln(r.Stdout, `{"ok":true,"stage":"validate"}`)
				return nil
			}
			_, _ = fmt.Fprintln(r.Stdout, "validate: ok")
			return nil
		},
	}
	addInputFlags(cmd, &in, &format)
	return cmd
}

func (r Runner) newLintCommand() *cobra.Command {
	lint := &cobra.Command{
		Use:   "lint",
		Short: "Run lint checks",
		RunE:  nil,
	}

	// Aggregate lint (no subcommand)
	{
		var in app.Inputs
		format := "text"
		addInputFlags(lint, &in, &format)
		lint.RunE = func(cmd *cobra.Command, args []string) error {
			entries := engine.ValidateInputs(in)
			if len(entries) > 0 {
				_ = diag.Write(entries, format, r.Stderr)
				return fmt.Errorf("lint failed")
			}
			lintEntries := engine.LintDiagnostics(in, "")
			if len(lintEntries) > 0 {
				_ = diag.Write(lintEntries, format, r.Stderr)
				return fmt.Errorf("lint failed")
			}
			if format == "json" {
				_, _ = fmt.Fprintln(r.Stdout, `{"ok":true,"stage":"lint","checks":["schema","config","sections","keys","translations","patterns","graph"]}`)
				return nil
			}
			_, _ = fmt.Fprintln(r.Stdout, "lint: ok (schema, config, sections, keys, translations, patterns, graph)")
			return nil
		}
	}

	for _, sub := range []string{"schema", "config", "sections", "keys", "translations", "patterns", "graph"} {
		check := sub
		var in app.Inputs
		format := "text"
		subCmd := &cobra.Command{
			Use:   sub,
			Short: fmt.Sprintf("Run lint %s checks", sub),
			RunE: func(cmd *cobra.Command, args []string) error {
				entries := engine.ValidateInputs(in)
				if len(entries) > 0 {
					_ = diag.Write(entries, format, r.Stderr)
					return fmt.Errorf("lint failed")
				}
				lintEntries := engine.LintDiagnostics(in, check)
				if len(lintEntries) > 0 {
					_ = diag.Write(lintEntries, format, r.Stderr)
					return fmt.Errorf("lint failed")
				}
				if format == "json" {
					_, _ = fmt.Fprintf(r.Stdout, "{\"ok\":true,\"stage\":\"lint\",\"checks\":[%q]}\n", check)
					return nil
				}
				_, _ = fmt.Fprintf(r.Stdout, "lint %s: ok\n", check)
				return nil
			},
		}
		addInputFlags(subCmd, &in, &format)
		lint.AddCommand(subCmd)
	}

	return lint
}

func (r Runner) newDryRunPreviewCommand() *cobra.Command {
	var in app.Inputs
	format := "text"
	cmd := &cobra.Command{
		Use:   "dry-run-preview",
		Short: "Preview planned artifacts",
		RunE: func(cmd *cobra.Command, args []string) error {
			entries := engine.ValidateInputs(in)
			if len(entries) > 0 {
				_ = diag.Write(entries, format, r.Stderr)
				return fmt.Errorf("dry-run-preview failed")
			}
			preview, prvEntries := engine.BuildPreview(in)
			if len(prvEntries) > 0 {
				_ = diag.Write(prvEntries, format, r.Stderr)
				return fmt.Errorf("dry-run-preview failed")
			}
			if format == "json" {
				enc := json.NewEncoder(r.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(preview)
			}
			for _, p := range preview {
				_, _ = fmt.Fprintf(r.Stdout, "%s -> %s (%s)\n", p.Target, p.ArtifactPattern, p.SchemaRef)
			}
			return nil
		},
	}
	addInputFlags(cmd, &in, &format)
	return cmd
}

func (r Runner) newVersionCommand() *cobra.Command {
	format := "text"
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print build metadata",
		RunE: func(cmd *cobra.Command, args []string) error {
			if format == "json" {
				payload := map[string]string{
					"version": r.Build.Version,
					"commit":  r.Build.Commit,
					"date":    r.Build.Date,
				}
				enc := json.NewEncoder(r.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(payload)
			}
			_, _ = fmt.Fprintf(r.Stdout, "%s\n", r.Build.Version)
			return nil
		},
	}
	cmd.Flags().StringVar(&format, "format", "text", "output format: text|json")
	return cmd
}

func (r Runner) newListCommand() *cobra.Command {
	var in app.Inputs
	format := "text"
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List schemas and generation capabilities",
		RunE: func(cmd *cobra.Command, args []string) error {
			entries := engine.ValidateInputs(in)
			if len(entries) > 0 {
				_ = diag.Write(entries, format, r.Stderr)
				return fmt.Errorf("list failed")
			}
			rows, listEntries := engine.BuildCatalog(in)
			if len(listEntries) > 0 {
				_ = diag.Write(listEntries, format, r.Stderr)
				return fmt.Errorf("list failed")
			}
			if format == "json" {
				enc := json.NewEncoder(r.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(rows)
			}
			if format == "table" {
				_, _ = fmt.Fprintln(r.Stdout, "SCHEMA         TARGET    ARTIFACT PATTERN                          NODE KINDS")
				_, _ = fmt.Fprintln(r.Stdout, "-------------  --------  ----------------------------------------  ---------------------")
				for _, row := range rows {
					_, _ = fmt.Fprintf(r.Stdout, "%-13s  %-8s  %-40s  %v\n", row.SchemaID, row.Target, row.ArtifactPattern, row.SupportsNode)
				}
				return nil
			}
			if len(rows) == 0 {
				_, _ = fmt.Fprintln(r.Stdout, "list: no capabilities")
				return nil
			}
			for _, row := range rows {
				_, _ = fmt.Fprintf(r.Stdout, "%s | %s | %s | %s\n", row.SchemaID, row.Target, row.ArtifactPattern, row.SupportsNode)
			}
			return nil
		},
	}
	addInputFlags(cmd, &in, &format)
	return cmd
}

func (r Runner) newExplainKeyCommand() *cobra.Command {
	var in app.Inputs
	var schemaID string
	var labels []string
	format := "text"
	cmd := &cobra.Command{
		Use:   "explain-key",
		Short: "Explain canonical key derivation from a label path",
		RunE: func(cmd *cobra.Command, args []string) error {
			entries := engine.ValidateInputs(in)
			if len(entries) > 0 {
				_ = diag.Write(entries, "text", r.Stderr)
				return fmt.Errorf("explain-key failed")
			}
			trace, expEntries := engine.ExplainKeyTrace(in, schemaID, labels)
			if len(expEntries) > 0 {
				_ = diag.Write(expEntries, "text", r.Stderr)
				return fmt.Errorf("explain-key failed")
			}
			if format == "json" {
				enc := json.NewEncoder(r.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(trace)
			}
			_, _ = fmt.Fprintf(r.Stdout, "schema: %s\n", trace.SchemaID)
			for idx, step := range trace.Steps {
				_, _ = fmt.Fprintf(r.Stdout, "step %d: %s (%s)\n", idx+1, step.Label, step.Kind)
			}
			_, _ = fmt.Fprintf(r.Stdout, "derived-key: %s\n", trace.Key)
			return nil
		},
	}
	cmd.Flags().StringVar(&in.RegistryPath, "registry", "", "registry cue path")
	cmd.Flags().StringVar(&in.RegistrySchemaPath, "registry-schema", "", "registry schema cue path")
	cmd.Flags().StringVar(&in.ConfigPath, "config", "", "config cue path")
	cmd.Flags().StringVar(&in.ConfigSchemaPath, "config-schema", "", "config schema cue path")
	cmd.Flags().StringVar(&format, "format", "text", "output format: text|json")
	cmd.Flags().StringVar(&schemaID, "schema-id", "", "schema id to use")
	cmd.Flags().StringSliceVar(&labels, "label-path", nil, "ordered labels for key derivation")
	return cmd
}

func (r Runner) newGenerateCommand() *cobra.Command {
	var in app.Inputs
	outputRoot := "."
	cmd := &cobra.Command{
		Use:   "generate <target>",
		Short: "Generate artifacts by target",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := args[0]
			entries := engine.ValidateInputs(in)
			if len(entries) > 0 {
				_ = diag.Write(entries, "text", r.Stderr)
				return fmt.Errorf("generate failed")
			}
			arts, genEntries := engine.GenerateArtifacts(in, target, outputRoot)
			if len(genEntries) > 0 {
				_ = diag.Write(genEntries, "text", r.Stderr)
				return fmt.Errorf("generate failed")
			}
			for _, a := range arts {
				_, _ = fmt.Fprintf(r.Stdout, "generated: %s\n", a.Path)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&in.RegistryPath, "registry", "", "registry cue path")
	cmd.Flags().StringVar(&in.RegistrySchemaPath, "registry-schema", "", "registry schema cue path")
	cmd.Flags().StringVar(&in.ConfigPath, "config", "", "config cue path")
	cmd.Flags().StringVar(&in.ConfigSchemaPath, "config-schema", "", "config schema cue path")
	cmd.Flags().StringVar(&outputRoot, "output-root", ".", "output root directory")
	return cmd
}
