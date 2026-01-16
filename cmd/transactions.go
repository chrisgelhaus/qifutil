/*
Copyright © 2025 Chris Gelhaus <chrisgelhaus@live.com>
*/
package cmd

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"qifutil/pkg/utils"

	"github.com/spf13/cobra"
)

var categoryMappingFile string
var accountMappingFile string
var payeeMappingFile string
var tagMappingFile string
var selectedAccounts string
var outputFields string
var outputPath string
var startDate string
var endDate string
var addTagForImport bool = false
var maxRecordsPerFile int = 5000
var csvColumns string

// Default columns for Monarch Money format
const DefaultMonarchColumns = "Date,Merchant,Category,Account,Original Statement,Notes,Amount,Tags"

type TransactionRecord struct {
	Date              string `json:"date" xml:"date"`
	Merchant          string `json:"merchant" xml:"merchant"`
	Category          string `json:"category" xml:"category"`
	Account           string `json:"account" xml:"account"`
	OriginalStatement string `json:"original_statement" xml:"original_statement"`
	Notes             string `json:"notes" xml:"notes"`
	Amount            string `json:"amount" xml:"amount"`
	Tags              string `json:"tags" xml:"tags"`
}
type transactionList struct {
	XMLName      xml.Name            `xml:"transactions"`
	Transactions []TransactionRecord `xml:"transaction"`
}

// transactionsCmd represents the transactions command
var transactionsCmd = &cobra.Command{
	Use:   "transactions",
	Short: "Convert your QIF file to CSV, JSON, or XML format",
	Long: `Convert your Quicken (QIF) file into a format you can use in other programs.

💡 First time user? Try our interactive guide:
   qifutil wizard

🔍 Want to see what accounts are in your file?
   qifutil list accounts --inputFile "YourFile.QIF"

DESCRIPTION:
  Reads transactions from a Quicken (QIF) file and exports them to CSV files.
  Each account's transactions are exported to separate CSV files.
  Large accounts are automatically split into multiple files for easier handling.

COMMON USES:
  1. Export to Monarch Money format:
     qifutil export transactions -i data.qif -o ./export/ -f MONARCH

  2. Export with custom CSV columns:
     qifutil export transactions -i data.qif -o ./export/ -f CSV \
       --csvColumns "Date,Merchant,Category,Amount"

  3. Export with mappings:
     qifutil export transactions -i data.qif -o ./export/ \
       -c categories.csv -p payees.csv

TIPS:
  - Use list-accounts command first to see available account names
  - Date filters accept YYYY-MM-DD format
  - Mapping files help standardize categories and payees
  - Set recordsPerFile=0 to keep all transactions in one file

OPTIONS:
  --inputFile          Required. Path to the QIF file to process
  --outputPath         Required. Directory where CSV files will be created
  --outputFormat       Optional. Output format: CSV, JSON, XML, or MONARCH (default: CSV)
  --csvColumns         Optional. Comma-separated column names for CSV output
                       (only applies to CSV format). Default is Monarch format.
  --accounts           Optional. Comma-separated list of accounts to process
  --categoryMapFile    Optional. CSV file mapping source to target categories
  --accountMapFile     Optional. CSV file mapping source to target account names
  --payeeMapFile       Optional. CSV file mapping source to target payee names
  --tagMapFile         Optional. CSV file mapping source to target tags
  --maxRecordsPerFile  Optional. Maximum transactions per output file (default: 5000)
  --addTagForImport    Optional. Add QIFIMPORT tag to all transactions

SUPPORTED FORMATS:
  CSV:     Generic CSV format. Column order is customizable via --csvColumns.
           Available columns: Date, Merchant, Category, Account,
           Original Statement, Notes, Amount, Tags

  MONARCH: Optimized for Monarch Money import. Equivalent to CSV format with
           all standard columns in the recommended order.

  JSON:    JSON array of transaction objects. One file per account.

  XML:     XML format with transaction elements. One file per account.

EXAMPLE COLUMNS:
  --csvColumns "Date,Merchant,Amount"
  --csvColumns "Date,Merchant,Category,Account,Amount"
  --csvColumns "Merchant,Amount,Category"

DEFAULT CSV COLUMNS (MONARCH):
  - Date (YYYY-MM-DD)
  - Merchant (Payee)
  - Category
  - Account
  - Original Statement
  - Notes (Memo)
  - Amount
  - Tags

MAPPING FILES:
  Mapping files should be CSV format with two columns:
  "source","target"
  
  Example category mapping:
  "Groceries","Food:Groceries"
  "Gas","Transportation:Fuel"`,
	PreRun: func(cmd *cobra.Command, args []string) {
		if inputFile == "" {
			fmt.Println("Error: Missing required flag --inputFile")
			os.Exit(1)
		}

		// Validate selected accounts format if provided
		if selectedAccounts != "" {
			accounts := strings.Split(selectedAccounts, ",")
			for _, account := range accounts {
				if strings.TrimSpace(account) == "" {
					fmt.Println("Error: Invalid account name in --accounts flag")
					os.Exit(1)
				}
			}
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
		fmt.Println("Starting transaction export...")

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

		// Try to create the output directory
		fmt.Printf("Creating output directory: %s\n", outputPath)
		if mkdirErr := os.MkdirAll(outputPath, 0755); mkdirErr != nil {
			fmt.Printf("Error creating output directory: %v\n", mkdirErr)
			os.Exit(1)
		}

		// Save current directory and change to output directory
		origDir, dirErr := os.Getwd()
		if dirErr != nil {
			fmt.Printf("Error getting current directory: %v\n", dirErr)
			os.Exit(1)
		}
		defer os.Chdir(origDir) // Restore original directory when done

		if chdirErr := os.Chdir(outputPath); chdirErr != nil {
			fmt.Printf("Error changing to output directory: %v\n", chdirErr)
			os.Exit(1)
		}

		// Validate input file exists and is readable
		if _, err := os.Stat(inputFile); os.IsNotExist(err) {
			fmt.Printf("Error: Input file not found: %s\n", inputFile)
			os.Exit(1)
		}

		// Process the selected accounts into a list
		var selectedAccountList []string
		if selectedAccounts != "" {
			selectedAccountList = strings.Split(selectedAccounts, ",")
			for i := range selectedAccountList {
				selectedAccountList[i] = strings.TrimSpace(selectedAccountList[i])
			}
		}

		// Output CSV Header
		var transactionRegexString string = `D(?<month>\d{1,2})\/(\s?(?<day>\d{1,2}))'(?<year>\d{2})[\r\n]+(U(?<amount1>.*?)[\r\n]+)(T(?<amount2>.*?)[\r\n]+)(C(?<cleared>.*?)[\r\n]+)((N(?<number>.*?)[\r\n]+)?)((P(?<payee>.*?)[\r\n]+)?)((M(?<memo>.*?)[\r\n]+)?)(L(?<category>.*?)[\r\n]+)`
		var accountBlockHeaderRegex string = `(?m)^!Account[^\n]*\n^N(.*?)\n^T(.*?)\n^\^\n^!Type:(Bank|CCard)\s*\n`

		// If MONARCH format is specified, use the default columns
		columnsToUse := csvColumns
		if strings.ToUpper(outputFormat) == "MONARCH" {
			columnsToUse = DefaultMonarchColumns
			outputFormat = "CSV" // Internally treat MONARCH as CSV
		}

		var outputCSVHeader string = columnsToUse + "\n"
		var categoryMapping map[string]string
		var payeeMapping map[string]string
		var accountMapping map[string]string
		var tagMapping map[string]string
		var err error

		// Load the Category Mapping
		if categoryMappingFile != "" {
			categoryMapping, err = loadMapping(categoryMappingFile)
			if err != nil {
				fmt.Println("Error loading category mapping:", err)
				return
			}
			fmt.Printf("%d Category Mappings Loaded:\n", len(categoryMapping))
			for k, v := range categoryMapping {
				fmt.Printf("  %s -> %s\n", k, v)
			}
		} else {
			fmt.Println("No category mapping file specified.")
		}

		// Load the Payee Mapping
		if payeeMappingFile != "" {
			payeeMapping, err = loadMapping(payeeMappingFile)
			if err != nil {
				fmt.Println("Error loading payee mapping:", err)
				return
			}
			fmt.Printf("%d Payee Mappings Loaded:\n", len(payeeMapping))
			for k, v := range payeeMapping {
				fmt.Printf("  %s -> %s\n", k, v)
			}
		} else {
			fmt.Println("No payee mapping file specified.")
		}

		// Load the Account Mapping
		if accountMappingFile != "" {
			accountMapping, err = loadMapping(accountMappingFile)
			if err != nil {
				fmt.Println("Error loading account mapping:", err)
				return
			}
			fmt.Printf("%d Account Mappings Loaded:\n", len(accountMapping))
			for k, v := range accountMapping {
				if v != "" {
					fmt.Printf("  %s -> %s\n", k, v)
				}
			}
		} else {
			fmt.Println("No account mapping file specified.")
		}

		// Load the Tag Mapping
		if tagMappingFile != "" {
			tagMapping, err = loadMapping(tagMappingFile)
			if err != nil {
				fmt.Println("Error loading tag mapping:", err)
				return
			}
			fmt.Printf("%d Tag Mappings Loaded:\n", len(tagMapping))
			for k, v := range tagMapping {
				fmt.Printf("  %s -> %s\n", k, v)
			}
		} else {
			fmt.Println("No tag mapping file specified.")
		}

		// Open the input file and find all the Bank and CCard blocks
		// Load input file
		inputBytes, err := os.ReadFile(inputFile)
		if err != nil {
			fmt.Println("Error reading file:", err)
		} else {
			fmt.Printf("Input file opened. Length: %d\n", len(inputBytes))
		}
		inputContent := string(inputBytes)

		// Standardize Line Endings to simplify Regex
		inputContent = strings.ReplaceAll(inputContent, "\r\n", "\n")

		// Gather the Account Blocks
		// Compile the regex
		regex, err := regexp.Compile(accountBlockHeaderRegex)
		if err != nil {
			return
		}
		accountBlocks := regex.FindAllStringSubmatchIndex(inputContent, -1)
		if len(accountBlocks) == 0 {
			fmt.Println("No matches found.")
		}

		// Initialize validation tracker for all accounts
		validator := utils.NewValidationTracker()

		// loop over each account block and Find all transaction matches
		for _, accountBlock := range accountBlocks {
			// Extract the account name from the matched block
			accountName := inputContent[accountBlock[2]:accountBlock[3]]

			// If specific accounts are selected, skip accounts that aren't in the list
			if len(selectedAccountList) > 0 {
				accountFound := false
				for _, selectedAccount := range selectedAccountList {
					if accountName == selectedAccount {
						accountFound = true
						break
					}
				}
				if !accountFound {
					continue
				}
			}

			// Map the account name using the account mapping if available
			var outputAccountName string
			if len(accountMapping[accountName]) > 0 {
				outputAccountName = accountMapping[accountName]
			} else {
				outputAccountName = accountName
			}

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

			fileIndex := 1
			count := 0
			var records []TransactionRecord

			// Determine file extension based on output format
			ext := ".csv"
			switch strings.ToUpper(outputFormat) {
			case "JSON":
				ext = ".json"
			case "XML":
				ext = ".xml"
			}

			// Create unique output file per Account
			outputFileName := fmt.Sprintf("%s_%d%s", accountName, fileIndex, ext)
			fmt.Printf("\nProcessing %s (File %d)\n", accountName, fileIndex)

			fullPath := filepath.Join(outputPath, outputFileName)
			if _, err := os.Stat(fullPath); err == nil {
				fmt.Printf("Warning: Overwriting existing file: %s\n", outputFileName)
			}

			outputFile, err := os.Create(fullPath)
			if err != nil {
				fmt.Printf("Error creating file %s: %v\n", outputFileName, err)
				return
			}
			if strings.ToUpper(outputFormat) == "XML" {
				outputFile.WriteString(xml.Header)
			}
			if strings.ToUpper(outputFormat) == "CSV" {
				if err := writeHeader(outputFile, outputCSVHeader); err != nil {
					outputFile.Close()
					fmt.Printf("Error: failed to write header to %s: %v\n", outputFileName, err)
					return
				}
			}

			// Extract the text between the type lines
			textBetweenTypes := inputContent[accountBlock[1]:endPos]

			// Use the existing pattern to match entries
			regex, err := regexp.Compile(transactionRegexString)
			if err != nil {
				return
			}

			// Find all transactions in the content.
			transactions := regex.FindAllStringSubmatch(textBetweenTypes, -1)

			// Print the number of transactions found
			fmt.Printf("Number of transactions found: %d\n", len(transactions))

			// Create a helper to extract named groups from regex matches
			getGroup := func(match []string, name string) string {
				for i, subexp := range regex.SubexpNames() {
					if subexp == name && i < len(match) {
						return match[i]
					}
				}
				return ""
			}

			for _, t := range transactions {
				if len(t) > 1 {
					month := strings.TrimSpace(getGroup(t, "month"))
					day := strings.TrimSpace(getGroup(t, "day"))
					year := strings.TrimSpace(getGroup(t, "year"))
					amount1 := strings.TrimSpace(getGroup(t, "amount1"))
					// Remove commas from amount for compatibility (e.g., "1,234.56" -> "1234.56")
					amount1 = strings.ReplaceAll(amount1, ",", "")
					// Parse amount to float and format with exactly 2 decimal places
					amountFloat, err := strconv.ParseFloat(amount1, 64)
					if err != nil {
						fmt.Printf("Warning: Could not parse amount '%s', using as-is\n", amount1)
					} else {
						amount1 = fmt.Sprintf("%.2f", amountFloat)
					}

					payee := strings.TrimSpace(getGroup(t, "payee"))
					// Apply the payee mapping
					payee = applyMapping(payee, payeeMapping)
					// Remove double quotes
					payee = strings.ReplaceAll(payee, "\"", "")

					transactionMemo := strings.TrimSpace(getGroup(t, "memo"))

					// Split the category and tag
					categoryRaw := strings.TrimSpace(getGroup(t, "category"))
					category, tag := utils.SplitCategoryAndTag(categoryRaw)

					// Trim whitespace
					category = strings.TrimSpace(category)
					// Apply the category mapping
					category = applyMapping(category, categoryMapping)

					// Trim whitespace
					tag = strings.TrimSpace(tag)
					// Apply the tag mapping
					tag = applyMapping(tag, tagMapping)

					// Prepend a custom Tag to the Category
					if addTagForImport {
						if tag != "" {
							tag = "QIFIMPORT," + tag
						} else {
							tag = "QIFIMPORT"
						}
					}

					// DATE FORMAT: YYYY-MM-DD
					fullYear := "20" + year
					month = "0" + month
					fullMonth := month[len(month)-2:]
					day = "0" + day
					fullDay := day[len(day)-2:]
					fullDate := fullYear + "-" + fullMonth + "-" + fullDay

					// Check if the transaction date is within the specified range
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

					// Validation tracking
					validator.RecordTransaction()
					if payee == "" {
						validator.AddMissingPayee()
					}
					if category == "" {
						validator.AddMissingCategory()
					}
					if amount1 == "0.00" || amount1 == "0" {
						validator.AddZeroAmount()
						validator.RecordTransactionIssue(fullDate, payee, amount1, category, "ZeroAmount")
					}

					record := TransactionRecord{
						Date:              fullDate,
						Merchant:          payee,
						Category:          category,
						Account:           outputAccountName,
						OriginalStatement: payee,
						Notes:             transactionMemo,
						Amount:            amount1,
						Tags:              tag,
					}

					if strings.ToUpper(outputFormat) == "JSON" {
						records = append(records, record)
					} else {
						line := buildCSVRow(record, columnsToUse)
						if err := writeTransaction(outputFile, line); err != nil {
							outputFile.Close()
							fmt.Printf("failed to write transaction: %v\n", err)
							return
						}
					}
					count++
					// Check if we need to split the file
					if maxRecordsPerFile != 0 && count%maxRecordsPerFile == 0 {
						// Close current file
						if strings.ToUpper(outputFormat) == "JSON" {
							jsonData, err := json.MarshalIndent(records, "", "  ")
							if err == nil {
								outputFile.Write(jsonData)
							}
							records = nil
						} else if strings.ToUpper(outputFormat) == "XML" {
							xmlData, err := xml.MarshalIndent(transactionList{Transactions: records}, "", "  ")
							if err == nil {
								outputFile.Write(xmlData)
							}
							records = nil
						}
						outputFile.Close()

						// Start new file
						fileIndex++
						outputFileName = fmt.Sprintf("%s_%d%s", accountName, fileIndex, ext)
						fullPath := filepath.Join(outputPath, outputFileName)
						fmt.Printf("\nCreating split file for %s (File %d) - Records %d to %d\n",
							accountName,
							fileIndex,
							(fileIndex-1)*maxRecordsPerFile+1,
							fileIndex*maxRecordsPerFile)

						outputFile, err = os.Create(fullPath)
						if err != nil {
							fmt.Printf("Error creating split file %s: %v\n", outputFileName, err)
							return
						}

						// Write appropriate headers for the new file
						if strings.ToUpper(outputFormat) == "XML" {
							outputFile.WriteString(xml.Header)
						}
						if strings.ToUpper(outputFormat) == "CSV" {
							if err := writeHeader(outputFile, outputCSVHeader); err != nil {
								outputFile.Close()
								fmt.Printf("Error: failed to write header to %s: %v\n", outputFileName, err)
								return
							}
						}
					}

				}
			}
			if strings.ToUpper(outputFormat) == "JSON" && len(records) > 0 {
				jsonData, err := json.MarshalIndent(records, "", "  ")
				if err == nil {
					outputFile.Write(jsonData)
				}
				records = nil
			} else if strings.ToUpper(outputFormat) == "XML" && len(records) > 0 {
				xmlData, err := xml.MarshalIndent(transactionList{Transactions: records}, "", "  ")
				if err == nil {
					outputFile.Write(xmlData)
				}
				records = nil
			}
			outputFile.Close()
		}

		// Print summary
		fmt.Println("\nExport Summary:")
		fmt.Printf("Input file: %s\n", inputFile)
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
		if len(selectedAccountList) > 0 {
			fmt.Printf("Processed accounts: %s\n", strings.Join(selectedAccountList, ", "))
		} else {
			fmt.Println("Processed all accounts")
		}
		fmt.Printf("Output directory: %s\n", outputPath)
		if maxRecordsPerFile > 0 {
			fmt.Printf("Split files: %d records per file (for Monarch compatibility)\n", maxRecordsPerFile)
		}
		fmt.Println("\nExport completed successfully!")

		// Print validation summary
		validator.PrintSummary()

		// Write detailed validation log (to transactions-specific log file)
		if err := validator.WriteValidationLogWithName(outputPath, "transactions_validation.log"); err != nil {
			fmt.Printf("Warning: Could not write validation log: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(transactionsCmd)

	// Add command-specific flags
	transactionsCmd.Flags().StringVarP(&outputFields, "outputFields", "", "", "Comma Separated list of fields to export from the QIF File.")
	transactionsCmd.Flags().StringVarP(&outputFormat, "outputFormat", "f", "CSV", "Output format (CSV, JSON, XML, MONARCH).")
	transactionsCmd.Flags().StringVarP(&csvColumns, "csvColumns", "", DefaultMonarchColumns, "Comma-separated list of columns for CSV output (only used with CSV format). Default is Monarch Money format.")
	transactionsCmd.Flags().StringVarP(&accountMappingFile, "accountMapFile", "a", "", "Supplied mapping file for accounts. Optional.")
	transactionsCmd.Flags().StringVarP(&categoryMappingFile, "categoryMapFile", "c", "", "Supplied mapping file for categories. Optional.")
	transactionsCmd.Flags().StringVarP(&payeeMappingFile, "payeeMapFile", "p", "", "Supplied mapping file for payees. Optional.")
	transactionsCmd.Flags().StringVarP(&tagMappingFile, "tagMapFile", "t", "", "Supplied mapping file for tags. Optional.")
	transactionsCmd.Flags().IntVarP(&maxRecordsPerFile, "recordsPerFile", "r", 5000, "Optional. Maximum number of records per CSV file. Default is 5000. If set to 0, all records will be written to a single file.")
	transactionsCmd.Flags().BoolVarP(&addTagForImport, "addTagForImport", "", true, "Add a custom tag to the transaction for import purposes")

	// Mark the shared required flags as required for this command
	transactionsCmd.MarkPersistentFlagRequired("inputFile")
	transactionsCmd.MarkPersistentFlagRequired("outputPath")
}

func loadMapping(filePath string) (map[string]string, error) {
	mapping := make(map[string]string)

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1 // Allow flexible line lengths

	for {
		record, err := reader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			fmt.Println("Error reading line in file:", filePath, ";", err)
			return nil, err
		}

		switch len(record) {
		case 1:
			// Single field - skip (move on)
			continue
		case 2:
			// Two fields - set key-value in map (but skip if value is empty)
			key := record[0]
			value := record[1]
			// Only add mapping if the target value is not empty
			if value != "" {
				mapping[key] = value
			}
		default:
			fmt.Println("Unexpected number of fields:", record)
		}
	}

	return mapping, nil
}

func applyMapping(input string, mapping map[string]string) string {
	// Loop through the mapping and look for the input value. If found, replace it with the mapped value.
	for oldValue, newValue := range mapping {
		if oldValue == input {
			fmt.Printf("Mapping: %s -> %s\n", input, newValue)
			return newValue
		}
	}
	// If no mapping is found, return the original input.
	return input
}

func writeHeader(f *os.File, h string) error {
	_, err := f.WriteString(h)
	return err
}

// buildCSVRow builds a CSV row from a TransactionRecord based on specified columns
func buildCSVRow(record TransactionRecord, columns string) string {
	columnList := strings.Split(columns, ",")
	values := make([]string, len(columnList))

	for i, col := range columnList {
		col = strings.TrimSpace(col)
		switch col {
		case "Date":
			values[i] = record.Date
		case "Merchant":
			values[i] = record.Merchant
		case "Category":
			values[i] = record.Category
		case "Account":
			values[i] = record.Account
		case "Original Statement":
			values[i] = record.OriginalStatement
		case "Notes":
			values[i] = record.Notes
		case "Amount":
			values[i] = record.Amount
		case "Tags":
			values[i] = record.Tags
		default:
			values[i] = ""
		}
	}

	// Build quoted CSV line
	var line strings.Builder
	for i, val := range values {
		if i > 0 {
			line.WriteString(",")
		}
		line.WriteString("\"" + strings.ReplaceAll(val, "\"", "\"\"") + "\"")
	}
	line.WriteString("\n")
	return line.String()
}

func writeTransaction(f *os.File, t string) error {
	_, err := f.WriteString(t)
	return err
}
