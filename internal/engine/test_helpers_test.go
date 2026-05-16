package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flarebyte/ash-splash-norn/internal/app"
)

func writeInputsFixture(t *testing.T, registrySchema, registryInput, configSchema, configInput string) app.Inputs {
	t.Helper()
	dir := t.TempDir()
	regSchema := filepath.Join(dir, "reg.schema.cue")
	regInput := filepath.Join(dir, "reg.cue")
	cfgSchema := filepath.Join(dir, "cfg.schema.cue")
	cfgInput := filepath.Join(dir, "cfg.cue")
	if err := os.WriteFile(regSchema, []byte(registrySchema), 0o644); err != nil {
		t.Fatalf("write reg schema: %v", err)
	}
	if err := os.WriteFile(regInput, []byte(registryInput), 0o644); err != nil {
		t.Fatalf("write reg input: %v", err)
	}
	if err := os.WriteFile(cfgSchema, []byte(configSchema), 0o644); err != nil {
		t.Fatalf("write cfg schema: %v", err)
	}
	if err := os.WriteFile(cfgInput, []byte(configInput), 0o644); err != nil {
		t.Fatalf("write cfg input: %v", err)
	}
	return app.Inputs{
		RegistryPath:       regInput,
		RegistrySchemaPath: regSchema,
		ConfigPath:         cfgInput,
		ConfigSchemaPath:   cfgSchema,
	}
}

func fixtureInputsWithConfigDirFromFile(t *testing.T, srcPath string) (app.Inputs, error) {
	t.Helper()
	in := fixtureInputs()
	raw, err := os.ReadFile(srcPath)
	if err != nil {
		return in, err
	}
	cfgDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(cfgDir, "part1.cue"), raw, 0o644); err != nil {
		return in, err
	}
	in.ConfigPath = cfgDir
	return in, nil
}
