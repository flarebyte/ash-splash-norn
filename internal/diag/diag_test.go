package diag

import (
	"bytes"
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
