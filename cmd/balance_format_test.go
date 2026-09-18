package cmd

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"qifutil/test"
)

// TestFormatBalanceNeverPrintsNegativeZero covers a running balance that lands
// on approximately zero. Formatting it directly gives "-0.00", which reads as
// wrong in a balance column.
func TestFormatBalanceNeverPrintsNegativeZero(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		want  string
	}{
		{"a hair below zero", -0.0000001, "0.00"},
		{"negative zero", math.Copysign(0, -1), "0.00"},
		{"plain zero", 0, "0.00"},
		{"a real negative", -12.34, "-12.34"},
		{"a real positive", 5000, "5000.00"},
		{"rounds to a cent", -0.006, "-0.01"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatBalance(tt.value); got != tt.want {
				t.Errorf("formatBalance(%v) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

// TestBalanceHistoryHasNoNegativeZero covers the same thing through a real run.
func TestBalanceHistoryHasNoNegativeZero(t *testing.T) {
	helper := test.NewHelper(t)
	tempDir := helper.CreateTempDir()

	prevInput, prevOutput := inputFile, outputPath
	prevAccounts, prevOpening, prevCurrent := selectedAccounts, openingBalance, currentBalance
	prevStart, prevEnd := startDate, endDate
	t.Cleanup(func() {
		inputFile, outputPath = prevInput, prevOutput
		selectedAccounts, openingBalance, currentBalance = prevAccounts, prevOpening, prevCurrent
		startDate, endDate = prevStart, prevEnd
	})

	sourceFile := filepath.Join(tempDir, "sample.qif")
	helper.CopyTestData("sample.qif", sourceFile)
	inputFile = sourceFile
	outputPath = tempDir
	selectedAccounts = "Checking Account"
	currentBalance = "5000.00"
	openingBalance = ""
	startDate, endDate = "", ""

	helper.CaptureOutput(func() {
		if err := balanceHistoryCmd.RunE(balanceHistoryCmd, []string{}); err != nil {
			t.Fatalf("balance-history failed: %v", err)
		}
	})

	content, err := os.ReadFile(filepath.Join(tempDir, "Checking Account_balance_history_1.csv"))
	if err != nil {
		t.Fatalf("failed to read the balance history: %v", err)
	}
	if strings.Contains(string(content), "-0.00") {
		t.Errorf("a balance printed as negative zero:\n%s", content)
	}
}
