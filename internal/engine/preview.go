package engine

import (
	"fmt"
	"os"
	"sort"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"github.com/flarebyte/ash-splash-norn/internal/app"
	"github.com/flarebyte/ash-splash-norn/internal/diag"
)

type PreviewRow struct {
	SchemaRef       string   `json:"schemaRef"`
	Target          string   `json:"target"`
	SupportsNode    []string `json:"supportsNodeKinds"`
	ArtifactPattern string   `json:"artifactPattern"`
}

type registryDoc struct {
	DesignRegistry struct {
		GeneratorCapabilities []struct {
			KeySchema         string   `cue:"keySchema"`
			Target            string   `cue:"target"`
			SupportsNodeKinds []string `cue:"supportsNodeKinds"`
			ArtifactPattern   string   `cue:"artifactPattern"`
		} `cue:"generatorCapabilities"`
	} `cue:"designRegistry"`
}

func BuildPreview(in app.Inputs) ([]PreviewRow, []diag.Entry) {
	src, entries := mergeCueSources(in.RegistrySchemaPath, in.RegistryPath, "preview")
	if len(entries) > 0 {
		return nil, entries
	}

	ctx := cuecontext.New()
	v := ctx.CompileString(src, cue.Filename("preview.cue"))
	if err := v.Err(); err != nil {
		return nil, []diag.Entry{{
			Stage:    "preview",
			ID:       "PRV-0001",
			Severity: diag.SeverityError,
			Message:  fmt.Sprintf("cue compile failed: %v", sanitizeCueErr(err)),
			Path:     in.RegistryPath,
		}}
	}

	regVal := v.LookupPath(cue.ParsePath("designRegistry"))
	if !regVal.Exists() {
		return nil, []diag.Entry{{
			Stage:    "preview",
			ID:       "PRV-0002",
			Severity: diag.SeverityError,
			Message:  "designRegistry is missing",
			Path:     in.RegistryPath,
		}}
	}

	var doc registryDoc
	if err := v.Decode(&doc); err != nil {
		return nil, []diag.Entry{{
			Stage:    "preview",
			ID:       "PRV-0003",
			Severity: diag.SeverityError,
			Message:  fmt.Sprintf("failed to decode registry: %v", sanitizeCueErr(err)),
			Path:     in.RegistryPath,
		}}
	}

	rows := make([]PreviewRow, 0, len(doc.DesignRegistry.GeneratorCapabilities))
	for _, c := range doc.DesignRegistry.GeneratorCapabilities {
		rows = append(rows, PreviewRow{
			SchemaRef:       c.KeySchema,
			Target:          c.Target,
			SupportsNode:    append([]string(nil), c.SupportsNodeKinds...),
			ArtifactPattern: c.ArtifactPattern,
		})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Target != rows[j].Target {
			return rows[i].Target < rows[j].Target
		}
		if rows[i].ArtifactPattern != rows[j].ArtifactPattern {
			return rows[i].ArtifactPattern < rows[j].ArtifactPattern
		}
		return rows[i].SchemaRef < rows[j].SchemaRef
	})
	return rows, nil
}

func mergeCueSources(schemaPath, inputPath, stage string) (string, []diag.Entry) {
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
	inputSrc, err := os.ReadFile(inputPath)
	if err != nil {
		return "", []diag.Entry{{
			Stage:    stage,
			ID:       "SCH-0102",
			Severity: diag.SeverityError,
			Message:  fmt.Sprintf("failed to read input file: %v", err),
			Path:     inputPath,
		}}
	}
	schemaPkg := extractPackage(string(schemaSrc))
	inputPkg := extractPackage(string(inputSrc))
	inputBody := stripPackageDecl(string(inputSrc))
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
