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

	t.Cleanup(func() {
		selectedAccounts, startDate, endDate = prevAccounts, prevStart, prevEnd
		outputFormat, csvColumns = prevFormat, prevColumns
		inputFile, outputPath = prevInput, prevOutput
		categoryMappingFile, payeeMappingFile = prevCategoryMap, prevPayeeMap
		preserveOriginalCategory = prevPreserve
	})

	selectedAccounts, startDate, endDate = "", "", ""
	outputFormat, csvColumns = "CSV", DefaultMonarchColumns
	categoryMappingFile, payeeMappingFile = "", ""
	preserveOriginalCategory = false
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
