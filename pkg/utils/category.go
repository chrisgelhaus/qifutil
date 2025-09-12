package utils

import "strings"

// SplitCategoryAndTag splits a category field into category and tag components
func SplitCategoryAndTag(categoryField string) (string, string) {
	// Split the category field by '/' if it exists
	parts := strings.Split(categoryField, "/")
	if len(parts) > 1 {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	return strings.TrimSpace(categoryField), ""
}
