package utils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeLog runs the tracker through a real log file and returns its contents.
func writeLog(t *testing.T, vt *ValidationTracker) string {
	t.Helper()
	dir := t.TempDir()
	if err := vt.WriteValidationLogWithName(dir, "validation.log"); err != nil {
		t.Fatalf("failed to write log: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(dir, "validation.log"))
	if err != nil {
		t.Fatalf("failed to read log: %v", err)
	}
	return string(content)
}

// TestUnmappedValuesAreOrderedByFrequency covers the order of the list. It came
// from ranging over a map, so it differed between runs of the same export.
func TestUnmappedValuesAreOrderedByFrequency(t *testing.T) {
	vt := NewValidationTracker()
	for i := 0; i < 3; i++ {
		vt.AddUnmatchedData("payee", "Seen Three Times")
	}
	for i := 0; i < 7; i++ {
		vt.AddUnmatchedData("payee", "Seen Seven Times")
	}
	vt.AddUnmatchedData("category", "Seen Once")

	log := writeLog(t, vt)
	seven := strings.Index(log, "Seen Seven Times")
	three := strings.Index(log, "Seen Three Times")
	once := strings.Index(log, "Seen Once")

	if seven < 0 || three < 0 || once < 0 {
		t.Fatalf("every unmapped value should be listed:\n%s", log)
	}
	if !(seven < three && three < once) {
		t.Errorf("expected most frequent first:\n%s", log)
	}
}

func TestUnmappedValuesNameTheKindOfValue(t *testing.T) {
	vt := NewValidationTracker()
	vt.AddUnmatchedData("category", "Food:Dining")

	log := writeLog(t, vt)
	if !strings.Contains(log, `category "Food:Dining"`) {
		t.Errorf("the kind and the value should read as separate things:\n%s", log)
	}
}

// TestSummaryDoesNotNameAFileThatIsNotWritten covers the closing line, which
// pointed at validation.log while the exports write transactions_validation.log
// and balance_history_validation.log.
func TestSummaryDoesNotNameAFileThatIsNotWritten(t *testing.T) {
	vt := NewValidationTracker()
	vt.AddMissingPayee()

	out := captureStdout(t, vt.PrintSummary)
	if strings.Contains(out, "validation.log") {
		t.Errorf("the summary should not name a specific log file:\n%s", out)
	}
	if !strings.Contains(out, "validation log") {
		t.Errorf("the summary should still point at the log:\n%s", out)
	}
}

// captureStdout collects what fn prints.
func captureStdout(t *testing.T, fn func()) (output string) {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w

	done := make(chan string)
	go func() {
		var sb strings.Builder
		buf := make([]byte, 4096)
		for {
			n, readErr := r.Read(buf)
			sb.Write(buf[:n])
			if readErr != nil {
				break
			}
		}
		done <- sb.String()
	}()

	defer func() {
		w.Close()
		os.Stdout = old
		output = <-done
	}()

	fn()
	return
}
