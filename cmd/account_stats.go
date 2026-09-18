/*
Copyright © 2025 Chris Gelhaus
*/
package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"qifutil/pkg/utils"
	"github.com/spf13/cobra"
)

type AccountStats struct {
	Name             string
	Type             string
	TransactionCount int
	EarliestDate     time.Time
	LatestDate       time.Time
}

// accountStatsCmd represents the account-stats command
var accountStatsCmd = &cobra.Command{
	Use:   "account-stats",
	Short: "Show transaction statistics for accounts",
	Long: `Display transaction statistics for accounts in a QIF file.

DESCRIPTION:
  Analyzes accounts in a Quicken (QIF) file and shows key statistics:
  - Number of transactions per account
  - Date range of transactions (earliest to latest)
  - Account types

USAGE EXAMPLES:
  1. Show stats for all accounts:
     qifutil account-stats -i data.qif

  2. Stats for specific accounts:
     qifutil account-stats -i data.qif -a "Checking,Credit Card"

  3. View specific account history:
     qifutil account-stats -i data.qif -a "Savings Account"

TIPS:
  - Use list-accounts first to get exact account names
  - Account names are case-sensitive
  - Use quotes around account names with spaces`,

	RunE: func(cmd *cobra.Command, args []string) error {
		// A failure past flag parsing is about the data, not how the command
		// was called, so cobra should print it without the usage text.
		cmd.SilenceUsage = true

		if inputFile == "" {
			return fmt.Errorf("missing required flag --inputFile")
		}

		// Validate input file exists
		if _, err := os.Stat(inputFile); os.IsNotExist(err) {
			return fmt.Errorf("input file not found: %s", inputFile)
		}

		fmt.Printf("Analyzing accounts in %s...\n\n", inputFile)

		// Load input file
		inputBytes, err := os.ReadFile(inputFile)
		if err != nil {
			return fmt.Errorf("reading %q: %w", inputFile, err)
		}

		inputContent := string(inputBytes)
		// Standardize Line Endings
		inputContent = strings.ReplaceAll(inputContent, "\r\n", "\n")

		// Process selected accounts
		var selectedAccountList []string
		if selectedAccounts != "" {
			selectedAccountList = strings.Split(selectedAccounts, ",")
			for i := range selectedAccountList {
				selectedAccountList[i] = strings.TrimSpace(selectedAccountList[i])
			}
		}

		// Regex pattern to match account blocks
		regex := anyAccountBlockPattern

		// Find all account blocks with positions
		accountBlocksIdx := regex.FindAllStringSubmatchIndex(inputContent, -1)
		if len(accountBlocksIdx) == 0 {
			fmt.Println("No accounts found in the file.")
			return nil
		}

		fmt.Printf("Account Statistics from %s:\n\n", inputFile)

		// Process each account
		for i, blockIdx := range accountBlocksIdx {
			// Groups: [0,1]=full [2,3]=name [4,5]=type-letter [6,7]=account-type
			accountName := strings.TrimSpace(inputContent[blockIdx[2]:blockIdx[3]])
			accountType := strings.TrimSpace(inputContent[blockIdx[6]:blockIdx[7]])

			// Skip if not in selected accounts
			if len(selectedAccountList) > 0 {
				found := false
				for _, selected := range selectedAccountList {
					if accountName == selected {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}

			// Find transactions in the section following this account
			var endPos int
			if i+1 < len(accountBlocksIdx) {
				endPos = accountBlocksIdx[i+1][0]
			} else {
				endPos = len(inputContent)
			}
			accountContent := inputContent[blockIdx[1]:endPos]

			// QIF records are read field by field: a single pattern cannot
			// express fields that are optional and may appear in any order, and
			// the separator before the year carries the century.
			transactionCount := 0
			var dates []time.Time
			for _, record := range utils.SplitRecords(accountContent) {
				// The next account's header block falls inside this account's text.
				if strings.HasPrefix(strings.TrimSpace(record), "!") {
					continue
				}
				fields := utils.ParseTransactionRecord(record)
				if fields.Date == "" {
					continue
				}

				transactionCount++
				if date, err := utils.ParseQIFDateField(fields.Date); err == nil {
					dates = append(dates, date)
				}
			}

			stats := AccountStats{
				Name:             accountName,
				Type:             accountType,
				TransactionCount: transactionCount,
			}

			if stats.TransactionCount > 0 {
				fmt.Printf("Account: %s (Type: %s)\n", stats.Name, stats.Type)
				fmt.Printf("  Transactions: %d\n", stats.TransactionCount)

				// A date that cannot be read leaves the range narrower rather
				// than dragging it to a sentinel year.
				if len(dates) > 0 {
					stats.EarliestDate, stats.LatestDate = dates[0], dates[0]
					for _, date := range dates[1:] {
						if date.Before(stats.EarliestDate) {
							stats.EarliestDate = date
						}
						if date.After(stats.LatestDate) {
							stats.LatestDate = date
						}
					}
					fmt.Printf("  Date Range: %s to %s\n",
						stats.EarliestDate.Format("2006-01-02"),
						stats.LatestDate.Format("2006-01-02"))
				}
				fmt.Println()
			} else {
				// Print statistics for accounts with no transactions
				fmt.Printf("Account: %s (Type: %s)\n", stats.Name, stats.Type)
				fmt.Printf("  No transactions found\n\n")
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(accountStatsCmd)

	// Add flags
	accountStatsCmd.PersistentFlags().StringVarP(&inputFile, "inputFile", "i", "", "Path to the QIF file to process")
	accountStatsCmd.Flags().StringVarP(&selectedAccounts, "accounts", "a", "", "Optional. Comma-separated list of accounts to analyze")

	// Mark required flags
	accountStatsCmd.MarkPersistentFlagRequired("inputFile")
}
