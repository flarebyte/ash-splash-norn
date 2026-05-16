package cli

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func fixtureArgs() []string {
	return []string{
		"--registry", "../../doc/design-meta/examples/input/design-registry.example.cue",
		"--registry-schema", "../../doc/design-meta/examples/model/design-registry.schema.cue",
		"--config", "../../doc/design-meta/examples/input/config-key.cue",
		"--config-schema", "../../doc/design-meta/examples/model/config-key.schema.cue",
	}
}

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

func TestRunLintAggregateJSONContract(t *testing.T) {
	var out bytes.Buffer
	var stderr bytes.Buffer
	r := Runner{Stdout: &out, Stderr: &stderr}
	args := append([]string{"lint"}, fixtureArgs()...)
	args = append(args, "--format", "json")
	code := r.Run(args)
	if code != 0 {
		t.Fatalf("expected exit 0 got %d stderr=%s", code, stderr.String())
	}
	var payload struct {
		OK     bool     `json:"ok"`
		Stage  string   `json:"stage"`
		Checks []string `json:"checks"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json: %v output=%s", err, out.String())
	}
	wantChecks := []string{"schema", "config", "sections", "keys", "translations", "patterns", "graph"}
	if !payload.OK || payload.Stage != "lint" || !reflect.DeepEqual(payload.Checks, wantChecks) {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}

func TestRunDryRunPreviewJSONDeterministicOrder(t *testing.T) {
	var out bytes.Buffer
	var stderr bytes.Buffer
	r := Runner{Stdout: &out, Stderr: &stderr}
	args := append([]string{"dry-run-preview"}, fixtureArgs()...)
	args = append(args, "--format", "json")
	code := r.Run(args)
	if code != 0 {
		t.Fatalf("expected exit 0 got %d stderr=%s", code, stderr.String())
	}
	var rows []struct {
		Target          string `json:"target"`
		ArtifactPattern string `json:"artifactPattern"`
		SchemaRef       string `json:"schemaRef"`
	}
	if err := json.Unmarshal(out.Bytes(), &rows); err != nil {
		t.Fatalf("invalid json: %v output=%s", err, out.String())
	}
	if len(rows) != 5 {
		t.Fatalf("expected 5 rows, got %d", len(rows))
	}
	wantTargets := []string{"arb.json", "cue", "dart", "go", "json"}
	for i, want := range wantTargets {
		if rows[i].Target != want {
			t.Fatalf("row %d target mismatch: got=%s want=%s", i, rows[i].Target, want)
		}
	}
}
