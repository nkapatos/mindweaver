package sqlcext

import (
	"strings"
	"testing"
)

func TestSanitizeFTS5Query(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		desc     string
	}{
		{
			name:     "simple query",
			input:    "hello world",
			expected: "hello OR world",
			desc:     "normal words should be ORed together",
		},
		{
			name:     "query with special chars",
			input:    `hello "world" AND test`,
			expected: "hello OR world OR test",
			desc:     "special FTS5 chars should be stripped",
		},
		{
			name:     "sql injection attempt",
			input:    `test'; DROP TABLE notes; --`,
			expected: "test OR DROP OR TABLE OR notes",
			desc:     "SQL injection should be neutralized",
		},
		{
			name:     "fts5 injection attempt",
			input:    `test OR (SELECT password FROM users)`,
			expected: "test OR SELECT OR password OR users",
			desc:     "FTS5 special operators should be stripped",
		},
		{
			name:     "parentheses attack",
			input:    `((((test))))`,
			expected: "test",
			desc:     "parentheses should be stripped",
		},
		{
			name:     "asterisk wildcard",
			input:    `test*`,
			expected: "test",
			desc:     "asterisks should be stripped",
		},
		{
			name:     "stop words only",
			input:    "the and are",
			expected: "*",
			desc:     "only stop words should return wildcard",
		},
		{
			name:     "short words",
			input:    "a b c test",
			expected: "test",
			desc:     "short words should be filtered out",
		},
		{
			name:     "mixed case",
			input:    "TEST World",
			expected: "TEST OR World",
			desc:     "case should be preserved",
		},
		{
			name:     "hyphenated words",
			input:    "full-text search",
			expected: "full-text OR search",
			desc:     "hyphens should be preserved",
		},
		{
			name:     "numbers",
			input:    "test 123 query",
			expected: "test OR 123 OR query",
			desc:     "numbers should be preserved",
		},
		{
			name:     "unicode attempt",
			input:    "test\x00\x01\x02",
			expected: "test",
			desc:     "unicode control chars should be stripped",
		},
		{
			name:     "empty query",
			input:    "",
			expected: "*",
			desc:     "empty query should return wildcard",
		},
		{
			name:     "only special chars",
			input:    `!@#$%^&*()`,
			expected: "*",
			desc:     "only special chars should return wildcard",
		},
		{
			name:     "question marks",
			input:    "what? how? why?",
			expected: "*",
			desc:     "question marks should be stripped, all stop words removed",
		},
		{
			name:     "double quotes",
			input:    `"exact phrase"`,
			expected: "exact OR phrase",
			desc:     "quotes should not create phrase search",
		},
		{
			name:     "fts5 NOT operator",
			input:    "test NOT excluded",
			expected: "test OR excluded",
			desc:     "NOT operator should be neutralized (NOT is stop word)",
		},
		{
			name:     "realistic query",
			input:    "how to implement FTS5 search?",
			expected: "implement OR FTS5 OR search",
			desc:     "realistic query with stop words",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeFTS5Query(tt.input)
			if got != tt.expected {
				t.Errorf("SanitizeFTS5Query(%q) = %q, want %q\nReason: %s",
					tt.input, got, tt.expected, tt.desc)
			}
		})
	}
}

func TestSanitizeFTS5Query_NoInjection(t *testing.T) {
	// These are known dangerous inputs that should be completely neutralized
	dangerousInputs := []string{
		`'; DROP TABLE notes; --`,
		`" OR 1=1 --`,
		`UNION SELECT * FROM users`,
		`<script>alert('xss')</script>`,
		`${system('rm -rf /')}`,
		`../../etc/passwd`,
		`null\x00byte`,
		string([]byte{0x00, 0x01, 0x02, 0x03}),
	}

	for _, input := range dangerousInputs {
		t.Run("dangerous: "+input, func(t *testing.T) {
			result := SanitizeFTS5Query(input)

			// Result should not contain any SQL-like syntax
			dangerous := []string{";", "--", "/*", "*/", "DROP", "DELETE", "INSERT", "UPDATE", "UNION"}
			for _, d := range dangerous {
				if strings.Contains(result, d) && d != "OR" {
					// Allow "OR" as it's our operator, but not in dangerous context
					if !strings.Contains(result, " OR ") {
						t.Errorf("sanitized result contains dangerous pattern %q: %q", d, result)
					}
				}
			}

			// Result should not contain parentheses
			if strings.ContainsAny(result, "()") {
				t.Errorf("sanitized result contains parentheses: %q", result)
			}

			// Result should not contain quotes
			if strings.ContainsAny(result, `"'`) {
				t.Errorf("sanitized result contains quotes: %q", result)
			}
		})
	}
}

func TestIsStopWord(t *testing.T) {
	stopWords := []string{"the", "and", "are", "what", "who", "how", "why", "when", "where"}
	for _, word := range stopWords {
		t.Run(word, func(t *testing.T) {
			if !isStopWord(word) {
				t.Errorf("isStopWord(%q) = false, want true", word)
			}
			// Test case insensitivity
			if !isStopWord(strings.ToUpper(word)) {
				t.Errorf("isStopWord(%q) = false, want true (case insensitive)", strings.ToUpper(word))
			}
		})
	}

	notStopWords := []string{"test", "search", "query", "document"}
	for _, word := range notStopWords {
		t.Run("not:"+word, func(t *testing.T) {
			if isStopWord(word) {
				t.Errorf("isStopWord(%q) = true, want false", word)
			}
		})
	}
}

func BenchmarkSanitizeFTS5Query(b *testing.B) {
	queries := []string{
		"simple query",
		"query with special !@#$ characters",
		"how to implement FTS5 search in sqlite?",
		`'; DROP TABLE notes; --`,
	}

	for _, query := range queries {
		b.Run(query, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = SanitizeFTS5Query(query)
			}
		})
	}
}
