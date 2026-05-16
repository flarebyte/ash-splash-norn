package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flarebyte/ash-splash-norn/internal/app"
)

func TestBuildPreviewFromRegistryFixture(t *testing.T) {
	in := app.Inputs{
		RegistryPath:       filepath.Clean("../../doc/design-meta/examples/input/design-registry.example.cue"),
		RegistrySchemaPath: filepath.Clean("../../doc/design-meta/examples/model/design-registry.schema.cue"),
	}
	rows, entries := BuildPreview(in)
	if len(entries) > 0 {
		t.Fatalf("expected no entries, got %v", entries)
	}
	if len(rows) != 5 {
		t.Fatalf("expected 5 rows, got %d", len(rows))
	}
	if rows[0].Target != "arb.json" {
		t.Fatalf("expected first target arb.json, got %s", rows[0].Target)
	}
}

func TestBuildPreviewRejectsUnsafePattern(t *testing.T) {
	dir := t.TempDir()
	schema := filepath.Join(dir, "reg.schema.cue")
	input := filepath.Join(dir, "reg.cue")
	if err := os.WriteFile(schema, []byte(`package designregistry
#DesignRegistrySpec: {
  keySchemaRegistry: [string]: _
  generatorCapabilities: [...{ keySchema: string, target: string, supportsNodeKinds: [...string], artifactPattern: string }]
}
`), 0o644); err != nil {
		t.Fatalf("write schema: %v", err)
	}
	if err := os.WriteFile(input, []byte(`package designregistry
designRegistry: #DesignRegistrySpec & {
  keySchemaRegistry: {"x": {}}
  generatorCapabilities: [{ keySchema: "x", target: "json", supportsNodeKinds: ["text"], artifactPattern: "../oops/<domain>.json" }]
}
`), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}
	_, entries := BuildPreview(app.Inputs{RegistryPath: input, RegistrySchemaPath: schema})
	if len(entries) == 0 || entries[0].ID != "PRV-0006" {
		t.Fatalf("expected PRV-0006, got %v", entries)
	}
}
