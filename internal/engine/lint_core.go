// purpose: Runs lint checks by category and aggregates deterministic diagnostics for CLI consumers.
// responsibilities: Compute active check set, collect schema/config/key/pattern/translation diagnostics, and sort output.
// architecture notes: The orchestration layer intentionally delegates all rule logic to specialized lint modules.

package engine

import (
	"sort"

	"github.com/flarebyte/ash-splash-norn/internal/app"
	"github.com/flarebyte/ash-splash-norn/internal/diag"
)

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
		for _, ks := range reg.DesignRegistry.KeySchemaRegistry {
			out = append(out, lintConfigQuality(cfg, ks)...)
			break
		}
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
				out = append(out, diag.Entry{Stage: "lint", ID: "LNT-0001", Severity: diag.SeverityError, Message: "unsupported validation section \"" + section + "\" for key \"" + v.Key + "\""})
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
			out = append(out, diag.Entry{Stage: "lint", ID: "LNT-0002", Severity: diag.SeverityError, Message: "missing required translation language \"" + lang + "\" for key \"" + e.Key + "\""})
		}
	}
	return out
}
