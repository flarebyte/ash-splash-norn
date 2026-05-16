package engine

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/flarebyte/ash-splash-norn/internal/app"
)

func fixtureInputs() app.Inputs {
	return app.Inputs{
		RegistryPath:       filepath.Clean("../../doc/design-meta/examples/input/design-registry.example.cue"),
		RegistrySchemaPath: filepath.Clean("../../doc/design-meta/examples/model/design-registry.schema.cue"),
		ConfigPath:         filepath.Clean("../../doc/design-meta/examples/input/config-key.cue"),
		ConfigSchemaPath:   filepath.Clean("../../doc/design-meta/examples/model/config-key.schema.cue"),
	}
}

func TestGenerateArtifactsJSONYAMLCUE(t *testing.T) {
	dir := t.TempDir()
	in := fixtureInputs()
	for _, target := range []string{"json", "yaml", "cue"} {
		outDir := filepath.Join(dir, target)
		arts, entries := GenerateArtifacts(in, target, outDir)
		if len(entries) > 0 {
			t.Fatalf("target=%s entries=%v", target, entries)
		}
		if len(arts) == 0 {
			t.Fatalf("expected artifacts for %s", target)
		}
		for _, a := range arts {
			if _, err := os.Stat(a.Path); err != nil {
				t.Fatalf("artifact missing for %s: %v", target, err)
			}
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

func TestGenerateArtifactsDeterministicBytes(t *testing.T) {
	in := fixtureInputs()
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
	in := fixtureInputs()
	outDir := t.TempDir()
	arts, entries := GenerateArtifacts(in, "json", outDir)
	if len(entries) > 0 {
		t.Fatalf("entries=%v", entries)
	}
	if len(arts) == 0 {
		t.Fatalf("expected artifacts")
	}

	paths := make([]string, 0, len(arts))
	for _, a := range arts {
		paths = append(paths, a.Path)
	}
	sort.Strings(paths)

	for _, p := range paths {
		raw, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read artifact: %v", err)
		}
		var payload map[string]any
		if err := json.Unmarshal(raw, &payload); err != nil {
			t.Fatalf("unmarshal json: %v", err)
		}
		if payload["schemaRef"] == nil || payload["target"] != "json" || payload["generatedAt"] == nil || payload["nodeKind"] == nil {
			t.Fatalf("missing envelope fields in payload: %v", payload)
		}
	}
}

func TestGenerateArtifactsPerKindFileNames(t *testing.T) {
	in := fixtureInputs()
	outDir := t.TempDir()
	arts, entries := GenerateArtifacts(in, "json", outDir)
	if len(entries) > 0 {
		t.Fatalf("entries=%v", entries)
	}
	if len(arts) != 3 {
		t.Fatalf("expected 3 json artifacts (i18n,text,validator), got %d", len(arts))
	}
	joined := "\n"
	for _, a := range arts {
		joined += a.Path + "\n"
	}
	for _, suffix := range []string{".i18n.json", ".text.json", ".validator.json"} {
		if !strings.Contains(joined, suffix) {
			t.Fatalf("missing expected suffix %s in artifacts:\n%s", suffix, joined)
		}
	}
}

func TestGenerateArtifactsJSONGoldenSnapshots(t *testing.T) {
	in := fixtureInputs()
	outDir := t.TempDir()
	arts, entries := GenerateArtifacts(in, "json", outDir)
	if len(entries) > 0 {
		t.Fatalf("entries=%v", entries)
	}
	if len(arts) != 3 {
		t.Fatalf("expected 3 artifacts, got %d", len(arts))
	}
	for _, a := range arts {
		got, err := os.ReadFile(a.Path)
		if err != nil {
			t.Fatalf("read generated file: %v", err)
		}
		base := filepath.Base(a.Path)
		wantPath := filepath.Join("testdata", "golden", base)
		want, err := os.ReadFile(wantPath)
		if err != nil {
			t.Fatalf("read golden file %s: %v", wantPath, err)
		}
		got = bytes.TrimRight(got, "\n")
		want = bytes.TrimRight(want, "\n")
		if !bytes.Equal(got, want) {
			t.Fatalf("golden mismatch for %s", base)
		}
	}
}

func TestGenerateArtifactsYAMLGoldenSnapshots(t *testing.T) {
	in := fixtureInputs()
	outDir := t.TempDir()
	arts, entries := GenerateArtifacts(in, "yaml", outDir)
	if len(entries) > 0 {
		t.Fatalf("entries=%v", entries)
	}
	if len(arts) != 3 {
		t.Fatalf("expected 3 artifacts, got %d", len(arts))
	}
	for _, a := range arts {
		got, err := os.ReadFile(a.Path)
		if err != nil {
			t.Fatalf("read generated file: %v", err)
		}
		base := filepath.Base(a.Path)
		wantPath := filepath.Join("testdata", "golden", base)
		want, err := os.ReadFile(wantPath)
		if err != nil {
			t.Fatalf("read golden file %s: %v", wantPath, err)
		}
		got = bytes.TrimRight(got, "\n")
		want = bytes.TrimRight(want, "\n")
		if !bytes.Equal(got, want) {
			t.Fatalf("golden mismatch for %s", base)
		}
	}
}

func TestGenerateArtifactsCUEGoldenSnapshots(t *testing.T) {
	in := fixtureInputs()
	outDir := t.TempDir()
	arts, entries := GenerateArtifacts(in, "cue", outDir)
	if len(entries) > 0 {
		t.Fatalf("entries=%v", entries)
	}
	if len(arts) != 3 {
		t.Fatalf("expected 3 artifacts, got %d", len(arts))
	}
	for _, a := range arts {
		got, err := os.ReadFile(a.Path)
		if err != nil {
			t.Fatalf("read generated file: %v", err)
		}
		base := filepath.Base(a.Path)
		wantPath := filepath.Join("testdata", "golden", base)
		want, err := os.ReadFile(wantPath)
		if err != nil {
			t.Fatalf("read golden file %s: %v", wantPath, err)
		}
		got = bytes.TrimRight(got, "\n")
		want = bytes.TrimRight(want, "\n")
		if !bytes.Equal(got, want) {
			t.Fatalf("golden mismatch for %s", base)
		}
	}
}
