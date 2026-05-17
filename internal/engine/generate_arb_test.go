package engine

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateArtifactsARBJsonGoldenSnapshots(t *testing.T) {
	in := fixtureInputs()
	outDir := t.TempDir()
	arts, entries := GenerateArtifacts(in, "arb.json", outDir)
	if len(entries) > 0 {
		t.Fatalf("entries=%v", entries)
	}
	if len(arts) != 2 {
		t.Fatalf("expected 2 locale artifacts, got %d", len(arts))
	}
	for _, a := range arts {
		got, err := os.ReadFile(a.Path)
		if err != nil {
			t.Fatalf("read generated: %v", err)
		}
		base := filepath.Base(a.Path)
		want, err := os.ReadFile(filepath.Join("testdata", "golden", base))
		if err != nil {
			t.Fatalf("read golden: %v", err)
		}
		if !bytes.Equal(bytes.TrimRight(got, "\n"), bytes.TrimRight(want, "\n")) {
			t.Fatalf("golden mismatch for %s", base)
		}
	}
}

func TestGenerateArtifactsARBJsonMissingRequiredLocaleFails(t *testing.T) {
	in := writeInputsFixture(t, `package designregistry
#DesignRegistrySpec: {
  keySchemaRegistry: [string]: {
    supportedLanguages: [...string]
    translationPolicy: { requireAllSupportedLanguages: bool }
  }
  generatorCapabilities: [...{
    keySchema: string
    target: "arb.json"
    supportsNodeKinds: [..."i18n"]
    artifactPattern: string
  }]
}
`, `package designregistry
designRegistry: #DesignRegistrySpec & {
  keySchemaRegistry: {
    "input-field": {
      supportedLanguages: ["en", "fr"]
      translationPolicy: { requireAllSupportedLanguages: true }
    }
  }
  generatorCapabilities: [{
    keySchema: "input-field"
    target: "arb.json"
    supportsNodeKinds: ["i18n"]
    artifactPattern: "lib/l10n/app_<locale>.arb.json"
  }]
}
`, `package configkey
i18nEntries: [...{
  key: string
  description: string
  metaArgs: [...string]
  translations: [string]: { text: string }
}]
textEntries: [..._]
validations: [..._]
`, `i18nEntries: [{
  key: "helloKey"
  description: "desc"
  metaArgs: ["meta"]
  translations: {
    en: { text: "Hello" }
  }
}]
textEntries: []
validations: []
`)
	_, entries := GenerateArtifacts(in, "arb.json", t.TempDir())
	if len(entries) == 0 {
		t.Fatalf("expected missing-locale error")
	}
	if entries[0].ID != "GEN-0301" {
		t.Fatalf("expected GEN-0301, got %v", entries)
	}
}

func TestGenerateArtifactsARBJsonMetadataShape(t *testing.T) {
	in := fixtureInputs()
	outDir := t.TempDir()
	arts, entries := GenerateArtifacts(in, "arb.json", outDir)
	if len(entries) > 0 {
		t.Fatalf("entries=%v", entries)
	}
	var enPath string
	for _, a := range arts {
		if filepath.Base(a.Path) == "app_en.arb.json" {
			enPath = a.Path
			break
		}
	}
	if enPath == "" {
		t.Fatalf("missing app_en.arb.json output")
	}
	raw, err := os.ReadFile(enPath)
	if err != nil {
		t.Fatalf("read en arb: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("decode en arb: %v", err)
	}
	metaRaw, ok := payload["@fieldsTextInputLabel"]
	if !ok {
		t.Fatalf("missing metadata entry for fieldsTextInputLabel")
	}
	meta, ok := metaRaw.(map[string]any)
	if !ok {
		t.Fatalf("invalid metadata shape: %T", metaRaw)
	}
	if _, ok := meta["description"]; !ok {
		t.Fatalf("expected description in metadata: %v", meta)
	}
	if _, ok := meta["context"]; !ok {
		t.Fatalf("expected context in metadata: %v", meta)
	}
	if _, ok := meta["placeholders"]; !ok {
		t.Fatalf("expected placeholders in metadata: %v", meta)
	}
}
