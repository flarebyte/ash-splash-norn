// purpose: Validates schema/input CUE pairs and shared merge semantics for engine consumers.
// responsibilities: Merge schema and input sources, compile/validate via CUE, and emit stable schema diagnostics.
// architecture notes: Package matching and merge behavior are intentionally shared to prevent command-level divergence.

package engine

import (
	"fmt"
	"os"
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
	merged, mergeEntries := mergeCueSourcesWithConfigDir(schemaPath, inputPath, stage, stage == "config")
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

func mergeCueSourcesWithConfigDir(schemaPath, inputPath, stage string, allowConfigDir bool) (string, []diag.Entry) {
	schemaSrc, err := os.ReadFile(schemaPath)
	if err != nil {
		return "", []diag.Entry{{
			Stage:    stage,
			ID:       "SCH-0101",
			Severity: diag.SeverityError,
			Message:  fmt.Sprintf("failed to read schema file: %v", err),
			Path:     schemaPath,
		}}
	}
	var inputSrc string
	if allowConfigDir {
		src, entries := readConfigInputSource(inputPath, stage)
		if len(entries) > 0 {
			return "", entries
		}
		inputSrc = src
	} else {
		b, err := os.ReadFile(inputPath)
		if err != nil {
			return "", []diag.Entry{{
				Stage:    stage,
				ID:       "SCH-0102",
				Severity: diag.SeverityError,
				Message:  fmt.Sprintf("failed to read input file: %v", err),
				Path:     inputPath,
			}}
		}
		inputSrc = string(b)
	}

	schemaPkg := extractPackage(string(schemaSrc))
	inputPkg := extractPackage(inputSrc)
	inputBody := stripPackageDecl(inputSrc)
	merged := string(schemaSrc)
	if inputPkg == "" && schemaPkg != "" {
		merged += "\n\n" + inputBody
		return merged, nil
	}
	if schemaPkg != "" && inputPkg != "" && schemaPkg != inputPkg {
		return "", []diag.Entry{{
			Stage:    stage,
			ID:       "SCH-0103",
			Severity: diag.SeverityError,
			Message:  fmt.Sprintf("package mismatch between schema (%s) and input (%s)", schemaPkg, inputPkg),
			Path:     inputPath,
		}}
	}
	merged += "\n\n" + inputBody
	return merged, nil
}

func mergeCueSources(schemaPath, inputPath, stage string) (string, []diag.Entry) {
	return mergeCueSourcesWithConfigDir(schemaPath, inputPath, stage, false)
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
