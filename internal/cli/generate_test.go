package cli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunGenerateJSON(t *testing.T) {
	var out bytes.Buffer
	var stderr bytes.Buffer
	r := Runner{Stdout: &out, Stderr: &stderr}
	dir := t.TempDir()
	code := r.Run([]string{
		"generate", "json",
		"--registry", "../../doc/design-meta/examples/input/design-registry.example.cue",
		"--registry-schema", "../../doc/design-meta/examples/model/design-registry.schema.cue",
		"--config", "../../doc/design-meta/examples/input/config-key.cue",
		"--config-schema", "../../doc/design-meta/examples/model/config-key.schema.cue",
		"--output-root", dir,
	})
	if code != 0 {
		t.Fatalf("expected exit 0 got %d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(out.String(), "generated:") {
		t.Fatalf("unexpected output: %s", out.String())
	}
	wantPath := filepath.Join(dir, "generated/config/input-field.json")
	if !strings.Contains(out.String(), wantPath) {
		t.Fatalf("expected generated path in output, got %s", out.String())
	}
}
