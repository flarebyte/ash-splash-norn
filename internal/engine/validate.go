package engine

import (
	"fmt"
	"os"

	"github.com/flarebyte/ash-splash-norn/internal/app"
	"github.com/flarebyte/ash-splash-norn/internal/diag"
)

func ValidateInputs(in app.Inputs) []diag.Entry {
	in = in.Cleaned()
	entries := make([]diag.Entry, 0)
	hasPathErrors := false
	for _, p := range in.Paths() {
		isConfigPath := p == in.ConfigPath
		if p == "" {
			entries = append(entries, diag.Entry{
				Stage:    "schema",
				ID:       "SCH-0001",
				Severity: diag.SeverityError,
				Message:  "required path flag is missing",
			})
			hasPathErrors = true
			continue
		}
		st, err := os.Stat(p)
		if err != nil {
			entries = append(entries, diag.Entry{
				Stage:    "schema",
				ID:       "SCH-0002",
				Severity: diag.SeverityError,
				Message:  fmt.Sprintf("path does not exist: %s", p),
				Path:     p,
			})
			hasPathErrors = true
			continue
		}
		if st.IsDir() {
			if isConfigPath {
				if _, cfgEntries := readConfigInputSource(p, "schema"); len(cfgEntries) > 0 {
					hasPathErrors = true
					for _, e := range cfgEntries {
						switch e.ID {
						case "CFG-0006":
							e.ID = "SCH-0004"
						case "CFG-0004":
							e.ID = "SCH-0005"
						}
						entries = append(entries, e)
					}
				}
				continue
			}
			entries = append(entries, diag.Entry{
				Stage:    "schema",
				ID:       "SCH-0003",
				Severity: diag.SeverityError,
				Message:  "path must be a file",
				Path:     p,
			})
			hasPathErrors = true
		}
	}
	if hasPathErrors {
		return entries
	}
	entries = append(entries, ValidateCuePairs(in)...)
	rows, listEntries := BuildCatalog(in)
	if len(listEntries) == 0 {
		for _, row := range rows {
			entries = append(entries, artifactPatternIssuesToDiag("schema", "SCH-0006", "SCH-0007", "SCH-0007", "SCH-0008", row.Target, row.ArtifactPattern)...)
		}
	}
	return entries
}
