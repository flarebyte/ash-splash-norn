package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flarebyte/ash-splash-norn/internal/app"
)

func TestLintDiagnosticsFixtureHappyPath(t *testing.T) {
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

func TestLintDiagnosticsConfigDirectoryPath(t *testing.T) {
	base := fixtureInputs()
	in, err := fixtureInputsWithConfigDirFromFile(t, base.ConfigPath)
	if err != nil {
		t.Fatalf("read fixture config: %v", err)
	}

	entries := LintDiagnostics(in, "config")
	if len(entries) > 0 {
		t.Fatalf("expected no diagnostics, got %v", entries)
	}
}

func TestLintDiagnosticsGraphCycle(t *testing.T) {
	dir := t.TempDir()
	regSchema := filepath.Join(dir, "reg.schema.cue")
	regInput := filepath.Join(dir, "reg.cue")
	cfgSchema := filepath.Join(dir, "cfg.schema.cue")
	cfgInput := filepath.Join(dir, "cfg.cue")

	mustWrite(t, regSchema, `package designregistry
#DesignRegistrySpec: {
  keySchemaRegistry: [string]: {
    metadata: { id: string, version: string }
    supportedLanguages: [...string]
    supportedCommandSections: [...string]
    translationPolicy: { requireAllSupportedLanguages: bool }
    metaArgsValidation: {
      args: [string]: {
        commandPath: [...string]
        adminOnly: bool
        flags: [...{ kind: string, name: string, schema: [...string], schemas?: [...[...string]] }]
      }
    }
    rootLabels: [...string]
    nodesByLabel: [string]: {
      label: string
      kind: "branch" | "i18n" | "text" | "validator"
      mandatory?: bool
      childLabels: [...string]
    }
  }
  generatorCapabilities: [...{ target: string, artifactPattern: string }]
}
`)
	mustWrite(t, regInput, `package designregistry
designRegistry: #DesignRegistrySpec & {
  keySchemaRegistry: {
    "input-field": {
      metadata: { id: "input-field", version: "1.0.0" }
      supportedLanguages: ["en"]
      supportedCommandSections: ["validation"]
      translationPolicy: { requireAllSupportedLanguages: false }
      metaArgsValidation: {
        args: {
          validation: {
            commandPath: ["meta"]
            adminOnly: false
            flags: []
          }
        }
      }
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
`)
	mustWrite(t, cfgSchema, `package configkey
i18nEntries: [..._]
textEntries: [..._]
validations: [..._]
`)
	mustWrite(t, cfgInput, `i18nEntries: []
textEntries: []
validations: []
`)

	in := app.Inputs{RegistryPath: regInput, RegistrySchemaPath: regSchema, ConfigPath: cfgInput, ConfigSchemaPath: cfgSchema}
	entries := LintDiagnostics(in, "graph")
	if len(entries) == 0 {
		t.Fatalf("expected graph lint errors")
	}
}

func TestLintDiagnosticsConfigInvalidMetaArgs(t *testing.T) {
	dir := t.TempDir()
	regSchema := filepath.Join(dir, "reg.schema.cue")
	regInput := filepath.Join(dir, "reg.cue")
	cfgSchema := filepath.Join(dir, "cfg.schema.cue")
	cfgInput := filepath.Join(dir, "cfg.cue")

	mustWrite(t, regSchema, `package designregistry
#DesignRegistrySpec: {
  keySchemaRegistry: [string]: {
    metadata: { id: string, version: string }
    supportedLanguages: [...string]
    supportedCommandSections: [...string]
    translationPolicy: { requireAllSupportedLanguages: bool }
    metaArgsValidation: {
      args: [string]: {
        commandPath: [...string]
        adminOnly: bool
        flags: [...{ kind: string, name: string, schema: [...string], schemas?: [...[...string]] }]
      }
    }
    rootLabels: [...string]
    nodesByLabel: [string]: {
      label: string
      kind: "branch" | "i18n" | "text" | "validator"
      mandatory?: bool
      childLabels: [...string]
    }
  }
  generatorCapabilities: [...{ target: string, artifactPattern: string }]
}
`)
	mustWrite(t, regInput, `package designregistry
designRegistry: #DesignRegistrySpec & {
  keySchemaRegistry: {
    "input-field": {
      metadata: { id: "input-field", version: "1.0.0" }
      supportedLanguages: ["en"]
      supportedCommandSections: ["validation"]
      translationPolicy: { requireAllSupportedLanguages: false }
      metaArgsValidation: {
        args: {
          validation: {
            commandPath: ["meta"]
            adminOnly: false
            flags: [
              {kind: "string", name: "status", schema: ["schema", "string", "--enum", "draft,stable,experimental", "--required"]},
              {kind: "string", name: "app", schema: ["schema", "string", "--enum", "v1,v2", "--required"]},
            ]
          }
        }
      }
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
`)
	mustWrite(t, cfgSchema, `package configkey
i18nEntries: [...{ key: string, metaArgs: [...string], translations: [string]: { text: string } }]
textEntries: [..._]
validations: [..._]
`)
	mustWrite(t, cfgInput, `i18nEntries: [
  {
    key: "fieldsTextInputLabel"
    metaArgs: ["meta", "--bad", "x"]
    translations: { en: { text: "A" } }
  },
]
textEntries: []
validations: []
`)

	in := app.Inputs{RegistryPath: regInput, RegistrySchemaPath: regSchema, ConfigPath: cfgInput, ConfigSchemaPath: cfgSchema}
	entries := LintDiagnostics(in, "config")
	if len(entries) == 0 {
		t.Fatalf("expected config lint errors")
	}
	found := false
	for _, e := range entries {
		if e.ID == "LNT-0020" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected LNT-0020 in diagnostics, got %v", entries)
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
