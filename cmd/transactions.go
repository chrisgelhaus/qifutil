/*
Copyright © 2025 Chris Gelhaus
*/
package cmd

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
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
var skipZeroAmounts bool = false
var preserveOriginalCategory bool = false
var expandSplits bool = false
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
   qifutil list-accounts --inputFile "YourFile.QIF"

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

  4. Export with mappings, keeping a back-reference to the original category:
     qifutil export transactions --inputFile data.qif --outputPath ./export/ \
       -c categories.csv --preserveOriginalCategory

TIPS:
  - Use list-accounts command first to see available account names
  - Date filters accept YYYY-MM-DD format
  - Mapping files help standardize categories and payees
  - Set recordsPerFile=0 to keep all transactions in one file

OPTIONS:
  --inputFile          Required. Path to the QIF file to process
  --outputPath         Required. Directory where CSV files will be created
  --outputFormat       Optional. Output format: CSV, JSON, XML, or MONARCH (default: CSV)
  --csvColumns         Optional. Comma-separated column names for CSV output.
                       An unrecognised name is refused. Only applies to CSV
                       output; ignored with a note for JSON and XML.
                       Default is Monarch format.
  --accounts           Optional. Comma-separated list of accounts to process
  --categoryMapFile    Optional. CSV file mapping source to target categories
  --accountMapFile     Optional. CSV file mapping source to target account names
  --payeeMapFile       Optional. CSV file mapping source to target payee names
  --tagMapFile         Optional. CSV file mapping source to target tags
  --maxRecordsPerFile  Optional. Maximum transactions per output file (default: 5000)
  --addTagForImport    Optional. Add a QIFIMPORT tag to every transaction.
                       ON BY DEFAULT. Pass --addTagForImport=false to
                       export without it.
  --preserveOriginalCategory
                       Optional. When a category mapping rewrites a category,
                       append the original to Notes so it can be traced back,
                       e.g. "Weekly shop [Original Category: Insurance:Auto]"
  --expandSplits       Optional. Export each line item of a split transaction
                       as its own row, so its category and amount are kept
                       instead of collapsing into the transaction total

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
  - Notes (Memo, plus the original category when --preserveOriginalCategory is set)
  - Amount
  - Tags

MAPPING FILES:
  Mapping files should be CSV format with two columns:
  "source","target"
  
  Example category mapping:
  "Groceries","Food:Groceries"
  "Gas","Transportation:Fuel"`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Past flag parsing, a failure is about the data rather than how the
		// command was called, so cobra should print the error without the usage
		// text. A bad flag still gets usage, because that is decided earlier.
		cmd.SilenceUsage = true

		if inputFile == "" {
			return fmt.Errorf("missing required flag --inputFile")
		}

		// Validate selected accounts format if provided
		if selectedAccounts != "" {
			accounts := strings.Split(selectedAccounts, ",")
			for _, account := range accounts {
				if strings.TrimSpace(account) == "" {
					return fmt.Errorf("invalid account name in --accounts flag")
				}
			}
		}

		// Validate date format if provided
		dateFormat := "2006-01-02"
		if startDate != "" {
			if _, err := time.Parse(dateFormat, startDate); err != nil {
				return fmt.Errorf("invalid start date %q: use YYYY-MM-DD", startDate)
			}
		}
		if endDate != "" {
			if _, err := time.Parse(dateFormat, endDate); err != nil {
				return fmt.Errorf("invalid end date %q: use YYYY-MM-DD", endDate)
			}
		}
		// Validate date range if both dates are provided
		if startDate != "" && endDate != "" {
			start, _ := time.Parse(dateFormat, startDate)
			end, _ := time.Parse(dateFormat, endDate)
			if end.Before(start) {
				return fmt.Errorf("end date %s is before start date %s", endDate, startDate)
			}
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Starting transaction export...")

		// Ensure we have a valid output path
		if outputPath == "" {
			return fmt.Errorf("missing required flag --outputPath")
		}

		// Clean and validate output path
		outputPath = filepath.Clean(outputPath)
		if !filepath.IsAbs(outputPath) {
			var absErr error
			outputPath, absErr = filepath.Abs(outputPath)
			if absErr != nil {
				return fmt.Errorf("resolving output path %q: %w", outputPath, absErr)
			}
		}

		// Try to create the output directory
		fmt.Printf("Creating output directory: %s\n", outputPath)
		if mkdirErr := os.MkdirAll(outputPath, 0755); mkdirErr != nil {
			return fmt.Errorf("creating output directory %q: %w", outputPath, mkdirErr)
		}

		// Validate input file exists and is readable
		if _, err := os.Stat(inputFile); os.IsNotExist(err) {
			return fmt.Errorf("input file not found: %s", inputFile)
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
		var accountBlockHeaderRegex string = `(?m)^!Account[^\n]*\n^N(.*?)\n^T(.*?)\n^\^\n^!Type:(Bank|CCard)\s*\n`

		// If MONARCH format is specified, use the default columns
		columnsToUse := csvColumns
		if strings.ToUpper(outputFormat) == "MONARCH" {
			columnsToUse = DefaultMonarchColumns
			outputFormat = "CSV" // Internally treat MONARCH as CSV
		}

		// A misspelled column used to produce a column that was blank in every
		// row, which is easy to miss in a file of thousands.
		if strings.ToUpper(outputFormat) == "CSV" {
			if err := validateCSVColumns(columnsToUse); err != nil {
				return err
			}
		} else if columnsToUse != DefaultMonarchColumns {
			fmt.Printf("Note: --csvColumns does not apply to %s output and is ignored\n",
				strings.ToUpper(outputFormat))
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
				return fmt.Errorf("loading category mapping %q: %w", categoryMappingFile, err)
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
				return fmt.Errorf("loading payee mapping %q: %w", payeeMappingFile, err)
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
				return fmt.Errorf("loading account mapping %q: %w", accountMappingFile, err)
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
				return fmt.Errorf("loading tag mapping %q: %w", tagMappingFile, err)
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
			return fmt.Errorf("compiling account block pattern: %w", err)
		}
		accountBlocks := regex.FindAllStringSubmatchIndex(inputContent, -1)
		if len(accountBlocks) == 0 {
			fmt.Println("No matches found.")
		}

		// Wrap each mapping so the rules it uses can be tracked, and the ones
		// it never uses reported at the end.
		categories := newMappingSet("Category", categoryMapping)
		payees := newMappingSet("Payee", payeeMapping)
		accounts := newMappingSet("Account", accountMapping)
		tags := newMappingSet("Tag", tagMapping)

		// Initialize validation tracker for all accounts
		validator := utils.NewValidationTracker()

		// Account names are not file names, so the stems already written are
		// tracked to keep two accounts from landing on the same file.
		usedFileBases := make(map[string]bool)
		skippedAccounts := 0

		// loop over each account block and Find all transaction matches
	accountLoop:
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
			outputAccountName = accounts.apply(accountName)
			if accounts.loaded() && outputAccountName == accountName {
				validator.AddUnmatchedData("account", accountName)
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

			// Create unique output file per Account. The account name may contain
			// characters a file name cannot, so it is sanitised first.
			fileBase := uniqueFileBase(accountName, usedFileBases)
			if fileBase != accountName {
				fmt.Printf("Note: account %q is written to files named %q\n", accountName, fileBase)
			}
			outputFileName := fmt.Sprintf("%s_%d%s", fileBase, fileIndex, ext)
			fmt.Printf("\nProcessing %s (File %d)\n", accountName, fileIndex)

			fullPath := filepath.Join(outputPath, outputFileName)
			outputFile, err := os.Create(fullPath)
			if err != nil {
				fmt.Printf("Error creating file %s: %v\n", outputFileName, err)
				skippedAccounts++
				continue accountLoop
			}
			if strings.ToUpper(outputFormat) == "XML" {
				_, err := outputFile.WriteString(xml.Header)
				if err != nil {
					outputFile.Close()
					fmt.Printf("Error: failed to write XML header to %s: %v\n", outputFileName, err)
					skippedAccounts++
					continue accountLoop
				}
			}
			if strings.ToUpper(outputFormat) == "CSV" {
				if err := writeHeader(outputFile, outputCSVHeader); err != nil {
					outputFile.Close()
					fmt.Printf("Error: failed to write header to %s: %v\n", outputFileName, err)
					skippedAccounts++
					continue accountLoop
				}
			}

			// Extract the text between the type lines
			textBetweenTypes := inputContent[accountBlock[1]:endPos]

			// QIF is line oriented: every line carries a one letter field code and
			// a value, and "^" ends a record. Most fields are optional and their
			// order is not fixed, which no single pattern can express, so records
			// are read field by field.
			var transactions []utils.TransactionFields
			undatedRecords := 0
			for _, record := range utils.SplitRecords(textBetweenTypes) {
				// The next account's header block falls inside this account's text.
				// It is QIF structure rather than a transaction, so it is neither
				// exported nor reported as one that failed to export.
				if strings.HasPrefix(strings.TrimSpace(record), "!") {
					continue
				}

				fields := utils.ParseTransactionRecord(record)
				if fields.Date == "" {
					undatedRecords++
					continue
				}
				transactions = append(transactions, fields)
			}

			// Print the number of transactions found
			fmt.Printf("Number of transactions found: %d\n", len(transactions))
			if undatedRecords > 0 {
				fmt.Printf("Note: %d record(s) had no date and were not exported\n", undatedRecords)
			}

			// Counted per account: the same date, payee and amount appearing in
			// two accounts is what a transfer between them looks like.
			seen := make(map[string]int)
			var seenOrder []string

			for _, fields := range transactions {
				// DATE FORMAT: YYYY-MM-DD. The separator before the year carries
				// the century, so the field is handed to the parser rather than
				// being pasted onto a fixed 20xx prefix.
				transDate, dateErr := utils.ParseQIFDateField(fields.Date)
				if dateErr != nil {
					fmt.Printf("Warning: Skipping transaction with unusable date: %v\n", dateErr)
					continue
				}
				fullDate := transDate.Format("2006-01-02")

				// Check if the transaction date is within the specified range
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

				payee := fields.Payee
				// Keep the payee exactly as it appeared in the QIF file so the
				// Original Statement column survives payee mapping.
				originalPayee := strings.ReplaceAll(payee, "\"", "")
				// Apply the payee mapping
				payee = payees.apply(payee)
				if payees.loaded() && payee == originalPayee && payee != "" {
					validator.AddUnmatchedData("payee", payee)
				}
				// Remove double quotes
				payee = strings.ReplaceAll(payee, "\"", "")

				// Tallied once per transaction, before a split fans it out into
				// several rows of differing amounts.
				key := fullDate + "|" + payee + "|" + strings.ReplaceAll(fields.Amount, ",", "")
				if seen[key] == 0 {
					seenOrder = append(seenOrder, key)
				}
				seen[key]++

				// A split transaction contributes one row per line item, so the
				// categories and amounts it records survive the export. Without
				// --expandSplits it stays a single row, as before.
				for _, row := range rowsForTransaction(fields, fullDate) {
					amount1 := strings.ReplaceAll(row.amount, ",", "")
					// Format the amount with exactly 2 decimal places
					if amountFloat, amountErr := parseAmount(row.amount); amountErr != nil {
						fmt.Printf("Warning: Could not parse amount '%s', using as-is\n", amount1)
					} else {
						amount1 = fmt.Sprintf("%.2f", amountFloat)
					}

					transactionMemo := row.memo

					// Split the category and tag
					category, tag := utils.SplitCategoryAndTag(row.category)

					// Trim whitespace
					category = strings.TrimSpace(category)
					// Keep the pre-mapping category so it can be referenced later
					originalCategory := category
					// Apply the category mapping
					category = categories.apply(category)
					if categories.loaded() && category == originalCategory && category != "" {
						validator.AddUnmatchedData("category", category)
					}

					// Record the pre-mapping category in the memo so a remapped
					// category can be traced back to what Quicken had.
					if preserveOriginalCategory && originalCategory != "" && category != originalCategory {
						transactionMemo = appendOriginalCategoryNote(transactionMemo, originalCategory)
					}

					// Trim whitespace
					tag = strings.TrimSpace(tag)
					// Apply the tag mapping
					originalTag := tag
					tag = tags.apply(tag)
					if tags.loaded() && tag == originalTag && tag != "" {
						validator.AddUnmatchedData("tag", tag)
					}

					// Prepend a custom Tag to the Category
					if addTagForImport {
						if tag != "" {
							tag = "QIFIMPORT," + tag
						} else {
							tag = "QIFIMPORT"
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
						// Skip this transaction if the skipZeroAmounts flag is set
						if skipZeroAmounts {
							validator.AddSkippedZeroAmount()
							continue
						}
					}

					record := TransactionRecord{
						Date:              fullDate,
						Merchant:          payee,
						Category:          category,
						Account:           outputAccountName,
						OriginalStatement: originalPayee,
						Notes:             transactionMemo,
						Amount:            amount1,
						Tags:              tag,
					}

					// Roll to the next file before writing, not after. Opening it at
					// the boundary created a file for records that never arrived,
					// leaving one holding nothing but a header.
					if maxRecordsPerFile != 0 && count > 0 && count%maxRecordsPerFile == 0 {
						// Close current file
						if strings.ToUpper(outputFormat) == "JSON" {
							jsonData, err := json.MarshalIndent(records, "", "  ")
							if err != nil {
								outputFile.Close()
								fmt.Printf("Error: failed to marshal JSON data: %v\n", err)
								skippedAccounts++
								continue accountLoop
							}
							_, err = outputFile.Write(jsonData)
							if err != nil {
								outputFile.Close()
								fmt.Printf("Error: failed to write JSON data to file: %v\n", err)
								skippedAccounts++
								continue accountLoop
							}
							records = nil
						} else if strings.ToUpper(outputFormat) == "XML" {
							xmlData, err := xml.MarshalIndent(transactionList{Transactions: records}, "", "  ")
							if err != nil {
								outputFile.Close()
								fmt.Printf("Error: failed to marshal XML data: %v\n", err)
								skippedAccounts++
								continue accountLoop
							}
							_, err = outputFile.Write(xmlData)
							if err != nil {
								outputFile.Close()
								fmt.Printf("Error: failed to write XML data to file: %v\n", err)
								skippedAccounts++
								continue accountLoop
							}
							records = nil
						}
						outputFile.Close()

						// Start new file
						fileIndex++
						outputFileName = fmt.Sprintf("%s_%d%s", fileBase, fileIndex, ext)
						fullPath := filepath.Join(outputPath, outputFileName)
						fmt.Printf("\nCreating split file for %s (File %d) - Records %d to %d\n",
							accountName,
							fileIndex,
							(fileIndex-1)*maxRecordsPerFile+1,
							fileIndex*maxRecordsPerFile)

						outputFile, err = os.Create(fullPath)
						if err != nil {
							fmt.Printf("Error creating split file %s: %v\n", outputFileName, err)
							skippedAccounts++
							continue accountLoop
						}

						// Write appropriate headers for the new file
						if strings.ToUpper(outputFormat) == "XML" {
							_, err := outputFile.WriteString(xml.Header)
							if err != nil {
								outputFile.Close()
								fmt.Printf("Error: failed to write XML header to split file %s: %v\n", outputFileName, err)
								skippedAccounts++
								continue accountLoop
							}
						}
						if strings.ToUpper(outputFormat) == "CSV" {
							if err := writeHeader(outputFile, outputCSVHeader); err != nil {
								outputFile.Close()
								fmt.Printf("Error: failed to write header to %s: %v\n", outputFileName, err)
								skippedAccounts++
								continue accountLoop
							}
						}
					}

					// JSON and XML are marshalled as a whole document, so their
					// records are collected and written when the file is closed.
					format := strings.ToUpper(outputFormat)
					if format == "JSON" || format == "XML" {
						records = append(records, record)
					} else {
						line := buildCSVRow(record, columnsToUse)
						if err := writeTransaction(outputFile, line); err != nil {
							outputFile.Close()
							fmt.Printf("failed to write transaction: %v\n", err)
							skippedAccounts++
							continue accountLoop
						}
					}
					count++

				}
			}
			if strings.ToUpper(outputFormat) == "JSON" && len(records) > 0 {
				jsonData, err := json.MarshalIndent(records, "", "  ")
				if err != nil {
					outputFile.Close()
					fmt.Printf("Error: failed to marshal final JSON data: %v\n", err)
					skippedAccounts++
					continue accountLoop
				}
				_, err = outputFile.Write(jsonData)
				if err != nil {
					outputFile.Close()
					fmt.Printf("Error: failed to write final JSON data to file: %v\n", err)
					skippedAccounts++
					continue accountLoop
				}
				records = nil
			} else if strings.ToUpper(outputFormat) == "XML" && len(records) > 0 {
				xmlData, err := xml.MarshalIndent(transactionList{Transactions: records}, "", "  ")
				if err != nil {
					outputFile.Close()
					fmt.Printf("Error: failed to marshal final XML data: %v\n", err)
					skippedAccounts++
					continue accountLoop
				}
				_, err = outputFile.Write(xmlData)
				if err != nil {
					outputFile.Close()
					fmt.Printf("Error: failed to write final XML data to file: %v\n", err)
					skippedAccounts++
					continue accountLoop
				}
				records = nil
			}
			outputFile.Close()

			// Report anything that appeared more than once in this account.
			for _, key := range seenOrder {
				if seen[key] < 2 {
					continue
				}
				parts := strings.SplitN(key, "|", 3)
				validator.AddDuplicate(parts[0], parts[1], parts[2], seen[key])
			}
		}

		// Print summary
		fmt.Println("\nExport Summary:")
		if skippedAccounts > 0 {
			fmt.Printf("Accounts skipped after a write error: %d\n", skippedAccounts)
		}
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
		// One line per mapping, in place of the line that used to be printed for
		// every value changed.
		for _, set := range []*mappingSet{categories, payees, accounts, tags} {
			if set.loaded() {
				plural := "values"
				if set.hits == 1 {
					plural = "value"
				}
				fmt.Printf("%s mapping: applied to %d %s\n", set.name, set.hits, plural)
			}
		}
		// This is on by default, so the export says so rather than leaving a tag
		// on every transaction to be discovered after the import.
		if addTagForImport {
			fmt.Println("Import tag: QIFIMPORT added to every transaction (turn this off with --addTagForImport=false)")
		}
		if maxRecordsPerFile > 0 {
			fmt.Printf("Split files: %d records per file (for Monarch compatibility)\n", maxRecordsPerFile)
		}
		fmt.Println("\nExport completed successfully!")

		// Rules that never matched anything are usually typos in the mapping
		// file, so they are worth naming.
		for _, set := range []*mappingSet{categories, payees, accounts, tags} {
			validator.RecordUnusedMapping(set.name, set.unusedRules())
		}

		// Print validation summary
		validator.PrintSummary()

		// Write detailed validation log (to transactions-specific log file)
		if err := validator.WriteValidationLogWithName(outputPath, "transactions_validation.log"); err != nil {
			fmt.Printf("Warning: Could not write validation log: %v\n", err)
		}

		return nil
	},
}

func init() {
	exportCmd.AddCommand(transactionsCmd)

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
	transactionsCmd.Flags().BoolVarP(&skipZeroAmounts, "skipZeroAmounts", "", false, "Skip transactions with zero amount (0.00 or 0)")
	transactionsCmd.Flags().BoolVarP(&preserveOriginalCategory, "preserveOriginalCategory", "", false, "Append the original (pre-mapping) category to the Notes field whenever a category mapping changes it")
	transactionsCmd.Flags().BoolVarP(&expandSplits, "expandSplits", "", false, "Export each line item of a split transaction as its own row, keeping its own category and amount")

	// Shorthands for the flags shared with the root command. Declaring them
	// locally shadows the persistent versions, which pflag then skips when
	// merging, so -i and -o work here without colliding with the -o that names
	// an output file on the list-style export commands.
	transactionsCmd.Flags().StringVarP(&inputFile, "inputFile", "i", "", "Path to input QIF file")
	transactionsCmd.Flags().StringVarP(&outputPath, "outputPath", "o", "", "Path to output directory")

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

// mappingSet holds the rules loaded from one mapping file and remembers
// which of them were used, so the rules that never matched can be reported
// afterwards. Those are usually typos in the mapping file.
type mappingSet struct {
	name  string
	rules map[string]string
	used  map[string]bool
	hits  int
}

func newMappingSet(name string, rules map[string]string) *mappingSet {
	return &mappingSet{name: name, rules: rules, used: make(map[string]bool)}
}

// loaded reports whether a mapping file was supplied for this kind of value.
func (m *mappingSet) loaded() bool {
	return m != nil && len(m.rules) > 0
}

// apply returns the mapped value, or the input unchanged when no rule
// matches.
func (m *mappingSet) apply(input string) string {
	if !m.loaded() {
		return input
	}
	target, ok := m.rules[input]
	if !ok {
		return input
	}
	m.used[input] = true
	m.hits++
	return target
}

// unusedRules lists the rules that never matched anything in the file.
func (m *mappingSet) unusedRules() []string {
	if !m.loaded() {
		return nil
	}
	var unused []string
	for source := range m.rules {
		if !m.used[source] {
			unused = append(unused, source)
		}
	}
	sort.Strings(unused)
	return unused
}

// uniqueFileBase turns an account name into a file name stem that is safe to
// write and distinct from the stems already used. Windows compares file names
// without regard to case, so what is taken is recorded in lower case.
func uniqueFileBase(accountName string, used map[string]bool) string {
	base := utils.SanitizeFileName(accountName)

	candidate := base
	for suffix := 2; used[strings.ToLower(candidate)]; suffix++ {
		candidate = fmt.Sprintf("%s~%d", base, suffix)
	}

	used[strings.ToLower(candidate)] = true
	return candidate
}

// transactionRow is the part of an output row that differs between a plain
// transaction and the individual line items of a split.
type transactionRow struct {
	amount   string
	memo     string
	category string
}

// rowsForTransaction returns the rows a QIF record contributes to the export.
// A split transaction becomes one row per line item when expandSplits is set,
// so the categories and amounts it records are not collapsed into the parent
// total; otherwise the record stays a single row.
func rowsForTransaction(fields utils.TransactionFields, date string) []transactionRow {
	if !expandSplits || len(fields.Splits) == 0 {
		return []transactionRow{{
			amount:   fields.Amount,
			memo:     fields.Memo,
			category: fields.Category,
		}}
	}

	rows := make([]transactionRow, 0, len(fields.Splits))
	for _, split := range fields.Splits {
		// A line item without its own memo inherits the transaction memo, so
		// the row is not left blank.
		memo := split.Memo
		if memo == "" {
			memo = fields.Memo
		}
		rows = append(rows, transactionRow{
			amount:   split.Amount,
			memo:     memo,
			category: split.Category,
		})
	}

	warnIfSplitsDoNotReconcile(fields, date)
	return rows
}

// warnIfSplitsDoNotReconcile reports a split whose line items do not add up to
// the transaction total, rather than quietly exporting an account that no
// longer balances.
func warnIfSplitsDoNotReconcile(fields utils.TransactionFields, date string) {
	total, err := parseAmount(fields.Amount)
	if err != nil {
		return
	}

	var sum float64
	for _, split := range fields.Splits {
		amount, splitErr := parseAmount(split.Amount)
		if splitErr != nil {
			return
		}
		sum += amount
	}

	// Half a cent is below anything QIF can express, so it is a safe tolerance.
	if difference := total - sum; math.Abs(difference) >= 0.005 {
		fmt.Printf("Warning: splits for %s \"%s\" do not sum to the transaction total; difference %.2f\n",
			date, fields.Payee, difference)
	}
}

// parseAmount reads a QIF amount, which may carry thousands separators.
func parseAmount(value string) (float64, error) {
	return strconv.ParseFloat(strings.ReplaceAll(value, ",", ""), 64)
}

// appendOriginalCategoryNote appends a back-reference to the pre-mapping
// category onto a transaction memo.
func appendOriginalCategoryNote(memo, originalCategory string) string {
	note := "[Original Category: " + originalCategory + "]"
	if memo == "" {
		return note
	}
	return memo + " " + note
}

func writeHeader(f *os.File, h string) error {
	_, err := f.WriteString(h)
	return err
}

// csvColumnValue maps a column name to the value it takes from a record. It
// is the single source both for the names --csvColumns accepts and for what
// each one writes, so the two cannot drift apart.
var csvColumnValue = map[string]func(TransactionRecord) string{
	"Date":               func(r TransactionRecord) string { return r.Date },
	"Merchant":           func(r TransactionRecord) string { return r.Merchant },
	"Category":           func(r TransactionRecord) string { return r.Category },
	"Account":            func(r TransactionRecord) string { return r.Account },
	"Original Statement": func(r TransactionRecord) string { return r.OriginalStatement },
	"Notes":              func(r TransactionRecord) string { return r.Notes },
	"Amount":             func(r TransactionRecord) string { return r.Amount },
	"Tags":               func(r TransactionRecord) string { return r.Tags },
}

// validateCSVColumns reports the first name --csvColumns does not recognise.
func validateCSVColumns(columns string) error {
	for _, name := range strings.Split(columns, ",") {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if _, ok := csvColumnValue[name]; !ok {
			return fmt.Errorf("unknown column %q in --csvColumns; valid names are %s",
				name, DefaultMonarchColumns)
		}
	}
	return nil
}

// buildCSVRow builds a CSV row from a TransactionRecord based on specified columns
func buildCSVRow(record TransactionRecord, columns string) string {
	columnList := strings.Split(columns, ",")
	values := make([]string, len(columnList))

	for i, col := range columnList {
		if value, ok := csvColumnValue[strings.TrimSpace(col)]; ok {
			values[i] = value(record)
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
