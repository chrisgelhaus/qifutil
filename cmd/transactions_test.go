package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"qifutil/test"
)

func TestMonarchFormat(t *testing.T) {
	helper := test.NewHelper(t)
	tempDir := helper.CreateTempDir()
	outputDir := filepath.Join(tempDir, "output")
	os.MkdirAll(outputDir, 0755)

	sourceFile := filepath.Join(tempDir, "sample.qif")
	helper.CopyTestData("sample.qif", sourceFile)

	// Reset and set flags for MONARCH format
	selectedAccounts = ""
	startDate = ""
	endDate = ""
	outputFormat = "MONARCH"
	csvColumns = DefaultMonarchColumns
	inputFile = sourceFile
	outputPath = outputDir

	helper.CaptureOutput(func() {
		transactionsCmd.Run(transactionsCmd, []string{})
	})

	checkingFile := filepath.Join(outputDir, "Checking Account_1.csv")
	helper.AssertFileExists(checkingFile)

	// Verify MONARCH format has all 8 columns in the correct order
	content, _ := os.ReadFile(checkingFile)
	lines := strings.Split(string(content), "\n")
	if len(lines) > 0 {
		header := lines[0]
		expectedHeader := "Date,Merchant,Category,Account,Original Statement,Notes,Amount,Tags"
		if header != expectedHeader {
			t.Errorf("MONARCH header mismatch.\nExpected: %s\nGot: %s", expectedHeader, header)
		}
	}

	// Verify MONARCH format includes all expected data
	helper.AssertFileContains(checkingFile, "Grocery Store")
	helper.AssertFileContains(checkingFile, "Food:Groceries")
	helper.AssertFileContains(checkingFile, "Checking Account")
	helper.AssertFileContains(checkingFile, "-45.23")
}

func TestCSVCustomColumns(t *testing.T) {
	helper := test.NewHelper(t)
	tempDir := helper.CreateTempDir()
	outputDir := filepath.Join(tempDir, "output")
	os.MkdirAll(outputDir, 0755)

	sourceFile := filepath.Join(tempDir, "sample.qif")
	helper.CopyTestData("sample.qif", sourceFile)

	// Test with custom column selection
	selectedAccounts = ""
	startDate = ""
	endDate = ""
	outputFormat = "CSV"
	csvColumns = "Date,Merchant,Amount,Category"
	inputFile = sourceFile
	outputPath = outputDir

	helper.CaptureOutput(func() {
		transactionsCmd.Run(transactionsCmd, []string{})
	})

	checkingFile := filepath.Join(outputDir, "Checking Account_1.csv")
	helper.AssertFileExists(checkingFile)

	// Verify custom columns are in the correct order
	content, _ := os.ReadFile(checkingFile)
	lines := strings.Split(string(content), "\n")
	if len(lines) > 0 {
		header := lines[0]
		expectedHeader := "Date,Merchant,Amount,Category"
		if header != expectedHeader {
			t.Errorf("CSV custom columns header mismatch.\nExpected: %s\nGot: %s", expectedHeader, header)
		}
	}

	// Verify data is present
	helper.AssertFileContains(checkingFile, "2023-01-15")
	helper.AssertFileContains(checkingFile, "Grocery Store")
	helper.AssertFileContains(checkingFile, "-45.23")
	helper.AssertFileContains(checkingFile, "Food:Groceries")
}

func TestCSVMinimalColumns(t *testing.T) {
	helper := test.NewHelper(t)
	tempDir := helper.CreateTempDir()
	outputDir := filepath.Join(tempDir, "output")
	os.MkdirAll(outputDir, 0755)

	sourceFile := filepath.Join(tempDir, "sample.qif")
	helper.CopyTestData("sample.qif", sourceFile)

	// Test with minimal column set
	selectedAccounts = ""
	startDate = ""
	endDate = ""
	outputFormat = "CSV"
	csvColumns = "Merchant,Amount"
	inputFile = sourceFile
	outputPath = outputDir

	helper.CaptureOutput(func() {
		transactionsCmd.Run(transactionsCmd, []string{})
	})

	checkingFile := filepath.Join(outputDir, "Checking Account_1.csv")
	helper.AssertFileExists(checkingFile)

	// Verify minimal columns
	content, _ := os.ReadFile(checkingFile)
	lines := strings.Split(string(content), "\n")
	if len(lines) > 0 {
		header := lines[0]
		expectedHeader := "Merchant,Amount"
		if header != expectedHeader {
			t.Errorf("CSV minimal columns header mismatch.\nExpected: %s\nGot: %s", expectedHeader, header)
		}
	}
}

func TestCSVDefaultEqualsMonarch(t *testing.T) {
	helper := test.NewHelper(t)
	tempDir := helper.CreateTempDir()
	outputDirA := filepath.Join(tempDir, "output_a")
	outputDirB := filepath.Join(tempDir, "output_b")
	os.MkdirAll(outputDirA, 0755)
	os.MkdirAll(outputDirB, 0755)

	sourceFileA := filepath.Join(tempDir, "sample_a.qif")
	helper.CopyTestData("sample.qif", sourceFileA)

	// Generate output with MONARCH format
	selectedAccounts = ""
	startDate = ""
	endDate = ""
	outputFormat = "MONARCH"
	csvColumns = DefaultMonarchColumns
	inputFile = sourceFileA
	outputPath = outputDirA

	helper.CaptureOutput(func() {
		transactionsCmd.Run(transactionsCmd, []string{})
	})

	// Generate output with CSV format (using defaults)
	sourceFileB := filepath.Join(tempDir, "sample_b.qif")
	helper.CopyTestData("sample.qif", sourceFileB)

	outputFormat = "CSV"
	csvColumns = DefaultMonarchColumns
	inputFile = sourceFileB
	outputPath = outputDirB

	helper.CaptureOutput(func() {
		transactionsCmd.Run(transactionsCmd, []string{})
	})

	// Compare the two outputs
	monarchFile := filepath.Join(outputDirA, "Checking Account_1.csv")
	csvFile := filepath.Join(outputDirB, "Checking Account_1.csv")

	monarchContent, _ := os.ReadFile(monarchFile)
	csvContent, _ := os.ReadFile(csvFile)

	if string(monarchContent) != string(csvContent) {
		t.Error("MONARCH format should produce identical output to CSV with default columns")
	}
}
