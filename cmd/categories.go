/*
Copyright © 2025 Chris Gelhaus
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

	"qifutil/pkg/utils"

	"github.com/spf13/cobra"
)

// categoriesCmd represents the categories command
var categoryOutputFile string

type categoryList struct {
	XMLName    xml.Name `xml:"categories"`
	Categories []string `xml:"category"`
}

var categoriesCmd = &cobra.Command{
	Use:   "categories",
	Short: "Extract categories from a QIF file",
	Long:  `Extract categories from a QIF file.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		if inputFile == "" {
			fmt.Println("Error: Missing required flag --inputFile")
			os.Exit(1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		var categories []string
		var catRecordRegex string = `(?m)(^N(.*)\n(^D(.*)\n)?(^T(.*)\n)?(^R(.*)\n)?(^E(.*)\n)?(^I(.*)\n)?^\^\n)`
		var catBlockHeaderRegex string = `(?m)^!Type:Cat\n`
		var accountBlockHeaderRegex string = `(?m)^!Account[^\n]*\n^N(.*?)\n^T(.*?)\n^\^\n^!Type:(Bank|CCard)\s*\n`

		// Build output file path using outputPath if provided
		outputFilePath := categoryOutputFile
		if outputPath != "" {
			outputFilePath = filepath.Join(outputPath, categoryOutputFile)
		}

		// Create the category output file
		categoryFile, err := os.Create(outputFilePath)
		if err != nil {
			fmt.Println("Error creating category file:", err)
			//return err
		} else {
			fmt.Println("Created category output file.")
		}
		defer categoryFile.Close()

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

		// Find the position of the Category Block
		catTypeRe, err := regexp.Compile(catBlockHeaderRegex)
		if err != nil {
			fmt.Println("Error compiling regular expression: ", err)
		}
		loc := catTypeRe.FindStringIndex(inputContent)
		var catBlockEnd int
		if loc == nil {
			fmt.Printf("No Category block found.\n")
		} else {
			fmt.Printf("Category block found at position: %d\n", loc[1])
			catBlockEnd = loc[1]
		}

		// Find the position of the next Type block
		restOfText := inputContent[catBlockEnd:]
		nextTypePattern := `(?mi)^\s*!Type:.*$`
		nextTypeRe := regexp.MustCompile(nextTypePattern)
		nextLoc := nextTypeRe.FindStringIndex(restOfText)
		var endPos int
		if nextLoc != nil {
			fmt.Printf("Next type found at:%d\n", nextLoc[1])
			endPos = catBlockEnd + nextLoc[0]
		} else {
			fmt.Printf("No next type block found.\n")
			endPos = len(inputContent)
		}

		// Extract the text between the Type lines
		textBetweenTypes := inputContent[catBlockEnd:endPos]

		// Use the existing pattern to match entries
		regex, _ := regexp.Compile(catRecordRegex)

		// Find all matches in the content.
		matches := regex.FindAllStringSubmatch(textBetweenTypes, -1)
		fmt.Printf("%d entries extracted from the category block.\n", len(matches))

		// Extract patterns to array from the Category block.
		for _, t := range matches {
			// Check if there is a captured group and extract the content.
			if len(t) > 1 {
				// Ensure there is a captured group.
				category := strings.TrimSpace(t[2])
				categories = append(categories, category)
			}
		}

		// Gather categories from the Accounts
		// Compile the regex
		regex, _ = regexp.Compile(accountBlockHeaderRegex)

		accountBlocks := regex.FindAllStringSubmatchIndex(inputContent, -1)
		if len(accountBlocks) == 0 {
			fmt.Println("No accounts found in input file.")
		}

		// loop over each account block and pull out categories
		for _, accountBlock := range accountBlocks {
			// Find the next Type Block
			accountName := inputContent[accountBlock[2]:accountBlock[3]]
			// print to console for debugging
			fmt.Printf("Processing Account: %s\n", accountName)

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
			// QIF records are read field by field. A single pattern cannot express
			// fields that are optional and may appear in any order, and the one
			// used here silently skipped any record that did not fit.
			var records []utils.TransactionFields
			for _, record := range utils.SplitRecords(textBetweenTypes) {
				// The next account's header block falls inside this account's text.
				if strings.HasPrefix(strings.TrimSpace(record), "!") {
					continue
				}
				fields := utils.ParseTransactionRecord(record)
				if fields.Date == "" {
					continue
				}
				records = append(records, fields)
			}

			fmt.Printf("%d categories extracted from account: %s\n\n", len(records), accountName)

			// A split line carries a category of its own, which may appear nowhere
			// else in the file.
			for _, fields := range records {
				raw := []string{fields.Category}
				for _, split := range fields.Splits {
					raw = append(raw, split.Category)
				}

				for _, value := range raw {
					category, _ := utils.SplitCategoryAndTag(value)
					if category != "" {
						categories = append(categories, strings.ReplaceAll(category, "\"", ""))
					}
				}
			}
		}

		// Sort and dedupe category list
		outputCategoryList := sortAndDedupStrings(categories)
		switch strings.ToUpper(outputFormat) {
		case "JSON":
			jsonData, err := json.MarshalIndent(outputCategoryList, "", "  ")
			if err != nil {
				fmt.Printf("Error marshaling JSON: %v\n", err)
				return
			}
			categoryFile.Write(jsonData)
		case "XML":
			xmlData, err := xml.MarshalIndent(categoryList{Categories: outputCategoryList}, "", "  ")
			if err != nil {
				fmt.Printf("Error marshaling XML: %v\n", err)
				return
			}
			categoryFile.Write([]byte(xml.Header))
			categoryFile.Write(xmlData)
		default:
			for _, item := range outputCategoryList {
				_, err := categoryFile.WriteString(fmt.Sprintf("\"%s\"\n", item))
				if err != nil {
					fmt.Printf("Error Writing to category file:\n")
				}
			}
		}

		fmt.Println("Unique Extracted Categories: ", len(outputCategoryList))

	},
}

func init() {
	exportCmd.AddCommand(categoriesCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// categoriesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// categoriesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	categoriesCmd.Flags().StringVarP(&inputFile, "inputFile", "i", "", "Input QIF file")
	categoriesCmd.Flags().StringVarP(&categoryOutputFile, "outputFile", "o", "categories.csv", "Output file for category names")
	categoriesCmd.Flags().StringVarP(&outputFormat, "outputFormat", "f", "CSV", "Output format (CSV, JSON, XML).")
}
