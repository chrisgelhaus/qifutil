package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// wizardCmd represents the wizard command
var wizardCmd = &cobra.Command{
	Use:   "wizard",
	Short: "Interactive guide to help you convert your QIF file",
	Long: `This command will guide you step-by-step through converting your QIF file.
It will ask you questions and help you create the right command for your needs.`,
	Run: func(cmd *cobra.Command, args []string) {
		reader := bufio.NewReader(os.Stdin)

		fmt.Println("\n=== QIFUTIL Conversion Wizard ===")
		fmt.Println("\nThis wizard will help you convert your Quicken (QIF) file.")
		fmt.Println("Press Enter after each response.")

		// Get input file
		fmt.Print("\nStep 1: Where is your QIF file? (Enter the file path): ")
		inputPath, _ := reader.ReadString('\n')
		// Clean the path from PowerShell artifacts
		inputPath = strings.TrimSpace(inputPath)
		inputPath = strings.TrimPrefix(inputPath, "& ") // Remove PowerShell invoke operator
		inputPath = strings.Trim(inputPath, "'\"")      // Remove both single and double quotes

		// Convert input path to absolute and clean it
		cleanInputPath := filepath.Clean(inputPath)
		if !filepath.IsAbs(cleanInputPath) {
			var err error
			cleanInputPath, err = filepath.Abs(cleanInputPath)
			if err != nil {
				fmt.Printf("Error with input path: %v\n", err)
				return
			}
		}
		inputFile = cleanInputPath

		if _, err := os.Stat(inputFile); os.IsNotExist(err) {
			fmt.Printf("\nError: Could not find file: %s\n", inputFile)
			fmt.Println("Please make sure the file exists and try again.")
			return
		}

		// Get output location
		defaultOutput := filepath.Join(filepath.Dir(cleanInputPath), "Exported")
		fmt.Printf("\nStep 2: Where should I save the converted files? (Press Enter for default: %s): ", defaultOutput)
		outputPath, _ := reader.ReadString('\n')
		outputPath = strings.TrimSpace(outputPath)
		outputPath = strings.TrimPrefix(outputPath, "& ") // Remove PowerShell invoke operator
		outputPath = strings.Trim(outputPath, "'\"")      // Remove quotes

		if outputPath == "" {
			outputPath = defaultOutput
		}

		// Clean and convert to absolute path
		outputPath = filepath.Clean(outputPath)
		if !filepath.IsAbs(outputPath) {
			var err error
			outputPath, err = filepath.Abs(outputPath)
			if err != nil {
				fmt.Printf("Error with path: %v\n", err)
				return
			}
		}

		// Create output directory if it doesn't exist
		fmt.Printf("Creating output directory: %s\n", outputPath)
		if err := os.MkdirAll(outputPath, 0755); err != nil {
			fmt.Printf("Error creating output directory: %v\n", err)
			return
		}

		// Show available accounts
		fmt.Println("\nStep 3: Let me check what accounts are in your file...")

		// Temporarily unset outputPath for account listing to avoid directory creation
		accounts := captureOutput(func() {
			// Create a new command instance for listing accounts
			tempCmd := *listAccountsCmd
			tempCmd.SetArgs([]string{"--inputFile", inputFile})
			tempCmd.Run(&tempCmd, []string{})
		})

		// Parse the accounts output to get a clean list
		accountList := parseAccountList(accounts)

		// Display the accounts with numbers
		fmt.Println("Found these accounts in your file:")
		for i, account := range accountList {
			fmt.Printf("%2d. %s\n", i+1, account)
		}

		// Ask about balance history generation early so we can capture account
		var generateBalanceHistoryLocal bool
		var balanceHistoryAccount string
		var balanceHistoryValue string
		var isBalanceHistoryOpening bool

		if getYesNoResponse(reader, "\nWould you like to generate balance history files? (y/n): ") {
			generateBalanceHistoryLocal = true

			fmt.Println("\nBalance history shows account balance changes over time (useful for Monarch Money).")
			fmt.Print("Enter the account number for balance history (see list above): ")
			accountInput, _ := reader.ReadString('\n')
			accountInput = strings.TrimSpace(accountInput)

			// Validate account number
			if accountNum, err := strconv.Atoi(accountInput); err == nil && accountNum > 0 && accountNum <= len(accountList) {
				balanceHistoryAccount = accountList[accountNum-1]
				fmt.Printf("Selected account: %s\n", balanceHistoryAccount)

				fmt.Println("\nChoose balance reference point:")
				fmt.Println("1. Current balance (ending - what the account totals at the end)")
				fmt.Println("2. Opening balance (starting - what the account had at the beginning)")
				fmt.Print("Choose a number (1-2, default: 1 - current balance): ")
				balanceChoice, _ := reader.ReadString('\n')
				balanceChoice = strings.TrimSpace(balanceChoice)

				isBalanceHistoryOpening = balanceChoice == "2"

				if isBalanceHistoryOpening {
					fmt.Print("Enter the opening balance (starting amount): ")
				} else {
					fmt.Print("Enter the current balance (ending amount): ")
				}
				balanceInput, _ := reader.ReadString('\n')
				balanceInput = strings.TrimSpace(balanceInput)

				// Validate balance is a number
				if _, err := strconv.ParseFloat(balanceInput, 64); err != nil {
					fmt.Printf("Invalid balance: %v. Balance history will not be generated.\n", err)
					generateBalanceHistoryLocal = false
				} else {
					balanceHistoryValue = balanceInput
				}
			} else {
				fmt.Println("Invalid account number. Balance history will not be generated.")
				generateBalanceHistoryLocal = false
			}
		}

		if getYesNoResponse(reader, "\nWould you like to convert specific accounts? (y/n, default: all accounts): ") {
			fmt.Print("Enter account numbers separated by commas (e.g., 1,3,5): ")
			numbersStr, _ := reader.ReadString('\n')
			numbersStr = strings.TrimSpace(numbersStr)

			// Convert selected numbers to account names
			var selectedAccountNames []string
			numbers := strings.Split(numbersStr, ",")
			for _, numStr := range numbers {
				numStr = strings.TrimSpace(numStr)
				if num, err := strconv.Atoi(numStr); err == nil && num > 0 && num <= len(accountList) {
					selectedAccountNames = append(selectedAccountNames, accountList[num-1])
				}
			}

			if len(selectedAccountNames) > 0 {
				selectedAccounts = strings.Join(selectedAccountNames, ",")
				fmt.Println("\nSelected accounts:")
				for _, name := range selectedAccountNames {
					fmt.Printf("  - %s\n", name)
				}
			} else {
				fmt.Println("\nNo valid account numbers entered. Processing all accounts.")
			}
		}

		// Ask about date filtering
		if getYesNoResponse(reader, "\nDo you want to filter by date range? (y/n): ") {
			startDate, endDate = getValidatedDateRange(reader)
		}

		// Ask about output format (balance history is already captured earlier)
		fmt.Print("\nWhat format would you like the output in?\n")
		fmt.Println("1. CSV (spreadsheet format, works with Excel) [default]")
		fmt.Println("2. JSON (technical format)")
		fmt.Println("3. XML (technical format)")
		fmt.Print("Choose a number (1-3): ")
		formatChoice, _ := reader.ReadString('\n')
		formatChoice = strings.TrimSpace(formatChoice)

		switch formatChoice {
		case "2":
			outputFormat = "JSON"
		case "3":
			outputFormat = "XML"
		default:
			outputFormat = "CSV"
		}

		// Ask about mapping files
		var categoryMapFile, payeeMapFile, accountMapFile, tagMapFile string

		if getYesNoResponse(reader, "\nWould you like to apply mapping files to transform data? (y/n): ") {
			fmt.Print("\nCategory mapping file (enter file path, or press Enter to skip): ")
			mapPath, _ := reader.ReadString('\n')
			mapPath = strings.TrimSpace(mapPath)
			mapPath = strings.TrimPrefix(mapPath, "& ")
			mapPath = strings.Trim(mapPath, "'\"")
			if mapPath != "" {
				absMapPath := filepath.Clean(mapPath)
				if !filepath.IsAbs(absMapPath) {
					var err error
					absMapPath, err = filepath.Abs(absMapPath)
					if err == nil {
						categoryMapFile = absMapPath
					}
				} else {
					categoryMapFile = absMapPath
				}
			}

			fmt.Print("Payee mapping file (enter file path, or press Enter to skip): ")
			mapPath, _ = reader.ReadString('\n')
			mapPath = strings.TrimSpace(mapPath)
			mapPath = strings.TrimPrefix(mapPath, "& ")
			mapPath = strings.Trim(mapPath, "'\"")
			if mapPath != "" {
				absMapPath := filepath.Clean(mapPath)
				if !filepath.IsAbs(absMapPath) {
					var err error
					absMapPath, err = filepath.Abs(absMapPath)
					if err == nil {
						payeeMapFile = absMapPath
					}
				} else {
					payeeMapFile = absMapPath
				}
			}

			fmt.Print("Account mapping file (enter file path, or press Enter to skip): ")
			mapPath, _ = reader.ReadString('\n')
			mapPath = strings.TrimSpace(mapPath)
			mapPath = strings.TrimPrefix(mapPath, "& ")
			mapPath = strings.Trim(mapPath, "'\"")
			if mapPath != "" {
				absMapPath := filepath.Clean(mapPath)
				if !filepath.IsAbs(absMapPath) {
					var err error
					absMapPath, err = filepath.Abs(absMapPath)
					if err == nil {
						accountMapFile = absMapPath
					}
				} else {
					accountMapFile = absMapPath
				}
			}

			fmt.Print("Tag mapping file (enter file path, or press Enter to skip): ")
			mapPath, _ = reader.ReadString('\n')
			mapPath = strings.TrimSpace(mapPath)
			mapPath = strings.TrimPrefix(mapPath, "& ")
			mapPath = strings.Trim(mapPath, "'\"")
			if mapPath != "" {
				absMapPath := filepath.Clean(mapPath)
				if !filepath.IsAbs(absMapPath) {
					var err error
					absMapPath, err = filepath.Abs(absMapPath)
					if err == nil {
						tagMapFile = absMapPath
					}
				} else {
					tagMapFile = absMapPath
				}
			}
		}

		fmt.Println("\nGreat! I'm ready to convert your file. Here's what I'm going to do:")
		fmt.Printf("- Read from: %s\n", inputFile)
		fmt.Printf("- Save to: %s\n", outputPath)
		if selectedAccounts != "" {
			fmt.Printf("- Convert these accounts: %s\n", selectedAccounts)
		} else {
			fmt.Println("- Convert all accounts")
		}
		if startDate != "" || endDate != "" {
			fmt.Printf("- Date range: %s to %s\n",
				ifEmpty(startDate, "beginning"),
				ifEmpty(endDate, "end"))
		}
		fmt.Printf("- Output format: %s\n", outputFormat)

		if generateBalanceHistoryLocal {
			if isBalanceHistoryOpening {
				fmt.Printf("- Generate balance history (opening balance: %s)\n", balanceHistoryValue)
			} else {
				fmt.Printf("- Generate balance history (current balance: %s)\n", balanceHistoryValue)
			}
		}

		if categoryMapFile != "" || payeeMapFile != "" || accountMapFile != "" || tagMapFile != "" {
			fmt.Println("- Mappings to apply:")
			if categoryMapFile != "" {
				fmt.Printf("  • Categories: %s\n", categoryMapFile)
			}
			if payeeMapFile != "" {
				fmt.Printf("  • Payees: %s\n", payeeMapFile)
			}
			if accountMapFile != "" {
				fmt.Printf("  • Accounts: %s\n", accountMapFile)
			}
			if tagMapFile != "" {
				fmt.Printf("  • Tags: %s\n", tagMapFile)
			}
		}

		fmt.Print("\nPress Enter to start the conversion (or Ctrl+C to cancel): ")
		reader.ReadString('\n')

		fmt.Println("\nStarting conversion...")

		// Make sure paths are absolute before running transactions
		if !filepath.IsAbs(outputPath) {
			var err error
			outputPath, err = filepath.Abs(outputPath)
			if err != nil {
				fmt.Printf("Error with output path: %v\n", err)
				return
			}
		}

		// Run the transactions command with explicit arguments
		transactionArgs := []string{
			"transactions",
			"--inputFile", inputFile,
			"--outputPath", outputPath,
		}

		if selectedAccounts != "" {
			transactionArgs = append(transactionArgs, "--accounts", selectedAccounts)
		}
		if startDate != "" {
			transactionArgs = append(transactionArgs, "--startDate", startDate)
		}
		if endDate != "" {
			transactionArgs = append(transactionArgs, "--endDate", endDate)
		}
		if categoryMapFile != "" {
			transactionArgs = append(transactionArgs, "--categoryMapFile", categoryMapFile)
		}
		if payeeMapFile != "" {
			transactionArgs = append(transactionArgs, "--payeeMapFile", payeeMapFile)
		}
		if accountMapFile != "" {
			transactionArgs = append(transactionArgs, "--accountMapFile", accountMapFile)
		}
		if tagMapFile != "" {
			transactionArgs = append(transactionArgs, "--tagMapFile", tagMapFile)
		}

		rootCmd.SetArgs(transactionArgs)
		rootCmd.Execute()

		// Generate balance history if requested
		if generateBalanceHistoryLocal && balanceHistoryAccount != "" {
			fmt.Println("\nGenerating balance history...")

			balanceHistoryArgs := []string{
				"export",
				"balance-history",
				"--inputFile", inputFile,
				"--outputPath", outputPath,
				"--accounts", balanceHistoryAccount,
			}

			if isBalanceHistoryOpening {
				balanceHistoryArgs = append(balanceHistoryArgs, "--openingBalance", balanceHistoryValue)
			} else {
				balanceHistoryArgs = append(balanceHistoryArgs, "--currentBalance", balanceHistoryValue)
			}

			if startDate != "" {
				balanceHistoryArgs = append(balanceHistoryArgs, "--startDate", startDate)
			}
			if endDate != "" {
				balanceHistoryArgs = append(balanceHistoryArgs, "--endDate", endDate)
			}

			rootCmd.SetArgs(balanceHistoryArgs)
			rootCmd.Execute()
		}
	},
}

func init() {
	rootCmd.AddCommand(wizardCmd)
}

func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = old

	return string(out)
}

func ifEmpty(str, fallback string) string {
	if str == "" {
		return fallback
	}
	return str
}

// parseAccountList extracts account names from the account listing output
func parseAccountList(output string) []string {
	var accounts []string

	// Find lines that start with a number followed by a period
	re := regexp.MustCompile(`(?m)^\d+\.\s+(.+)$`)
	matches := re.FindAllStringSubmatch(output, -1)

	for _, match := range matches {
		if len(match) > 1 {
			accounts = append(accounts, match[1])
		}
	}

	return accounts
}

// validateDate checks if a date string is in YYYY-MM-DD format
// Returns the parsed time.Time or an error
func validateDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, nil // Empty is valid (skip date)
	}

	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format, use YYYY-MM-DD")
	}

	return t, nil
}

// getValidatedDate reads and validates a date from user input
// Allows empty input to represent "skip" and returns an empty string in that case
func getValidatedDate(reader *bufio.Reader, prompt string) string {
	for {
		fmt.Print(prompt)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		// Empty input is allowed (skip filtering)
		if input == "" {
			return ""
		}

		// Validate the date format
		_, err := validateDate(input)
		if err != nil {
			fmt.Printf("Invalid date: %v. Please enter a date in YYYY-MM-DD format or press Enter to skip.\n", err)
			continue
		}

		return input
	}
}

// getValidatedDateRange reads and validates both start and end dates
// Ensures start date is not after end date
func getValidatedDateRange(reader *bufio.Reader) (string, string) {
	for {
		startDate := getValidatedDate(reader, "Start date (YYYY-MM-DD, or press Enter to skip): ")
		endDate := getValidatedDate(reader, "End date (YYYY-MM-DD, or press Enter to skip): ")

		// If both are empty, that's valid
		if startDate == "" && endDate == "" {
			return startDate, endDate
		}

		// If only one is specified, that's valid
		if startDate == "" || endDate == "" {
			return startDate, endDate
		}

		// Both are specified, validate ordering
		startTime, _ := time.Parse("2006-01-02", startDate)
		endTime, _ := time.Parse("2006-01-02", endDate)

		if startTime.After(endTime) {
			fmt.Println("Error: Start date cannot be after end date. Please try again.")
			continue
		}

		return startDate, endDate
	}
}

// getYesNoResponse reads and validates a yes/no response from the user
// Returns true for yes (y/Y), false for no (n/N or empty), and reprompts for invalid input
func getYesNoResponse(reader *bufio.Reader, prompt string) bool {
	for {
		fmt.Print(prompt)
		answer, _ := reader.ReadString('\n')
		answer = strings.ToLower(strings.TrimSpace(answer))

		switch answer {
		case "y":
			return true
		case "n", "":
			return false
		default:
			fmt.Println("Invalid input. Please enter 'y' for yes, 'n' for no, or press Enter for no.")
		}
	}
}
