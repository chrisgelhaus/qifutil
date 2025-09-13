package cmd

import (
	"fmt"
	"strings"
)

// friendlyError returns a user-friendly error message for common errors
func friendlyError(err error) string {
	msg := err.Error()

	switch {
	case contains(msg, "no such file"):
		return "Could not find the QIF file. Make sure:\n" +
			"1. The file path is correct\n" +
			"2. You included the full path (e.g., C:\\Users\\Name\\Downloads\\file.qif)\n" +
			"3. The file exists in that location\n\n" +
			"💡 Tip: You can drag and drop the file into the command window instead of typing the path"

	case contains(msg, "permission denied"):
		return "Cannot access the file or folder. Make sure:\n" +
			"1. You have permission to access the file\n" +
			"2. The file isn't open in another program\n" +
			"3. You have permission to create files in the output folder"

	case contains(msg, "invalid date"):
		return "The date format is incorrect. Please use:\n" +
			"YYYY-MM-DD format (for example: 2025-09-12)\n\n" +
			"💡 Tip: Dates must include leading zeros (01 instead of 1)"

	default:
		return fmt.Sprintf("An error occurred: %v\n\n"+
			"Need help? Try:\n"+
			"1. Run 'qifutil wizard' for an interactive guide\n"+
			"2. Check our documentation at https://github.com/chrisgelhaus/qifutil\n"+
			"3. Open an issue on GitHub if you think you found a bug", err)
	}
}

func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
