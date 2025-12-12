package sqlcext

import (
	"strings"
)

// SanitizeFTS5Query escapes special FTS5 characters that could cause syntax errors.
//
// FTS5 special characters: " * ( ) AND OR NOT
// Also strips question marks which can cause syntax errors in certain positions.
//
// SECURITY: This function extracts meaningful words and ORs them together,
// preventing any FTS5 syntax injection. All special characters are removed,
// and the result is used in parameterized queries with ? placeholders.
//
// Strategy: Extract meaningful words and OR them together.
// This finds documents containing ANY of the keywords (more flexible than phrase search).
func SanitizeFTS5Query(query string) string {
	// Remove special characters that could break FTS5 syntax
	clean := ""
	for _, r := range query {
		// Keep alphanumeric, spaces, and hyphens
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == ' ' || r == '-' {
			clean += string(r)
		} else {
			// Replace special chars with space
			clean += " "
		}
	}

	// Split into words
	words := make([]string, 0)
	for _, word := range strings.Fields(clean) {
		// Skip short words and common stop words
		if len(word) >= 3 && !isStopWord(word) {
			words = append(words, word)
		}
	}

	// If no meaningful words, return a safe wildcard
	if len(words) == 0 {
		return "*"
	}

	// Join with OR for flexible matching
	return strings.Join(words, " OR ")
}

// isStopWord checks if a word is too common to be useful in search.
// These are filtered out to improve search quality and performance.
func isStopWord(word string) bool {
	word = strings.ToLower(word)
	stopWords := map[string]bool{
		"the": true, "and": true, "are": true, "what": true, "who": true,
		"how": true, "why": true, "when": true, "where": true, "which": true,
		"that": true, "this": true, "with": true, "for": true, "from": true,
		"was": true, "were": true, "been": true, "have": true, "has": true,
		"had": true, "can": true, "could": true, "will": true, "would": true,
		"shall": true, "should": true, "may": true, "might": true, "must": true,
		"not": true, // FTS5 operator
	}
	return stopWords[word]
}
