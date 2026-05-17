package engine

import (
	"encoding/json"
	"fmt"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"github.com/flarebyte/ash-splash-norn/internal/app"
	"github.com/flarebyte/ash-splash-norn/internal/diag"
	"gopkg.in/yaml.v3"
)

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

func buildArbDoc(schemaRef, locale string, requireAll bool, cfg genConfig) (map[string]any, []diag.Entry) {
	_ = schemaRef
	doc := map[string]any{}
	for _, e := range cfg.I18nEntries {
		tr, ok := e.Translations[locale]
		if !ok {
			if requireAll {
				return nil, []diag.Entry{{
					Stage:    "generate",
					ID:       "GEN-0301",
					Severity: diag.SeverityError,
					Message:  fmt.Sprintf("missing translation for locale %s and key %s", locale, e.Key),
				}}
			}
			continue
		}
		txt, _ := tr["text"].(string)
		doc[e.Key] = txt
		meta := map[string]any{}
		if strings.TrimSpace(e.Description) != "" {
			meta["description"] = e.Description
		}
		if ctx, ok := tr["context"].(map[string]any); ok && len(ctx) > 0 {
			meta["context"] = ctx
			placeholders := map[string]any{}
			for k := range ctx {
				placeholders[k] = map[string]any{"type": "String"}
			}
			if len(placeholders) > 0 {
				meta["placeholders"] = placeholders
			}
		}
		if len(meta) > 0 {
			doc["@"+e.Key] = meta
		}
	}
	return doc, nil
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

	cfgInputBody, cfgEntries := readConfigInputSource(in.ConfigPath, "generate")
	if len(cfgEntries) > 0 {
		out := make([]diag.Entry, 0, len(cfgEntries))
		for _, e := range cfgEntries {
			switch e.ID {
			case "CFG-0001", "CFG-0002", "CFG-0003", "CFG-0005":
				e.ID = "GEN-0103"
			case "CFG-0004":
				e.ID = "GEN-0106"
			case "CFG-0006":
				e.ID = "GEN-0107"
			default:
				e.ID = "GEN-0103"
			}
			e.Stage = "generate"
			out = append(out, e)
		}
		return reg, cfg, out
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
