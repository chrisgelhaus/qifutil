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
		executeConversion(reader, wizardChoices{
			exportTransactions: true,
			categoryMapFile:    categoryMap,
			usingLoadedConfig:  true,
		})
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

// runWizard drives the built binary with the given keystrokes.
func runWizard(t *testing.T, dir, input string) string {
	t.Helper()
	cmd := exec.Command(qifutilBin, "wizard")
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader(input)
	out, _ := cmd.CombinedOutput()
	return string(out)
}

// TestWizardReAsksForAMistypedPath covers the most likely mistake at the most
// used prompt. Every other prompt re-asks on bad input; this one printed the
// problem and ended the wizard.
func TestWizardReAsksForAMistypedPath(t *testing.T) {
	helper := test.NewHelper(t)
	tempDir := helper.CreateTempDir()
	helper.CopyTestData("sample.qif", filepath.Join(tempDir, "sample.qif"))

	// a name that does not exist, then the real one, then run it through.
	out := runWizard(t, tempDir,
		"n\nnope.qif\nsample.qif\n./out\n1\nn\nn\n1\nn\nn\n\nn\n")

	if !strings.Contains(out, "Could not find file") {
		t.Errorf("the wizard should say the file was not found:\n%s", out)
	}
	if !strings.Contains(out, "Export completed successfully") {
		t.Errorf("the wizard should carry on after a mistyped path:\n%s", out)
	}
}

// TestWizardRejectsAnEmptyPath covers pressing Enter at the file prompt. The
// empty path resolved to the current directory, which exists, so it passed the
// check and every later answer shifted up by one.
func TestWizardRejectsAnEmptyPath(t *testing.T) {
	helper := test.NewHelper(t)
	tempDir := helper.CreateTempDir()
	helper.CopyTestData("sample.qif", filepath.Join(tempDir, "sample.qif"))

	out := runWizard(t, tempDir,
		"n\n\nsample.qif\n./out\n1\nn\nn\n1\nn\nn\n\nn\n")

	// The QIF file must not be taken as the place to write output.
	for _, line := range strings.Split(out, "\n") {
		if !strings.Contains(line, "Creating output directory") {
			continue
		}
		if strings.HasSuffix(strings.TrimSpace(line), "sample.qif") {
			t.Errorf("the QIF file was taken as the output directory: %s", line)
		}
	}
	if !strings.Contains(out, "Export completed successfully") {
		t.Errorf("the wizard should re-ask and then carry on:\n%s", out)
	}
}
