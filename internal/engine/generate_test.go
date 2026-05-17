package engine

import (
	"bytes"
	"encoding/json"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/flarebyte/ash-splash-norn/internal/app"
)

func fixtureInputs() app.Inputs {
	return app.Inputs{
		RegistryPath:       filepath.Clean("../../doc/design-meta/examples/input/design-registry.example.cue"),
		RegistrySchemaPath: filepath.Clean("../../doc/design-meta/examples/model/design-registry.schema.cue"),
		ConfigPath:         filepath.Clean("../../doc/design-meta/examples/input/config-key.cue"),
		ConfigSchemaPath:   filepath.Clean("../../doc/design-meta/examples/model/config-key.schema.cue"),
	}
}

func TestGenerateArtifactsJSONYAMLCUE(t *testing.T) {
	dir := t.TempDir()
	in := fixtureInputs()
	for _, target := range []string{"json", "yaml", "cue"} {
		outDir := filepath.Join(dir, target)
		arts, entries := GenerateArtifacts(in, target, outDir)
		if len(entries) > 0 {
			t.Fatalf("target=%s entries=%v", target, entries)
		}
		if len(arts) == 0 {
			t.Fatalf("expected artifacts for %s", target)
		}
		for _, a := range arts {
			if _, err := os.Stat(a.Path); err != nil {
				t.Fatalf("artifact missing for %s: %v", target, err)
			}
		}
	}
}

func TestGenerateArtifactsConfigDirectoryPath(t *testing.T) {
	base := fixtureInputs()
	in, err := fixtureInputsWithConfigDirFromFile(t, base.ConfigPath)
	if err != nil {
		t.Fatalf("read fixture config: %v", err)
	}

	arts, entries := GenerateArtifacts(in, "json", t.TempDir())
	if len(entries) > 0 {
		t.Fatalf("entries=%v", entries)
	}
	if len(arts) == 0 {
		t.Fatal("expected artifacts")
	}
}

func assertGoldenSnapshotsForTarget(t *testing.T, in app.Inputs, target string, expectedCount int) {
	t.Helper()
	outDir := t.TempDir()
	arts, entries := GenerateArtifacts(in, target, outDir)
	if len(entries) > 0 {
		t.Fatalf("entries=%v", entries)
	}
	if len(arts) != expectedCount {
		t.Fatalf("expected %d artifacts, got %d", expectedCount, len(arts))
	}
	for _, a := range arts {
		got, err := os.ReadFile(a.Path)
		if err != nil {
			t.Fatalf("read generated file: %v", err)
		}
		base := filepath.Base(a.Path)
		wantPath := filepath.Join("testdata", "golden", base)
		want, err := os.ReadFile(wantPath)
		if err != nil {
			t.Fatalf("read golden file %s: %v", wantPath, err)
		}
		got = bytes.TrimRight(got, "\n")
		want = bytes.TrimRight(want, "\n")
		if !bytes.Equal(got, want) {
			t.Fatalf("golden mismatch for %s", base)
		}
	}
}

func TestGenerateArtifactsUnknownTarget(t *testing.T) {
	in := app.Inputs{}
	_, entries := GenerateArtifacts(in, "tsx", t.TempDir())
	if len(entries) == 0 || entries[0].ID != "GEN-0002" {
		t.Fatalf("expected GEN-0002, got %v", entries)
	}
}

func TestGenerateArtifactsDeterministicBytes(t *testing.T) {
	in := fixtureInputs()
	for _, target := range []string{"json", "yaml", "cue"} {
		dir1 := t.TempDir()
		dir2 := t.TempDir()
		arts1, entries1 := GenerateArtifacts(in, target, dir1)
		arts2, entries2 := GenerateArtifacts(in, target, dir2)
		if len(entries1) > 0 || len(entries2) > 0 {
			t.Fatalf("target=%s entries1=%v entries2=%v", target, entries1, entries2)
		}
		if len(arts1) != len(arts2) {
			t.Fatalf("target=%s artifact count mismatch", target)
		}
		for i := range arts1 {
			b1, err := os.ReadFile(arts1[i].Path)
			if err != nil {
				t.Fatalf("read run1 artifact: %v", err)
			}
			b2, err := os.ReadFile(arts2[i].Path)
			if err != nil {
				t.Fatalf("read run2 artifact: %v", err)
			}
			if !bytes.Equal(b1, b2) {
				t.Fatalf("target=%s artifact bytes differ across runs", target)
			}
		}
	}
}

func TestGenerateArtifactsJSONContractFields(t *testing.T) {
	in := fixtureInputs()
	outDir := t.TempDir()
	arts, entries := GenerateArtifacts(in, "json", outDir)
	if len(entries) > 0 {
		t.Fatalf("entries=%v", entries)
	}
	if len(arts) == 0 {
		t.Fatalf("expected artifacts")
	}

	paths := make([]string, 0, len(arts))
	for _, a := range arts {
		paths = append(paths, a.Path)
	}
	sort.Strings(paths)

	for _, p := range paths {
		raw, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read artifact: %v", err)
		}
		var payload map[string]any
		if err := json.Unmarshal(raw, &payload); err != nil {
			t.Fatalf("unmarshal json: %v", err)
		}
		if payload["schemaRef"] == nil || payload["target"] != "json" || payload["generatedAt"] == nil || payload["nodeKind"] == nil {
			t.Fatalf("missing envelope fields in payload: %v", payload)
		}
	}
}

func TestGenerateArtifactsPerKindFileNames(t *testing.T) {
	in := fixtureInputs()
	outDir := t.TempDir()
	arts, entries := GenerateArtifacts(in, "json", outDir)
	if len(entries) > 0 {
		t.Fatalf("entries=%v", entries)
	}
	if len(arts) != 3 {
		t.Fatalf("expected 3 json artifacts (i18n,text,validator), got %d", len(arts))
	}
	joined := "\n"
	for _, a := range arts {
		joined += a.Path + "\n"
	}
	for _, suffix := range []string{".i18n.json", ".text.json", ".validator.json"} {
		if !strings.Contains(joined, suffix) {
			t.Fatalf("missing expected suffix %s in artifacts:\n%s", suffix, joined)
		}
	}
}

func TestGenerateArtifactsJSONGoldenSnapshots(t *testing.T) {
	assertGoldenSnapshotsForTarget(t, fixtureInputs(), "json", 3)
}

func TestGenerateArtifactsYAMLGoldenSnapshots(t *testing.T) {
	assertGoldenSnapshotsForTarget(t, fixtureInputs(), "yaml", 3)
}

func TestGenerateArtifactsCUEGoldenSnapshots(t *testing.T) {
	assertGoldenSnapshotsForTarget(t, fixtureInputs(), "cue", 3)
}

func TestGenerateArtifactsGoAndDart(t *testing.T) {
	in := fixtureInputs()
	for _, target := range []string{"go", "dart"} {
		outDir := t.TempDir()
		arts, entries := GenerateArtifacts(in, target, outDir)
		if len(entries) > 0 {
			t.Fatalf("target=%s entries=%v", target, entries)
		}
		if len(arts) != 1 {
			t.Fatalf("target=%s expected 1 artifact, got %d", target, len(arts))
		}
		raw, err := os.ReadFile(arts[0].Path)
		if err != nil {
			t.Fatalf("target=%s read artifact: %v", target, err)
		}
		content := string(raw)
		if target == "go" {
			if !strings.Contains(content, "// Code generated by splash. DO NOT EDIT.") {
				t.Fatalf("missing go generated header")
			}
			if !strings.Contains(content, "package generatedconfig") {
				t.Fatalf("missing go package name")
			}
		}
		if target == "dart" {
			if !strings.Contains(content, "// Generated by splash. Do not edit.") {
				t.Fatalf("missing dart generated header")
			}
			if !strings.Contains(content, "library generated_config;") {
				t.Fatalf("missing dart library name")
			}
		}
	}
}

func TestGenerateArtifactsGoAndDartGoldenSnapshots(t *testing.T) {
	in := fixtureInputs()
	for _, target := range []string{"go", "dart"} {
		outDir := t.TempDir()
		arts, entries := GenerateArtifacts(in, target, outDir)
		if len(entries) > 0 {
			t.Fatalf("target=%s entries=%v", target, entries)
		}
		if len(arts) != 1 {
			t.Fatalf("target=%s expected 1 artifact, got %d", target, len(arts))
		}
		got, err := os.ReadFile(arts[0].Path)
		if err != nil {
			t.Fatalf("target=%s read generated: %v", target, err)
		}
		base := filepath.Base(arts[0].Path)
		want, err := os.ReadFile(filepath.Join("testdata", "golden", base))
		if err != nil {
			t.Fatalf("target=%s read golden: %v", target, err)
		}
		got = bytes.TrimRight(got, "\n")
		want = bytes.TrimRight(want, "\n")
		if !bytes.Equal(got, want) {
			t.Fatalf("target=%s golden mismatch for %s", target, base)
		}
	}
}

func TestGenerateArtifactsGoSourceParses(t *testing.T) {
	in := fixtureInputs()
	outDir := t.TempDir()
	arts, entries := GenerateArtifacts(in, "go", outDir)
	if len(entries) > 0 {
		t.Fatalf("entries=%v", entries)
	}
	if len(arts) != 1 {
		t.Fatalf("expected 1 artifact, got %d", len(arts))
	}
	src, err := os.ReadFile(arts[0].Path)
	if err != nil {
		t.Fatalf("read generated go source: %v", err)
	}
	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, arts[0].Path, src, parser.AllErrors); err != nil {
		t.Fatalf("generated go is not parsable: %v", err)
	}
}

func TestGenerateArtifactsIdentifierCollision(t *testing.T) {
	dir := t.TempDir()
	regSchema := filepath.Join(dir, "reg.schema.cue")
	regInput := filepath.Join(dir, "reg.cue")
	cfgSchema := filepath.Join(dir, "cfg.schema.cue")
	cfgInput := filepath.Join(dir, "cfg.cue")
	mustWrite(t, regSchema, `package designregistry
#DesignRegistrySpec: {
  keySchemaRegistry: [string]: _
  generatorCapabilities: [...{
    keySchema: string
    target: "go"
    supportsNodeKinds: [..."text"]
    artifactPattern: string
    artifactNaming?: { goPackageName?: string }
  }]
}
`)
	mustWrite(t, regInput, `package designregistry
designRegistry: #DesignRegistrySpec & {
  keySchemaRegistry: {"input-field": {}}
  generatorCapabilities: [{
    keySchema: "input-field"
    target: "go"
    supportsNodeKinds: ["text"]
    artifactPattern: "internal/generated/<domain>_config.go"
    artifactNaming: { goPackageName: "generatedconfig" }
  }]
}
`)
	mustWrite(t, cfgSchema, `package configkey
textEntries: [...{ key: string, value: string, metaArgs: [...string] }]
i18nEntries: [..._]
validations: [..._]
`)
	mustWrite(t, cfgInput, `textEntries: [
  {key: "aB", value: "v1", metaArgs: ["meta"]},
  {key: "a-b", value: "v2", metaArgs: ["meta"]},
]
i18nEntries: []
validations: []
`)
	in := app.Inputs{RegistryPath: regInput, RegistrySchemaPath: regSchema, ConfigPath: cfgInput, ConfigSchemaPath: cfgSchema}
	_, entries := GenerateArtifacts(in, "go", t.TempDir())
	if len(entries) == 0 {
		t.Fatalf("expected collision diagnostics")
	}
	if entries[0].ID != "GEN-0201" {
		t.Fatalf("expected GEN-0201, got %v", entries)
	}
}

func TestGenerateArtifactsEscapingAndUnicode(t *testing.T) {
	dir := t.TempDir()
	regSchema := filepath.Join(dir, "reg.schema.cue")
	regInput := filepath.Join(dir, "reg.cue")
	cfgSchema := filepath.Join(dir, "cfg.schema.cue")
	cfgInput := filepath.Join(dir, "cfg.cue")
	mustWrite(t, regSchema, `package designregistry
#DesignRegistrySpec: {
  keySchemaRegistry: [string]: _
  generatorCapabilities: [...{
    keySchema: string
    target: "go" | "dart"
    supportsNodeKinds: [..."text"]
    artifactPattern: string
    artifactNaming?: {
      goPackageName?: string
      dartLibraryName?: string
    }
  }]
}
`)
	mustWrite(t, regInput, `package designregistry
designRegistry: #DesignRegistrySpec & {
  keySchemaRegistry: {"input-field": {}}
  generatorCapabilities: [
    {
      keySchema: "input-field"
      target: "go"
      supportsNodeKinds: ["text"]
      artifactPattern: "internal/generated/<domain>_config.go"
      artifactNaming: { goPackageName: "generatedconfig" }
    },
    {
      keySchema: "input-field"
      target: "dart"
      supportsNodeKinds: ["text"]
      artifactPattern: "lib/generated/<domain>_config.dart"
      artifactNaming: { dartLibraryName: "generated_config" }
    },
  ]
}
`)
	mustWrite(t, cfgSchema, `package configkey
textEntries: [...{ key: string, value: string, metaArgs: [...string] }]
i18nEntries: [..._]
validations: [..._]
`)
	mustWrite(t, cfgInput, "textEntries: [{key: \"textKey\", value: \"Line 1\\n\\\"quote\\\" - café\", metaArgs: [\"meta\"]}]\ni18nEntries: []\nvalidations: []\n")
	in := app.Inputs{RegistryPath: regInput, RegistrySchemaPath: regSchema, ConfigPath: cfgInput, ConfigSchemaPath: cfgSchema}
	goArts, goEntries := GenerateArtifacts(in, "go", dir)
	if len(goEntries) > 0 || len(goArts) != 1 {
		t.Fatalf("go generation failed: entries=%v arts=%v", goEntries, goArts)
	}
	dartArts, dartEntries := GenerateArtifacts(in, "dart", dir)
	if len(dartEntries) > 0 || len(dartArts) != 1 {
		t.Fatalf("dart generation failed: entries=%v arts=%v", dartEntries, dartArts)
	}
	goRaw, _ := os.ReadFile(goArts[0].Path)
	dartRaw, _ := os.ReadFile(dartArts[0].Path)
	if !strings.Contains(string(goRaw), "\\n") || !strings.Contains(string(goRaw), "\\\"quote\\\"") {
		t.Fatalf("expected escaped newline/quotes in go output: %s", string(goRaw))
	}
	if !strings.Contains(string(goRaw), "café") {
		t.Fatalf("expected unicode preserved in go output")
	}
	if !strings.Contains(string(dartRaw), "\\n") || !strings.Contains(string(dartRaw), "\\\"quote\\\"") {
		t.Fatalf("expected escaped newline/quotes in dart output: %s", string(dartRaw))
	}
	if !strings.Contains(string(dartRaw), "café") {
		t.Fatalf("expected unicode preserved in dart output")
	}
}

func TestGenerateArtifactsRejectsUnsafePattern(t *testing.T) {
	in := writeInputsFixture(t, `package designregistry
#DesignRegistrySpec: {
  keySchemaRegistry: [string]: _
  generatorCapabilities: [...{
    keySchema: string
    target: "json"
    supportsNodeKinds: [..."text"]
    artifactPattern: string
  }]
}
`, `package designregistry
designRegistry: #DesignRegistrySpec & {
  keySchemaRegistry: {"input-field": {}}
  generatorCapabilities: [{
    keySchema: "input-field"
    target: "json"
    supportsNodeKinds: ["text"]
    artifactPattern: "../bad/<domain>.json"
  }]
}
`, `package configkey
textEntries: [...{ key: string, value: string, metaArgs: [...string] }]
i18nEntries: [..._]
validations: [..._]
`, `textEntries: [{key: "k", value: "v", metaArgs: []}]
i18nEntries: []
validations: []
`)
	_, entries := GenerateArtifacts(in, "json", t.TempDir())
	if len(entries) == 0 || entries[0].ID != "GEN-0009" {
		t.Fatalf("expected GEN-0009, got %v", entries)
	}
}
