// purpose: Validates config-entry quality rules including duplicate keys and meta-args command schemas.
// responsibilities: Check duplicate key sets and validate metaArgs values against compiled snake-knot-picker specs.
// architecture notes: Meta-args compilation is intentionally single-source to mirror runtime argv validation semantics.

package engine

import (
	"fmt"
	"sort"

	"github.com/flarebyte/ash-splash-norn/internal/diag"
	picker "github.com/flarebyte/snake-knot-picker"
)

func lintConfigQuality(cfg lintConfig, ks keySchemaSpec) []diag.Entry {
	out := make([]diag.Entry, 0)
	out = append(out, lintDuplicateKeys("i18nEntries", collectI18nKeys(cfg))...)
	out = append(out, lintDuplicateKeys("textEntries", collectTextKeys(cfg))...)
	out = append(out, lintDuplicateKeys("validations", collectValidationKeys(cfg))...)
	out = append(out, lintMetaArgs(cfg, ks.MetaArgsValidation.Args)...)
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
			out = append(out, diag.Entry{Stage: "lint", ID: "LNT-0017", Severity: diag.SeverityError, Message: fmt.Sprintf("duplicate key %q in %s", key, section)})
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

func lintMetaArgs(cfg lintConfig, args map[string]struct {
	CommandPath []string `cue:"commandPath"`
	AdminOnly   bool     `cue:"adminOnly"`
	Flags       []struct {
		Kind    string     `cue:"kind"`
		Name    string     `cue:"name"`
		Schema  []string   `cue:"schema"`
		Schemas [][]string `cue:"schemas"`
	} `cue:"flags"`
}) []diag.Entry {
	if len(args) == 0 {
		return nil
	}
	sections := make([]string, 0, len(args))
	for s := range args {
		sections = append(sections, s)
	}
	sort.Strings(sections)
	base := args[sections[0]]
	doc := picker.CommandDocument{
		Version:     "1",
		CommandPath: append([]string(nil), base.CommandPath...),
		AdminOnly:   base.AdminOnly,
		Flags:       make([]picker.CommandFlagDef, 0, len(base.Flags)),
	}
	for _, f := range base.Flags {
		doc.Flags = append(doc.Flags, picker.CommandFlagDef{Kind: f.Kind, Name: f.Name, Schema: append([]string(nil), f.Schema...), Schemas: append([][]string(nil), f.Schemas...)})
	}
	compiled, err := picker.CompileCommandDocument(doc)
	if err != nil {
		return []diag.Entry{{Stage: "lint", ID: "LNT-0018", Severity: diag.SeverityError, Message: fmt.Sprintf("metaArgsValidation compile failed: %v", err)}}
	}

	validateEntry := func(section, key string, metaArgs []string, out *[]diag.Entry) {
		if len(metaArgs) == 0 {
			*out = append(*out, diag.Entry{Stage: "lint", ID: "LNT-0019", Severity: diag.SeverityError, Message: fmt.Sprintf("metaArgs is required for key %q in %s", key, section)})
			return
		}
		if _, err := picker.Validate(compiled, metaArgs); err != nil {
			*out = append(*out, diag.Entry{Stage: "lint", ID: "LNT-0020", Severity: diag.SeverityError, Message: fmt.Sprintf("invalid metaArgs for key %q in %s: %v", key, section, err)})
		}
	}

	out := make([]diag.Entry, 0)
	for _, e := range cfg.I18nEntries {
		validateEntry("i18nEntries", e.Key, e.MetaArgs, &out)
	}
	for _, e := range cfg.TextEntries {
		validateEntry("textEntries", e.Key, e.MetaArgs, &out)
	}
	for _, e := range cfg.Validations {
		validateEntry("validations", e.Key, e.MetaArgs, &out)
	}
	return out
}
