package testutils

import "time"

// ContextKey is a custom type for context keys to avoid collisions.
type ContextKey string

const (
	// ContextKeyLogger is the key for the logger in the request context.
	ContextKeyLogger ContextKey = "logger"
)

// ContextKeyPathValue returns a ContextKey for a given path value.
func ContextKeyPathValue(key string) ContextKey {
	return ContextKey("pathValue_" + key)
}

// stringPtr is a helper function to return a pointer to a string literal.
func StringPtr(s string) *string {
	return &s
}

// boolPtr is a helper function to return a pointer to a bool literal.
func BoolPtr(b bool) *bool {
	return &b
}

// timePtr is a helper function to return a pointer to a time.Time literal.
func TimePtr(t time.Time) *time.Time {
	return &t
}

// IntPtr is a helper function to return a pointer to an int literal.
func IntPtr(i int) *int {
	return &i
}