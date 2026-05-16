package diag

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestSortStableOrder(t *testing.T) {
	entries := []Entry{
		{Stage: "lint", ID: "LNT-2", Message: "b"},
		{Stage: "config", ID: "CFG-1", Message: "a"},
	}
	Sort(entries)
	if entries[0].Stage != "config" {
		t.Fatalf("unexpected first stage: %s", entries[0].Stage)
	}
}

func TestWriteJSON(t *testing.T) {
	var buf bytes.Buffer
	err := Write([]Entry{{Stage: "schema", ID: "SCH-1", Severity: SeverityError, Message: "m"}}, "json", &buf)
	if err != nil {
		t.Fatalf("write json: %v", err)
	}
	if !strings.Contains(buf.String(), "\"SCH-1\"") {
		t.Fatalf("missing id in output: %s", buf.String())
	}
}

func TestWriteJSONIncludesLocationFromPath(t *testing.T) {
	var buf bytes.Buffer
	err := Write([]Entry{{Stage: "schema", ID: "SCH-1", Severity: SeverityError, Message: "m", Path: "a.cue"}}, "json", &buf)
	if err != nil {
		t.Fatalf("write json: %v", err)
	}
	var entries []Entry
	if err := json.Unmarshal(buf.Bytes(), &entries); err != nil {
		t.Fatalf("unmarshal json: %v", err)
	}
	if len(entries) != 1 || entries[0].Location != "a.cue" {
		t.Fatalf("expected location from path, got %+v", entries)
	}
}

func TestWriteJSONDeterministic(t *testing.T) {
	in := []Entry{
		{Stage: "schema", ID: "SCH-2", Severity: SeverityError, Message: "b", Path: "b"},
		{Stage: "schema", ID: "SCH-1", Severity: SeverityError, Message: "a", Path: "a"},
	}
	var b1 bytes.Buffer
	var b2 bytes.Buffer
	if err := Write(append([]Entry(nil), in...), "json", &b1); err != nil {
		t.Fatalf("write1: %v", err)
	}
	if err := Write(append([]Entry(nil), in...), "json", &b2); err != nil {
		t.Fatalf("write2: %v", err)
	}
	if !reflect.DeepEqual(b1.String(), b2.String()) {
		t.Fatalf("expected deterministic output")
	}
}
