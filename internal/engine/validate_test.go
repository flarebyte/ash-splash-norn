package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flarebyte/ash-splash-norn/internal/app"
)

func TestValidateInputsMissingPath(t *testing.T) {
	entries := ValidateInputs(app.Inputs{})
	if len(entries) == 0 {
		t.Fatal("expected errors")
	}
}

func TestValidateInputsOK(t *testing.T) {
	dir := t.TempDir()
	mk := func(name, content string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatalf("write file: %v", err)
		}
		return p
	}
	registrySchema := `package designregistry
#DesignRegistrySpec: {
  keySchemaRegistry: [string]: _
  generatorCapabilities: [..._]
}
`
	registryInput := `package designregistry
designRegistry: #DesignRegistrySpec & {
  keySchemaRegistry: {}
  generatorCapabilities: []
}
`
	configSchema := `package configkey
textEntries: [...{
  key: string
}]
`
	configInput := `textEntries: [{key: "ok"}]
`
	entries := ValidateInputs(app.Inputs{
		RegistryPath:       mk("registry.cue", registryInput),
		RegistrySchemaPath: mk("registry.schema.cue", registrySchema),
		ConfigPath:         mk("config.cue", configInput),
		ConfigSchemaPath:   mk("config.schema.cue", configSchema),
	})
	if len(entries) != 0 {
		t.Fatalf("expected no errors, got %d", len(entries))
	}
}

func TestValidateInputsConfigDirNoCue(t *testing.T) {
	dir := t.TempDir()
	cfgDir := filepath.Join(dir, "cfg")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("mkdir cfg: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "README.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write marker: %v", err)
	}
	entries := ValidateInputs(app.Inputs{
		RegistryPath:       filepath.Clean("../../doc/design-meta/examples/input/design-registry.example.cue"),
		RegistrySchemaPath: filepath.Clean("../../doc/design-meta/examples/model/design-registry.schema.cue"),
		ConfigPath:         cfgDir,
		ConfigSchemaPath:   filepath.Clean("../../doc/design-meta/examples/model/config-key.schema.cue"),
	})
	if len(entries) == 0 || entries[0].ID != "SCH-0005" {
		t.Fatalf("expected SCH-0005, got %v", entries)
	}
}

func TestValidateInputsConfigDirMixedPackage(t *testing.T) {
	dir := t.TempDir()
	cfgDir := filepath.Join(dir, "cfg")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("mkdir cfg: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "a.cue"), []byte("package a\nx: 1\n"), 0o644); err != nil {
		t.Fatalf("write a.cue: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "b.cue"), []byte("package b\ny: 2\n"), 0o644); err != nil {
		t.Fatalf("write b.cue: %v", err)
	}
	entries := ValidateInputs(app.Inputs{
		RegistryPath:       filepath.Clean("../../doc/design-meta/examples/input/design-registry.example.cue"),
		RegistrySchemaPath: filepath.Clean("../../doc/design-meta/examples/model/design-registry.schema.cue"),
		ConfigPath:         cfgDir,
		ConfigSchemaPath:   filepath.Clean("../../doc/design-meta/examples/model/config-key.schema.cue"),
	})
	if len(entries) == 0 || entries[0].ID != "SCH-0004" {
		t.Fatalf("expected SCH-0004, got %v", entries)
	}
}

func TestValidateInputsRejectsUnsafeArtifactPattern(t *testing.T) {
	dir := t.TempDir()
	regSchema := filepath.Join(dir, "reg.schema.cue")
	regInput := filepath.Join(dir, "reg.cue")
	cfgSchema := filepath.Join(dir, "cfg.schema.cue")
	cfgInput := filepath.Join(dir, "cfg.cue")
	if err := os.WriteFile(regSchema, []byte(`package designregistry
#DesignRegistrySpec: {
  keySchemaRegistry: [string]: _
  generatorCapabilities: [...{ keySchema: string, target: string, supportsNodeKinds: [...string], artifactPattern: string }]
}
`), 0o644); err != nil {
		t.Fatalf("write reg schema: %v", err)
	}
	if err := os.WriteFile(regInput, []byte(`package designregistry
designRegistry: #DesignRegistrySpec & {
  keySchemaRegistry: {"x": {}}
  generatorCapabilities: [{ keySchema: "x", target: "json", supportsNodeKinds: ["text"], artifactPattern: "../bad/<domain>.json" }]
}
`), 0o644); err != nil {
		t.Fatalf("write reg input: %v", err)
	}
	if err := os.WriteFile(cfgSchema, []byte(`package configkey
textEntries: [...{ key: string }]
`), 0o644); err != nil {
		t.Fatalf("write cfg schema: %v", err)
	}
	if err := os.WriteFile(cfgInput, []byte(`textEntries: [{key: "ok"}]
`), 0o644); err != nil {
		t.Fatalf("write cfg input: %v", err)
	}
	entries := ValidateInputs(app.Inputs{
		RegistryPath:       regInput,
		RegistrySchemaPath: regSchema,
		ConfigPath:         cfgInput,
		ConfigSchemaPath:   cfgSchema,
	})
	found := false
	for _, e := range entries {
		if e.ID == "SCH-0008" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected SCH-0008, got %v", entries)
	}
}
