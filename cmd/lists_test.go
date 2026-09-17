package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"qifutil/test"
)

// listFixtureQIF exercises what the old single pattern could not read: a record
// whose payee precedes its number, a record missing the optional U and C lines,
// a split carrying categories of its own, and a 1900s date.
const listFixtureQIF = `!Account
NChecking
TBank
^
!Type:Bank
D1/5'23
U-10.00
T-10.00
CX
POut Of Order Payee
NCheque 1234
LGroceries:Weekly/Vacation
^
D1/6'23
T-20.00
PSparse Payee
LSparse:Category
^
D1/7'23
T-90.00
PSplit Payee
LTop:Level
SSplit:Only:Category/SplitTag
Ea split line
$-90.00
^
D3/15/99
T-5.00
PLast Century Payee
LOld:Category
^
`

// runListCommand runs one of the list-style export commands over the fixture
// and returns the contents of the file it wrote.
func runListCommand(t *testing.T, cmdName string) string {
	helper := test.NewHelper(t)
	tempDir := helper.CreateTempDir()

	prevInput, prevOutput, prevFormat := inputFile, outputPath, outputFormat
	t.Cleanup(func() { inputFile, outputPath, outputFormat = prevInput, prevOutput, prevFormat })

	sourceFile := filepath.Join(tempDir, "fixture.qif")
	if err := os.WriteFile(sourceFile, []byte(listFixtureQIF), 0644); err != nil {
		t.Fatalf("failed to write qif fixture: %v", err)
	}
	inputFile = sourceFile
	outputPath = tempDir
	outputFormat = "CSV"

	var target, outName string
	switch cmdName {
	case "categories":
		categoryOutputFile, outName = "categories.csv", "categories.csv"
	case "payees":
		payeeOutputFile, outName = "payees.csv", "payees.csv"
	case "tags":
		tagsOutputFile, outName = "tags.csv", "tags.csv"
	default:
		t.Fatalf("unknown command %q", cmdName)
	}
	target = filepath.Join(tempDir, outName)

	helper.CaptureOutput(func() {
		switch cmdName {
		case "categories":
			categoriesCmd.Run(categoriesCmd, []string{})
		case "payees":
			payeesCmd.Run(payeesCmd, []string{})
		case "tags":
			tagsCmd.Run(tagsCmd, []string{})
		}
	})

	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("failed to read %s: %v", target, err)
	}
	return string(content)
}

func TestPayeesIncludeRecordsTheOldPatternMissed(t *testing.T) {
	got := runListCommand(t, "payees")
	for _, payee := range []string{
		"Out Of Order Payee", "Sparse Payee", "Split Payee", "Last Century Payee",
	} {
		if !strings.Contains(got, payee) {
			t.Errorf("payee list is missing %q:\n%s", payee, got)
		}
	}
}

func TestCategoriesIncludeRecordsTheOldPatternMissed(t *testing.T) {
	got := runListCommand(t, "categories")
	for _, category := range []string{
		"Groceries:Weekly", "Sparse:Category", "Top:Level", "Old:Category",
	} {
		if !strings.Contains(got, category) {
			t.Errorf("category list is missing %q:\n%s", category, got)
		}
	}
}

func TestCategoriesIncludeSplitCategories(t *testing.T) {
	got := runListCommand(t, "categories")
	if !strings.Contains(got, "Split:Only:Category") {
		t.Errorf("a category used only by a split line is missing:\n%s", got)
	}
}

func TestTagsIncludeTagsFromTransactionsAndSplits(t *testing.T) {
	got := runListCommand(t, "tags")
	for _, tag := range []string{"Vacation", "SplitTag"} {
		if !strings.Contains(got, tag) {
			t.Errorf("tag list is missing %q:\n%s", tag, got)
		}
	}
}

func TestAccountStatsCountsRecordsTheOldPatternMissed(t *testing.T) {
	helper := test.NewHelper(t)
	tempDir := helper.CreateTempDir()

	prevInput, prevAccounts := inputFile, selectedAccounts
	t.Cleanup(func() { inputFile, selectedAccounts = prevInput, prevAccounts })

	sourceFile := filepath.Join(tempDir, "fixture.qif")
	if err := os.WriteFile(sourceFile, []byte(listFixtureQIF), 0644); err != nil {
		t.Fatalf("failed to write qif fixture: %v", err)
	}
	inputFile = sourceFile
	selectedAccounts = ""

	output := helper.CaptureOutput(func() {
		accountStatsCmd.Run(accountStatsCmd, []string{})
	})

	if !strings.Contains(output, "Transactions: 4") {
		t.Errorf("expected all four transactions to be counted, got:\n%s", output)
	}
	// The 1900s date must widen the range rather than being skipped.
	if !strings.Contains(output, "1999-03-15") {
		t.Errorf("expected the 1900s date in the range, got:\n%s", output)
	}
}

// TestBalanceHistoryIncludesRecordsTheOldPatternMissed matters more than the
// list commands: a record the pattern could not read was left out of the
// arithmetic, so the reported balance was simply wrong.
func TestBalanceHistoryIncludesRecordsTheOldPatternMissed(t *testing.T) {
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

	qif := `!Account
NChecking
TBank
^
!Type:Bank
D1/5'23
U-10.00
T-10.00
CX
PNormal
Mok
LFood:Dining
^
D1/6'23
T-20.00
POut Of Order
NCheque 5
LFood:Dining
^
`

	sourceFile := filepath.Join(tempDir, "bh.qif")
	if err := os.WriteFile(sourceFile, []byte(qif), 0644); err != nil {
		t.Fatalf("failed to write qif fixture: %v", err)
	}

	inputFile = sourceFile
	outputPath = tempDir
	selectedAccounts = "Checking"
	openingBalance = "0"
	currentBalance = ""
	startDate, endDate = "", ""

	helper.CaptureOutput(func() {
		if err := balanceHistoryCmd.RunE(balanceHistoryCmd, []string{}); err != nil {
			t.Fatalf("balance-history failed: %v", err)
		}
	})

	result := filepath.Join(tempDir, "Checking_balance_history_1.csv")
	helper.AssertFileExists(result)
	content, err := os.ReadFile(result)
	if err != nil {
		t.Fatalf("failed to read %s: %v", result, err)
	}

	// Both transactions must be counted, so the balance runs to -30.00.
	if !strings.Contains(string(content), "2023-01-06") {
		t.Errorf("the out-of-order record is missing from the history:\n%s", content)
	}
	if !strings.Contains(string(content), "-30.00") {
		t.Errorf("closing balance should be -30.00:\n%s", content)
	}
}

// TestBalanceHistorySanitizesTheAccountNameForItsFileName covers the same trap
// the transaction export had: a colon makes NTFS treat the rest of the name as
// an alternate data stream, so the run reports success and writes nothing
// visible.
func TestBalanceHistorySanitizesTheAccountNameForItsFileName(t *testing.T) {
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

	qif := "!Account\nNAmex / Joint\nTCCard\n^\n!Type:CCard\n" +
		"D1/5'23\nT-10.00\nPSlash Name\nLFood:Dining\n^\n"
	sourceFile := filepath.Join(tempDir, "slash.qif")
	if err := os.WriteFile(sourceFile, []byte(qif), 0644); err != nil {
		t.Fatalf("failed to write qif fixture: %v", err)
	}

	inputFile = sourceFile
	outputPath = tempDir
	selectedAccounts = "Amex / Joint"
	openingBalance = "0"
	currentBalance = ""
	startDate, endDate = "", ""

	helper.CaptureOutput(func() {
		if err := balanceHistoryCmd.RunE(balanceHistoryCmd, []string{}); err != nil {
			t.Fatalf("balance-history failed for an account with a slash: %v", err)
		}
	})

	result := filepath.Join(tempDir, "Amex _ Joint_balance_history_1.csv")
	helper.AssertFileExists(result)
	helper.AssertFileContains(result, "-10.00")
}

// TestBalanceHistoryReportsFailuresInsteadOfExiting covers failure paths that
// previously ended the process, and so could not be asserted on at all.
func TestBalanceHistoryReportsFailuresInsteadOfExiting(t *testing.T) {
	tests := []struct {
		name     string
		accounts string
		opening  string
		current  string
		wantErr  string
	}{
		{"no account", "", "0", "", "requires exactly one account"},
		{"several accounts", "A,B", "0", "", "requires exactly one account"},
		{"no balance given", "Checking", "", "", "either --currentBalance or --openingBalance"},
		{"both balances given", "Checking", "0", "0", "mutually exclusive"},
		{"balance is not a number", "Checking", "abc", "", "invalid balance"},
		{"account is absent from the file", "Nowhere", "0", "", "not found in"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			selectedAccounts = tt.accounts
			openingBalance = tt.opening
			currentBalance = tt.current
			startDate, endDate = "", ""

			// Cobra runs PreRunE before RunE, and the argument checks live there.
			var err error
			helper.CaptureOutput(func() {
				if err = balanceHistoryCmd.PreRunE(balanceHistoryCmd, []string{}); err != nil {
					return
				}
				err = balanceHistoryCmd.RunE(balanceHistoryCmd, []string{})
			})

			if err == nil {
				t.Fatalf("expected an error mentioning %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want it to mention %q", err, tt.wantErr)
			}
		})
	}
}
