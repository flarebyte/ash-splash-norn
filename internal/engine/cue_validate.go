package engine

import (
	"fmt"
	"regexp"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"github.com/flarebyte/ash-splash-norn/internal/app"
	"github.com/flarebyte/ash-splash-norn/internal/diag"
)

var pkgPattern = regexp.MustCompile(`(?m)^\s*package\s+([A-Za-z_][A-Za-z0-9_]*)\s*$`)

func ValidateCuePairs(in app.Inputs) []diag.Entry {
	in = in.Cleaned()
	entries := make([]diag.Entry, 0)

	cfgEntries := validateCuePair(in.ConfigSchemaPath, in.ConfigPath, "config")
	entries = append(entries, cfgEntries...)

	regEntries := validateCuePair(in.RegistrySchemaPath, in.RegistryPath, "registry")
	entries = append(entries, regEntries...)

	return entries
}

func validateCuePair(schemaPath, inputPath, stage string) []diag.Entry {
	entries := make([]diag.Entry, 0)
	merged, mergeEntries := mergeCueSources(schemaPath, inputPath, stage)
	if len(mergeEntries) > 0 {
		return mergeEntries
	}

	ctx := cuecontext.New()
	v := ctx.CompileString(merged, cue.Filename(stage+".cue"))
	if err := v.Err(); err != nil {
		entries = append(entries, diag.Entry{
			Stage:    stage,
			ID:       "SCH-0104",
			Severity: diag.SeverityError,
			Message:  fmt.Sprintf("cue compile failed: %v", sanitizeCueErr(err)),
			Path:     inputPath,
		})
		return entries
	}
	if err := v.Validate(cue.All()); err != nil {
		entries = append(entries, diag.Entry{
			Stage:    stage,
			ID:       "SCH-0105",
			Severity: diag.SeverityError,
			Message:  fmt.Sprintf("cue validation failed: %v", sanitizeCueErr(err)),
			Path:     inputPath,
		})
	}
	return entries
}

func extractPackage(src string) string {
	m := pkgPattern.FindStringSubmatch(src)
	if len(m) < 2 {
		return ""
	}
	return m[1]
}

func sanitizeCueErr(err error) string {
	msg := err.Error()
	msg = strings.ReplaceAll(msg, "\n", " | ")
	return strings.TrimSpace(msg)
}

func stripPackageDecl(src string) string {
	return pkgPattern.ReplaceAllString(src, "")
}
