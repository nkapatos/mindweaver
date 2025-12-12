package utils

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	// nonAlphanumeric matches any character that is not alphanumeric or space
	nonAlphanumeric = regexp.MustCompile(`[^\p{L}\p{N}\s-]+`)
	// multipleHyphens matches multiple consecutive hyphens
	multipleHyphens = regexp.MustCompile(`-+`)
)

// GenerateSlug converts a string into a URL-safe slug.
// It handles unicode characters, lowercases, replaces spaces with hyphens,
// and removes special characters.
//
// Examples:
//   - "My Collection" -> "my-collection"
//   - "Project 2024!" -> "project-2024"
//   - "Café & Bar" -> "caf-bar"
//   - "東京 Notes" -> "notes"
func GenerateSlug(s string) string {
	// Convert to lowercase
	slug := strings.ToLower(s)

	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove non-alphanumeric characters (except hyphens)
	slug = nonAlphanumeric.ReplaceAllString(slug, "")

	// Replace multiple consecutive hyphens with a single hyphen
	slug = multipleHyphens.ReplaceAllString(slug, "-")

	// Trim hyphens from both ends
	slug = strings.Trim(slug, "-")

	// If the slug is empty after cleaning (e.g., all special characters),
	// generate a fallback
	if slug == "" {
		slug = "collection"
	}

	return slug
}

// NormalizePathComponent normalizes a path component for storage.
// Unlike GenerateSlug, this preserves the original casing and more characters,
// but ensures it's safe for path construction.
func NormalizePathComponent(s string) string {
	// Trim whitespace
	normalized := strings.TrimSpace(s)

	// Replace forward slashes (path separators) with hyphens to avoid path confusion
	normalized = strings.ReplaceAll(normalized, "/", "-")

	return normalized
}

// IsValidSlug checks if a string is a valid slug format.
// Valid slugs:
//   - Contain only lowercase letters, numbers, and hyphens
//   - Do not start or end with a hyphen
//   - Do not contain consecutive hyphens
func IsValidSlug(s string) bool {
	if s == "" {
		return false
	}

	// Check if it starts or ends with a hyphen
	if strings.HasPrefix(s, "-") || strings.HasSuffix(s, "-") {
		return false
	}

	// Check if it contains consecutive hyphens
	if strings.Contains(s, "--") {
		return false
	}

	// Check if all characters are valid (lowercase alphanumeric or hyphen)
	for _, r := range s {
		if !unicode.IsLower(r) && !unicode.IsDigit(r) && r != '-' {
			return false
		}
	}

	return true
}
