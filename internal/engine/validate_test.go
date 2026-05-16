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
	mk := func(name string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte("ok"), 0o644); err != nil {
			t.Fatalf("write file: %v", err)
		}
		return p
	}
	entries := ValidateInputs(app.Inputs{
		RegistryPath:       mk("registry.cue"),
		RegistrySchemaPath: mk("registry.schema.cue"),
		ConfigPath:         mk("config.cue"),
		ConfigSchemaPath:   mk("config.schema.cue"),
	})
	if len(entries) != 0 {
		t.Fatalf("expected no errors, got %d", len(entries))
	}
}
