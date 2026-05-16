package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flarebyte/ash-splash-norn/internal/app"
)

func TestLintDiagnosticsSectionsAndTranslations(t *testing.T) {
	in := app.Inputs{
		RegistryPath:       filepath.Clean("../../doc/design-meta/examples/input/design-registry.example.cue"),
		RegistrySchemaPath: filepath.Clean("../../doc/design-meta/examples/model/design-registry.schema.cue"),
		ConfigPath:         filepath.Clean("../../doc/design-meta/examples/input/config-key.cue"),
		ConfigSchemaPath:   filepath.Clean("../../doc/design-meta/examples/model/config-key.schema.cue"),
	}

	sections := LintDiagnostics(in, "sections")
	if len(sections) != 0 {
		t.Fatalf("expected no section diagnostics, got %v", sections)
	}

	translations := LintDiagnostics(in, "translations")
	if len(translations) != 0 {
		t.Fatalf("expected no translation diagnostics, got %v", translations)
	}
}

func TestLintDiagnosticsTranslationsMissingLanguage(t *testing.T) {
	dir := t.TempDir()
	regSchema := filepath.Join(dir, "reg.schema.cue")
	regInput := filepath.Join(dir, "reg.cue")
	cfgSchema := filepath.Join(dir, "cfg.schema.cue")
	cfgInput := filepath.Join(dir, "cfg.cue")

	if err := os.WriteFile(regSchema, []byte(`package designregistry
#DesignRegistrySpec: {
  keySchemaRegistry: [string]: {
    supportedLanguages: [...string]
    supportedCommandSections: [...string]
    translationPolicy: {
      requireAllSupportedLanguages: bool
    }
  }
  generatorCapabilities: [..._]
}
`), 0o644); err != nil {
		t.Fatalf("write reg schema: %v", err)
	}
	if err := os.WriteFile(regInput, []byte(`package designregistry
designRegistry: #DesignRegistrySpec & {
  keySchemaRegistry: {
    "input-field": {
      supportedLanguages: ["en", "fr"]
      supportedCommandSections: ["validation"]
      translationPolicy: { requireAllSupportedLanguages: true }
    }
  }
  generatorCapabilities: []
}
`), 0o644); err != nil {
		t.Fatalf("write reg input: %v", err)
	}
	if err := os.WriteFile(cfgSchema, []byte(`package configkey
i18nEntries: [...{
  key: string
  translations: [string]: {
    text: string
  }
}]
validations: [..._]
`), 0o644); err != nil {
		t.Fatalf("write cfg schema: %v", err)
	}
	if err := os.WriteFile(cfgInput, []byte(`i18nEntries: [{
  key: "fooBar"
  translations: {
    en: { text: "Hello" }
  }
}]
validations: []
`), 0o644); err != nil {
		t.Fatalf("write cfg input: %v", err)
	}

	in := app.Inputs{
		RegistryPath:       regInput,
		RegistrySchemaPath: regSchema,
		ConfigPath:         cfgInput,
		ConfigSchemaPath:   cfgSchema,
	}
	entries := LintDiagnostics(in, "translations")
	if len(entries) == 0 {
		t.Fatalf("expected translation lint errors")
	}
	if entries[0].ID != "LNT-0002" {
		t.Fatalf("unexpected id: %s", entries[0].ID)
	}
}
