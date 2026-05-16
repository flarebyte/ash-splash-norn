package engine

import (
	"bytes"
	"encoding/json"
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

func TestGenerateArtifactsDeterministicBytes(t *testing.T) {
	in := app.Inputs{
		RegistryPath:       filepath.Clean("../../doc/design-meta/examples/input/design-registry.example.cue"),
		RegistrySchemaPath: filepath.Clean("../../doc/design-meta/examples/model/design-registry.schema.cue"),
		ConfigPath:         filepath.Clean("../../doc/design-meta/examples/input/config-key.cue"),
		ConfigSchemaPath:   filepath.Clean("../../doc/design-meta/examples/model/config-key.schema.cue"),
	}
	for _, target := range []string{"json", "yaml", "cue"} {
		dir1 := t.TempDir()
		dir2 := t.TempDir()
		arts1, entries1 := GenerateArtifacts(in, target, dir1)
		arts2, entries2 := GenerateArtifacts(in, target, dir2)
		if len(entries1) > 0 || len(entries2) > 0 {
			t.Fatalf("target=%s entries1=%v entries2=%v", target, entries1, entries2)
		}
		if len(arts1) != len(arts2) {
			t.Fatalf("target=%s artifact count mismatch", target)
		}
		for i := range arts1 {
			b1, err := os.ReadFile(arts1[i].Path)
			if err != nil {
				t.Fatalf("read run1 artifact: %v", err)
			}
			b2, err := os.ReadFile(arts2[i].Path)
			if err != nil {
				t.Fatalf("read run2 artifact: %v", err)
			}
			if !bytes.Equal(b1, b2) {
				t.Fatalf("target=%s artifact bytes differ across runs", target)
			}
		}
	}
}

func TestGenerateArtifactsJSONContractFields(t *testing.T) {
	in := app.Inputs{
		RegistryPath:       filepath.Clean("../../doc/design-meta/examples/input/design-registry.example.cue"),
		RegistrySchemaPath: filepath.Clean("../../doc/design-meta/examples/model/design-registry.schema.cue"),
		ConfigPath:         filepath.Clean("../../doc/design-meta/examples/input/config-key.cue"),
		ConfigSchemaPath:   filepath.Clean("../../doc/design-meta/examples/model/config-key.schema.cue"),
	}
	outDir := t.TempDir()
	arts, entries := GenerateArtifacts(in, "json", outDir)
	if len(entries) > 0 {
		t.Fatalf("entries=%v", entries)
	}
	if len(arts) == 0 {
		t.Fatalf("expected artifacts")
	}
	raw, err := os.ReadFile(arts[0].Path)
	if err != nil {
		t.Fatalf("read artifact: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal json: %v", err)
	}
	if payload["schemaRef"] == nil || payload["target"] != "json" || payload["generatedAt"] == nil {
		t.Fatalf("missing envelope fields in payload: %v", payload)
	}
}

func TestGenerateArtifactsUnknownTarget(t *testing.T) {
	in := app.Inputs{}
	_, entries := GenerateArtifacts(in, "go", t.TempDir())
	if len(entries) == 0 || entries[0].ID != "GEN-0002" {
		t.Fatalf("expected GEN-0002, got %v", entries)
	}
}
