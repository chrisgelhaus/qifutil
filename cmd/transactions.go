/*
Copyright © 2025 Chris Gelhaus <chrisgelhaus@live.com>
*/
package cmd

import (
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var categoryMappingFile string
var accountMappingFile string
var payeeMappingFile string
var tagMappingFile string
var outputPath string
var addTagForImport bool = false

// transactionsCmd represents the transactions command
var transactionsCmd = &cobra.Command{
	Use:   "transactions",
	Short: "Export transactions from a QIF file",
	Long: `Export transactions from a QIF file.
	Command takes a QIF file as input and exports transactions to CSV files.
	Optional mapping files can be used to map categories, payees, accounts and tags.
	Each account in the QIF file will have its own CSV file.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		if inputFile == "" {
			fmt.Println("Error: Missing required flag --inputFile")
			os.Exit(1)
		}
		if outputPath == "" {
			fmt.Println("Error: Missing required flag --outputPath")
			os.Exit(1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Begin Export Transactions")

		// Output CSV Header
		var transactionRegexString string = `D(?<month>\d{1,2})\/(\s?(?<day>\d{1,2}))'(?<year>\d{2})[\r\n]+(U(?<amount1>.*?)[\r\n]+)(T(?<amount2>.*?)[\r\n]+)(C(?<cleared>.*?)[\r\n]+)((N(?<number>.*?)[\r\n]+)?)(P(?<payee>.*?)[\r\n]+)((M(?<memo>.*?)[\r\n]+)?)(L(?<category>.*?)[\r\n]+)`
		var accountBlockHeaderRegex string = `(?m)^!Account[^\n]*\n^N(.*?)\n^T(.*?)\n^\^\n^!Type:(Bank|CCard)\s*\n`
		outputCSVHeader := "Date,Merchant,Category,Account,Original Statement,Notes,Amount,Tags\n"
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

		// loop over each account block
		// Find all matches for transactions
		for _, accountBlock := range accountBlocks {
			var outputAccountName string
			accountName := inputContent[accountBlock[2]:accountBlock[3]]
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

			// Create unique output file per Account
			outputFile, err := os.Create(outputPath + accountName + ".csv")
			if err != nil {
				fmt.Println("Error creating file:", err)
				return
			}

			// Write header to the output file.
			_, err = outputFile.WriteString(outputCSVHeader)
			if err != nil {
				fmt.Println("Error writing to file:", err)
				return
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

			for _, t := range transactions {
				// Check if there is a captured group and extract the content.
				if len(t) > 1 { // Ensure there is a captured group.
					month := strings.TrimSpace(t[1])
					day := strings.TrimSpace(t[2])
					year := strings.TrimSpace(t[4])
					amount1 := strings.TrimSpace(t[6])
					//amount2 := strings.TrimSpace(t[8])
					//cleared := strings.TrimSpace(t[10])
					//number := strings.TrimSpace(t[13])

					payee := strings.TrimSpace(t[15])
					// Apply the payee mapping
					payee = applyMapping(payee, payeeMapping)
					// Remove double quotes
					payee = strings.ReplaceAll(payee, "\"", "")

					transactionMemo := strings.TrimSpace(t[18])

					// Split the category and tag
					category, tag := splitCategoryAndTag(t[20])

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

					// Surround output values with double quotes to ensure they are treated as strings
					// Write the transaction to the output file

					// This output is compatiple with the Monarch CSV Importer
					_, err := outputFile.WriteString("\"" + fullDate + "\",\"" + payee + "\",\"" + category + "\",\"" + outputAccountName + "\",\"" + payee + "\",\"" + transactionMemo + "\",\"" + amount1 + "\",\"" + tag + "\"\n")

					if err != nil {
						fmt.Println("Error writing to file:", err)
						return
					}
				}
			}
			outputFile.Close()
		}

	},
}

func init() {
	exportCmd.AddCommand(transactionsCmd)
	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// transactionsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	transactionsCmd.Flags().StringVarP(&inputFile, "inputFile", "i", "", "Input QIF file")
	transactionsCmd.Flags().StringVarP(&outputFormat, "outputFormat", "f", "CSV", "Output format (CSV, JSON, etc.). Currently only CSV is supported.")
	transactionsCmd.Flags().StringVarP(&accountMappingFile, "accountMapFile", "a", "", "Supplied mapping file for accounts. Optional.")
	transactionsCmd.Flags().StringVarP(&categoryMappingFile, "categoryMapFile", "c", "", "Supplied mapping file for categories. Optional.")
	transactionsCmd.Flags().StringVarP(&payeeMappingFile, "payeeMapFile", "p", "", "Supplied mapping file for payees. Optional.")
	transactionsCmd.Flags().StringVarP(&tagMappingFile, "tagMapFile", "t", "", "Supplied mapping file for tags. Optional.")
	transactionsCmd.Flags().StringVarP(&outputPath, "outputPath", "", "", "Output path for transaction file")
	transactionsCmd.Flags().BoolVarP(&addTagForImport, "addTagForImport", "", true, "Add a custom tag to the transaction for import purposes")
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
			// Two fields - set key-value in map
			key := record[0]
			value := record[1]
			mapping[key] = value
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
