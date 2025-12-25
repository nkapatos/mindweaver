package notes

// ============================================================================
// Shared Test Helpers
// ============================================================================
// Helper functions used across handler tests in this package.

// stringPtr returns a pointer to the given string.
func stringPtr(s string) *string {
	return &s
}

// int64Ptr returns a pointer to the given int64.
func int64Ptr(i int64) *int64 {
	return &i
}

// boolPtr returns a pointer to the given bool.
func boolPtr(b bool) *bool {
	return &b
}
