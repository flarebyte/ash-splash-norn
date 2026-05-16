package engine

import (
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
