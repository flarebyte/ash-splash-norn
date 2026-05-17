// purpose: Applies schema governance and artifact-pattern policy checks for lint workflows.
// responsibilities: Validate registry metadata conventions and route pattern policy diagnostics.
// architecture notes: Policy checks are isolated to keep governance evolution independent from graph/config rules.

package engine

import (
	"fmt"
	"strings"

	"github.com/flarebyte/ash-splash-norn/internal/diag"
)

func lintPatterns(reg lintRegistry) []diag.Entry {
	out := make([]diag.Entry, 0)
	for _, c := range reg.DesignRegistry.GeneratorCapabilities {
		out = append(out, artifactPatternIssuesToDiag("lint", "LNT-0009", "LNT-0010", "LNT-0011", "LNT-0011", c.Target, c.ArtifactPattern)...)
	}
	return out
}

func lintSchemaGovernance(reg lintRegistry) []diag.Entry {
	out := make([]diag.Entry, 0)
	if len(reg.DesignRegistry.KeySchemaRegistry) == 0 {
		out = append(out, diag.Entry{Stage: "lint", ID: "LNT-0012", Severity: diag.SeverityError, Message: "keySchemaRegistry must contain at least one schema"})
		return out
	}
	for key, ks := range reg.DesignRegistry.KeySchemaRegistry {
		if strings.TrimSpace(ks.Metadata.ID) == "" {
			out = append(out, diag.Entry{Stage: "lint", ID: "LNT-0013", Severity: diag.SeverityError, Message: fmt.Sprintf("schema %q metadata.id is required", key)})
		}
		if ks.Metadata.ID != "" && ks.Metadata.ID != key {
			out = append(out, diag.Entry{Stage: "lint", ID: "LNT-0014", Severity: diag.SeverityError, Message: fmt.Sprintf("schema key %q must match metadata.id %q", key, ks.Metadata.ID)})
		}
		if len(ks.SupportedLanguages) == 0 {
			out = append(out, diag.Entry{Stage: "lint", ID: "LNT-0015", Severity: diag.SeverityError, Message: fmt.Sprintf("schema %q must declare supportedLanguages", key)})
		}
		if len(ks.SupportedCommandSections) == 0 {
			out = append(out, diag.Entry{Stage: "lint", ID: "LNT-0016", Severity: diag.SeverityError, Message: fmt.Sprintf("schema %q must declare supportedCommandSections", key)})
		}
	}
	return out
}
