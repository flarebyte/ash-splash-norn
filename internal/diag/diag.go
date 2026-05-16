package diag

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
)

// Severity is a diagnostic severity level.
type Severity string

const (
	SeverityError Severity = "error"
	SeverityWarn  Severity = "warn"
)

// Entry is a stable machine-readable diagnostic.
type Entry struct {
	Stage    string   `json:"stage"`
	ID       string   `json:"id"`
	Severity Severity `json:"severity"`
	Message  string   `json:"message"`
	Path     string   `json:"path,omitempty"`
}

func Sort(entries []Entry) {
	sort.Slice(entries, func(i, j int) bool {
		a := entries[i]
		b := entries[j]
		if a.Stage != b.Stage {
			return a.Stage < b.Stage
		}
		if a.ID != b.ID {
			return a.ID < b.ID
		}
		if a.Path != b.Path {
			return a.Path < b.Path
		}
		return a.Message < b.Message
	})
}

func Write(entries []Entry, format string, w io.Writer) error {
	Sort(entries)
	if format == "json" {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(entries)
	}
	for _, e := range entries {
		if e.Path != "" {
			if _, err := fmt.Fprintf(w, "%s %s [%s] %s (%s)\n", e.Severity, e.ID, e.Stage, e.Message, e.Path); err != nil {
				return err
			}
			continue
		}
		if _, err := fmt.Fprintf(w, "%s %s [%s] %s\n", e.Severity, e.ID, e.Stage, e.Message); err != nil {
			return err
		}
	}
	return nil
}
