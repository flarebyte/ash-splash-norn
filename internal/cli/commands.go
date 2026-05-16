package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/flarebyte/ash-splash-norn/internal/app"
	"github.com/flarebyte/ash-splash-norn/internal/diag"
	"github.com/flarebyte/ash-splash-norn/internal/engine"
)

func (r Runner) runValidate(args []string) error {
	fs := newFlagSet("validate")
	var in app.Inputs
	var format string
	fs.StringVar(&in.RegistryPath, "registry", "", "registry cue path")
	fs.StringVar(&in.RegistrySchemaPath, "registry-schema", "", "registry schema cue path")
	fs.StringVar(&in.ConfigPath, "config", "", "config cue path")
	fs.StringVar(&in.ConfigSchemaPath, "config-schema", "", "config schema cue path")
	fs.StringVar(&format, "format", "text", "output format: text|json")
	if err := fs.Parse(args); err != nil {
		return ErrUsage
	}
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
}

func (r Runner) runLint(args []string) error {
	lintSubcommand := ""
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		lintSubcommand = args[0]
		args = args[1:]
	}

	fs := newFlagSet("lint")
	var in app.Inputs
	var format string
	fs.StringVar(&in.RegistryPath, "registry", "", "registry cue path")
	fs.StringVar(&in.ConfigPath, "config", "", "config cue path")
	fs.StringVar(&in.RegistrySchemaPath, "registry-schema", "", "registry schema cue path")
	fs.StringVar(&in.ConfigSchemaPath, "config-schema", "", "config schema cue path")
	fs.StringVar(&format, "format", "text", "output format: text|json")
	if err := fs.Parse(args); err != nil {
		return ErrUsage
	}
	if lintSubcommand != "" {
		switch lintSubcommand {
		case "schema", "config", "sections", "keys", "translations", "patterns", "graph":
		default:
			return fmt.Errorf("lint failed: unknown lint subcommand: %s", lintSubcommand)
		}
	}
	entries := engine.ValidateInputs(in)
	if len(entries) > 0 {
		_ = diag.Write(entries, format, r.Stderr)
		return fmt.Errorf("lint failed")
	}
	lintEntries := engine.LintDiagnostics(in, lintSubcommand)
	if len(lintEntries) > 0 {
		_ = diag.Write(lintEntries, format, r.Stderr)
		return fmt.Errorf("lint failed")
	}
	if format == "json" {
		if lintSubcommand == "" {
			_, _ = fmt.Fprintln(r.Stdout, `{"ok":true,"stage":"lint","checks":["schema","config","sections","keys","translations","patterns","graph"]}`)
			return nil
		}
		_, _ = fmt.Fprintf(r.Stdout, "{\"ok\":true,\"stage\":\"lint\",\"checks\":[%q]}\n", lintSubcommand)
		return nil
	}
	if lintSubcommand == "" {
		_, _ = fmt.Fprintln(r.Stdout, "lint: ok (schema, config, sections, keys, translations, patterns, graph)")
		return nil
	}
	_, _ = fmt.Fprintf(r.Stdout, "lint %s: ok\n", lintSubcommand)
	return nil
}

func (r Runner) runDryRunPreview(args []string) error {
	fs := newFlagSet("dry-run-preview")
	var in app.Inputs
	var format string
	fs.StringVar(&in.RegistryPath, "registry", "", "registry cue path")
	fs.StringVar(&in.ConfigPath, "config", "", "config cue path")
	fs.StringVar(&in.RegistrySchemaPath, "registry-schema", "", "registry schema cue path")
	fs.StringVar(&in.ConfigSchemaPath, "config-schema", "", "config schema cue path")
	fs.StringVar(&format, "format", "text", "output format: text|json")
	if err := fs.Parse(args); err != nil {
		return ErrUsage
	}
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
}

func (r Runner) runVersion(args []string) error {
	fs := newFlagSet("version")
	var format string
	fs.StringVar(&format, "format", "text", "output format: text|json")
	if err := fs.Parse(args); err != nil {
		return ErrUsage
	}
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
}
