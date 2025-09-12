package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"qifutil/test"
)

func TestTransactionsCmd(t *testing.T) {
	helper := test.NewHelper(t)

	t.Run("basic export", func(t *testing.T) {
		// Set up test environment
		tempDir := helper.CreateTempDir()
		outputDir := filepath.Join(tempDir, "output")
		os.MkdirAll(outputDir, 0755)

		inputFile := filepath.Join(tempDir, "sample.qif")
		helper.CopyTestData("sample.qif", inputFile)

		// Reset command flags
		selectedAccounts = ""
		startDate = ""
		endDate = ""
		outputFormat = "CSV"

		// Set required flags
		inputFile = inputFile
		outputPath = outputDir

		// Execute command
		output := helper.CaptureOutput(func() {
			transactionsCmd.Run(transactionsCmd, []string{})
		})

		// Verify output files were created
		helper.AssertFileExists(filepath.Join(outputDir, "Checking Account_1.csv"))
		helper.AssertFileExists(filepath.Join(outputDir, "Credit Card_1.csv"))

		// Verify file contents
		helper.AssertFileContains(filepath.Join(outputDir, "Checking Account_1.csv"), "Grocery Store")
		helper.AssertFileContains(filepath.Join(outputDir, "Credit Card_1.csv"), "Electric Company")

		// Verify command output
		helper.AssertOutputContains(output, "Export completed successfully")
	})

	t.Run("account filtering", func(t *testing.T) {
		tempDir := helper.CreateTempDir()
		outputDir := filepath.Join(tempDir, "output")
		os.MkdirAll(outputDir, 0755)

		inputFile := filepath.Join(tempDir, "sample.qif")
		helper.CopyTestData("sample.qif", inputFile)

		// Reset and set flags
		selectedAccounts = "Checking Account"
		startDate = ""
		endDate = ""
		outputFormat = "CSV"
		inputFile = inputFile
		outputPath = outputDir

		output := helper.CaptureOutput(func() {
			transactionsCmd.Run(transactionsCmd, []string{})
		})

		// Verify only Checking Account file was created
		helper.AssertFileExists(filepath.Join(outputDir, "Checking Account_1.csv"))
		if _, err := os.Stat(filepath.Join(outputDir, "Credit Card_1.csv")); !os.IsNotExist(err) {
			t.Error("Credit Card file should not exist when filtering for Checking Account")
		}

		helper.AssertOutputContains(output, "Processed accounts: Checking Account")
	})

	t.Run("date filtering", func(t *testing.T) {
		tempDir := helper.CreateTempDir()
		outputDir := filepath.Join(tempDir, "output")
		os.MkdirAll(outputDir, 0755)

		inputFile := filepath.Join(tempDir, "sample.qif")
		helper.CopyTestData("sample.qif", inputFile)

		// Reset and set flags
		selectedAccounts = ""
		startDate = "2023-01-16"
		endDate = "2023-01-16"
		outputFormat = "CSV"
		inputFile = inputFile
		outputPath = outputDir

		output := helper.CaptureOutput(func() {
			transactionsCmd.Run(transactionsCmd, []string{})
		})

		// Verify date filtered content
		checkingFile := filepath.Join(outputDir, "Checking Account_1.csv")
		helper.AssertFileExists(checkingFile)
		helper.AssertFileContains(checkingFile, "2023-01-16")

		content, _ := os.ReadFile(checkingFile)
		if strings.Contains(string(content), "2023-01-15") {
			t.Error("File should not contain transactions from 2023-01-15")
		}

		helper.AssertOutputContains(output, "Date range: 2023-01-16 to 2023-01-16")
	})

	t.Run("output formats", func(t *testing.T) {
		formats := []string{"CSV", "JSON", "XML"}

		for _, format := range formats {
			t.Run(format, func(t *testing.T) {
				tempDir := helper.CreateTempDir()
				outputDir := filepath.Join(tempDir, "output")
				os.MkdirAll(outputDir, 0755)

				inputFile := filepath.Join(tempDir, "sample.qif")
				helper.CopyTestData("sample.qif", inputFile)

				// Reset and set flags
				selectedAccounts = ""
				startDate = ""
				endDate = ""
				outputFormat = format
				inputFile = inputFile
				outputPath = outputDir

				helper.CaptureOutput(func() {
					transactionsCmd.Run(transactionsCmd, []string{})
				})

				ext := "." + strings.ToLower(format)
				helper.AssertFileExists(filepath.Join(outputDir, "Checking Account_1"+ext))
				helper.AssertFileExists(filepath.Join(outputDir, "Credit Card_1"+ext))
			})
		}
	})
}
