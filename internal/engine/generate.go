package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	goformat "go/format"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

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
			ArtifactNaming    struct {
				GoPackageName   string `cue:"goPackageName"`
				DartLibraryName string `cue:"dartLibraryName"`
			} `cue:"artifactNaming"`
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
	GoPackageName   string
	DartLibraryName string
}

type commandFlagLit struct {
	Kind    string
	Name    string
	Schema  []string
	Schemas [][]string
}

type commandSpecLit struct {
	CommandPath []string
	AdminOnly   bool
	Flags       []commandFlagLit
}

func GenerateArtifacts(in app.Inputs, target, outputRoot string) ([]GeneratedArtifact, []diag.Entry) {
	if target != "json" && target != "yaml" && target != "cue" && target != "go" && target != "dart" {
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

func buildGoSource(cap capResolved, cfg genConfig) ([]byte, error) {
	pkg := cap.GoPackageName
	if strings.TrimSpace(pkg) == "" {
		pkg = "generatedconfig"
	}
	keys := make([]string, 0)
	textByKey := map[string]string{}
	metaByKey := map[string][]string{}
	validatorByKey := map[string]commandSpecLit{}
	allowed := map[string]struct{}{}
	for _, k := range cap.SupportsNode {
		allowed[k] = struct{}{}
	}
	if _, ok := allowed["text"]; ok {
		for _, e := range cfg.TextEntries {
			keys = append(keys, e.Key)
			textByKey[e.Key] = e.Value
			metaByKey[e.Key] = append([]string(nil), e.MetaArgs...)
		}
	}
	if _, ok := allowed["validator"]; ok {
		for _, e := range cfg.Validations {
			keys = append(keys, e.Key)
			metaByKey[e.Key] = append([]string(nil), e.MetaArgs...)
			if len(e.Commands) > 0 {
				spec, ok := extractCommandSpec(e.Commands[0].Args)
				if ok {
					validatorByKey[e.Key] = spec
				}
			}
		}
	}
	sort.Strings(keys)
	ids := map[string]string{}
	for _, key := range keys {
		id := toPascal(key) + "Key"
		if prev, ok := ids[id]; ok && prev != key {
			return nil, fmt.Errorf("identifier collision between %q and %q (%s)", prev, key, id)
		}
		ids[id] = key
	}
	var b bytes.Buffer
	b.WriteString("package " + pkg + "\n\n")
	b.WriteString("// Code generated by flyb. DO NOT EDIT.\n\n")
	for _, key := range keys {
		id := toPascal(key) + "Key"
		b.WriteString(fmt.Sprintf("const %s = %q\n", id, key))
	}
	b.WriteString("\n")
	if len(textByKey) > 0 {
		b.WriteString("var TextByKey = map[string]string{\n")
		for _, key := range sortedMapKeysString(textByKey) {
			id := toPascal(key) + "Key"
			b.WriteString(fmt.Sprintf("\t%s: %q,\n", id, textByKey[key]))
		}
		b.WriteString("}\n\n")
	}
	if len(validatorByKey) > 0 {
		b.WriteString("type CommandFlag struct {\n")
		b.WriteString("\tKind    string\n")
		b.WriteString("\tName    string\n")
		b.WriteString("\tSchema  []string\n")
		b.WriteString("\tSchemas [][]string\n")
		b.WriteString("}\n\n")
		b.WriteString("type CommandSpec struct {\n")
		b.WriteString("\tCommandPath []string\n")
		b.WriteString("\tAdminOnly   bool\n")
		b.WriteString("\tFlags       []CommandFlag\n")
		b.WriteString("}\n\n")
		b.WriteString("var ValidationByKey = map[string]CommandSpec{\n")
		for _, key := range sortedMapKeysSpec(validatorByKey) {
			id := toPascal(key) + "Key"
			spec := validatorByKey[key]
			b.WriteString(fmt.Sprintf("\t%s: {\n", id))
			b.WriteString(fmt.Sprintf("\t\tCommandPath: %s,\n", goStringSlice(spec.CommandPath)))
			if spec.AdminOnly {
				b.WriteString("\t\tAdminOnly:   true,\n")
			} else {
				b.WriteString("\t\tAdminOnly:   false,\n")
			}
			b.WriteString("\t\tFlags: []CommandFlag{\n")
			for _, fl := range spec.Flags {
				b.WriteString("\t\t\t{\n")
				b.WriteString(fmt.Sprintf("\t\t\t\tKind:   %q,\n", fl.Kind))
				b.WriteString(fmt.Sprintf("\t\t\t\tName:   %q,\n", fl.Name))
				b.WriteString(fmt.Sprintf("\t\t\t\tSchema: %s,\n", goStringSlice(fl.Schema)))
				if len(fl.Schemas) > 0 {
					b.WriteString(fmt.Sprintf("\t\t\t\tSchemas: %s,\n", go2DStringSlice(fl.Schemas)))
				}
				b.WriteString("\t\t\t},\n")
			}
			b.WriteString("\t\t},\n")
			b.WriteString("\t},\n")
		}
		b.WriteString("}\n\n")
	}
	if len(metaByKey) > 0 {
		b.WriteString("var MetaArgsByKey = map[string][]string{\n")
		for _, key := range sortedMapKeysSlice(metaByKey) {
			id := toPascal(key) + "Key"
			b.WriteString(fmt.Sprintf("\t%s: {", id))
			for i, tok := range metaByKey[key] {
				if i > 0 {
					b.WriteString(", ")
				}
				b.WriteString(fmt.Sprintf("%q", tok))
			}
			b.WriteString("},\n")
		}
		b.WriteString("}\n")
	}
	formatted, err := goformat.Source(b.Bytes())
	if err != nil {
		return nil, fmt.Errorf("generated go source is not gofmt-compatible: %w", err)
	}
	return formatted, nil
}

func buildDartSource(cap capResolved, cfg genConfig) ([]byte, error) {
	lib := cap.DartLibraryName
	if strings.TrimSpace(lib) == "" {
		lib = "generated_config"
	}
	keys := make([]string, 0)
	textByKey := map[string]string{}
	metaByKey := map[string][]string{}
	validatorByKey := map[string]commandSpecLit{}
	allowed := map[string]struct{}{}
	for _, k := range cap.SupportsNode {
		allowed[k] = struct{}{}
	}
	if _, ok := allowed["text"]; ok {
		for _, e := range cfg.TextEntries {
			keys = append(keys, e.Key)
			textByKey[e.Key] = e.Value
			metaByKey[e.Key] = append([]string(nil), e.MetaArgs...)
		}
	}
	if _, ok := allowed["validator"]; ok {
		for _, e := range cfg.Validations {
			keys = append(keys, e.Key)
			metaByKey[e.Key] = append([]string(nil), e.MetaArgs...)
			if len(e.Commands) > 0 {
				spec, ok := extractCommandSpec(e.Commands[0].Args)
				if ok {
					validatorByKey[e.Key] = spec
				}
			}
		}
	}
	sort.Strings(keys)
	ids := map[string]string{}
	for _, key := range keys {
		id := toLowerCamel(key) + "Key"
		if prev, ok := ids[id]; ok && prev != key {
			return nil, fmt.Errorf("identifier collision between %q and %q (%s)", prev, key, id)
		}
		ids[id] = key
	}
	var b bytes.Buffer
	b.WriteString("library " + lib + ";\n\n")
	b.WriteString("// Generated by flyb. Do not edit.\n\n")
	for _, key := range keys {
		id := toLowerCamel(key) + "Key"
		b.WriteString(fmt.Sprintf("const String %s = %q;\n", id, key))
	}
	b.WriteString("\n")
	if len(textByKey) > 0 {
		b.WriteString("const Map<String, String> textByKey = {\n")
		for _, key := range sortedMapKeysString(textByKey) {
			id := toLowerCamel(key) + "Key"
			b.WriteString(fmt.Sprintf("  %s: %q,\n", id, textByKey[key]))
		}
		b.WriteString("};\n\n")
	}
	if len(validatorByKey) > 0 {
		b.WriteString("class CommandFlag {\n")
		b.WriteString("  final String kind;\n")
		b.WriteString("  final String name;\n")
		b.WriteString("  final List<String> schema;\n")
		b.WriteString("  final List<List<String>> schemas;\n\n")
		b.WriteString("  const CommandFlag({\n")
		b.WriteString("    required this.kind,\n")
		b.WriteString("    required this.name,\n")
		b.WriteString("    required this.schema,\n")
		b.WriteString("    this.schemas = const [],\n")
		b.WriteString("  });\n")
		b.WriteString("}\n\n")
		b.WriteString("class CommandSpec {\n")
		b.WriteString("  final List<String> commandPath;\n")
		b.WriteString("  final bool adminOnly;\n")
		b.WriteString("  final List<CommandFlag> flags;\n\n")
		b.WriteString("  const CommandSpec({\n")
		b.WriteString("    required this.commandPath,\n")
		b.WriteString("    required this.adminOnly,\n")
		b.WriteString("    required this.flags,\n")
		b.WriteString("  });\n")
		b.WriteString("}\n\n")
		b.WriteString("const Map<String, CommandSpec> validationByKey = {\n")
		for _, key := range sortedMapKeysSpec(validatorByKey) {
			id := toLowerCamel(key) + "Key"
			spec := validatorByKey[key]
			b.WriteString(fmt.Sprintf("  %s: CommandSpec(\n", id))
			b.WriteString(fmt.Sprintf("    commandPath: %s,\n", dartStringSlice(spec.CommandPath)))
			if spec.AdminOnly {
				b.WriteString("    adminOnly: true,\n")
			} else {
				b.WriteString("    adminOnly: false,\n")
			}
			b.WriteString("    flags: [\n")
			for _, fl := range spec.Flags {
				b.WriteString("      CommandFlag(\n")
				b.WriteString(fmt.Sprintf("        kind: %q,\n", fl.Kind))
				b.WriteString(fmt.Sprintf("        name: %q,\n", fl.Name))
				b.WriteString(fmt.Sprintf("        schema: %s,\n", dartStringSlice(fl.Schema)))
				if len(fl.Schemas) > 0 {
					b.WriteString(fmt.Sprintf("        schemas: %s,\n", dart2DStringSlice(fl.Schemas)))
				}
				b.WriteString("      ),\n")
			}
			b.WriteString("    ],\n")
			b.WriteString("  ),\n")
		}
		b.WriteString("};\n\n")
	}
	if len(metaByKey) > 0 {
		b.WriteString("const Map<String, List<String>> metaArgsByKey = {\n")
		for _, key := range sortedMapKeysSlice(metaByKey) {
			id := toLowerCamel(key) + "Key"
			b.WriteString(fmt.Sprintf("  %s: [", id))
			for i, tok := range metaByKey[key] {
				if i > 0 {
					b.WriteString(", ")
				}
				b.WriteString(fmt.Sprintf("%q", tok))
			}
			b.WriteString("],\n")
		}
		b.WriteString("};\n")
	}
	return b.Bytes(), nil
}

func toPascal(s string) string {
	var b strings.Builder
	upperNext := true
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			upperNext = true
			continue
		}
		if upperNext {
			b.WriteRune(unicode.ToUpper(r))
			upperNext = false
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func toLowerCamel(s string) string {
	p := toPascal(s)
	if p == "" {
		return p
	}
	return strings.ToLower(p[:1]) + p[1:]
}

func sortedMapKeysString(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func sortedMapKeysAny(m map[string]map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func sortedMapKeysSpec(m map[string]commandSpecLit) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func sortedMapKeysSlice(m map[string][]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func goStringSlice(items []string) string {
	if len(items) == 0 {
		return "[]string{}"
	}
	var b strings.Builder
	b.WriteString("[]string{")
	for i, s := range items {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(fmt.Sprintf("%q", s))
	}
	b.WriteString("}")
	return b.String()
}

func go2DStringSlice(items [][]string) string {
	if len(items) == 0 {
		return "[][]string{}"
	}
	var b strings.Builder
	b.WriteString("[][]string{")
	for i, row := range items {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(goStringSlice(row))
	}
	b.WriteString("}")
	return b.String()
}

func dartStringSlice(items []string) string {
	if len(items) == 0 {
		return "[]"
	}
	var b strings.Builder
	b.WriteString("[")
	for i, s := range items {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(fmt.Sprintf("%q", s))
	}
	b.WriteString("]")
	return b.String()
}

func dart2DStringSlice(items [][]string) string {
	if len(items) == 0 {
		return "[]"
	}
	var b strings.Builder
	b.WriteString("[")
	for i, row := range items {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(dartStringSlice(row))
	}
	b.WriteString("]")
	return b.String()
}

func extractCommandSpec(args map[string]any) (commandSpecLit, bool) {
	if len(args) == 0 {
		return commandSpecLit{}, false
	}
	sections := make([]string, 0, len(args))
	for k := range args {
		sections = append(sections, k)
	}
	sort.Strings(sections)
	first := sections[0]
	raw, ok := args[first].(map[string]any)
	if !ok {
		return commandSpecLit{}, false
	}
	spec := commandSpecLit{}
	if cp, ok := raw["commandPath"].([]any); ok {
		for _, it := range cp {
			if s, ok := it.(string); ok {
				spec.CommandPath = append(spec.CommandPath, s)
			}
		}
	}
	if adminOnly, ok := raw["adminOnly"].(bool); ok {
		spec.AdminOnly = adminOnly
	}
	if flags, ok := raw["flags"].([]any); ok {
		for _, rf := range flags {
			m, ok := rf.(map[string]any)
			if !ok {
				continue
			}
			fl := commandFlagLit{}
			if s, ok := m["kind"].(string); ok {
				fl.Kind = s
			}
			if s, ok := m["name"].(string); ok {
				fl.Name = s
			}
			if sc, ok := m["schema"].([]any); ok {
				for _, it := range sc {
					if s, ok := it.(string); ok {
						fl.Schema = append(fl.Schema, s)
					}
				}
			}
			if scs, ok := m["schemas"].([]any); ok {
				for _, r := range scs {
					rowAny, ok := r.([]any)
					if !ok {
						continue
					}
					row := make([]string, 0, len(rowAny))
					for _, it := range rowAny {
						if s, ok := it.(string); ok {
							row = append(row, s)
						}
					}
					fl.Schemas = append(fl.Schemas, row)
				}
			}
			spec.Flags = append(spec.Flags, fl)
		}
	}
	return spec, true
}
