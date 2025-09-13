package utils

import (
	"path/filepath"
	"runtime"
	"strings"
)

// CleanPath sanitizes a file path, handling PowerShell and other shell artifacts
func CleanPath(path string) string {
	// Remove PowerShell and general shell artifacts
	path = strings.TrimSpace(path)
	path = strings.TrimPrefix(path, "& ") // Remove PowerShell invoke operator
	path = strings.Trim(path, "'\"")      // Remove both single and double quotes

	// Normalize path separators for the current OS
	if runtime.GOOS == "windows" {
		path = strings.ReplaceAll(path, "/", "\\")
	} else {
		path = strings.ReplaceAll(path, "\\", "/")
	}

	// Try to convert to absolute path
	if !filepath.IsAbs(path) {
		if abs, err := filepath.Abs(path); err == nil {
			path = abs
		}
	}

	return path
}
