package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"qifutil/pkg/utils"
)

var generateBalanceHistory bool
var currentBalance string
var openingBalance string

// balanceHistoryCmd represents the balance-history export command
var balanceHistoryCmd = &cobra.Command{
	Use:   "balance-history",
	Short: "Generate balance history files for imported accounts",
	Long: `Generate daily balance history files for Monarch Money imports.

Balance history shows the account balance at the end of each day based on transaction sums.
This is useful when migrating historical data to Monarch Money.

REQUIREMENTS:
  - Exactly one account must be specified (--accounts)
  - Either --currentBalance or --openingBalance must be provided (mutually exclusive)

BALANCE OPTIONS:
  --currentBalance    The ending balance (as of the last transaction date or --endDate)
                      Works backward from this known balance. Use when you know what the
                      account balance should be at a specific point in time.

  --openingBalance    The starting balance before the first transaction date or --startDate
                      Works forward from this known balance. Use when you know the account
                      balance at the beginning of the period.

EXAMPLE:
  qifutil export balance-history \
    --inputFile data.qif \
    --outputPath ./export/ \
    --accounts "Checking Account" \
    --currentBalance 5000.00 \
    --startDate 2025-01-01 \
    --endDate 2025-12-31

CSV FORMAT:
  Date,Balance
  2025-01-01,5025.50
  2025-01-02,5010.25
  ...

TIPS:
  - Only dates with transactions are included
  - File naming: {AccountName}_balance_history_1.csv
  - If exceeded maxRecordsPerFile, creates _2.csv, _3.csv, etc.
  - Use list-accounts to find exact account names`,

	PreRun: func(cmd *cobra.Command, args []string) {
		// Validate input file exists
		if inputFile == "" {
			fmt.Println("Error: Missing required flag --inputFile")
			os.Exit(1)
		}

		// Validate single account is specified
		if selectedAccounts == "" {
			fmt.Println("Error: balance-history requires exactly one account (--accounts)")
			os.Exit(1)
		}

		accountList := strings.Split(selectedAccounts, ",")
		if len(accountList) != 1 {
			fmt.Println("Error: balance-history requires exactly one account. Multiple accounts specified.")
			os.Exit(1)
		}

		accountName := strings.TrimSpace(accountList[0])
		if accountName == "" {
			fmt.Println("Error: Account name cannot be empty")
			os.Exit(1)
		}

		// Validate mutually exclusive balance options
		hasCurrentBalance := currentBalance != ""
		hasOpeningBalance := openingBalance != ""

		if !hasCurrentBalance && !hasOpeningBalance {
			fmt.Println("Error: Either --currentBalance or --openingBalance must be specified")
			os.Exit(1)
		}

		if hasCurrentBalance && hasOpeningBalance {
			fmt.Println("Error: --currentBalance and --openingBalance are mutually exclusive. Use only one.")
			os.Exit(1)
		}

		// Validate balance is a valid number
		balanceStr := currentBalance
		if openingBalance != "" {
			balanceStr = openingBalance
		}

		if _, err := strconv.ParseFloat(balanceStr, 64); err != nil {
			fmt.Printf("Error: Invalid balance value '%s': must be a valid number\n", balanceStr)
			os.Exit(1)
		}

		// Validate date format if provided
		dateFormat := "2006-01-02"
		if startDate != "" {
			if _, err := time.Parse(dateFormat, startDate); err != nil {
				fmt.Println("Error: Invalid start date format. Use YYYY-MM-DD")
				os.Exit(1)
			}
		}
		if endDate != "" {
			if _, err := time.Parse(dateFormat, endDate); err != nil {
				fmt.Println("Error: Invalid end date format. Use YYYY-MM-DD")
				os.Exit(1)
			}
		}
		// Validate date range if both dates are provided
		if startDate != "" && endDate != "" {
			start, _ := time.Parse(dateFormat, startDate)
			end, _ := time.Parse(dateFormat, endDate)
			if end.Before(start) {
				fmt.Println("Error: End date cannot be before start date")
				os.Exit(1)
			}
		}
	},

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting balance history generation...")

		// Initialize validation tracker
		validator := utils.NewValidationTracker()

		// Ensure we have a valid output path
		if outputPath == "" {
			fmt.Println("Error: No output path specified")
			os.Exit(1)
		}

		// Clean and validate output path
		outputPath = filepath.Clean(outputPath)
		if !filepath.IsAbs(outputPath) {
			var absErr error
			outputPath, absErr = filepath.Abs(outputPath)
			if absErr != nil {
				fmt.Printf("Error with output path: %v\n", absErr)
				os.Exit(1)
			}
		}

		// Create the output directory
		fmt.Printf("Creating output directory: %s\n", outputPath)
		if mkdirErr := os.MkdirAll(outputPath, 0755); mkdirErr != nil {
			fmt.Printf("Error creating output directory: %v\n", mkdirErr)
			os.Exit(1)
		}

		// Validate input file exists
		if _, err := os.Stat(inputFile); os.IsNotExist(err) {
			fmt.Printf("Error: Input file not found: %s\n", inputFile)
			os.Exit(1)
		}

		// Get the account name (already validated to be single account)
		accountName := strings.TrimSpace(selectedAccounts)

		// Load and parse QIF file
		inputBytes, err := os.ReadFile(inputFile)
		if err != nil {
			fmt.Println("Error reading file:", err)
			os.Exit(1)
		}

		fmt.Printf("Input file opened. Length: %d\n", len(inputBytes))
		inputContent := string(inputBytes)

		// Standardize Line Endings
		inputContent = strings.ReplaceAll(inputContent, "\r\n", "\n")

		// Find the account block for the selected account
		accountBlockHeaderRegex := `(?m)^!Account[^\n]*\n^N(.*?)\n^T(.*?)\n^\^\n^!Type:(Bank|CCard)\s*\n`
		regex, err := regexp.Compile(accountBlockHeaderRegex)
		if err != nil {
			fmt.Println("Error compiling regex:", err)
			os.Exit(1)
		}

		accountBlocks := regex.FindAllStringSubmatchIndex(inputContent, -1)
		var selectedBlockContent string
		var foundAccount bool

		for _, accountBlock := range accountBlocks {
			currentAccountName := inputContent[accountBlock[2]:accountBlock[3]]
			if currentAccountName == accountName {
				foundAccount = true

				// Find the end of this account block
				restOfText := inputContent[accountBlock[1]:]
				nextTypePattern := `(?mi)^\s*!Type:.*$`
				nextTypeRe := regexp.MustCompile(nextTypePattern)
				nextLoc := nextTypeRe.FindStringIndex(restOfText)
				var endPos int
				if nextLoc != nil {
					endPos = accountBlock[1] + nextLoc[0]
				} else {
					endPos = len(inputContent)
				}

				selectedBlockContent = inputContent[accountBlock[1]:endPos]
				break
			}
		}

		if !foundAccount {
			fmt.Printf("Error: Account '%s' not found in file\n", accountName)
			os.Exit(1)
		}

		// Extract transactions for this account
		transactionRegexString := `D(?<month>\d{1,2})\/(\s?(?<day>\d{1,2}))'(?<year>\d{2})[\r\n]+(U(?<amount1>.*?)[\r\n]+)(T(?<amount2>.*?)[\r\n]+)(C(?<cleared>.*?)[\r\n]+)((N(?<number>.*?)[\r\n]+)?)(P(?<payee>.*?)[\r\n]+)((M(?<memo>.*?)[\r\n]+)?)(L(?<category>.*?)[\r\n]+)`
		transactionRegex, err := regexp.Compile(transactionRegexString)
		if err != nil {
			fmt.Println("Error compiling transaction regex:", err)
			os.Exit(1)
		}

		transactions := transactionRegex.FindAllStringSubmatch(selectedBlockContent, -1)
		fmt.Printf("Number of transactions found: %d\n", len(transactions))

		// Build daily balance map
		dailyBalances := make(map[string]float64)
		var dateKeys []string
		dateKeySet := make(map[string]bool)

		for _, t := range transactions {
			if len(t) > 1 {
				month := strings.TrimSpace(t[1])
				day := strings.TrimSpace(t[2])
				year := strings.TrimSpace(t[4])
				amount := strings.TrimSpace(t[6])

				// Remove commas from amount (for US-formatted numbers like 1,234.56)
				amount = strings.ReplaceAll(amount, ",", "")

				// Parse amount
				amountFloat, err := strconv.ParseFloat(amount, 64)
				if err != nil {
					fmt.Printf("Warning: Could not parse amount '%s' in transaction\n", amount)
					continue
				}

				// Validation tracking
				validator.RecordTransaction()
				if amountFloat == 0.0 {
					validator.AddZeroAmount()
				}

				// Format date
				fullYear := "20" + year
				month = "0" + month
				fullMonth := month[len(month)-2:]
				day = "0" + day
				fullDay := day[len(day)-2:]
				fullDate := fullYear + "-" + fullMonth + "-" + fullDay

				// Check date filtering
				transDate, _ := time.Parse("2006-01-02", fullDate)
				if startDate != "" {
					startDateTime, _ := time.Parse("2006-01-02", startDate)
					if transDate.Before(startDateTime) {
						continue
					}
				}
				if endDate != "" {
					endDateTime, _ := time.Parse("2006-01-02", endDate)
					if transDate.After(endDateTime) {
						continue
					}
				}

				// Accumulate daily balance
				dailyBalances[fullDate] += amountFloat

				// Track unique dates in order
				if !dateKeySet[fullDate] {
					dateKeySet[fullDate] = true
					dateKeys = append(dateKeys, fullDate)
				}
			}
		}

		if len(dailyBalances) == 0 {
			fmt.Println("Warning: No transactions found for balance history")
			return
		}

		// Sort dates
		sortDates(dateKeys)

		// Calculate running balances
		balanceFloat, _ := strconv.ParseFloat(currentBalance+openingBalance, 64) // One will be empty string
		isForwardCalculation := openingBalance != ""

		balanceRecords := make([]BalanceRecord, 0)

		if isForwardCalculation {
			// Forward calculation: opening balance + daily deltas
			for _, dateStr := range dateKeys {
				balanceFloat += dailyBalances[dateStr]
				balanceRecords = append(balanceRecords, BalanceRecord{
					Date:    dateStr,
					Balance: fmt.Sprintf("%.2f", balanceFloat),
				})
			}
		} else {
			// Backward calculation: start from current balance and work backward
			// First, sum all transactions to know total change
			var totalChange float64
			for _, dailyAmount := range dailyBalances {
				totalChange += dailyAmount
			}

			// Now calculate running balance from current balance going backward
			currentBal := balanceFloat - totalChange
			for _, dateStr := range dateKeys {
				currentBal += dailyBalances[dateStr]
				balanceRecords = append(balanceRecords, BalanceRecord{
					Date:    dateStr,
					Balance: fmt.Sprintf("%.2f", currentBal),
				})
			}
		}

		// Write balance history files
		fileIndex := 1
		count := 0
		outputFileName := fmt.Sprintf("%s_balance_history_%d.csv", accountName, fileIndex)
		fmt.Printf("\nGenerating balance history for %s (File %d)\n", accountName, fileIndex)

		fullPath := filepath.Join(outputPath, outputFileName)
		if _, err := os.Stat(fullPath); err == nil {
			fmt.Printf("Warning: Overwriting existing file: %s\n", outputFileName)
		}

		outputFile, err := os.Create(fullPath)
		if err != nil {
			fmt.Printf("Error creating file %s: %v\n", outputFileName, err)
			os.Exit(1)
		}

		// Write header
		if _, err := outputFile.WriteString("Date,Balance\n"); err != nil {
			outputFile.Close()
			fmt.Printf("Error writing header to %s: %v\n", outputFileName, err)
			os.Exit(1)
		}

		// Write balance records
		for _, record := range balanceRecords {
			line := fmt.Sprintf("%s,%s\n", record.Date, record.Balance)
			if _, err := outputFile.WriteString(line); err != nil {
				outputFile.Close()
				fmt.Printf("Error writing record to %s: %v\n", outputFileName, err)
				os.Exit(1)
			}

			count++

			// Check if we need to split the file
			if maxRecordsPerFile != 0 && count%maxRecordsPerFile == 0 {
				outputFile.Close()

				// Start new file
				fileIndex++
				outputFileName = fmt.Sprintf("%s_balance_history_%d.csv", accountName, fileIndex)
				fmt.Printf("Creating continuation file: %s (File %d)\n", outputFileName, fileIndex)

				fullPath := filepath.Join(outputPath, outputFileName)
				outputFile, err = os.Create(fullPath)
				if err != nil {
					fmt.Printf("Error creating file %s: %v\n", outputFileName, err)
					os.Exit(1)
				}

				// Write header for new file
				if _, err := outputFile.WriteString("Date,Balance\n"); err != nil {
					outputFile.Close()
					fmt.Printf("Error writing header to %s: %v\n", outputFileName, err)
					os.Exit(1)
				}
			}
		}

		outputFile.Close()

		// Print summary
		fmt.Println("\nBalance History Summary:")
		fmt.Printf("Input file: %s\n", inputFile)
		fmt.Printf("Account: %s\n", accountName)
		if isForwardCalculation {
			fmt.Printf("Opening balance: %s\n", openingBalance)
		} else {
			fmt.Printf("Current balance (as of last transaction): %s\n", currentBalance)
		}
		if startDate != "" || endDate != "" {
			start := "earliest"
			if startDate != "" {
				start = startDate
			}
			end := "latest"
			if endDate != "" {
				end = endDate
			}
			fmt.Printf("Date range: %s to %s\n", start, end)
		}
		fmt.Printf("Balance records generated: %d\n", len(balanceRecords))
		fmt.Printf("Output directory: %s\n", outputPath)
		if maxRecordsPerFile > 0 {
			fmt.Printf("Split files: %d records per file\n", maxRecordsPerFile)
		}
		fmt.Println("\nBalance history generation completed successfully!")

		// Print validation summary
		validator.PrintSummary()
	},
}

// BalanceRecord represents a daily balance entry
type BalanceRecord struct {
	Date    string
	Balance string
}

func init() {
	exportCmd.AddCommand(balanceHistoryCmd)

	// Add command-specific flags
	balanceHistoryCmd.Flags().StringVarP(&currentBalance, "currentBalance", "", "", "The ending account balance (as of the last transaction date). Use for backward calculation. Mutually exclusive with --openingBalance.")
	balanceHistoryCmd.Flags().StringVarP(&openingBalance, "openingBalance", "", "", "The starting account balance (before the first transaction date). Use for forward calculation. Mutually exclusive with --currentBalance.")
}

// sortDates sorts a slice of date strings in YYYY-MM-DD format
func sortDates(dates []string) {
	for i := 0; i < len(dates); i++ {
		for j := i + 1; j < len(dates); j++ {
			if dates[j] < dates[i] {
				dates[i], dates[j] = dates[j], dates[i]
			}
		}
	}
}
