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

func runExplainKey(t *testing.T, format string) (int, string, string) {
	t.Helper()
	var out bytes.Buffer
	var stderr bytes.Buffer
	r := Runner{Stdout: &out, Stderr: &stderr}
	args := []string{
		"explain-key",
		"--registry", "../../doc/design-meta/examples/input/design-registry.example.cue",
		"--registry-schema", "../../doc/design-meta/examples/model/design-registry.schema.cue",
		"--config", "../../doc/design-meta/examples/input/config-key.cue",
		"--config-schema", "../../doc/design-meta/examples/model/config-key.schema.cue",
		"--schema-id", "input-field",
		"--label-path", "fields,textInput,label",
	}
	if format != "" {
		args = append(args, "--format", format)
	}
	code := r.Run(args)
	return code, out.String(), stderr.String()
}

func TestRunVersionJSON(t *testing.T) {
	var out bytes.Buffer
	var err bytes.Buffer
	r := Runner{Stdout: &out, Stderr: &err, Build: BuildInfo{Version: "1.2.3", Commit: "abc", Date: "today"}}
	code := r.Run([]string{"version", "--format", "json"})
	if code != 0 {
		t.Fatalf("expected exit 0 got %d stderr=%s", code, err.String())
	}
	var payload map[string]string
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json: %v output=%s", err, out.String())
	}
	if payload["version"] != "1.2.3" || payload["commitId"] != "abc" || payload["date"] != "today" {
		t.Fatalf("unexpected payload: %#v", payload)
	}
	for _, key := range []string{"os", "arch", "goVersion"} {
		if payload[key] == "" {
			t.Fatalf("missing %s in payload: %#v", key, payload)
		}
	}
}

func TestRunVersionText(t *testing.T) {
	var out bytes.Buffer
	var err bytes.Buffer
	r := Runner{Stdout: &out, Stderr: &err, Build: BuildInfo{Version: "1.2.3", Commit: "abc", Date: "today"}}
	code := r.Run([]string{"version"})
	if code != 0 {
		t.Fatalf("expected exit 0 got %d stderr=%s", code, err.String())
	}
	txt := out.String()
	for _, tok := range []string{"version=1.2.3", "commitId=abc", "date=today", "os=", "arch=", "goVersion="} {
		if !strings.Contains(txt, tok) {
			t.Fatalf("missing token %q in output: %s", tok, txt)
		}
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

func TestRunValidateJSONDiagnosticsDeterministic(t *testing.T) {
	var out1 bytes.Buffer
	var err1 bytes.Buffer
	r1 := Runner{Stdout: &out1, Stderr: &err1}
	code1 := r1.Run([]string{"validate", "--format", "json"})
	if code1 == 0 {
		t.Fatalf("expected non-zero code")
	}
	var out2 bytes.Buffer
	var err2 bytes.Buffer
	r2 := Runner{Stdout: &out2, Stderr: &err2}
	code2 := r2.Run([]string{"validate", "--format", "json"})
	if code2 == 0 {
		t.Fatalf("expected non-zero code")
	}
	if err1.String() != err2.String() {
		t.Fatalf("expected deterministic diagnostics json")
	}
	if !strings.Contains(err1.String(), "\"id\": \"SCH-0001\"") {
		t.Fatalf("expected schema diagnostics: %s", err1.String())
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

func TestRunListJSON(t *testing.T) {
	var out bytes.Buffer
	var stderr bytes.Buffer
	r := Runner{Stdout: &out, Stderr: &stderr}
	args := append([]string{"list"}, fixtureArgs()...)
	args = append(args, "--format", "json")
	code := r.Run(args)
	if code != 0 {
		t.Fatalf("expected exit 0 got %d stderr=%s", code, stderr.String())
	}
	var rows []map[string]any
	if err := json.Unmarshal(out.Bytes(), &rows); err != nil {
		t.Fatalf("invalid json: %v output=%s", err, out.String())
	}
	if len(rows) == 0 {
		t.Fatalf("expected non-empty rows")
	}
}

func TestRunListTable(t *testing.T) {
	var out bytes.Buffer
	var stderr bytes.Buffer
	r := Runner{Stdout: &out, Stderr: &stderr}
	args := append([]string{"list"}, fixtureArgs()...)
	args = append(args, "--format", "table")
	code := r.Run(args)
	if code != 0 {
		t.Fatalf("expected exit 0 got %d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(out.String(), "SCHEMA") || !strings.Contains(out.String(), "ARTIFACT PATTERN") {
		t.Fatalf("unexpected table output: %s", out.String())
	}
}

func TestRunExplainKey(t *testing.T) {
	code, out, stderr := runExplainKey(t, "")
	if code != 0 {
		t.Fatalf("expected exit 0 got %d stderr=%s", code, stderr)
	}
	if !strings.Contains(out, "derived-key: fieldsTextInputLabel") {
		t.Fatalf("unexpected explain-key output: %s", out)
	}
}

func TestRunExplainKeyJSON(t *testing.T) {
	code, out, stderr := runExplainKey(t, "json")
	if code != 0 {
		t.Fatalf("expected exit 0 got %d stderr=%s", code, stderr)
	}
	var payload struct {
		SchemaID string `json:"schemaId"`
		Key      string `json:"key"`
		Steps    []struct {
			Label string `json:"label"`
			Kind  string `json:"kind"`
		} `json:"steps"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("invalid json: %v output=%s", err, out)
	}
	if payload.SchemaID != "input-field" || payload.Key != "fieldsTextInputLabel" || len(payload.Steps) != 3 {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}
