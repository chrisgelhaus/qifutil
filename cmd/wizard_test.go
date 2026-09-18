package cmd

import (
	"os/exec"
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

// TestWizardReachesTheAccountList drives the wizard the way a person does,
// through a pipe, because it reads from os.Stdin. Step 3 called Run on a copy
// of the list-accounts command; once that command moved to RunE, Run was nil
// and the wizard panicked before showing a single account.
func TestWizardReachesTheAccountList(t *testing.T) {
	helper := test.NewHelper(t)
	tempDir := helper.CreateTempDir()
	helper.CopyTestData("sample.qif", filepath.Join(tempDir, "sample.qif"))

	// no saved config, the input file, the output directory, then stop.
	input := "n\nsample.qif\n./out\n"

	cmd := exec.Command(qifutilBin, "wizard")
	cmd.Dir = tempDir
	cmd.Stdin = strings.NewReader(input)
	out, _ := cmd.CombinedOutput()

	if strings.Contains(string(out), "panic:") {
		t.Fatalf("the wizard panicked:\n%s", out)
	}
	if !strings.Contains(string(out), "Found these accounts in your file:") {
		t.Errorf("the wizard should list the accounts it found:\n%s", out)
	}
	for _, account := range []string{"Checking Account", "Savings Account", "CreditCard Account"} {
		if !strings.Contains(string(out), account) {
			t.Errorf("account %q missing from the listing:\n%s", account, out)
		}
	}
}
