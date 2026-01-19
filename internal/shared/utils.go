package shared

import "slices"

// Contains checks if a string is in a slice of strings.
// This is a utility function used for validation throughout the package.
func Contains(slice []string, item string) bool {
	return slices.Contains(slice, item)
}
