package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flarebyte/ash-splash-norn/internal/app"
)

func TestGenerateArtifactsJSONYAMLCUE(t *testing.T) {
	dir := t.TempDir()
	in := app.Inputs{
		RegistryPath:       filepath.Clean("../../doc/design-meta/examples/input/design-registry.example.cue"),
		RegistrySchemaPath: filepath.Clean("../../doc/design-meta/examples/model/design-registry.schema.cue"),
		ConfigPath:         filepath.Clean("../../doc/design-meta/examples/input/config-key.cue"),
		ConfigSchemaPath:   filepath.Clean("../../doc/design-meta/examples/model/config-key.schema.cue"),
	}
	for _, target := range []string{"json", "yaml", "cue"} {
		outDir := filepath.Join(dir, target)
		arts, entries := GenerateArtifacts(in, target, outDir)
		if len(entries) > 0 {
			t.Fatalf("target=%s entries=%v", target, entries)
		}
		if len(arts) == 0 {
			t.Fatalf("expected artifacts for %s", target)
		}
		if _, err := os.Stat(arts[0].Path); err != nil {
			t.Fatalf("artifact missing for %s: %v", target, err)
		}
	}
}

func TestGenerateArtifactsUnknownTarget(t *testing.T) {
	in := app.Inputs{}
	_, entries := GenerateArtifacts(in, "go", t.TempDir())
	if len(entries) == 0 || entries[0].ID != "GEN-0002" {
		t.Fatalf("expected GEN-0002, got %v", entries)
	}
}
