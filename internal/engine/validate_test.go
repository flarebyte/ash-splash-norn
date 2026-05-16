package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flarebyte/ash-splash-norn/internal/app"
)

func TestValidateInputsMissingPath(t *testing.T) {
	entries := ValidateInputs(app.Inputs{})
	if len(entries) == 0 {
		t.Fatal("expected errors")
	}
}

func TestValidateInputsOK(t *testing.T) {
	dir := t.TempDir()
	mk := func(name, content string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatalf("write file: %v", err)
		}
		return p
	}
	registrySchema := `package designregistry
#DesignRegistrySpec: {
  keySchemaRegistry: [string]: _
  generatorCapabilities: [..._]
}
`
	registryInput := `package designregistry
designRegistry: #DesignRegistrySpec & {
  keySchemaRegistry: {}
  generatorCapabilities: []
}
`
	configSchema := `package configkey
textEntries: [...{
  key: string
}]
`
	configInput := `textEntries: [{key: "ok"}]
`
	entries := ValidateInputs(app.Inputs{
		RegistryPath:       mk("registry.cue", registryInput),
		RegistrySchemaPath: mk("registry.schema.cue", registrySchema),
		ConfigPath:         mk("config.cue", configInput),
		ConfigSchemaPath:   mk("config.schema.cue", configSchema),
	})
	if len(entries) != 0 {
		t.Fatalf("expected no errors, got %d", len(entries))
	}
}
