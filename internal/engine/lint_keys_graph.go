package engine

import (
	"fmt"
	"strings"

	"github.com/flarebyte/ash-splash-norn/internal/diag"
)

func deriveExpectedKeys(rootLabels []string, nodes map[string]struct {
	Label       string   `cue:"label"`
	Kind        string   `cue:"kind"`
	Mandatory   bool     `cue:"mandatory"`
	ChildLabels []string `cue:"childLabels"`
}) (expectedByKind, mandatoryByKind, []diag.Entry) {
	expected := expectedByKind{I18n: map[string]struct{}{}, Text: map[string]struct{}{}, Validator: map[string]struct{}{}}
	mandatory := mandatoryByKind{I18n: map[string]struct{}{}, Text: map[string]struct{}{}, Validator: map[string]struct{}{}}
	out := make([]diag.Entry, 0)
	visiting := map[string]bool{}
	visited := map[string]bool{}
	seenKeys := map[string]string{}

	var walk func(string, []string)
	walk = func(label string, path []string) {
		if visiting[label] {
			out = append(out, diag.Entry{Stage: "lint", ID: "LNT-0006", Severity: diag.SeverityError, Message: fmt.Sprintf("graph cycle detected at label %q", label)})
			return
		}
		node, ok := nodes[label]
		if !ok {
			out = append(out, diag.Entry{Stage: "lint", ID: "LNT-0008", Severity: diag.SeverityError, Message: fmt.Sprintf("child label %q not found in nodesByLabel", label)})
			return
		}
		visiting[label] = true
		nextPath := append(path, node.Label)
		if node.Kind != "branch" {
			key := toPathCamel(nextPath)
			if prev, exists := seenKeys[key]; exists && prev != strings.Join(nextPath, ".") {
				out = append(out, diag.Entry{Stage: "lint", ID: "LNT-0007", Severity: diag.SeverityError, Message: fmt.Sprintf("generated key collision for %q", key)})
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
		out = append(out, diag.Entry{Stage: "lint", ID: "LNT-0003", Severity: diag.SeverityError, Message: fmt.Sprintf("missing mandatory key %q in %s", key, section)})
	}
	return out
}

func lintKeySet(section string, expected, actual map[string]struct{}) []diag.Entry {
	out := make([]diag.Entry, 0)
	for key := range expected {
		if _, ok := actual[key]; ok {
			continue
		}
		out = append(out, diag.Entry{Stage: "lint", ID: "LNT-0004", Severity: diag.SeverityError, Message: fmt.Sprintf("missing expected key %q in %s", key, section)})
	}
	for key := range actual {
		if _, ok := expected[key]; ok {
			continue
		}
		out = append(out, diag.Entry{Stage: "lint", ID: "LNT-0005", Severity: diag.SeverityError, Message: fmt.Sprintf("unexpected key %q in %s", key, section)})
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
