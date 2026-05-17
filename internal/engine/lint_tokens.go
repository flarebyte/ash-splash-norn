// purpose: Provides lightweight token extraction helpers used by artifact-pattern policy checks.
// responsibilities: Scan placeholder tokens and test token membership in pattern strings.
// architecture notes: Simple local helpers are kept separate to avoid coupling policy logic to parser details.

package engine

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
