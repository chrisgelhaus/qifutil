package cmd

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"qifutil/test"
)

// TestWizardForwardsMappingFiles covers the wizard handing the mapping files it
// collected to the export. The paths arrive as parameters, so reading the
// package-level mapping variables instead silently drops them.
func TestWizardForwardsMappingFiles(t *testing.T) {
	helper := test.NewHelper(t)
	tempDir := helper.CreateTempDir()
	setupTransactionExport(t)

	sourceFile := filepath.Join(tempDir, "sample.qif")
	helper.CopyTestData("sample.qif", sourceFile)
	categoryMap := writeMappingFile(t, tempDir, "categories.csv", "\"Food:Groceries\",\"Everyday\"\n")

	outputDir := filepath.Join(tempDir, "output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("failed to create output dir: %v", err)
	}
	inputFile = sourceFile
	outputPath = outputDir

	// The wizard collects mapping paths locally and passes them in; the package
	// level variables stay empty, exactly as on a fresh wizard run.
	reader := bufio.NewReader(strings.NewReader("\n"))
	helper.CaptureOutput(func() {
		executeConversion(reader, true, false, "", "", false,
			categoryMap, "", "", "", false, true)
	})

	checkingFile := filepath.Join(outputDir, "Checking Account_1.csv")
	helper.AssertFileExists(checkingFile)

	line := lineContaining(t, checkingFile, "Grocery Store")
	if !strings.Contains(line, "\"Everyday\"") {
		t.Errorf("wizard did not apply the category mapping it was given, got: %s", line)
	}
}
