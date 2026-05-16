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

type lintRegistry struct {
	DesignRegistry struct {
		KeySchemaRegistry map[string]struct {
			SupportedLanguages       []string `cue:"supportedLanguages"`
			SupportedCommandSections []string `cue:"supportedCommandSections"`
			TranslationPolicy        struct {
				RequireAllSupportedLanguages bool `cue:"requireAllSupportedLanguages"`
			} `cue:"translationPolicy"`
		} `cue:"keySchemaRegistry"`
	} `cue:"designRegistry"`
}

type lintConfig struct {
	I18nEntries []struct {
		Key          string                    `cue:"key"`
		Translations map[string]map[string]any `cue:"translations"`
	} `cue:"i18nEntries"`
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

	if runSections {
		out = append(out, lintSections(cfg, supportedSections)...)
	}
	if runTranslations {
		out = append(out, lintTranslations(cfg, supportedLanguages, requireAllLanguages)...)
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
