/*
Copyright © 2025 Chris Gelhaus <chrisgelhaus@live.com>
*/
package cmd

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// payeesCmd represents the payees command
var payeeOutputFile string

type payeeList struct {
	XMLName xml.Name `xml:"payees"`
	Payees  []string `xml:"payee"`
}

var payeesCmd = &cobra.Command{
	Use:   "payees",
	Short: "Extract payees from a QIF file",
	Long:  `Extract payees from a QIF file.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		if inputFile == "" {
			fmt.Println("Error: Missing required flag --inputFile")
			os.Exit(1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {

		var payees []string
		var transactionRegexString string = `D(?<month>\d{1,2})\/(\s?(?<day>\d{1,2}))'(?<year>\d{2})[\r\n]+(U(?<amount1>.*?)[\r\n]+)(T(?<amount2>.*?)[\r\n]+)(C(?<cleared>.*?)[\r\n]+)((N(?<number>.*?)[\r\n]+)?)(P(?<payee>.*?)[\r\n]+)((M(?<memo>.*?)[\r\n]+)?)(L(?<category>.*?)[\r\n]+)`
		var accountBlockHeaderRegex string = `(?m)^!Account[^\n]*\n^N(.*?)\n^T(.*?)\n^\^\n^!Type:(Bank|CCard)\s*\n`

		// Build output file path using outputPath if provided
		outputFilePath := payeeOutputFile
		if outputPath != "" {
			outputFilePath = filepath.Join(outputPath, payeeOutputFile)
		}

		// Create the category output file
		payeeFile, err := os.Create(outputFilePath)
		if err != nil {
			fmt.Println("Error creating category file:", err)
		} else {
			fmt.Println("Created catergory output file.")
		}
		defer payeeFile.Close()

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

		// Gather payees from the Accounts
		// Compile the regex
		regex, _ := regexp.Compile(accountBlockHeaderRegex)

		accountBlocks := regex.FindAllStringSubmatchIndex(inputContent, -1)
		if len(accountBlocks) == 0 {
			fmt.Println("No matches found.")
		}

		// loop over each account block and pull out payees
		for _, accountBlock := range accountBlocks {
			accountName := inputContent[accountBlock[2]:accountBlock[3]]

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

			// Extract the text between the type lines
			textBetweenTypes := inputContent[accountBlock[1]:endPos]

			// Use the existing pattern to match entries
			regex, _ := regexp.Compile(transactionRegexString)

			// Find all matches in the content.
			transactions := regex.FindAllStringSubmatch(textBetweenTypes, -1)
			fmt.Printf("%d payees extracted from account: %s\n", len(transactions), accountName)

			// Loop through matches and add payees to the array
			for _, t := range transactions {
				if len(t) > 1 {
					payee := strings.TrimSpace(t[15])
					// Remove double quotes
					payee = strings.ReplaceAll(payee, "\"", "")
					// Add payee to the list
					payees = append(payees, payee)
				}
			}
		}

		// Sort and dedupe payee list
		outputPayeeList := sortAndDedupStrings(payees)
		// Write payees to the file
		switch strings.ToUpper(outputFormat) {
		case "JSON":
			jsonData, err := json.MarshalIndent(outputPayeeList, "", "  ")
			if err != nil {
				fmt.Printf("Error marshaling JSON: %v\n", err)
				return
			}
			payeeFile.Write(jsonData)
		case "XML":
			xmlData, err := xml.MarshalIndent(payeeList{Payees: outputPayeeList}, "", "  ")
			if err != nil {
				fmt.Printf("Error marshaling XML: %v\n", err)
				return
			}
			payeeFile.Write([]byte(xml.Header))
			payeeFile.Write(xmlData)
		default:
			for _, item := range outputPayeeList {
				_, err := payeeFile.WriteString(fmt.Sprintf("\"%s\"\n", item))
				if err != nil {
					fmt.Printf("Error Writing to category file:\n")
				}
			}
		}

		fmt.Println("Unique Extracted Payees: ", len(outputPayeeList))

	},
}

func init() {
	exportCmd.AddCommand(payeesCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// payeesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// payeesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	payeesCmd.Flags().StringVarP(&inputFile, "inputFile", "i", "", "Input QIF file")
	payeesCmd.Flags().StringVarP(&payeeOutputFile, "outputFile", "o", "payees.csv", "Output file for payee names")
	payeesCmd.Flags().StringVarP(&outputFormat, "outputFormat", "f", "CSV", "Output format (CSV, JSON, XML).")
}
