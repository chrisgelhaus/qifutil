package cmd

import (
	"path/filepath"
	"strings"
	"testing"

	"qifutil/test"
)

// TestUnknownColumnNameIsRefused covers a misspelled column, which used to
// produce a silently blank column in every row.
func TestUnknownColumnNameIsRefused(t *testing.T) {
	helper := test.NewHelper(t)
	tempDir := helper.CreateTempDir()
	setupTransactionExport(t)

	csvColumns = "Date,Merchent,Amount"

	var err error
	helper.CaptureOutput(func() {
		err = runExportReturningError(t, tempDir)
	})

	if err == nil {
		t.Fatal("an unknown column name should stop the export")
	}
	if !strings.Contains(err.Error(), "Merchent") {
		t.Errorf("the error should name the offending column, got: %v", err)
	}
	if !strings.Contains(err.Error(), "Original Statement") {
		t.Errorf("the error should list the valid names, got: %v", err)
	}
}

func TestKnownColumnNamesAreAccepted(t *testing.T) {
	helper := test.NewHelper(t)
	tempDir := helper.CreateTempDir()
	setupTransactionExport(t)

	// Including the two-word name and some surrounding space.
	csvColumns = "Date, Original Statement ,Amount"

	var err error
	helper.CaptureOutput(func() {
		err = runExportReturningError(t, tempDir)
	})
	if err != nil {
		t.Fatalf("valid column names should be accepted: %v", err)
	}
}

// TestColumnsAreReportedAsIgnoredForOtherFormats covers passing --csvColumns
// with JSON or XML, where it does nothing at all.
func TestColumnsAreReportedAsIgnoredForOtherFormats(t *testing.T) {
	helper := test.NewHelper(t)
	tempDir := helper.CreateTempDir()
	setupTransactionExport(t)

	outputFormat = "JSON"
	csvColumns = "Date,Merchant"

	output, _ := exportInlineQIF(t, helper, tempDir, mixedCategoriesQIF)
	if !strings.Contains(output, "--csvColumns") {
		t.Errorf("the export should say the flag does nothing here:\n%s", output)
	}
}

func TestColumnsAreNotReportedForCSV(t *testing.T) {
	helper := test.NewHelper(t)
	tempDir := helper.CreateTempDir()
	setupTransactionExport(t)

	outputFormat = "CSV"
	csvColumns = "Date,Merchant"

	output, _ := exportInlineQIF(t, helper, tempDir, mixedCategoriesQIF)
	if strings.Contains(output, "--csvColumns") {
		t.Errorf("nothing to warn about when the flag applies:\n%s", output)
	}
}

// runExportReturningError runs the export over the sample file and hands back
// whatever it reports.
func runExportReturningError(t *testing.T, tempDir string) error {
	t.Helper()
	helper := test.NewHelper(t)
	sourceFile := filepath.Join(tempDir, "sample.qif")
	helper.CopyTestData("sample.qif", sourceFile)
	inputFile = sourceFile
	outputPath = filepath.Join(tempDir, "output")
	return transactionsCmd.RunE(transactionsCmd, []string{})
}
