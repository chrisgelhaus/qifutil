package test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestHelper provides utilities for testing CLI commands
type TestHelper struct {
	t *testing.T
}

// NewHelper creates a new TestHelper
func NewHelper(t *testing.T) *TestHelper {
	return &TestHelper{t: t}
}

// CreateTempDir creates a temporary directory and returns its path.
// The directory will be automatically cleaned up when the test ends.
func (h *TestHelper) CreateTempDir() string {
	dir, err := os.MkdirTemp("", "qifutil-test-*")
	if err != nil {
		h.t.Fatalf("failed to create temp dir: %v", err)
	}
	h.t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

// CopyTestData copies a test data file from testdata directory to the target path
func (h *TestHelper) CopyTestData(filename string, targetPath string) {
	// testdata is located at test/testdata relative to the project root
	// Go test runners are always executed from the project root
	src := filepath.Join("test", "testdata", filename)

	data, err := os.ReadFile(src)
	if err != nil {
		h.t.Fatalf("failed to read test data file %s: %v", src, err)
	}
	err = os.WriteFile(targetPath, data, 0644)
	if err != nil {
		h.t.Fatalf("failed to write test data to %s: %v", targetPath, err)
	}
}

// CaptureOutput captures stdout during the execution of the given function
func (h *TestHelper) CaptureOutput(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	done := make(chan bool)
	var buf bytes.Buffer
	go func() {
		io.Copy(&buf, r)
		done <- true
	}()

	fn()

	w.Close()
	os.Stdout = old
	<-done

	return buf.String()
}

// AssertFileExists checks if a file exists at the given path
func (h *TestHelper) AssertFileExists(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		h.t.Errorf("expected file to exist at %s", path)
	}
}

// AssertFileContains checks if a file contains the expected content
func (h *TestHelper) AssertFileContains(path string, expected string) {
	content, err := os.ReadFile(path)
	if err != nil {
		h.t.Errorf("failed to read file %s: %v", path, err)
		return
	}
	if !strings.Contains(string(content), expected) {
		h.t.Errorf("file %s does not contain expected content %q", path, expected)
	}
}

// AssertOutputContains checks if the captured output contains the expected string
func (h *TestHelper) AssertOutputContains(output string, expected string) {
	if !strings.Contains(output, expected) {
		h.t.Errorf("expected output to contain %q, got %q", expected, output)
	}
}

// AssertError checks if an error contains the expected message
func (h *TestHelper) AssertError(err error, expected string) {
	if err == nil {
		h.t.Errorf("expected error containing %q, got nil", expected)
		return
	}
	if !strings.Contains(err.Error(), expected) {
		h.t.Errorf("expected error containing %q, got %q", expected, err.Error())
	}
}
