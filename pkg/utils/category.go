package utils

import "strings"

// SplitCategoryAndTag splits a category field into category and tag components
// The function splits on the first '/' only, preserving any subsequent '/' in the tag
func SplitCategoryAndTag(categoryField string) (string, string) {
	// Split the category field by the first '/' only
	idx := strings.Index(categoryField, "/")
	if idx != -1 {
		// Found a slash - split on first occurrence
		return strings.TrimSpace(categoryField[:idx]), strings.TrimSpace(categoryField[idx+1:])
	}
	// No slash found
	return strings.TrimSpace(categoryField), ""
}
