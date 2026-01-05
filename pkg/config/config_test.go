package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWizardConfigCreate(t *testing.T) {
	cfg := &WizardConfig{
		InputFile:  "test.qif",
		OutputPath: "/output",
	}

	if cfg.InputFile != "test.qif" {
		t.Error("InputFile not set correctly")
	}
	if cfg.OutputPath != "/output" {
		t.Error("OutputPath not set correctly")
	}
}

func TestWizardConfigIsEmpty(t *testing.T) {
	// Empty config
	cfg := &WizardConfig{}
	if !cfg.IsEmpty() {
		t.Error("Empty config should return true")
	}

	// Config with input file
	cfg.InputFile = "test.qif"
	if cfg.IsEmpty() {
		t.Error("Config with InputFile should not be empty")
	}

	// Config with output path
	cfg2 := &WizardConfig{OutputPath: "/output"}
	if cfg2.IsEmpty() {
		t.Error("Config with OutputPath should not be empty")
	}
}

func TestWizardConfigSaveAndLoad(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.json")

	// Create and save config
	originalConfig := &WizardConfig{
		InputFile:            "test.qif",
		OutputPath:           "/output",
		ExportTransactions:   true,
		ExportBalanceHistory: false,
		SelectedAccounts:     "Checking,Savings",
		StartDate:            "2025-01-01",
		EndDate:              "2025-12-31",
		OutputFormat:         "CSV",
		AddTagForImport:      true,
	}

	err := originalConfig.SaveConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	// Load config
	loadedConfig, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify all fields match
	if loadedConfig.InputFile != originalConfig.InputFile {
		t.Errorf("InputFile mismatch: %s vs %s", loadedConfig.InputFile, originalConfig.InputFile)
	}
	if loadedConfig.OutputPath != originalConfig.OutputPath {
		t.Errorf("OutputPath mismatch: %s vs %s", loadedConfig.OutputPath, originalConfig.OutputPath)
	}
	if loadedConfig.ExportTransactions != originalConfig.ExportTransactions {
		t.Error("ExportTransactions mismatch")
	}
	if loadedConfig.ExportBalanceHistory != originalConfig.ExportBalanceHistory {
		t.Error("ExportBalanceHistory mismatch")
	}
	if loadedConfig.SelectedAccounts != originalConfig.SelectedAccounts {
		t.Errorf("SelectedAccounts mismatch: %s vs %s", loadedConfig.SelectedAccounts, originalConfig.SelectedAccounts)
	}
	if loadedConfig.StartDate != originalConfig.StartDate {
		t.Errorf("StartDate mismatch: %s vs %s", loadedConfig.StartDate, originalConfig.StartDate)
	}
	if loadedConfig.EndDate != originalConfig.EndDate {
		t.Errorf("EndDate mismatch: %s vs %s", loadedConfig.EndDate, originalConfig.EndDate)
	}
	if loadedConfig.OutputFormat != originalConfig.OutputFormat {
		t.Errorf("OutputFormat mismatch: %s vs %s", loadedConfig.OutputFormat, originalConfig.OutputFormat)
	}
	if loadedConfig.AddTagForImport != originalConfig.AddTagForImport {
		t.Error("AddTagForImport mismatch")
	}
}

func TestWizardConfigLoadNonexistentFile(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/config.json")
	if err == nil {
		t.Error("Loading nonexistent file should return error")
	}
}

func TestWizardConfigLoadInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid-config.json")

	// Write invalid JSON
	err := os.WriteFile(configPath, []byte("{ invalid json"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err = LoadConfig(configPath)
	if err == nil {
		t.Error("Loading invalid JSON should return error")
	}
}

func TestWizardConfigString(t *testing.T) {
	cfg := &WizardConfig{
		InputFile:    "test.qif",
		OutputPath:   "/output",
		OutputFormat: "CSV",
	}

	summary := cfg.String()
	if summary == "" {
		t.Error("String() returned empty string")
	}
	if !contains(summary, "test.qif") {
		t.Error("String() should contain InputFile")
	}
	if !contains(summary, "/output") {
		t.Error("String() should contain OutputPath")
	}
}

func TestWizardConfigStringWithAllFields(t *testing.T) {
	cfg := &WizardConfig{
		InputFile:             "test.qif",
		OutputPath:            "/output",
		ExportTransactions:    true,
		ExportBalanceHistory:  true,
		BalanceHistoryAccount: "Checking",
		SelectedAccounts:      "Checking,Savings",
		StartDate:             "2025-01-01",
		EndDate:               "2025-12-31",
		OutputFormat:          "CSV",
		CategoryMapFile:       "/mappings/categories.csv",
		PayeeMapFile:          "/mappings/payees.csv",
		AccountMapFile:        "/mappings/accounts.csv",
		TagMapFile:            "/mappings/tags.csv",
	}

	summary := cfg.String()
	if summary == "" {
		t.Error("String() returned empty string")
	}

	// Verify key sections are present
	if !contains(summary, "test.qif") {
		t.Error("Should contain InputFile")
	}
	if !contains(summary, "Both transactions and balance history") {
		t.Error("Should indicate both exports")
	}
	if !contains(summary, "Checking") {
		t.Error("Should contain BalanceHistoryAccount")
	}
	if !contains(summary, "categories.csv") {
		t.Error("Should contain CategoryMapFile")
	}
}

func TestWizardConfigSaveAndLoadWithMappings(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config-with-mappings.json")

	cfg := &WizardConfig{
		InputFile:       "test.qif",
		OutputPath:      "/output",
		CategoryMapFile: "/path/to/categories.csv",
		PayeeMapFile:    "/path/to/payees.csv",
		AccountMapFile:  "/path/to/accounts.csv",
		TagMapFile:      "/path/to/tags.csv",
	}

	err := cfg.SaveConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	loaded, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if loaded.CategoryMapFile != cfg.CategoryMapFile {
		t.Error("CategoryMapFile not preserved")
	}
	if loaded.PayeeMapFile != cfg.PayeeMapFile {
		t.Error("PayeeMapFile not preserved")
	}
	if loaded.AccountMapFile != cfg.AccountMapFile {
		t.Error("AccountMapFile not preserved")
	}
	if loaded.TagMapFile != cfg.TagMapFile {
		t.Error("TagMapFile not preserved")
	}
}

func TestWizardConfigSaveAndLoadBalanceHistory(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config-balance-history.json")

	cfg := &WizardConfig{
		InputFile:             "test.qif",
		OutputPath:            "/output",
		ExportTransactions:    false,
		ExportBalanceHistory:  true,
		BalanceHistoryAccount: "Savings Account",
		BalanceHistoryOpening: true,
		BalanceHistoryValue:   "5000.00",
		StartDate:             "2025-01-01",
		EndDate:               "2025-12-31",
	}

	err := cfg.SaveConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	loaded, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if !loaded.ExportBalanceHistory {
		t.Error("ExportBalanceHistory should be true")
	}
	if loaded.ExportTransactions {
		t.Error("ExportTransactions should be false")
	}
	if loaded.BalanceHistoryAccount != "Savings Account" {
		t.Error("BalanceHistoryAccount not preserved")
	}
	if !loaded.BalanceHistoryOpening {
		t.Error("BalanceHistoryOpening should be true")
	}
	if loaded.BalanceHistoryValue != "5000.00" {
		t.Error("BalanceHistoryValue not preserved")
	}
}

// Helper function to check if string contains substring
func contains(str, substr string) bool {
	return len(str) > 0 && len(substr) > 0 && (str == substr || len(str) >= len(substr) && (substr == str[:len(substr)] || substr == str[len(str)-len(substr):] || len(str) > len(substr)+1))
	// Simple contains check
	for i := 0; i <= len(str)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if str[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
