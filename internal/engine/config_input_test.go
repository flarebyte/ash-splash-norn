package engine

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadConfigInputSourceFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.cue")
	if err := os.WriteFile(path, []byte("textEntries: []\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	src, entries := readConfigInputSource(path, "config")
	if len(entries) > 0 {
		t.Fatalf("expected no entries, got %v", entries)
	}
	if src == "" {
		t.Fatal("expected non-empty source")
	}
}

func TestReadConfigInputSourceDirectoryNoCue(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "README.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	_, entries := readConfigInputSource(dir, "config")
	if len(entries) == 0 || entries[0].ID != "CFG-0004" {
		t.Fatalf("expected CFG-0004, got %v", entries)
	}
}

func TestReadConfigInputSourceDirectoryMixedPackage(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.cue"), []byte("package foo\n\na: 1\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.cue"), []byte("package bar\n\nb: 2\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	_, entries := readConfigInputSource(dir, "config")
	if len(entries) == 0 || entries[0].ID != "CFG-0006" {
		t.Fatalf("expected CFG-0006, got %v", entries)
	}
}
