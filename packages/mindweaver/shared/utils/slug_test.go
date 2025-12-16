package utils

import "testing"

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple lowercase",
			input:    "my collection",
			expected: "my-collection",
		},
		{
			name:     "mixed case",
			input:    "My Collection",
			expected: "my-collection",
		},
		{
			name:     "with special characters",
			input:    "Project 2024!",
			expected: "project-2024",
		},
		{
			name:     "with ampersand",
			input:    "Café & Bar",
			expected: "café-bar",
		},
		{
			name:     "multiple spaces",
			input:    "my   collection",
			expected: "my-collection",
		},
		{
			name:     "leading and trailing spaces",
			input:    "  my collection  ",
			expected: "my-collection",
		},
		{
			name:     "only special characters",
			input:    "!!!",
			expected: "collection",
		},
		{
			name:     "with hyphens already",
			input:    "my-collection",
			expected: "my-collection",
		},
		{
			name:     "unicode characters",
			input:    "東京 Notes",
			expected: "東京-notes",
		},
		{
			name:     "numbers",
			input:    "2024 Q1",
			expected: "2024-q1",
		},
		{
			name:     "complex case",
			input:    "My Project's: 2024 (Q1)",
			expected: "my-projects-2024-q1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateSlug(tt.input)
			if result != tt.expected {
				t.Errorf("GenerateSlug(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsValidSlug(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid slug",
			input:    "my-collection",
			expected: true,
		},
		{
			name:     "valid single word",
			input:    "collection",
			expected: true,
		},
		{
			name:     "valid with numbers",
			input:    "project-2024",
			expected: true,
		},
		{
			name:     "invalid uppercase",
			input:    "My-Collection",
			expected: false,
		},
		{
			name:     "invalid leading hyphen",
			input:    "-collection",
			expected: false,
		},
		{
			name:     "invalid trailing hyphen",
			input:    "collection-",
			expected: false,
		},
		{
			name:     "invalid consecutive hyphens",
			input:    "my--collection",
			expected: false,
		},
		{
			name:     "invalid special characters",
			input:    "my_collection",
			expected: false,
		},
		{
			name:     "invalid spaces",
			input:    "my collection",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidSlug(tt.input)
			if result != tt.expected {
				t.Errorf("IsValidSlug(%q) = %v; want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizePathComponent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple string",
			input:    "my collection",
			expected: "my collection",
		},
		{
			name:     "with forward slash",
			input:    "my/collection",
			expected: "my-collection",
		},
		{
			name:     "with leading/trailing spaces",
			input:    "  my collection  ",
			expected: "my collection",
		},
		{
			name:     "multiple slashes",
			input:    "my//collection",
			expected: "my--collection",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizePathComponent(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizePathComponent(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}
