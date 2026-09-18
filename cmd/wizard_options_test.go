package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"qifutil/test"
)

// wizardFixture writes a QIF holding a mapped category and a split, plus a
// category mapping file, and returns the directory.
func wizardFixture(t *testing.T) string {
	t.Helper()
	helper := test.NewHelper(t)
	dir := helper.CreateTempDir()

	qif := "!Account\nNChecking\nTBank\n^\n!Type:Bank\n" +
		"D1/5'23\nT-100.00\nPSuperstore\nMbig shop\nLFood:Groceries\n" +
		"SFood:Groceries\nEGroceries portion\n$-60.00\n" +
		"SShopping:Home\nEHousehold portion\n$-40.00\n^\n"
	if err := os.WriteFile(filepath.Join(dir, "w.qif"), []byte(qif), 0644); err != nil {
		t.Fatalf("failed to write qif fixture: %v", err)
	}
	mapping := "\"Food:Groceries\",\"Everyday\"\n"
	if err := os.WriteFile(filepath.Join(dir, "cats.csv"), []byte(mapping), 0644); err != nil {
		t.Fatalf("failed to write mapping fixture: %v", err)
	}
	return dir
}

// TestWizardOffersTheOriginalCategoryNote covers the option being reachable at
// all from the wizard, which is where people who do not use flags start.
func TestWizardOffersTheOriginalCategoryNote(t *testing.T) {
	helper := test.NewHelper(t)
	dir := wizardFixture(t)

	// config=n, file, out, type=1, accounts=n, dates=n, format=1,
	// mappings=y, cats.csv, payee, account, tag, note=y, zero=n, split=n,
	// ENTER, save=n
	out := runWizard(t, dir,
		"n\nw.qif\n./out\n1\nn\nn\n1\ny\ncats.csv\n\n\n\ny\nn\nn\n\nn\n")

	if !strings.Contains(out, "Export completed successfully") {
		t.Fatalf("the wizard did not finish:\n%s", out)
	}
	result := filepath.Join(dir, "out", "Checking_1.csv")
	helper.AssertFileContains(result, "[Original Category: Food:Groceries]")
}

// TestWizardDoesNotAskAboutTheNoteWithoutAMapping covers the option being
// pointless without a category mapping to record the original of.
func TestWizardDoesNotAskAboutTheNoteWithoutAMapping(t *testing.T) {
	dir := wizardFixture(t)

	out := runWizard(t, dir,
		"n\nw.qif\n./out\n1\nn\nn\n1\nn\nn\nn\n\nn\n")

	if strings.Contains(out, "original category") {
		t.Errorf("the note question should not be asked without a mapping:\n%s", out)
	}
	if !strings.Contains(out, "Export completed successfully") {
		t.Errorf("the wizard did not finish:\n%s", out)
	}
}

func TestWizardOffersSplitExpansion(t *testing.T) {
	helper := test.NewHelper(t)
	dir := wizardFixture(t)

	// no mappings; zero=n, split=y
	out := runWizard(t, dir,
		"n\nw.qif\n./out\n1\nn\nn\n1\nn\nn\ny\n\nn\n")

	if !strings.Contains(out, "Export completed successfully") {
		t.Fatalf("the wizard did not finish:\n%s", out)
	}
	result := filepath.Join(dir, "out", "Checking_1.csv")
	helper.AssertFileContains(result, "Groceries portion")
	helper.AssertFileContains(result, "Shopping:Home")
}

// TestWizardConfigKeepsTheNewOptions covers a saved configuration reproducing
// the run that produced it. Dropping settings on save is the bug the mapping
// files already had.
func TestWizardConfigKeepsTheNewOptions(t *testing.T) {
	dir := wizardFixture(t)

	runWizard(t, dir,
		"n\nw.qif\n./out\n1\nn\nn\n1\ny\ncats.csv\n\n\n\ny\nn\ny\n\ny\ncfg.json\n")

	saved, err := os.ReadFile(filepath.Join(dir, "out", "cfg.json"))
	if err != nil {
		t.Fatalf("the wizard did not save a config: %v", err)
	}
	for _, want := range []string{
		`"preserveOriginalCategory": true`,
		`"expandSplits": true`,
	} {
		if !strings.Contains(string(saved), want) {
			t.Errorf("saved config missing %s:\n%s", want, saved)
		}
	}
}

// TestWizardDoesNotClaimSuccessWhenTheExportFails covers the wizard printing
// "Conversion completed!" unconditionally. The export reports failure now, and
// the wizard ignored what came back.
func TestWizardDoesNotClaimSuccessWhenTheExportFails(t *testing.T) {
	dir := wizardFixture(t)

	// a mapping file that is not there
	out := runWizard(t, dir,
		"n\nw.qif\n./out\n1\nn\nn\n1\ny\nnosuch.csv\n\n\n\ny\nn\nn\n\nn\n")

	if strings.Contains(out, "Conversion completed!") {
		t.Errorf("the wizard claimed success after a failed export:\n%s", out)
	}
	if !strings.Contains(out, "loading category mapping") {
		t.Errorf("the failure should be reported:\n%s", out)
	}
}

// TestWizardReAsksForTheBalance covers a mistyped balance. It printed
// "Balance history will not be generated", carried on, and finished with
// "Conversion completed!" having produced nothing.
func TestWizardReAsksForTheBalance(t *testing.T) {
	helper := test.NewHelper(t)
	dir := helper.CreateTempDir()
	helper.CopyTestData("sample.qif", filepath.Join(dir, "sample.qif"))

	// balance history only, account 1, current balance, a typo, then a number
	out := runWizard(t, dir,
		"n\nsample.qif\n./out\n2\n1\n1\nnot-a-number\n5000.00\n\nn\n")

	if strings.Contains(out, "Balance history will not be generated") {
		t.Errorf("the wizard should re-ask rather than abandon the balance:\n%s", out)
	}
	if !strings.Contains(out, "Balance history generation completed successfully") {
		t.Errorf("the balance history should still be produced:\n%s", out)
	}
}
