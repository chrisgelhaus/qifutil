package utils

import (
	"runtime"
	"strings"
)

// sanitizeFilePath cleans and validates a file path
func sanitizePath(path string) string {
	// Remove PowerShell artifacts
	path = strings.TrimSpace(path)
	path = strings.TrimPrefix(path, "& ") // Remove PowerShell invoke operator
	path = strings.Trim(path, "'\"")      // Remove both single and double quotes

	// Convert to proper path separators for the OS
	if runtime.GOOS == "windows" {
		path = strings.ReplaceAll(path, "/", "\\")
	} else {
		path = strings.ReplaceAll(path, "\\", "/")
	}

	return path
}
