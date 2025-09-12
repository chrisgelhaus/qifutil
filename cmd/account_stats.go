/*
Copyright © 2025 Chris Gelhaus <chrisgelhaus@live.com>
*/
package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

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

	Run: func(cmd *cobra.Command, args []string) {
		if inputFile == "" {
			fmt.Println("Error: Missing required flag --inputFile")
			fmt.Println("Usage: qifutil account-stats -i <qif-file> [-a <account-names>]")
			os.Exit(1)
		}

		// Validate input file exists
		if _, err := os.Stat(inputFile); os.IsNotExist(err) {
			fmt.Printf("Error: Input file not found: %s\n", inputFile)
			os.Exit(1)
		}

		fmt.Printf("Analyzing accounts in %s...\n\n", inputFile)

		// Load input file
		inputBytes, err := os.ReadFile(inputFile)
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			return
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
		accountBlockHeaderRegex := `!Account\nN(.*?)\nT(.*?)\n\^\n!Type:(.*?)\n`
		regex, err := regexp.Compile(accountBlockHeaderRegex)
		if err != nil {
			fmt.Println("Error compiling regex:", err)
			return
		}

		// Transaction regex - only match dates
		transactionRegex := regexp.MustCompile(`D(\d{1,2})/(\d{1,2})'(\d{2})`)

		// Find all account blocks
		accountBlocks := regex.FindAllStringSubmatch(inputContent, -1)
		if len(accountBlocks) == 0 {
			fmt.Println("No accounts found in the file.")
			return
		}

		fmt.Printf("Account Statistics from %s:\n\n", inputFile)

		// Process each account
		for _, block := range accountBlocks {
			accountName := strings.TrimSpace(block[1]) // Name group
			accountType := strings.TrimSpace(block[3]) // AccountType group

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
			loc := regex.FindStringIndex(inputContent)
			if loc == nil {
				continue
			}

			nextAccountLoc := regex.FindStringIndex(inputContent[loc[1]:])
			var accountContent string
			if nextAccountLoc != nil {
				accountContent = inputContent[loc[1] : loc[1]+nextAccountLoc[0]]
			} else {
				accountContent = inputContent[loc[1]:]
			}

			// Find all transactions
			transactions := transactionRegex.FindAllStringSubmatch(accountContent, -1)

			// Process transactions
			stats := AccountStats{
				Name:             accountName,
				Type:             accountType,
				TransactionCount: len(transactions),
				EarliestDate:     time.Date(2099, 12, 31, 0, 0, 0, 0, time.UTC),
				LatestDate:       time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
			} // Process dates if we have transactions
			if stats.TransactionCount > 0 {
				for _, t := range transactions {
					// Process date
					month, _ := strconv.Atoi(t[1])
					day, _ := strconv.Atoi(t[2])
					year, _ := strconv.Atoi("20" + t[3])
					date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)

					if date.Before(stats.EarliestDate) {
						stats.EarliestDate = date
					}
					if date.After(stats.LatestDate) {
						stats.LatestDate = date
					}
				}

				// Print statistics
				fmt.Printf("Account: %s (Type: %s)\n", stats.Name, stats.Type)
				fmt.Printf("  Transactions: %d\n", stats.TransactionCount)
				fmt.Printf("  Date Range: %s to %s\n",
					stats.EarliestDate.Format("2006-01-02"),
					stats.LatestDate.Format("2006-01-02"))
				fmt.Println()
			} else {
				// Print statistics for accounts with no transactions
				fmt.Printf("Account: %s (Type: %s)\n", stats.Name, stats.Type)
				fmt.Printf("  No transactions found\n\n")
			}
		}
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
