package cli

import (
	"encoding/json"
	"fmt"

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
	entries := engine.ValidateInputs(in)
	if len(entries) > 0 {
		_ = diag.Write(entries, format, r.Stderr)
		return fmt.Errorf("lint failed")
	}
	if format == "json" {
		_, _ = fmt.Fprintln(r.Stdout, `{"ok":true,"stage":"lint","checks":["schema-inputs"]}`)
		return nil
	}
	_, _ = fmt.Fprintln(r.Stdout, "lint: ok (schema-inputs)")
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
	type row struct {
		SchemaRef       string   `json:"schemaRef"`
		Target          string   `json:"target"`
		SupportsNode    []string `json:"supportsNodeKinds"`
		ArtifactPattern string   `json:"artifactPattern"`
	}
	preview := []row{
		{SchemaRef: "input-field", Target: "arb.json", SupportsNode: []string{"i18n"}, ArtifactPattern: "lib/l10n/app_<locale>.arb.json"},
		{SchemaRef: "input-field", Target: "cue", SupportsNode: []string{"i18n", "text", "validator"}, ArtifactPattern: "generated/config/<domain>.cue"},
		{SchemaRef: "input-field", Target: "dart", SupportsNode: []string{"text", "validator"}, ArtifactPattern: "lib/generated/<domain>_config.dart"},
		{SchemaRef: "input-field", Target: "go", SupportsNode: []string{"text", "validator"}, ArtifactPattern: "internal/generated/<domain>_config.go"},
		{SchemaRef: "input-field", Target: "json", SupportsNode: []string{"i18n", "text", "validator"}, ArtifactPattern: "generated/config/<domain>.json"},
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
