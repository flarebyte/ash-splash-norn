// purpose: Validates artifact pattern tokens and path safety rules before file generation occurs.
// responsibilities: Check required/allowed tokens per target and map policy violations to diagnostics.
// architecture notes: Unsafe path rejection and token policy are centralized to keep lint/validate/generate behavior aligned.

package engine

import (
	"path/filepath"
	"strings"

	"github.com/flarebyte/ash-splash-norn/internal/diag"
)

type artifactPatternIssue struct {
	Code    string
	Message string
}

func validateArtifactPatternPolicy(target, pattern string) []artifactPatternIssue {
	out := make([]artifactPatternIssue, 0)
	if strings.TrimSpace(pattern) == "" {
		return []artifactPatternIssue{{Code: "empty_pattern", Message: "artifact pattern must not be empty"}}
	}
	tokens := extractPatternTokens(pattern)
	allowedByTarget := map[string]map[string]struct{}{
		"arb.json": {"<locale>": {}},
		"default":  {"<domain>": {}},
	}
	allowed := allowedByTarget["default"]
	if target == "arb.json" {
		allowed = allowedByTarget["arb.json"]
	}
	for _, tok := range tokens {
		if _, ok := allowed[tok]; ok {
			continue
		}
		out = append(out, artifactPatternIssue{
			Code:    "unsupported_token",
			Message: `unsupported artifact pattern token ` + `"` + tok + `" for target ` + `"` + target + `"`,
		})
	}
	if target == "arb.json" && !containsToken(tokens, "<locale>") {
		out = append(out, artifactPatternIssue{
			Code:    "missing_locale",
			Message: "artifact pattern for arb.json target must include <locale>",
		})
	}
	if target != "arb.json" && !containsToken(tokens, "<domain>") {
		out = append(out, artifactPatternIssue{
			Code:    "missing_domain",
			Message: `artifact pattern for target ` + `"` + target + `"` + " must include <domain>",
		})
	}
	if filepath.IsAbs(pattern) {
		out = append(out, artifactPatternIssue{
			Code:    "absolute_path",
			Message: "artifact pattern must be relative, absolute paths are not allowed",
		})
	}
	for _, seg := range strings.Split(filepath.ToSlash(pattern), "/") {
		if seg == ".." {
			out = append(out, artifactPatternIssue{
				Code:    "path_traversal",
				Message: "artifact pattern must not include path traversal segments (..)",
			})
			break
		}
	}
	return out
}

func artifactPatternIssuesToDiag(stage, idUnsupported, idMissingLocale, idMissingDomain, idUnsafe, target, pattern string) []diag.Entry {
	issues := validateArtifactPatternPolicy(target, pattern)
	out := make([]diag.Entry, 0, len(issues))
	for _, issue := range issues {
		id := idUnsupported
		if issue.Code == "missing_locale" {
			id = idMissingLocale
		}
		if issue.Code == "missing_domain" {
			id = idMissingDomain
		}
		if issue.Code == "absolute_path" || issue.Code == "path_traversal" {
			id = idUnsafe
		}
		out = append(out, diag.Entry{
			Stage:    stage,
			ID:       id,
			Severity: diag.SeverityError,
			Message:  issue.Message,
			Path:     pattern,
		})
	}
	return out
}
