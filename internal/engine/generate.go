package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"github.com/flarebyte/ash-splash-norn/internal/app"
	"github.com/flarebyte/ash-splash-norn/internal/diag"
	"gopkg.in/yaml.v3"
)

type GeneratedArtifact struct {
	Path string
}

type genRegistry struct {
	DesignRegistry struct {
		GeneratorCapabilities []struct {
			KeySchema         string   `cue:"keySchema"`
			Target            string   `cue:"target"`
			SupportsNodeKinds []string `cue:"supportsNodeKinds"`
			ArtifactPattern   string   `cue:"artifactPattern"`
		} `cue:"generatorCapabilities"`
	} `cue:"designRegistry"`
}

type genConfig struct {
	I18nEntries []struct {
		Key          string                    `cue:"key"`
		MetaArgs     []string                  `cue:"metaArgs"`
		Translations map[string]map[string]any `cue:"translations"`
	} `cue:"i18nEntries"`
	TextEntries []struct {
		Key      string   `cue:"key"`
		Value    string   `cue:"value"`
		MetaArgs []string `cue:"metaArgs"`
	} `cue:"textEntries"`
	Validations []struct {
		Key      string   `cue:"key"`
		MetaArgs []string `cue:"metaArgs"`
		Commands []struct {
			Args map[string]any `cue:"args"`
		} `cue:"commands"`
	} `cue:"validations"`
}

type emitDoc map[string]any

type capResolved struct {
	KeySchema       string
	Target          string
	SupportsNode    []string
	ArtifactPattern string
}

func GenerateArtifacts(in app.Inputs, target, outputRoot string) ([]GeneratedArtifact, []diag.Entry) {
	if target != "json" && target != "yaml" && target != "cue" {
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

	artifacts := make([]GeneratedArtifact, 0)
	for _, cap := range caps {
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
			caps = append(caps, capResolved{KeySchema: c.KeySchema, Target: c.Target, SupportsNode: append([]string(nil), c.SupportsNodeKinds...), ArtifactPattern: c.ArtifactPattern})
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
			caps = append(caps, capResolved{KeySchema: c.KeySchema, Target: "yaml", SupportsNode: append([]string(nil), c.SupportsNodeKinds...), ArtifactPattern: pattern})
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

func buildNodeKindDoc(schemaRef, target, nodeKind string, cfg genConfig) emitDoc {
	doc := emitDoc{
		"schemaRef":   schemaRef,
		"target":      target,
		"nodeKind":    nodeKind,
		"generatedAt": "deterministic",
	}
	switch nodeKind {
	case "i18n":
		for _, e := range cfg.I18nEntries {
			doc[e.Key] = e.Translations
			doc["@"+e.Key] = map[string]any{"metaArgs": append([]string(nil), e.MetaArgs...)}
		}
	case "text":
		for _, e := range cfg.TextEntries {
			doc[e.Key] = e.Value
			doc["@"+e.Key] = map[string]any{"metaArgs": append([]string(nil), e.MetaArgs...)}
		}
	case "validator":
		for _, e := range cfg.Validations {
			args := map[string]any{}
			if len(e.Commands) > 0 {
				args = e.Commands[0].Args
			}
			doc[e.Key] = args
			doc["@"+e.Key] = map[string]any{"metaArgs": append([]string(nil), e.MetaArgs...)}
		}
	}
	return doc
}

func encodeDocByTarget(doc emitDoc, target string) ([]byte, error) {
	switch target {
	case "json":
		return json.MarshalIndent(doc, "", "  ")
	case "yaml":
		return yaml.Marshal(doc)
	case "cue":
		b, err := json.Marshal(doc)
		if err != nil {
			return nil, err
		}
		return []byte("output: " + string(b)), nil
	default:
		return nil, fmt.Errorf("unsupported target: %s", target)
	}
}

func loadGenerateDocs(in app.Inputs) (genRegistry, genConfig, []diag.Entry) {
	var reg genRegistry
	var cfg genConfig

	regMerged, entries := mergeCueSources(in.RegistrySchemaPath, in.RegistryPath, "generate")
	if len(entries) > 0 {
		return reg, cfg, entries
	}

	ctx := cuecontext.New()
	regVal := ctx.CompileString(regMerged, cue.Filename("generate-registry.cue"))
	if err := regVal.Err(); err != nil {
		return reg, cfg, []diag.Entry{{Stage: "generate", ID: "GEN-0100", Severity: diag.SeverityError, Message: fmt.Sprintf("registry compile failed: %v", sanitizeCueErr(err)), Path: in.RegistryPath}}
	}
	if err := regVal.Decode(&reg); err != nil {
		return reg, cfg, []diag.Entry{{Stage: "generate", ID: "GEN-0101", Severity: diag.SeverityError, Message: fmt.Sprintf("registry decode failed: %v", sanitizeCueErr(err)), Path: in.RegistryPath}}
	}

	cfgSchemaSrc, err := os.ReadFile(in.ConfigSchemaPath)
	if err != nil {
		return reg, cfg, []diag.Entry{{Stage: "generate", ID: "GEN-0102", Severity: diag.SeverityError, Message: fmt.Sprintf("failed to read config schema: %v", err), Path: in.ConfigSchemaPath}}
	}
	cfgInputSrc, err := os.ReadFile(in.ConfigPath)
	if err != nil {
		return reg, cfg, []diag.Entry{{Stage: "generate", ID: "GEN-0103", Severity: diag.SeverityError, Message: fmt.Sprintf("failed to read config input: %v", err), Path: in.ConfigPath}}
	}

	cfgSchemaPkg := extractPackage(string(cfgSchemaSrc))
	cfgInputBody := string(cfgInputSrc)
	if extractPackage(cfgInputBody) == "" && cfgSchemaPkg != "" {
		cfgInputBody = "package " + cfgSchemaPkg + "\n\n" + cfgInputBody
	}
	cfgVal := ctx.CompileString(cfgInputBody, cue.Filename("generate-config.cue"))
	if err := cfgVal.Err(); err != nil {
		return reg, cfg, []diag.Entry{{Stage: "generate", ID: "GEN-0104", Severity: diag.SeverityError, Message: fmt.Sprintf("config compile failed: %v", sanitizeCueErr(err)), Path: in.ConfigPath}}
	}
	if err := cfgVal.Decode(&cfg); err != nil {
		return reg, cfg, []diag.Entry{{Stage: "generate", ID: "GEN-0105", Severity: diag.SeverityError, Message: fmt.Sprintf("config decode failed: %v", sanitizeCueErr(err)), Path: in.ConfigPath}}
	}
	return reg, cfg, nil
}
