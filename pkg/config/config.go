package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// WizardConfig stores wizard settings that can be saved and loaded
type WizardConfig struct {
	// Basic settings
	InputFile  string `json:"inputFile"`
	OutputPath string `json:"outputPath"`

	// Export type selection
	ExportTransactions    bool   `json:"exportTransactions"`
	ExportBalanceHistory  bool   `json:"exportBalanceHistory"`
	BalanceHistoryAccount string `json:"balanceHistoryAccount"`
	BalanceHistoryOpening bool   `json:"balanceHistoryOpening"` // true = opening balance, false = current balance
	BalanceHistoryValue   string `json:"balanceHistoryValue"`

	// Transaction settings
	SelectedAccounts string `json:"selectedAccounts"`
	StartDate        string `json:"startDate"`
	EndDate          string `json:"endDate"`
	OutputFormat     string `json:"outputFormat"`

	// Mapping files
	CategoryMapFile string `json:"categoryMapFile"`
	PayeeMapFile    string `json:"payeeMapFile"`
	AccountMapFile  string `json:"accountMapFile"`
	TagMapFile      string `json:"tagMapFile"`

	// Options
	AddTagForImport  bool `json:"addTagForImport"`
	SkipZeroAmounts  bool `json:"skipZeroAmounts"`
}

// LoadConfig loads configuration from a JSON file
func LoadConfig(filePath string) (*WizardConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config WizardConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// SaveConfig saves configuration to a JSON file
func (c *WizardConfig) SaveConfig(filePath string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// IsEmpty checks if the config has been populated
func (c *WizardConfig) IsEmpty() bool {
	return c.InputFile == "" && c.OutputPath == ""
}

// String returns a human-readable summary of the config
func (c *WizardConfig) String() string {
	summary := "Configuration Summary:\n"
	summary += fmt.Sprintf("  Input File: %s\n", c.InputFile)
	summary += fmt.Sprintf("  Output Path: %s\n", c.OutputPath)

	if c.ExportTransactions && c.ExportBalanceHistory {
		summary += fmt.Sprintf("  Export Type: Both transactions and balance history\n")
	} else if c.ExportTransactions {
		summary += fmt.Sprintf("  Export Type: Transactions only\n")
	} else if c.ExportBalanceHistory {
		summary += fmt.Sprintf("  Export Type: Balance history only\n")
	}

	if c.SelectedAccounts != "" {
		summary += fmt.Sprintf("  Accounts: %s\n", c.SelectedAccounts)
	}

	if c.BalanceHistoryAccount != "" {
		summary += fmt.Sprintf("  Balance History Account: %s\n", c.BalanceHistoryAccount)
	}

	if c.StartDate != "" || c.EndDate != "" {
		summary += fmt.Sprintf("  Date Range: %s to %s\n", c.StartDate, c.EndDate)
	}

	if c.OutputFormat != "" {
		summary += fmt.Sprintf("  Format: %s\n", c.OutputFormat)
	}

	if c.CategoryMapFile != "" || c.PayeeMapFile != "" || c.AccountMapFile != "" || c.TagMapFile != "" {
		summary += "  Mappings Applied:\n"
		if c.CategoryMapFile != "" {
			summary += fmt.Sprintf("    • Categories: %s\n", c.CategoryMapFile)
		}
		if c.PayeeMapFile != "" {
			summary += fmt.Sprintf("    • Payees: %s\n", c.PayeeMapFile)
		}
		if c.AccountMapFile != "" {
			summary += fmt.Sprintf("    • Accounts: %s\n", c.AccountMapFile)
		}
		if c.TagMapFile != "" {
			summary += fmt.Sprintf("    • Tags: %s\n", c.TagMapFile)
		}
	}

	return summary
}
