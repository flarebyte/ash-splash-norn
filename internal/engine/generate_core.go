// purpose: Coordinates artifact generation across supported targets from validated registry/config inputs.
// responsibilities: Resolve capabilities, enforce pattern preflight, and route writes to target-specific emit paths.
// architecture notes: Target routing is explicit to keep failure modes stable and diagnostics deterministic.

package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/flarebyte/ash-splash-norn/internal/app"
	"github.com/flarebyte/ash-splash-norn/internal/diag"
)

func GenerateArtifacts(in app.Inputs, target, outputRoot string) ([]GeneratedArtifact, []diag.Entry) {
	if target != "json" && target != "yaml" && target != "cue" && target != "go" && target != "dart" && target != "arb.json" {
		return nil, []diag.Entry{{Stage: "generate", ID: "GEN-0002", Severity: diag.SeverityError, Message: fmt.Sprintf("unsupported target for P02: %s", target)}}
	}
	reg, cfg, entries := loadGenerateDocs(in)
	if len(entries) > 0 {
		return nil, entries
	}

	caps := resolveCapabilities(reg, target)
	if len(caps) == 0 {
		return nil, []diag.Entry{{Stage: "generate", ID: "GEN-0001", Severity: diag.SeverityError, Message: fmt.Sprintf("no generator capability for target %s", target)}}
	}
	for _, cap := range caps {
		patternEntries := artifactPatternIssuesToDiag("generate", "GEN-0007", "GEN-0008", "GEN-0010", "GEN-0009", cap.Target, cap.ArtifactPattern)
		if len(patternEntries) > 0 {
			return nil, patternEntries
		}
	}

	artifacts := make([]GeneratedArtifact, 0)
	for _, cap := range caps {
		if target == "arb.json" {
			langs := []string{}
			requireAll := true
			if ks, ok := reg.DesignRegistry.KeySchemaRegistry[cap.KeySchema]; ok {
				langs = append(langs, ks.SupportedLanguages...)
				requireAll = ks.TranslationPolicy.RequireAllSupportedLanguages
			}
			sort.Strings(langs)
			if len(langs) == 0 {
				return nil, []diag.Entry{{Stage: "generate", ID: "GEN-0300", Severity: diag.SeverityError, Message: fmt.Sprintf("no supportedLanguages found for key schema %s", cap.KeySchema)}}
			}
			for _, lang := range langs {
				doc, dEntries := buildArbDoc(cap.KeySchema, lang, requireAll, cfg)
				if len(dEntries) > 0 {
					return nil, dEntries
				}
				rel := strings.ReplaceAll(cap.ArtifactPattern, "<locale>", lang)
				outPath := filepath.Clean(filepath.Join(outputRoot, rel))
				if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
					return nil, []diag.Entry{{Stage: "generate", ID: "GEN-0004", Severity: diag.SeverityError, Message: fmt.Sprintf("failed to create output directory: %v", err), Path: outPath}}
				}
				data, err := json.MarshalIndent(doc, "", "  ")
				if err != nil {
					return nil, []diag.Entry{{Stage: "generate", ID: "GEN-0005", Severity: diag.SeverityError, Message: fmt.Sprintf("failed to encode output: %v", err), Path: outPath}}
				}
				if err := os.WriteFile(outPath, append(data, '\n'), 0o644); err != nil {
					return nil, []diag.Entry{{Stage: "generate", ID: "GEN-0006", Severity: diag.SeverityError, Message: fmt.Sprintf("failed to write output file: %v", err), Path: outPath}}
				}
				artifacts = append(artifacts, GeneratedArtifact{Path: outPath})
			}
			continue
		}
		if target == "go" || target == "dart" {
			rel := strings.ReplaceAll(cap.ArtifactPattern, "<domain>", cap.KeySchema)
			outPath := filepath.Clean(filepath.Join(outputRoot, rel))
			if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
				return nil, []diag.Entry{{Stage: "generate", ID: "GEN-0004", Severity: diag.SeverityError, Message: fmt.Sprintf("failed to create output directory: %v", err), Path: outPath}}
			}
			var data []byte
			var err error
			if target == "go" {
				data, err = buildGoSource(cap, cfg)
			} else {
				data, err = buildDartSource(cap, cfg)
			}
			if err != nil {
				return nil, []diag.Entry{{Stage: "generate", ID: "GEN-0201", Severity: diag.SeverityError, Message: err.Error(), Path: outPath}}
			}
			if err := os.WriteFile(outPath, data, 0o644); err != nil {
				return nil, []diag.Entry{{Stage: "generate", ID: "GEN-0006", Severity: diag.SeverityError, Message: fmt.Sprintf("failed to write output file: %v", err), Path: outPath}}
			}
			artifacts = append(artifacts, GeneratedArtifact{Path: outPath})
			continue
		}
		supports := append([]string(nil), cap.SupportsNode...)
		sort.Strings(supports)
		for _, kind := range supports {
			doc := buildNodeKindDoc(cap.KeySchema, target, kind, cfg)
			rel := strings.ReplaceAll(cap.ArtifactPattern, "<domain>", cap.KeySchema)
			if strings.Contains(rel, "<locale>") {
				return nil, []diag.Entry{{Stage: "generate", ID: "GEN-0003", Severity: diag.SeverityError, Message: "<locale> artifact patterns are not supported in P02 targets"}}
			}
			if len(supports) > 1 {
				rel = injectNodeKindInPath(rel, kind)
			}
			outPath := filepath.Clean(filepath.Join(outputRoot, rel))
			if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
				return nil, []diag.Entry{{Stage: "generate", ID: "GEN-0004", Severity: diag.SeverityError, Message: fmt.Sprintf("failed to create output directory: %v", err), Path: outPath}}
			}
			data, err := encodeDocByTarget(doc, target)
			if err != nil {
				return nil, []diag.Entry{{Stage: "generate", ID: "GEN-0005", Severity: diag.SeverityError, Message: fmt.Sprintf("failed to encode output: %v", err), Path: outPath}}
			}
			if err := os.WriteFile(outPath, append(data, '\n'), 0o644); err != nil {
				return nil, []diag.Entry{{Stage: "generate", ID: "GEN-0006", Severity: diag.SeverityError, Message: fmt.Sprintf("failed to write output file: %v", err), Path: outPath}}
			}
			artifacts = append(artifacts, GeneratedArtifact{Path: outPath})
		}
	}
	return artifacts, nil
}

func resolveCapabilities(reg genRegistry, target string) []capResolved {
	caps := make([]capResolved, 0)
	for _, c := range reg.DesignRegistry.GeneratorCapabilities {
		if c.Target == target {
			caps = append(caps, capResolved{
				KeySchema:       c.KeySchema,
				Target:          c.Target,
				SupportsNode:    append([]string(nil), c.SupportsNodeKinds...),
				ArtifactPattern: c.ArtifactPattern,
				GoPackageName:   c.ArtifactNaming.GoPackageName,
				DartLibraryName: c.ArtifactNaming.DartLibraryName,
			})
		}
	}
	if target == "yaml" && len(caps) == 0 {
		for _, c := range reg.DesignRegistry.GeneratorCapabilities {
			if c.Target != "json" {
				continue
			}
			pattern := c.ArtifactPattern
			if strings.HasSuffix(pattern, ".json") {
				pattern = strings.TrimSuffix(pattern, ".json") + ".yaml"
			}
			caps = append(caps, capResolved{
				KeySchema:       c.KeySchema,
				Target:          "yaml",
				SupportsNode:    append([]string(nil), c.SupportsNodeKinds...),
				ArtifactPattern: pattern,
				GoPackageName:   c.ArtifactNaming.GoPackageName,
				DartLibraryName: c.ArtifactNaming.DartLibraryName,
			})
		}
	}
	sort.Slice(caps, func(i, j int) bool {
		if caps[i].Target != caps[j].Target {
			return caps[i].Target < caps[j].Target
		}
		if caps[i].ArtifactPattern != caps[j].ArtifactPattern {
			return caps[i].ArtifactPattern < caps[j].ArtifactPattern
		}
		return caps[i].KeySchema < caps[j].KeySchema
	})
	return caps
}

func injectNodeKindInPath(path, kind string) string {
	ext := filepath.Ext(path)
	if ext == "" {
		return path + "." + kind
	}
	base := strings.TrimSuffix(path, ext)
	return base + "." + kind + ext
}
