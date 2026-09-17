package utils

import "strings"

// forbiddenInFileNames lists the characters Windows does not allow in a file
// name. The colon matters most: NTFS reads "Fidelity: Roth_1.csv" as an
// alternate data stream on a file called "Fidelity", so the export appears to
// succeed while the data becomes invisible.
const forbiddenInFileNames = `<>:"/\\|?*`

// SanitizeFileName replaces every character that cannot appear in a file name
// with an underscore, one for one, so the result still resembles the name it
// came from. An empty name yields a usable placeholder rather than a file
// called only by its suffix.
func SanitizeFileName(name string) string {
	sanitized := strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f || strings.ContainsRune(forbiddenInFileNames, r) {
			return '_'
		}
		return r
	}, name)

	if strings.TrimSpace(sanitized) == "" {
		return "account"
	}
	return sanitized
}
