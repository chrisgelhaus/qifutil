/*
Copyright © 2025 Chris Gelhaus <chrisgelhaus@live.com>
*/
package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// listAccountsCmd represents the listAccounts command
var listAccountsCmd = &cobra.Command{
	Use:   "list-accounts",
	Short: "List all accounts in a QIF file",
	Long: `List all accounts found in a QIF file along with their types.

DESCRIPTION:
  Scans a Quicken (QIF) file and displays all accounts found within it.
  This command is typically used before running other commands to:
  - View available accounts for export
  - Verify account names for filtering
  - Check account types (Bank, Credit Card, etc.)

USAGE EXAMPLES:
  1. List account names only:
     qifutil list-accounts -i data.qif

  2. Show account types:
     qifutil list-accounts -i data.qif --showTypes

TIPS:
  - Use this command first to get correct account names for other commands
  - Account names are case-sensitive
  - Copy/paste account names to ensure exact matches`,

	Run: func(cmd *cobra.Command, args []string) {
		if inputFile == "" {
			fmt.Println("Error: Missing required flag --inputFile")
			fmt.Println("Usage: qifutil list-accounts -i <qif-file>")
			os.Exit(1)
		}

		// Validate input file exists
		if _, err := os.Stat(inputFile); os.IsNotExist(err) {
			fmt.Printf("Error: Input file not found: %s\n", inputFile)
			os.Exit(1)
		}

		fmt.Printf("Reading accounts from %s...\n\n", inputFile)

		// Load input file
		inputBytes, err := os.ReadFile(inputFile)
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			return
		}

		inputContent := string(inputBytes)
		// Standardize Line Endings
		inputContent = strings.ReplaceAll(inputContent, "\r\n", "\n")

		// Regex pattern to match account blocks
		accountBlockHeaderRegex := `(?m)^!Account[^\n]*\n^N(.*?)\n^T(.*?)\n^\^\n^!Type:(Bank|CCard)\s*\n`
		regex, err := regexp.Compile(accountBlockHeaderRegex)
		if err != nil {
			fmt.Println("Error compiling regex:", err)
			return
		}

		// Find all account blocks
		accountBlocks := regex.FindAllStringSubmatch(inputContent, -1)
		if len(accountBlocks) == 0 {
			fmt.Println("No accounts found in the file.")
			return
		}

		fmt.Printf("Found %d accounts in %s:\n\n", len(accountBlocks), inputFile)

		// Print each account
		for i, match := range accountBlocks {
			accountName := strings.TrimSpace(match[1])
			accountType := strings.TrimSpace(match[3])

			if showTypes {
				fmt.Printf("%d. %s (Type: %s)\n", i+1, accountName, accountType)
			} else {
				fmt.Printf("%d. %s\n", i+1, accountName)
			}
		}
	},
}

var showTypes bool

func init() {
	rootCmd.AddCommand(listAccountsCmd)

	// Add flags
	listAccountsCmd.Flags().BoolVar(&showTypes, "showTypes", false, "Show account types along with names")

	// Inherit persistent flags from root command
	listAccountsCmd.PersistentFlags().StringVarP(&inputFile, "inputFile", "i", "", "Path to the QIF file to process")

	// Mark required flags
	listAccountsCmd.MarkPersistentFlagRequired("inputFile")
}
