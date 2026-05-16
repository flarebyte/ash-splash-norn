package engine

import (
	"fmt"
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
	src, entries := mergeCueSourcesWithConfigDir(in.RegistrySchemaPath, in.RegistryPath, "preview", false)
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
	for _, c := range doc.DesignRegistry.GeneratorCapabilities {
		entries := artifactPatternIssuesToDiag("preview", "PRV-0004", "PRV-0005", "PRV-0007", "PRV-0006", c.Target, c.ArtifactPattern)
		if len(entries) > 0 {
			return nil, entries
		}
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
