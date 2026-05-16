package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunVersionJSON(t *testing.T) {
	var out bytes.Buffer
	var err bytes.Buffer
	r := Runner{Stdout: &out, Stderr: &err, Build: BuildInfo{Version: "1.2.3", Commit: "abc", Date: "today"}}
	code := r.Run([]string{"version", "--format", "json"})
	if code != 0 {
		t.Fatalf("expected exit 0 got %d stderr=%s", code, err.String())
	}
	if !strings.Contains(out.String(), "1.2.3") {
		t.Fatalf("missing version: %s", out.String())
	}
}

func TestRunValidateMissingFlags(t *testing.T) {
	var out bytes.Buffer
	var stderr bytes.Buffer
	r := Runner{Stdout: &out, Stderr: &stderr}
	code := r.Run([]string{"validate"})
	if code == 0 {
		t.Fatalf("expected non-zero code")
	}
	if !strings.Contains(stderr.String(), "SCH-0001") {
		t.Fatalf("expected diagnostic id, got %s", stderr.String())
	}
}

func TestRunLintSubcommandJSON(t *testing.T) {
	var out bytes.Buffer
	var stderr bytes.Buffer
	r := Runner{Stdout: &out, Stderr: &stderr}
	code := r.Run([]string{
		"lint", "translations",
		"--registry", "../../doc/design-meta/examples/input/design-registry.example.cue",
		"--registry-schema", "../../doc/design-meta/examples/model/design-registry.schema.cue",
		"--config", "../../doc/design-meta/examples/input/config-key.cue",
		"--config-schema", "../../doc/design-meta/examples/model/config-key.schema.cue",
		"--format", "json",
	})
	if code != 0 {
		t.Fatalf("expected exit 0 got %d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(out.String(), "\"translations\"") {
		t.Fatalf("missing subcommand check in output: %s", out.String())
	}
}
