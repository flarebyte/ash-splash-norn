package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flarebyte/ash-splash-norn/internal/app"
)

func TestValidateCuePairsWithFixtures(t *testing.T) {
	in := app.Inputs{
		RegistryPath:       filepath.Clean("../../doc/design-meta/examples/input/design-registry.example.cue"),
		RegistrySchemaPath: filepath.Clean("../../doc/design-meta/examples/model/design-registry.schema.cue"),
		ConfigPath:         filepath.Clean("../../doc/design-meta/examples/input/config-key.cue"),
		ConfigSchemaPath:   filepath.Clean("../../doc/design-meta/examples/model/config-key.schema.cue"),
	}
	entries := ValidateCuePairs(in)
	if len(entries) != 0 {
		t.Fatalf("expected no schema errors, got %v", entries)
	}
}

func TestValidateCuePairsInvalidType(t *testing.T) {
	dir := t.TempDir()
	schema := filepath.Join(dir, "schema.cue")
	input := filepath.Join(dir, "input.cue")
	if err := os.WriteFile(schema, []byte("package p\nvalue: string\n"), 0o644); err != nil {
		t.Fatalf("write schema: %v", err)
	}
	if err := os.WriteFile(input, []byte("value: 42\n"), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	entries := validateCuePair(schema, input, "config")
	if len(entries) == 0 {
		t.Fatal("expected cue validation error")
	}
	if entries[0].ID != "SCH-0104" && entries[0].ID != "SCH-0105" {
		t.Fatalf("unexpected id: %s", entries[0].ID)
	}
}
