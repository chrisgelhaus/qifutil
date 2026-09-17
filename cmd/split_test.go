package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"qifutil/test"
)

// numberedQIF builds one account holding n transactions, one per day.
func numberedQIF(n int) string {
	var b strings.Builder
	b.WriteString("!Account\nNChecking\nTBank\n^\n!Type:Bank\n")
	for i := 0; i < n; i++ {
		fmt.Fprintf(&b, "D1/%d'23\nT-1.00\nPVendor %d\nLFood:Dining\n^\n", i%28+1, i)
	}
	return b.String()
}

// countSplitFiles returns the names of the numbered outputs for an account.
func countSplitFiles(t *testing.T, dir, prefix string) []string {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(dir, prefix))
	if err != nil {
		t.Fatalf("failed to list output files: %v", err)
	}
	return matches
}

// TestSplitDoesNotLeaveAnEmptyTrailingFile covers a record count that lands
// exactly on the boundary. The next file was opened for records that never
// arrived, leaving a file holding nothing but a header.
func TestSplitDoesNotLeaveAnEmptyTrailingFile(t *testing.T) {
	helper := test.NewHelper(t)
	tempDir := helper.CreateTempDir()
	setupTransactionExport(t)

	prevMax := maxRecordsPerFile
	t.Cleanup(func() { maxRecordsPerFile = prevMax })
	maxRecordsPerFile = 5

	_, outputDir := exportInlineQIF(t, helper, tempDir, numberedQIF(10))

	files := countSplitFiles(t, outputDir, "Checking_*.csv")
	if len(files) != 2 {
		t.Errorf("10 records at 5 per file should be 2 files, got %d: %v", len(files), files)
	}
	for _, f := range files {
		content, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("failed to read %s: %v", f, err)
		}
		if strings.Count(strings.TrimSpace(string(content)), "\n") < 1 {
			t.Errorf("%s holds only a header", filepath.Base(f))
		}
	}
}

// TestSplitStillDividesWhenTheCountIsUneven guards the ordinary case.
func TestSplitStillDividesWhenTheCountIsUneven(t *testing.T) {
	helper := test.NewHelper(t)
	tempDir := helper.CreateTempDir()
	setupTransactionExport(t)

	prevMax := maxRecordsPerFile
	t.Cleanup(func() { maxRecordsPerFile = prevMax })
	maxRecordsPerFile = 5

	_, outputDir := exportInlineQIF(t, helper, tempDir, numberedQIF(12))

	files := countSplitFiles(t, outputDir, "Checking_*.csv")
	if len(files) != 3 {
		t.Errorf("12 records at 5 per file should be 3 files, got %d: %v", len(files), files)
	}
	last := filepath.Join(outputDir, "Checking_3.csv")
	helper.AssertFileContains(last, "Vendor 11")
}

// TestBalanceHistorySplitDoesNotLeaveAnEmptyTrailingFile covers the same
// boundary in the other command that splits its output.
func TestBalanceHistorySplitDoesNotLeaveAnEmptyTrailingFile(t *testing.T) {
	helper := test.NewHelper(t)
	tempDir := helper.CreateTempDir()

	prevInput, prevOutput, prevMax := inputFile, outputPath, maxRecordsPerFile
	prevAccounts, prevOpening, prevCurrent := selectedAccounts, openingBalance, currentBalance
	prevStart, prevEnd := startDate, endDate
	t.Cleanup(func() {
		inputFile, outputPath, maxRecordsPerFile = prevInput, prevOutput, prevMax
		selectedAccounts, openingBalance, currentBalance = prevAccounts, prevOpening, prevCurrent
		startDate, endDate = prevStart, prevEnd
	})

	sourceFile := filepath.Join(tempDir, "bh.qif")
	if err := os.WriteFile(sourceFile, []byte(numberedQIF(10)), 0644); err != nil {
		t.Fatalf("failed to write qif fixture: %v", err)
	}

	inputFile = sourceFile
	outputPath = tempDir
	maxRecordsPerFile = 5
	selectedAccounts = "Checking"
	openingBalance = "0"
	currentBalance = ""
	startDate, endDate = "", ""

	helper.CaptureOutput(func() {
		if err := balanceHistoryCmd.RunE(balanceHistoryCmd, []string{}); err != nil {
			t.Fatalf("balance-history failed: %v", err)
		}
	})

	files := countSplitFiles(t, tempDir, "Checking_balance_history_*.csv")
	if len(files) != 2 {
		t.Errorf("10 records at 5 per file should be 2 files, got %d: %v", len(files), files)
	}
}
