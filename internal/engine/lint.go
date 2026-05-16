package engine

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"github.com/flarebyte/ash-splash-norn/internal/app"
	"github.com/flarebyte/ash-splash-norn/internal/diag"
)

type lintRegistry struct {
	DesignRegistry struct {
		KeySchemaRegistry map[string]struct {
			Metadata struct {
				ID      string `cue:"id"`
				Version string `cue:"version"`
			} `cue:"metadata"`
			SupportedLanguages       []string `cue:"supportedLanguages"`
			SupportedCommandSections []string `cue:"supportedCommandSections"`
			TranslationPolicy        struct {
				RequireAllSupportedLanguages bool `cue:"requireAllSupportedLanguages"`
			} `cue:"translationPolicy"`
			RootLabels   []string `cue:"rootLabels"`
			NodesByLabel map[string]struct {
				Label       string   `cue:"label"`
				Kind        string   `cue:"kind"`
				Mandatory   bool     `cue:"mandatory"`
				ChildLabels []string `cue:"childLabels"`
			} `cue:"nodesByLabel"`
		} `cue:"keySchemaRegistry"`
		GeneratorCapabilities []struct {
			Target          string `cue:"target"`
			ArtifactPattern string `cue:"artifactPattern"`
		} `cue:"generatorCapabilities"`
	} `cue:"designRegistry"`
}

type lintConfig struct {
	I18nEntries []struct {
		Key          string                    `cue:"key"`
		Translations map[string]map[string]any `cue:"translations"`
	} `cue:"i18nEntries"`
	TextEntries []struct {
		Key string `cue:"key"`
	} `cue:"textEntries"`
	Validations []struct {
		Key      string `cue:"key"`
		Commands []struct {
			Args map[string]any `cue:"args"`
		} `cue:"commands"`
	} `cue:"validations"`
}

func LintDiagnostics(in app.Inputs, check string) []diag.Entry {
	reg, cfg, entries := loadLintDocs(in)
	if len(entries) > 0 {
		return entries
	}

	supportedSections := map[string]struct{}{}
	supportedLanguages := map[string]struct{}{}
	requireAllLanguages := true
	for _, ks := range reg.DesignRegistry.KeySchemaRegistry {
		for _, s := range ks.SupportedCommandSections {
			supportedSections[s] = struct{}{}
		}
		for _, l := range ks.SupportedLanguages {
			supportedLanguages[l] = struct{}{}
		}
		requireAllLanguages = ks.TranslationPolicy.RequireAllSupportedLanguages
		break
	}

	out := make([]diag.Entry, 0)
	runSections := check == "" || check == "sections"
	runTranslations := check == "" || check == "translations"
	runKeys := check == "" || check == "keys"
	runGraph := check == "" || check == "graph"
	runPatterns := check == "" || check == "patterns"
	runSchema := check == "" || check == "schema"
	runConfig := check == "" || check == "config"

	if runSchema {
		out = append(out, lintSchemaGovernance(reg)...)
	}
	if runConfig {
		out = append(out, lintConfigQuality(cfg)...)
	}

	if runSections {
		out = append(out, lintSections(cfg, supportedSections)...)
	}
	if runTranslations {
		out = append(out, lintTranslations(cfg, supportedLanguages, requireAllLanguages)...)
	}
	if runKeys || runGraph {
		for _, ks := range reg.DesignRegistry.KeySchemaRegistry {
			expected, mandatory, graphEntries := deriveExpectedKeys(ks.RootLabels, ks.NodesByLabel)
			if runGraph {
				out = append(out, graphEntries...)
			}
			if runKeys {
				out = append(out, lintKeys(cfg, expected, mandatory)...)
			}
			break
		}
	}
	if runPatterns {
		out = append(out, lintPatterns(reg)...)
	}
	diag.Sort(out)
	return out
}

func lintSections(cfg lintConfig, supported map[string]struct{}) []diag.Entry {
	out := make([]diag.Entry, 0)
	for _, v := range cfg.Validations {
		for _, c := range v.Commands {
			for section := range c.Args {
				if _, ok := supported[section]; ok {
					continue
				}
				out = append(out, diag.Entry{
					Stage:    "lint",
					ID:       "LNT-0001",
					Severity: diag.SeverityError,
					Message:  fmt.Sprintf("unsupported validation section %q for key %q", section, v.Key),
				})
			}
		}
	}
	return out
}

func lintTranslations(cfg lintConfig, supported map[string]struct{}, requireAll bool) []diag.Entry {
	if !requireAll {
		return nil
	}
	langs := make([]string, 0, len(supported))
	for l := range supported {
		langs = append(langs, l)
	}
	sort.Strings(langs)
	out := make([]diag.Entry, 0)
	for _, e := range cfg.I18nEntries {
		for _, lang := range langs {
			if _, ok := e.Translations[lang]; ok {
				continue
			}
			out = append(out, diag.Entry{
				Stage:    "lint",
				ID:       "LNT-0002",
				Severity: diag.SeverityError,
				Message:  fmt.Sprintf("missing required translation language %q for key %q", lang, e.Key),
			})
		}
	}
	return out
}

type expectedByKind struct {
	I18n      map[string]struct{}
	Text      map[string]struct{}
	Validator map[string]struct{}
}

type mandatoryByKind struct {
	I18n      map[string]struct{}
	Text      map[string]struct{}
	Validator map[string]struct{}
}

func deriveExpectedKeys(rootLabels []string, nodes map[string]struct {
	Label       string   `cue:"label"`
	Kind        string   `cue:"kind"`
	Mandatory   bool     `cue:"mandatory"`
	ChildLabels []string `cue:"childLabels"`
}) (expectedByKind, mandatoryByKind, []diag.Entry) {
	expected := expectedByKind{
		I18n:      map[string]struct{}{},
		Text:      map[string]struct{}{},
		Validator: map[string]struct{}{},
	}
	mandatory := mandatoryByKind{
		I18n:      map[string]struct{}{},
		Text:      map[string]struct{}{},
		Validator: map[string]struct{}{},
	}
	out := make([]diag.Entry, 0)
	visiting := map[string]bool{}
	visited := map[string]bool{}
	seenKeys := map[string]string{}

	var walk func(string, []string)
	walk = func(label string, path []string) {
		if visiting[label] {
			out = append(out, diag.Entry{
				Stage:    "lint",
				ID:       "LNT-0006",
				Severity: diag.SeverityError,
				Message:  fmt.Sprintf("graph cycle detected at label %q", label),
			})
			return
		}
		node, ok := nodes[label]
		if !ok {
			out = append(out, diag.Entry{
				Stage:    "lint",
				ID:       "LNT-0008",
				Severity: diag.SeverityError,
				Message:  fmt.Sprintf("child label %q not found in nodesByLabel", label),
			})
			return
		}
		if visited[label] {
			// Already validated from another path; still traverse path-specific key leaves below in caller recursion.
		}
		visiting[label] = true
		nextPath := append(path, node.Label)
		if node.Kind != "branch" {
			key := toPathCamel(nextPath)
			if prev, exists := seenKeys[key]; exists && prev != strings.Join(nextPath, ".") {
				out = append(out, diag.Entry{
					Stage:    "lint",
					ID:       "LNT-0007",
					Severity: diag.SeverityError,
					Message:  fmt.Sprintf("generated key collision for %q", key),
				})
			}
			seenKeys[key] = strings.Join(nextPath, ".")
			switch node.Kind {
			case "i18n":
				expected.I18n[key] = struct{}{}
				if node.Mandatory {
					mandatory.I18n[key] = struct{}{}
				}
			case "text":
				expected.Text[key] = struct{}{}
				if node.Mandatory {
					mandatory.Text[key] = struct{}{}
				}
			case "validator":
				expected.Validator[key] = struct{}{}
				if node.Mandatory {
					mandatory.Validator[key] = struct{}{}
				}
			}
		}
		for _, child := range node.ChildLabels {
			walk(child, nextPath)
		}
		visiting[label] = false
		visited[label] = true
	}
	for _, root := range rootLabels {
		walk(root, nil)
	}
	return expected, mandatory, out
}

func lintKeys(cfg lintConfig, expected expectedByKind, mandatory mandatoryByKind) []diag.Entry {
	out := make([]diag.Entry, 0)
	actualI18n := map[string]struct{}{}
	for _, e := range cfg.I18nEntries {
		actualI18n[e.Key] = struct{}{}
	}
	actualText := map[string]struct{}{}
	for _, e := range cfg.TextEntries {
		actualText[e.Key] = struct{}{}
	}
	actualValidator := map[string]struct{}{}
	for _, e := range cfg.Validations {
		actualValidator[e.Key] = struct{}{}
	}

	out = append(out, lintMandatory("i18nEntries", mandatory.I18n, actualI18n)...)
	out = append(out, lintMandatory("textEntries", mandatory.Text, actualText)...)
	out = append(out, lintMandatory("validations", mandatory.Validator, actualValidator)...)
	out = append(out, lintKeySet("i18nEntries", expected.I18n, actualI18n)...)
	out = append(out, lintKeySet("textEntries", expected.Text, actualText)...)
	out = append(out, lintKeySet("validations", expected.Validator, actualValidator)...)
	return out
}

func lintMandatory(section string, required, actual map[string]struct{}) []diag.Entry {
	out := make([]diag.Entry, 0)
	for key := range required {
		if _, ok := actual[key]; ok {
			continue
		}
		out = append(out, diag.Entry{
			Stage:    "lint",
			ID:       "LNT-0003",
			Severity: diag.SeverityError,
			Message:  fmt.Sprintf("missing mandatory key %q in %s", key, section),
		})
	}
	return out
}

func lintKeySet(section string, expected, actual map[string]struct{}) []diag.Entry {
	out := make([]diag.Entry, 0)
	for key := range expected {
		if _, ok := actual[key]; ok {
			continue
		}
		out = append(out, diag.Entry{
			Stage:    "lint",
			ID:       "LNT-0004",
			Severity: diag.SeverityError,
			Message:  fmt.Sprintf("missing expected key %q in %s", key, section),
		})
	}
	for key := range actual {
		if _, ok := expected[key]; ok {
			continue
		}
		out = append(out, diag.Entry{
			Stage:    "lint",
			ID:       "LNT-0005",
			Severity: diag.SeverityError,
			Message:  fmt.Sprintf("unexpected key %q in %s", key, section),
		})
	}
	return out
}

func toPathCamel(path []string) string {
	if len(path) == 0 {
		return ""
	}
	out := path[0]
	for _, p := range path[1:] {
		if p == "" {
			continue
		}
		out += strings.ToUpper(p[:1]) + p[1:]
	}
	return out
}

func lintPatterns(reg lintRegistry) []diag.Entry {
	out := make([]diag.Entry, 0)
	allowedTokens := map[string]struct{}{
		"<domain>": {},
		"<locale>": {},
	}
	for _, c := range reg.DesignRegistry.GeneratorCapabilities {
		tokens := extractPatternTokens(c.ArtifactPattern)
		for _, tok := range tokens {
			if _, ok := allowedTokens[tok]; ok {
				continue
			}
			out = append(out, diag.Entry{
				Stage:    "lint",
				ID:       "LNT-0009",
				Severity: diag.SeverityError,
				Message:  fmt.Sprintf("unsupported artifact pattern token %q for target %q", tok, c.Target),
			})
		}
		if c.Target == "arb.json" && !containsToken(tokens, "<locale>") {
			out = append(out, diag.Entry{
				Stage:    "lint",
				ID:       "LNT-0010",
				Severity: diag.SeverityError,
				Message:  "artifact pattern for arb.json target must include <locale>",
			})
		}
		if c.Target != "arb.json" && !containsToken(tokens, "<domain>") {
			out = append(out, diag.Entry{
				Stage:    "lint",
				ID:       "LNT-0011",
				Severity: diag.SeverityError,
				Message:  fmt.Sprintf("artifact pattern for target %q must include <domain>", c.Target),
			})
		}
	}
	return out
}

func lintSchemaGovernance(reg lintRegistry) []diag.Entry {
	out := make([]diag.Entry, 0)
	if len(reg.DesignRegistry.KeySchemaRegistry) == 0 {
		out = append(out, diag.Entry{
			Stage:    "lint",
			ID:       "LNT-0012",
			Severity: diag.SeverityError,
			Message:  "keySchemaRegistry must contain at least one schema",
		})
		return out
	}
	for key, ks := range reg.DesignRegistry.KeySchemaRegistry {
		if strings.TrimSpace(ks.Metadata.ID) == "" {
			out = append(out, diag.Entry{
				Stage:    "lint",
				ID:       "LNT-0013",
				Severity: diag.SeverityError,
				Message:  fmt.Sprintf("schema %q metadata.id is required", key),
			})
		}
		if ks.Metadata.ID != "" && ks.Metadata.ID != key {
			out = append(out, diag.Entry{
				Stage:    "lint",
				ID:       "LNT-0014",
				Severity: diag.SeverityError,
				Message:  fmt.Sprintf("schema key %q must match metadata.id %q", key, ks.Metadata.ID),
			})
		}
		if len(ks.SupportedLanguages) == 0 {
			out = append(out, diag.Entry{
				Stage:    "lint",
				ID:       "LNT-0015",
				Severity: diag.SeverityError,
				Message:  fmt.Sprintf("schema %q must declare supportedLanguages", key),
			})
		}
		if len(ks.SupportedCommandSections) == 0 {
			out = append(out, diag.Entry{
				Stage:    "lint",
				ID:       "LNT-0016",
				Severity: diag.SeverityError,
				Message:  fmt.Sprintf("schema %q must declare supportedCommandSections", key),
			})
		}
	}
	return out
}

func lintConfigQuality(cfg lintConfig) []diag.Entry {
	out := make([]diag.Entry, 0)
	out = append(out, lintDuplicateKeys("i18nEntries", collectI18nKeys(cfg))...)
	out = append(out, lintDuplicateKeys("textEntries", collectTextKeys(cfg))...)
	out = append(out, lintDuplicateKeys("validations", collectValidationKeys(cfg))...)
	return out
}

func lintDuplicateKeys(section string, keys []string) []diag.Entry {
	seen := map[string]int{}
	for _, k := range keys {
		seen[k]++
	}
	out := make([]diag.Entry, 0)
	for key, count := range seen {
		if count > 1 {
			out = append(out, diag.Entry{
				Stage:    "lint",
				ID:       "LNT-0017",
				Severity: diag.SeverityError,
				Message:  fmt.Sprintf("duplicate key %q in %s", key, section),
			})
		}
	}
	return out
}

func collectI18nKeys(cfg lintConfig) []string {
	keys := make([]string, 0, len(cfg.I18nEntries))
	for _, e := range cfg.I18nEntries {
		keys = append(keys, e.Key)
	}
	return keys
}

func collectTextKeys(cfg lintConfig) []string {
	keys := make([]string, 0, len(cfg.TextEntries))
	for _, e := range cfg.TextEntries {
		keys = append(keys, e.Key)
	}
	return keys
}

func collectValidationKeys(cfg lintConfig) []string {
	keys := make([]string, 0, len(cfg.Validations))
	for _, e := range cfg.Validations {
		keys = append(keys, e.Key)
	}
	return keys
}

func extractPatternTokens(pattern string) []string {
	tokens := make([]string, 0)
	for i := 0; i < len(pattern); i++ {
		if pattern[i] != '<' {
			continue
		}
		j := i + 1
		for ; j < len(pattern) && pattern[j] != '>'; j++ {
		}
		if j < len(pattern) && pattern[j] == '>' {
			tokens = append(tokens, pattern[i:j+1])
			i = j
		}
	}
	return tokens
}

func containsToken(tokens []string, token string) bool {
	for _, t := range tokens {
		if t == token {
			return true
		}
	}
	return false
}

func loadLintDocs(in app.Inputs) (lintRegistry, lintConfig, []diag.Entry) {
	var reg lintRegistry
	var cfg lintConfig

	regMerged, entries := mergeCueSources(in.RegistrySchemaPath, in.RegistryPath, "lint")
	if len(entries) > 0 {
		return reg, cfg, entries
	}
	regVal, regErr := compileCue("lint-registry", regMerged)
	if regErr != nil {
		return reg, cfg, []diag.Entry{*regErr}
	}
	if err := regVal.Decode(&reg); err != nil {
		return reg, cfg, []diag.Entry{{
			Stage:    "lint",
			ID:       "LNT-0101",
			Severity: diag.SeverityError,
			Message:  fmt.Sprintf("failed to decode registry: %v", sanitizeCueErr(err)),
			Path:     in.RegistryPath,
		}}
	}

	cfgSchemaSrc, err := os.ReadFile(in.ConfigSchemaPath)
	if err != nil {
		return reg, cfg, []diag.Entry{{Stage: "lint", ID: "LNT-0102", Severity: diag.SeverityError, Message: fmt.Sprintf("failed to read config schema: %v", err), Path: in.ConfigSchemaPath}}
	}
	cfgInputSrc, err := os.ReadFile(in.ConfigPath)
	if err != nil {
		return reg, cfg, []diag.Entry{{Stage: "lint", ID: "LNT-0103", Severity: diag.SeverityError, Message: fmt.Sprintf("failed to read config input: %v", err), Path: in.ConfigPath}}
	}

	cfgSchemaPkg := extractPackage(string(cfgSchemaSrc))
	cfgInputBody := string(cfgInputSrc)
	if extractPackage(cfgInputBody) == "" && cfgSchemaPkg != "" {
		cfgInputBody = "package " + cfgSchemaPkg + "\n\n" + cfgInputBody
	}
	cfgVal, cfgErr := compileCue("lint-config", cfgInputBody)
	if cfgErr != nil {
		return reg, cfg, []diag.Entry{*cfgErr}
	}
	if err := cfgVal.Decode(&cfg); err != nil {
		return reg, cfg, []diag.Entry{{
			Stage:    "lint",
			ID:       "LNT-0104",
			Severity: diag.SeverityError,
			Message:  fmt.Sprintf("failed to decode config: %v", sanitizeCueErr(err)),
			Path:     in.ConfigPath,
		}}
	}
	return reg, cfg, nil
}

func compileCue(stage, src string) (cue.Value, *diag.Entry) {
	ctx := cuecontext.New()
	v := ctx.CompileString(src, cue.Filename(stage+".cue"))
	if err := v.Err(); err != nil {
		return cue.Value{}, &diag.Entry{Stage: "lint", ID: "LNT-0100", Severity: diag.SeverityError, Message: fmt.Sprintf("cue compile failed: %v", sanitizeCueErr(err))}
	}
	return v, nil
}
