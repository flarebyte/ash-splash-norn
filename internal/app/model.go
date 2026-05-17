// purpose: Defines CLI input/build models so command and engine layers share a stable contract.
// responsibilities: Normalize and expose input path sets for downstream validation and execution.
// architecture notes: Model logic is intentionally thin to keep business rules in engine and command wiring in cli.

package app

import "path/filepath"

// Inputs captures user-provided config paths for validation/linting/preview.
type Inputs struct {
	RegistryPath       string
	RegistrySchemaPath string
	ConfigPath         string
	ConfigSchemaPath   string
}

func (i Inputs) Paths() []string {
	return []string{i.RegistryPath, i.RegistrySchemaPath, i.ConfigPath, i.ConfigSchemaPath}
}

func (i Inputs) Cleaned() Inputs {
	if i.RegistryPath != "" {
		i.RegistryPath = filepath.Clean(i.RegistryPath)
	}
	if i.RegistrySchemaPath != "" {
		i.RegistrySchemaPath = filepath.Clean(i.RegistrySchemaPath)
	}
	if i.ConfigPath != "" {
		i.ConfigPath = filepath.Clean(i.ConfigPath)
	}
	if i.ConfigSchemaPath != "" {
		i.ConfigSchemaPath = filepath.Clean(i.ConfigSchemaPath)
	}
	return i
}
