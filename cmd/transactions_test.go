package cmd

import (
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

// setupTransactionExport resets the shared command globals to a known state and
// restores them when the test ends, so tests in this package do not leak flag
// values into each other.
func setupTransactionExport(t *testing.T) {
	prevAccounts, prevStart, prevEnd := selectedAccounts, startDate, endDate
	prevFormat, prevColumns := outputFormat, csvColumns
	prevInput, prevOutput := inputFile, outputPath
	prevCategoryMap, prevPayeeMap := categoryMappingFile, payeeMappingFile
	prevPreserve := preserveOriginalCategory
	prevExpand := expandSplits

	t.Cleanup(func() {
		selectedAccounts, startDate, endDate = prevAccounts, prevStart, prevEnd
		outputFormat, csvColumns = prevFormat, prevColumns
		inputFile, outputPath = prevInput, prevOutput
		categoryMappingFile, payeeMappingFile = prevCategoryMap, prevPayeeMap
		preserveOriginalCategory = prevPreserve
		expandSplits = prevExpand
	})

	selectedAccounts, startDate, endDate = "", "", ""
	outputFormat, csvColumns = "CSV", DefaultMonarchColumns
	categoryMappingFile, payeeMappingFile = "", ""
	preserveOriginalCategory = false
	expandSplits = false
}

// exportSampleDir runs the transactions command over sample.qif and returns the
// directory the output was written to.
func exportSampleDir(t *testing.T, helper *test.TestHelper, tempDir string) string {
	outputDir := filepath.Join(tempDir, "output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("failed to create output dir: %v", err)
	}

	sourceFile := filepath.Join(tempDir, "sample.qif")
	helper.CopyTestData("sample.qif", sourceFile)

	inputFile = sourceFile
	outputPath = outputDir

	helper.CaptureOutput(func() {
		transactionsCmd.Run(transactionsCmd, []string{})
	})

	return outputDir
}

// exportSample runs the transactions command over sample.qif and returns the
// path of the Checking Account CSV output file.
func exportSample(t *testing.T, helper *test.TestHelper, tempDir string) string {
	return filepath.Join(exportSampleDir(t, helper, tempDir), "Checking Account_1.csv")
}

// writeMappingFile writes a two-column mapping CSV and returns its path.
func writeMappingFile(t *testing.T, dir, name, contents string) string {
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(contents), 0644); err != nil {
		t.Fatalf("failed to write mapping file %s: %v", path, err)
	}
	return path
}

// countLinesContaining reports how many lines of content contain needle.
func countLinesContaining(content, needle string) int {
	count := 0
	for _, line := range strings.Split(content, "\n") {
		if strings.Contains(line, needle) {
			count++
		}
	}
	return count
}

// lineContaining returns the first line of a file containing needle.
func lineContaining(t *testing.T, path, needle string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	for _, line := range strings.Split(string(content), "\n") {
		if strings.Contains(line, needle) {
			return line
		}
	}
	t.Fatalf("no line in %s contains %q", path, needle)
	return ""
}

func TestPreserveOriginalCategoryAppendsRemappedCategoryToNotes(t *testing.T) {
	helper := test.NewHelper(t)
	tempDir := helper.CreateTempDir()
	setupTransactionExport(t)

	categoryMappingFile = writeMappingFile(t, tempDir, "categories.csv", "\"Food:Groceries\",\"Food\"\n")
	preserveOriginalCategory = true

	checkingFile := exportSample(t, helper, tempDir)
	helper.AssertFileExists(checkingFile)

	line := lineContaining(t, checkingFile, "Grocery Store")
	if !strings.Contains(line, "\"Food shopping [Original Category: Food:Groceries]\"") {
		t.Errorf("expected notes to carry the pre-mapping category, got: %s", line)
	}
}

func TestPreserveOriginalCategoryOmitsNoteWhenCategoryUnchanged(t *testing.T) {
	helper := test.NewHelper(t)
	tempDir := helper.CreateTempDir()
	setupTransactionExport(t)

	// Transportation:Fuel is deliberately absent from the mapping.
	categoryMappingFile = writeMappingFile(t, tempDir, "categories.csv", "\"Food:Groceries\",\"Food\"\n")
	preserveOriginalCategory = true

	checkingFile := exportSample(t, helper, tempDir)

	line := lineContaining(t, checkingFile, "COSTCO GAS STATION")
	if strings.Contains(line, "Original Category") {
		t.Errorf("expected no back-reference on an unmapped category, got: %s", line)
	}
}

func TestPreserveOriginalCategoryDisabledByDefault(t *testing.T) {
	helper := test.NewHelper(t)
	tempDir := helper.CreateTempDir()
	setupTransactionExport(t)

	categoryMappingFile = writeMappingFile(t, tempDir, "categories.csv", "\"Food:Groceries\",\"Food\"\n")

	checkingFile := exportSample(t, helper, tempDir)

	content, err := os.ReadFile(checkingFile)
	if err != nil {
		t.Fatalf("failed to read %s: %v", checkingFile, err)
	}
	if strings.Contains(string(content), "Original Category") {
		t.Error("expected no back-reference when --preserveOriginalCategory is off")
	}
}

func TestOriginalStatementKeepsPreMappingPayee(t *testing.T) {
	helper := test.NewHelper(t)
	tempDir := helper.CreateTempDir()
	setupTransactionExport(t)

	payeeMappingFile = writeMappingFile(t, tempDir, "payees.csv", "\"Grocery Store\",\"Wholesale Club\"\n")

	checkingFile := exportSample(t, helper, tempDir)

	line := lineContaining(t, checkingFile, "Wholesale Club")
	if !strings.Contains(line, "\"Grocery Store\"") {
		t.Errorf("expected Original Statement to hold the pre-mapping payee, got: %s", line)
	}
}

func TestPreserveOriginalCategorySkipsNoteWhenOriginalIsBlank(t *testing.T) {
	helper := test.NewHelper(t)
	tempDir := helper.CreateTempDir()
	setupTransactionExport(t)

	// An uncategorized transaction, plus a rule that maps the blank category.
	qif := "!Account\nNTest Account\nTBank\n^\n!Type:Bank\n" +
		"D1/5'23\nU-10.00\nT-10.00\nCX\nPSome Store\nMA memo\nL\n^\n"
	sourceFile := filepath.Join(tempDir, "blank.qif")
	if err := os.WriteFile(sourceFile, []byte(qif), 0644); err != nil {
		t.Fatalf("failed to write qif fixture: %v", err)
	}

	categoryMappingFile = writeMappingFile(t, tempDir, "categories.csv", "\"\",\"Uncategorized\"\n")
	preserveOriginalCategory = true

	outputDir := filepath.Join(tempDir, "output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("failed to create output dir: %v", err)
	}
	inputFile = sourceFile
	outputPath = outputDir

	helper.CaptureOutput(func() {
		transactionsCmd.Run(transactionsCmd, []string{})
	})

	line := lineContaining(t, filepath.Join(outputDir, "Test Account_1.csv"), "Some Store")
	if strings.Contains(line, "Original Category") {
		t.Errorf("expected no back-reference when the original category is blank, got: %s", line)
	}
}

// qifutilBin is a freshly built CLI binary, used by tests that exercise
// behaviour ending in os.Exit and so cannot run in-process.
var qifutilBin string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "qifutil-bin-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}

	qifutilBin = filepath.Join(dir, "qifutil")
	if runtime.GOOS == "windows" {
		qifutilBin += ".exe"
	}

	build := exec.Command("go", "build", "-o", qifutilBin, ".")
	build.Dir = ".."
	if out, buildErr := build.CombinedOutput(); buildErr != nil {
		fmt.Fprintf(os.Stderr, "failed to build qifutil: %v\n%s", buildErr, out)
		os.RemoveAll(dir)
		os.Exit(1)
	}

	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

func TestRelativeInputFileResolvesAgainstWorkingDirectory(t *testing.T) {
	helper := test.NewHelper(t)
	tempDir := helper.CreateTempDir()
	helper.CopyTestData("sample.qif", filepath.Join(tempDir, "sample.qif"))

	cmd := exec.Command(qifutilBin, "export", "transactions",
		"--inputFile", "sample.qif", "--outputPath", "./out")
	cmd.Dir = tempDir

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("export with a relative --inputFile failed: %v\n%s", err, out)
	}

	helper.AssertFileExists(filepath.Join(tempDir, "out", "Checking Account_1.csv"))
}

func TestXMLOutputIsWellFormedXML(t *testing.T) {
	helper := test.NewHelper(t)
	tempDir := helper.CreateTempDir()
	setupTransactionExport(t)
	outputFormat = "XML"

	xmlFile := filepath.Join(exportSampleDir(t, helper, tempDir), "Checking Account_1.xml")
	helper.AssertFileExists(xmlFile)

	content, err := os.ReadFile(xmlFile)
	if err != nil {
		t.Fatalf("failed to read %s: %v", xmlFile, err)
	}

	var parsed transactionList
	if err := xml.Unmarshal(content, &parsed); err != nil {
		t.Fatalf("XML output does not parse: %v\n%s", err, content)
	}
	if len(parsed.Transactions) == 0 {
		t.Fatal("XML output contained no transaction elements")
	}

	first := parsed.Transactions[0]
	if first.Date != "2023-01-05" || first.Merchant != "Employee Payroll" {
		t.Errorf("unexpected first transaction: %+v", first)
	}
}

func TestTransactionsAcceptsShortInputAndOutputFlags(t *testing.T) {
	helper := test.NewHelper(t)
	tempDir := helper.CreateTempDir()
	helper.CopyTestData("sample.qif", filepath.Join(tempDir, "sample.qif"))

	// The form every usage example in the help text and README documents.
	cmd := exec.Command(qifutilBin, "export", "transactions",
		"-i", "sample.qif", "-o", "./out", "-f", "MONARCH")
	cmd.Dir = tempDir

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("documented -i/-o form failed: %v\n%s", err, out)
	}

	helper.AssertFileExists(filepath.Join(tempDir, "out", "Checking Account_1.csv"))
}

// TestCategoriesShortOutputFlagStillNamesAFile guards the meaning of -o on the
// list-style commands, where it selects an output file rather than a directory.
func TestCategoriesShortOutputFlagStillNamesAFile(t *testing.T) {
	helper := test.NewHelper(t)
	tempDir := helper.CreateTempDir()
	helper.CopyTestData("sample.qif", filepath.Join(tempDir, "sample.qif"))
	if err := os.MkdirAll(filepath.Join(tempDir, "out"), 0755); err != nil {
		t.Fatalf("failed to create output dir: %v", err)
	}

	cmd := exec.Command(qifutilBin, "export", "categories",
		"-i", "sample.qif", "-o", "cats.csv", "--outputPath", "./out")
	cmd.Dir = tempDir

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("export categories with -o failed: %v\n%s", err, out)
	}

	helper.AssertFileExists(filepath.Join(tempDir, "out", "cats.csv"))
}

func TestNineteenHundredsDatesAreExported(t *testing.T) {
	helper := test.NewHelper(t)
	tempDir := helper.CreateTempDir()
	setupTransactionExport(t)

	// Quicken marks the century with the separator before the year: a slash for
	// the 1900s, an apostrophe for the 2000s.
	qif := "!Account\nNTest Account\nTBank\n^\n!Type:Bank\n" +
		"D3/15/99\nU-10.00\nT-10.00\nCX\nPLast Century\nMslash form\nLFood:Dining\n^\n" +
		"D1/5'23\nU-30.00\nT-30.00\nCX\nPThis Century\nMapostrophe form\nLFood:Dining\n^\n"

	sourceFile := filepath.Join(tempDir, "century.qif")
	if err := os.WriteFile(sourceFile, []byte(qif), 0644); err != nil {
		t.Fatalf("failed to write qif fixture: %v", err)
	}

	outputDir := filepath.Join(tempDir, "output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("failed to create output dir: %v", err)
	}
	inputFile = sourceFile
	outputPath = outputDir

	helper.CaptureOutput(func() {
		transactionsCmd.Run(transactionsCmd, []string{})
	})

	checkingFile := filepath.Join(outputDir, "Test Account_1.csv")
	helper.AssertFileExists(checkingFile)

	slash := lineContaining(t, checkingFile, "Last Century")
	if !strings.Contains(slash, "1999-03-15") {
		t.Errorf("slash-form date should export as 1999-03-15, got: %s", slash)
	}

	apostrophe := lineContaining(t, checkingFile, "This Century")
	if !strings.Contains(apostrophe, "2023-01-05") {
		t.Errorf("apostrophe-form date should export as 2023-01-05, got: %s", apostrophe)
	}
}

// TestOptionalQIFFieldsAreNotDropped covers records that omit fields QIF treats
// as optional, or that order them differently. A single pattern requiring
// D, U, T, C and L in sequence silently loses all of these.
func TestOptionalQIFFieldsAreNotDropped(t *testing.T) {
	helper := test.NewHelper(t)
	tempDir := helper.CreateTempDir()
	setupTransactionExport(t)

	qif := `!Account
NChecking
TBank
^
!Type:Bank
D1/5'23
U-10.00
T-10.00
CX
PComplete
Mall fields
LFood:Dining
^
D1/6'23
T-20.00
CX
PNo U field
Mmissing U
LFood:Dining
^
D1/7'23
U-30.00
T-30.00
PNo C field
Mmissing C
LFood:Dining
^
D1/8'23
U-40.00
T-40.00
CX
PNo L field
Mmissing L
^
D1/9'23
U-50.00
T-50.00
CX
LFood:Dining
PReordered after L
^
D1/10'23
T-60.00
POnly T and P
^
`

	sourceFile := filepath.Join(tempDir, "mixed.qif")
	if err := os.WriteFile(sourceFile, []byte(qif), 0644); err != nil {
		t.Fatalf("failed to write qif fixture: %v", err)
	}

	outputDir := filepath.Join(tempDir, "output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("failed to create output dir: %v", err)
	}
	inputFile = sourceFile
	outputPath = outputDir

	helper.CaptureOutput(func() {
		transactionsCmd.Run(transactionsCmd, []string{})
	})

	checkingFile := filepath.Join(outputDir, "Checking_1.csv")
	helper.AssertFileExists(checkingFile)

	for _, payee := range []string{
		"Complete", "No U field", "No C field", "No L field",
		"Reordered after L", "Only T and P",
	} {
		helper.AssertFileContains(checkingFile, payee)
	}

	// The amount must survive the fallback from U to T.
	noU := lineContaining(t, checkingFile, "No U field")
	if !strings.Contains(noU, "-20.00") {
		t.Errorf("amount should fall back to the T field, got: %s", noU)
	}

	// A payee written after the category must still be read.
	reordered := lineContaining(t, checkingFile, "Reordered after L")
	if !strings.Contains(reordered, "Food:Dining") {
		t.Errorf("reordered record lost its category, got: %s", reordered)
	}
}

	const splitQIF = `!Account
NChecking
TBank
^
!Type:Bank
D1/5'23
U-100.00
T-100.00
CX
PSuperstore
Mbig shop
LFood:Groceries
SFood:Groceries
EGroceries portion
$-60.00
SShopping:Home
$-40.00
^
`

// exportInlineQIF writes a QIF fixture, exports it, and returns the console
// output together with the directory the export was written to.
func exportInlineQIF(t *testing.T, helper *test.TestHelper, tempDir, qif string) (string, string) {
	sourceFile := filepath.Join(tempDir, "fixture.qif")
	if err := os.WriteFile(sourceFile, []byte(qif), 0644); err != nil {
		t.Fatalf("failed to write qif fixture: %v", err)
	}

	outputDir := filepath.Join(tempDir, "output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("failed to create output dir: %v", err)
	}
	inputFile = sourceFile
	outputPath = outputDir

	out := helper.CaptureOutput(func() {
		transactionsCmd.Run(transactionsCmd, []string{})
	})
	return out, outputDir
}

func TestSplitsCollapseToOneRowByDefault(t *testing.T) {
	helper := test.NewHelper(t)
	tempDir := helper.CreateTempDir()
	setupTransactionExport(t)

	_, outputDir := exportInlineQIF(t, helper, tempDir, splitQIF)
	checkingFile := filepath.Join(outputDir, "Checking_1.csv")

	content, err := os.ReadFile(checkingFile)
	if err != nil {
		t.Fatalf("failed to read %s: %v", checkingFile, err)
	}
	if rows := countLinesContaining(string(content), "Superstore"); rows != 1 {
		t.Errorf("expected one row without --expandSplits, got %d\n%s", rows, content)
	}
	helper.AssertFileContains(checkingFile, "-100.00")
}

func TestExpandSplitsEmitsOneRowPerSplit(t *testing.T) {
	helper := test.NewHelper(t)
	tempDir := helper.CreateTempDir()
	setupTransactionExport(t)
	expandSplits = true

	_, outputDir := exportInlineQIF(t, helper, tempDir, splitQIF)
	checkingFile := filepath.Join(outputDir, "Checking_1.csv")

	groceries := lineContaining(t, checkingFile, "Groceries portion")
	if !strings.Contains(groceries, "-60.00") || !strings.Contains(groceries, "Food:Groceries") {
		t.Errorf("first split row wrong: %s", groceries)
	}

	household := lineContaining(t, checkingFile, "Shopping:Home")
	if !strings.Contains(household, "-40.00") {
		t.Errorf("second split row wrong: %s", household)
	}

	// The parent row must not also be written, or the account total doubles.
	content, err := os.ReadFile(checkingFile)
	if err != nil {
		t.Fatalf("failed to read %s: %v", checkingFile, err)
	}
	if strings.Contains(string(content), "-100.00") {
		t.Errorf("parent row was written alongside its splits:\n%s", content)
	}
}

func TestExpandSplitsFallsBackToTheParentMemo(t *testing.T) {
	helper := test.NewHelper(t)
	tempDir := helper.CreateTempDir()
	setupTransactionExport(t)
	expandSplits = true

	_, outputDir := exportInlineQIF(t, helper, tempDir, splitQIF)
	checkingFile := filepath.Join(outputDir, "Checking_1.csv")

	// The Shopping:Home split carries no E line, so it inherits "big shop".
	household := lineContaining(t, checkingFile, "Shopping:Home")
	if !strings.Contains(household, "big shop") {
		t.Errorf("split without its own memo should inherit the parent memo: %s", household)
	}
}

func TestExpandSplitsWarnsWhenSplitsDoNotReconcile(t *testing.T) {
	helper := test.NewHelper(t)
	tempDir := helper.CreateTempDir()
	setupTransactionExport(t)
	expandSplits = true

	unbalanced := strings.Replace(splitQIF, "$-40.00", "$-30.00", 1)
	output, outputDir := exportInlineQIF(t, helper, tempDir, unbalanced)

	if !strings.Contains(output, "do not sum") {
		t.Errorf("expected a warning that the splits do not reconcile, got:\n%s", output)
	}
	if !strings.Contains(output, "10.00") {
		t.Errorf("warning should name the difference, got:\n%s", output)
	}

	// The split rows are still exported.
	helper.AssertFileContains(filepath.Join(outputDir, "Checking_1.csv"), "-30.00")
}

const hostileNamesQIF = `!Account
NAmex / Joint
TCCard
^
!Type:CCard
D1/5'23
T-10.00
PSlash Name
LFood:Dining
^
!Account
NFidelity: Roth
TBank
^
!Type:Bank
D1/6'23
T-20.00
PColon Name
LFood:Dining
^
!Account
NNormal Account
TBank
^
!Type:Bank
D1/7'23
T-30.00
PPlain Name
LFood:Dining
^
`

func TestAccountNamesAreSanitizedForFileNames(t *testing.T) {
	helper := test.NewHelper(t)
	tempDir := helper.CreateTempDir()
	setupTransactionExport(t)

	_, outputDir := exportInlineQIF(t, helper, tempDir, hostileNamesQIF)

	// A slash would be read as a directory and a colon as an NTFS stream.
	for name, payee := range map[string]string{
		"Amex _ Joint_1.csv":   "Slash Name",
		"Fidelity_ Roth_1.csv": "Colon Name",
		"Normal Account_1.csv": "Plain Name",
	} {
		path := filepath.Join(outputDir, name)
		helper.AssertFileExists(path)
		helper.AssertFileContains(path, payee)
	}
}

func TestCollidingAccountFileNamesAreDisambiguated(t *testing.T) {
	helper := test.NewHelper(t)
	tempDir := helper.CreateTempDir()
	setupTransactionExport(t)

	// "A/B" and "A_B" both sanitise to "A_B".
	qif := `!Account
NA/B
TBank
^
!Type:Bank
D1/5'23
T-10.00
PFirst Account
LFood:Dining
^
!Account
NA_B
TBank
^
!Type:Bank
D1/6'23
T-20.00
PSecond Account
LFood:Dining
^
`

	_, outputDir := exportInlineQIF(t, helper, tempDir, qif)

	first := filepath.Join(outputDir, "A_B_1.csv")
	helper.AssertFileExists(first)
	helper.AssertFileContains(first, "First Account")

	// The second account must not overwrite the first.
	second := filepath.Join(outputDir, "A_B~2_1.csv")
	helper.AssertFileExists(second)
	helper.AssertFileContains(second, "Second Account")
}

func TestOneUnwritableAccountDoesNotAbortTheExport(t *testing.T) {
	helper := test.NewHelper(t)
	tempDir := helper.CreateTempDir()
	setupTransactionExport(t)

	sourceFile := filepath.Join(tempDir, "fixture.qif")
	if err := os.WriteFile(sourceFile, []byte(hostileNamesQIF), 0644); err != nil {
		t.Fatalf("failed to write qif fixture: %v", err)
	}
	outputDir := filepath.Join(tempDir, "output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("failed to create output dir: %v", err)
	}

	// A directory where the first account's file belongs makes os.Create fail.
	blocked := filepath.Join(outputDir, "Amex _ Joint_1.csv")
	if err := os.MkdirAll(blocked, 0755); err != nil {
		t.Fatalf("failed to create blocking directory: %v", err)
	}

	inputFile = sourceFile
	outputPath = outputDir
	output := helper.CaptureOutput(func() {
		transactionsCmd.Run(transactionsCmd, []string{})
	})

	// The accounts after the failure must still be exported.
	helper.AssertFileExists(filepath.Join(outputDir, "Fidelity_ Roth_1.csv"))
	helper.AssertFileExists(filepath.Join(outputDir, "Normal Account_1.csv"))

	if !strings.Contains(output, "Amex / Joint") {
		t.Errorf("the skipped account should be named in the output, got:\n%s", output)
	}
}
