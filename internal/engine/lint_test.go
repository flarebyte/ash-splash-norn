package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flarebyte/ash-splash-norn/internal/app"
)

func TestLintDiagnosticsSectionsTranslationsKeysGraphPatternsSchemaConfig(t *testing.T) {
	in := app.Inputs{
		RegistryPath:       filepath.Clean("../../doc/design-meta/examples/input/design-registry.example.cue"),
		RegistrySchemaPath: filepath.Clean("../../doc/design-meta/examples/model/design-registry.schema.cue"),
		ConfigPath:         filepath.Clean("../../doc/design-meta/examples/input/config-key.cue"),
		ConfigSchemaPath:   filepath.Clean("../../doc/design-meta/examples/model/config-key.schema.cue"),
	}

	for _, check := range []string{"sections", "translations", "keys", "graph", "patterns", "schema", "config"} {
		entries := LintDiagnostics(in, check)
		if len(entries) != 0 {
			t.Fatalf("expected no diagnostics for %s, got %v", check, entries)
		}
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
    metadata: {
      id: string
      version: string
    }
    supportedLanguages: [...string]
    supportedCommandSections: [...string]
    translationPolicy: {
      requireAllSupportedLanguages: bool
    }
    rootLabels: [...string]
    nodesByLabel: [string]: {
      label: string
      kind: "branch" | "i18n" | "text" | "validator"
      childLabels: [...string]
    }
  }
  generatorCapabilities: [...{
    target: string
    artifactPattern: string
  }]
}
`), 0o644); err != nil {
		t.Fatalf("write reg schema: %v", err)
	}
	if err := os.WriteFile(regInput, []byte(`package designregistry
designRegistry: #DesignRegistrySpec & {
  keySchemaRegistry: {
    "input-field": {
      metadata: { id: "input-field", version: "1.0.0" }
      supportedLanguages: ["en", "fr"]
      supportedCommandSections: ["validation"]
      translationPolicy: { requireAllSupportedLanguages: true }
      rootLabels: ["fields"]
      nodesByLabel: {
        fields: { label: "fields", kind: "branch", childLabels: [] }
      }
    }
  }
  generatorCapabilities: [
    {target: "json", artifactPattern: "generated/config/<domain>.json"},
  ]
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
textEntries: [..._]
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
textEntries: []
validations: []
`), 0o644); err != nil {
		t.Fatalf("write cfg input: %v", err)
	}

	in := app.Inputs{RegistryPath: regInput, RegistrySchemaPath: regSchema, ConfigPath: cfgInput, ConfigSchemaPath: cfgSchema}
	entries := LintDiagnostics(in, "translations")
	if len(entries) == 0 {
		t.Fatalf("expected translation lint errors")
	}
	if entries[0].ID != "LNT-0002" {
		t.Fatalf("unexpected id: %s", entries[0].ID)
	}
}

func TestLintDiagnosticsGraphCycle(t *testing.T) {
	dir := t.TempDir()
	regSchema := filepath.Join(dir, "reg.schema.cue")
	regInput := filepath.Join(dir, "reg.cue")
	cfgSchema := filepath.Join(dir, "cfg.schema.cue")
	cfgInput := filepath.Join(dir, "cfg.cue")

	if err := os.WriteFile(regSchema, []byte(`package designregistry
#DesignRegistrySpec: {
  keySchemaRegistry: [string]: {
    metadata: {
      id: string
      version: string
    }
    supportedLanguages: [...string]
    supportedCommandSections: [...string]
    translationPolicy: {
      requireAllSupportedLanguages: bool
    }
    rootLabels: [...string]
    nodesByLabel: [string]: {
      label: string
      kind: "branch" | "i18n" | "text" | "validator"
      childLabels: [...string]
    }
  }
  generatorCapabilities: [...{
    target: string
    artifactPattern: string
  }]
}
`), 0o644); err != nil {
		t.Fatalf("write reg schema: %v", err)
	}
	if err := os.WriteFile(regInput, []byte(`package designregistry
designRegistry: #DesignRegistrySpec & {
  keySchemaRegistry: {
    "input-field": {
      metadata: { id: "input-field", version: "1.0.0" }
      supportedLanguages: ["en"]
      supportedCommandSections: ["validation"]
      translationPolicy: { requireAllSupportedLanguages: false }
      rootLabels: ["a"]
      nodesByLabel: {
        a: { label: "a", kind: "branch", childLabels: ["b"] }
        b: { label: "b", kind: "branch", childLabels: ["a"] }
      }
    }
  }
  generatorCapabilities: [
    {target: "json", artifactPattern: "generated/config/<domain>.json"},
  ]
}
`), 0o644); err != nil {
		t.Fatalf("write reg input: %v", err)
	}
	if err := os.WriteFile(cfgSchema, []byte(`package configkey
i18nEntries: [..._]
textEntries: [..._]
validations: [..._]
`), 0o644); err != nil {
		t.Fatalf("write cfg schema: %v", err)
	}
	if err := os.WriteFile(cfgInput, []byte(`i18nEntries: []
textEntries: []
validations: []
`), 0o644); err != nil {
		t.Fatalf("write cfg input: %v", err)
	}

	in := app.Inputs{RegistryPath: regInput, RegistrySchemaPath: regSchema, ConfigPath: cfgInput, ConfigSchemaPath: cfgSchema}
	entries := LintDiagnostics(in, "graph")
	if len(entries) == 0 {
		t.Fatalf("expected graph lint errors")
	}
	if entries[0].ID != "LNT-0006" {
		t.Fatalf("unexpected first id: %s", entries[0].ID)
	}
}

func TestLintDiagnosticsPatternPolicy(t *testing.T) {
	dir := t.TempDir()
	regSchema := filepath.Join(dir, "reg.schema.cue")
	regInput := filepath.Join(dir, "reg.cue")
	cfgSchema := filepath.Join(dir, "cfg.schema.cue")
	cfgInput := filepath.Join(dir, "cfg.cue")

	if err := os.WriteFile(regSchema, []byte(`package designregistry
#DesignRegistrySpec: {
  keySchemaRegistry: [string]: {
    metadata: {
      id: string
      version: string
    }
    supportedLanguages: [...string]
    supportedCommandSections: [...string]
    translationPolicy: {
      requireAllSupportedLanguages: bool
    }
    rootLabels: [...string]
    nodesByLabel: [string]: {
      label: string
      kind: "branch" | "i18n" | "text" | "validator"
      childLabels: [...string]
    }
  }
  generatorCapabilities: [...{
    target: string
    artifactPattern: string
  }]
}
`), 0o644); err != nil {
		t.Fatalf("write reg schema: %v", err)
	}
	if err := os.WriteFile(regInput, []byte(`package designregistry
designRegistry: #DesignRegistrySpec & {
  keySchemaRegistry: {
    "input-field": {
      metadata: { id: "input-field", version: "1.0.0" }
      supportedLanguages: ["en"]
      supportedCommandSections: ["validation"]
      translationPolicy: { requireAllSupportedLanguages: false }
      rootLabels: ["fields"]
      nodesByLabel: {
        fields: { label: "fields", kind: "branch", childLabels: [] }
      }
    }
  }
  generatorCapabilities: [
    {target: "arb.json", artifactPattern: "lib/l10n/app.arb.json"},
    {target: "json", artifactPattern: "generated/config/<tenant>.json"},
  ]
}
`), 0o644); err != nil {
		t.Fatalf("write reg input: %v", err)
	}
	if err := os.WriteFile(cfgSchema, []byte(`package configkey
i18nEntries: [..._]
textEntries: [..._]
validations: [..._]
`), 0o644); err != nil {
		t.Fatalf("write cfg schema: %v", err)
	}
	if err := os.WriteFile(cfgInput, []byte(`i18nEntries: []
textEntries: []
validations: []
`), 0o644); err != nil {
		t.Fatalf("write cfg input: %v", err)
	}

	in := app.Inputs{RegistryPath: regInput, RegistrySchemaPath: regSchema, ConfigPath: cfgInput, ConfigSchemaPath: cfgSchema}
	entries := LintDiagnostics(in, "patterns")
	if len(entries) == 0 {
		t.Fatalf("expected pattern lint errors")
	}
}

func TestLintDiagnosticsSchemaMismatch(t *testing.T) {
	dir := t.TempDir()
	regSchema := filepath.Join(dir, "reg.schema.cue")
	regInput := filepath.Join(dir, "reg.cue")
	cfgSchema := filepath.Join(dir, "cfg.schema.cue")
	cfgInput := filepath.Join(dir, "cfg.cue")

	if err := os.WriteFile(regSchema, []byte(`package designregistry
#DesignRegistrySpec: {
  keySchemaRegistry: [string]: {
    metadata: {
      id: string
      version: string
    }
    supportedLanguages: [...string]
    supportedCommandSections: [...string]
    translationPolicy: {
      requireAllSupportedLanguages: bool
    }
    rootLabels: [...string]
    nodesByLabel: [string]: {
      label: string
      kind: "branch" | "i18n" | "text" | "validator"
      childLabels: [...string]
    }
  }
  generatorCapabilities: [...{
    target: string
    artifactPattern: string
  }]
}
`), 0o644); err != nil {
		t.Fatalf("write reg schema: %v", err)
	}
	if err := os.WriteFile(regInput, []byte(`package designregistry
designRegistry: #DesignRegistrySpec & {
  keySchemaRegistry: {
    "input-field": {
      metadata: { id: "different-id", version: "1.0.0" }
      supportedLanguages: ["en"]
      supportedCommandSections: ["validation"]
      translationPolicy: { requireAllSupportedLanguages: false }
      rootLabels: ["fields"]
      nodesByLabel: {
        fields: { label: "fields", kind: "branch", childLabels: [] }
      }
    }
  }
  generatorCapabilities: [
    {target: "json", artifactPattern: "generated/config/<domain>.json"},
  ]
}
`), 0o644); err != nil {
		t.Fatalf("write reg input: %v", err)
	}
	if err := os.WriteFile(cfgSchema, []byte(`package configkey
i18nEntries: [..._]
textEntries: [..._]
validations: [..._]
`), 0o644); err != nil {
		t.Fatalf("write cfg schema: %v", err)
	}
	if err := os.WriteFile(cfgInput, []byte(`i18nEntries: []
textEntries: []
validations: []
`), 0o644); err != nil {
		t.Fatalf("write cfg input: %v", err)
	}

	in := app.Inputs{RegistryPath: regInput, RegistrySchemaPath: regSchema, ConfigPath: cfgInput, ConfigSchemaPath: cfgSchema}
	entries := LintDiagnostics(in, "schema")
	if len(entries) == 0 {
		t.Fatalf("expected schema lint errors")
	}
}

func TestLintDiagnosticsConfigDuplicateKeys(t *testing.T) {
	dir := t.TempDir()
	regSchema := filepath.Join(dir, "reg.schema.cue")
	regInput := filepath.Join(dir, "reg.cue")
	cfgSchema := filepath.Join(dir, "cfg.schema.cue")
	cfgInput := filepath.Join(dir, "cfg.cue")

	if err := os.WriteFile(regSchema, []byte(`package designregistry
#DesignRegistrySpec: {
  keySchemaRegistry: [string]: {
    metadata: {
      id: string
      version: string
    }
    supportedLanguages: [...string]
    supportedCommandSections: [...string]
    translationPolicy: {
      requireAllSupportedLanguages: bool
    }
    rootLabels: [...string]
    nodesByLabel: [string]: {
      label: string
      kind: "branch" | "i18n" | "text" | "validator"
      childLabels: [...string]
    }
  }
  generatorCapabilities: [...{
    target: string
    artifactPattern: string
  }]
}
`), 0o644); err != nil {
		t.Fatalf("write reg schema: %v", err)
	}
	if err := os.WriteFile(regInput, []byte(`package designregistry
designRegistry: #DesignRegistrySpec & {
  keySchemaRegistry: {
    "input-field": {
      metadata: { id: "input-field", version: "1.0.0" }
      supportedLanguages: ["en"]
      supportedCommandSections: ["validation"]
      translationPolicy: { requireAllSupportedLanguages: false }
      rootLabels: ["fields"]
      nodesByLabel: {
        fields: { label: "fields", kind: "branch", childLabels: [] }
      }
    }
  }
  generatorCapabilities: [
    {target: "json", artifactPattern: "generated/config/<domain>.json"},
  ]
}
`), 0o644); err != nil {
		t.Fatalf("write reg input: %v", err)
	}
	if err := os.WriteFile(cfgSchema, []byte(`package configkey
i18nEntries: [...{ key: string, translations: [string]: { text: string } }]
textEntries: [...{ key: string }]
validations: [...{ key: string, commands: [..._] }]
`), 0o644); err != nil {
		t.Fatalf("write cfg schema: %v", err)
	}
	if err := os.WriteFile(cfgInput, []byte(`i18nEntries: [
  { key: "dupKey", translations: { en: { text: "A" } } },
  { key: "dupKey", translations: { en: { text: "B" } } },
]
textEntries: []
validations: []
`), 0o644); err != nil {
		t.Fatalf("write cfg input: %v", err)
	}

	in := app.Inputs{RegistryPath: regInput, RegistrySchemaPath: regSchema, ConfigPath: cfgInput, ConfigSchemaPath: cfgSchema}
	entries := LintDiagnostics(in, "config")
	if len(entries) == 0 {
		t.Fatalf("expected config lint errors")
	}
}
