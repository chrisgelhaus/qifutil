package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"qifutil/test"
)

// readValidationLog returns the log the transactions export writes beside its
// output.
func readValidationLog(t *testing.T, outputDir string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(outputDir, "transactions_validation.log"))
	if err != nil {
		t.Fatalf("failed to read validation log: %v", err)
	}
	return string(content)
}

func TestDuplicateTransactionsAreReported(t *testing.T) {
	helper := test.NewHelper(t)
	tempDir := helper.CreateTempDir()
	setupTransactionExport(t)

	// The same date, payee and amount twice in one account.
	qif := `!Account
NChecking
TBank
^
!Type:Bank
D1/15'23
T-45.23
PGrocery Store
LFood:Groceries
^
D1/15'23
T-45.23
PGrocery Store
LFood:Groceries
^
D1/16'23
T-10.00
PCoffee Shop
LFood:Dining
^
`

	output, outputDir := exportInlineQIF(t, helper, tempDir, qif)

	if !strings.Contains(output, "Potential duplicates") {
		t.Errorf("the summary should report the duplicate:\n%s", output)
	}
	log := readValidationLog(t, outputDir)
	if !strings.Contains(log, "Grocery Store") || !strings.Contains(log, "appears 2 times") {
		t.Errorf("the log should name the duplicated transaction:\n%s", log)
	}
	if strings.Contains(log, "Coffee Shop") {
		t.Errorf("a transaction appearing once is not a duplicate:\n%s", log)
	}
}

// TestTransfersBetweenAccountsAreNotDuplicates covers the same date, payee and
// amount in two accounts, which is what a transfer between them looks like.
func TestTransfersBetweenAccountsAreNotDuplicates(t *testing.T) {
	helper := test.NewHelper(t)
	tempDir := helper.CreateTempDir()
	setupTransactionExport(t)

	qif := `!Account
NChecking
TBank
^
!Type:Bank
D1/15'23
T-500.00
PTransfer
LTransfers
^
!Account
NSavings
TBank
^
!Type:Bank
D1/15'23
T-500.00
PTransfer
LTransfers
^
`

	output, _ := exportInlineQIF(t, helper, tempDir, qif)

	if strings.Contains(output, "Potential duplicates") {
		t.Errorf("the same entry in two accounts is a transfer, not a duplicate:\n%s", output)
	}
}

// mixedCategoriesQIF holds one category a mapping will cover and one it will
// not, so "unmapped" means something.
const mixedCategoriesQIF = `!Account
NChecking
TBank
^
!Type:Bank
D1/15'23
T-45.23
PGrocery Store
LFood:Groceries
^
D1/16'23
T-85.00
PPizza Place
LFood:Dining
^
`

func TestUnmappedValuesAreReportedWhenAMappingFileIsGiven(t *testing.T) {
	helper := test.NewHelper(t)
	tempDir := helper.CreateTempDir()
	setupTransactionExport(t)

	// The mapping covers Food:Groceries but not Food:Dining.
	categoryMappingFile = writeMappingFile(t, tempDir, "categories.csv",
		"\"Food:Groceries\",\"Everyday\"\n")

	_, outputDir := exportInlineQIF(t, helper, tempDir, mixedCategoriesQIF)
	log := readValidationLog(t, outputDir)

	if !strings.Contains(log, "Unmapped values") {
		t.Errorf("the log should list values the mapping did not cover:\n%s", log)
	}
}

func TestUnmappedValuesAreNotReportedWithoutAMappingFile(t *testing.T) {
	helper := test.NewHelper(t)
	tempDir := helper.CreateTempDir()
	setupTransactionExport(t)

	// With no mapping file every value is unmapped, which tells nobody anything.
	_, outputDir := exportInlineQIF(t, helper, tempDir, mixedCategoriesQIF)
	log := readValidationLog(t, outputDir)

	if strings.Contains(log, "Unmapped values") {
		t.Errorf("unmapped values need a mapping file to be meaningful:\n%s", log)
	}
}

func TestUnusedMappingRulesAreReported(t *testing.T) {
	helper := test.NewHelper(t)
	tempDir := helper.CreateTempDir()
	setupTransactionExport(t)

	// The second rule matches nothing in the file, which usually means a typo.
	categoryMappingFile = writeMappingFile(t, tempDir, "categories.csv",
		"\"Food:Groceries\",\"Everyday\"\n\"Typo:Categry\",\"Something\"\n")

	output, outputDir := exportInlineQIF(t, helper, tempDir, mixedCategoriesQIF)
	log := readValidationLog(t, outputDir)

	if !strings.Contains(log, "Typo:Categry") {
		t.Errorf("the log should name the rule that never matched:\n%s", log)
	}
	if !strings.Contains(output, "rules never used") {
		t.Errorf("the summary should mention unused rules:\n%s", output)
	}
	// The rule that did match must not be listed as unused.
	if strings.Contains(log, "- \"Food:Groceries\"") {
		t.Errorf("a rule that matched is not unused:\n%s", log)
	}
}

// TestMappingCountsAreSummarised covers the replacement for the line that used
// to be printed for every single mapping applied, which buried the summary on a
// file of any size.
func TestMappingCountsAreSummarised(t *testing.T) {
	helper := test.NewHelper(t)
	tempDir := helper.CreateTempDir()
	setupTransactionExport(t)

	categoryMappingFile = writeMappingFile(t, tempDir, "categories.csv",
		"\"Food:Groceries\",\"Everyday\"\n")

	output, _ := exportInlineQIF(t, helper, tempDir, mixedCategoriesQIF)

	// One of the two categories is covered by the mapping.
	if !strings.Contains(output, "Category mapping: applied to 1 value") {
		t.Errorf("the summary should count the mappings applied:\n%s", output)
	}
	// The per-hit line is gone.
	if strings.Contains(output, "Mapping: Food:Groceries ->") {
		t.Errorf("the per-hit mapping line should no longer be printed:\n%s", output)
	}
}

func TestMappingCountsAreAbsentWithoutAMappingFile(t *testing.T) {
	helper := test.NewHelper(t)
	tempDir := helper.CreateTempDir()
	setupTransactionExport(t)

	output, _ := exportInlineQIF(t, helper, tempDir, mixedCategoriesQIF)

	if strings.Contains(output, "mapping: applied to") {
		t.Errorf("nothing to report when no mapping file was given:\n%s", output)
	}
}
