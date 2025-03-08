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

// categoriesCmd represents the categories command
var categoriesCmd = &cobra.Command{
	Use:   "categories",
	Short: "Extract categories from a QIF file",
	Long:  `Extract categories from a QIF file.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		if inputFile == "" {
			fmt.Println("Error: Missing required flag --inputFile")
			os.Exit(1)
		}
		if outputFile == "" {
			fmt.Println("Error: Missing required flag --outputFile")
			os.Exit(1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		var categories []string
		var transactionRegexString string = `D(?<month>\d{1,2})\/(\s?(?<day>\d{1,2}))'(?<year>\d{2})[\r\n]+(U(?<amount1>.*?)[\r\n]+)(T(?<amount2>.*?)[\r\n]+)(C(?<cleared>.*?)[\r\n]+)((N(?<number>.*?)[\r\n]+)?)(P(?<payee>.*?)[\r\n]+)((M(?<memo>.*?)[\r\n]+)?)(L(?<category>.*?)[\r\n]+)`
		var catRecordRegex string = `(?m)(^N(.*)\n(^D(.*)\n)?(^T(.*)\n)?(^R(.*)\n)?(^E(.*)\n)?(^I(.*)\n)?^\^\n)`
		var catBlockHeaderRegex string = `(?m)^!Type:Cat\n`
		var accountBlockHeaderRegex string = `(?m)^!Account[^\n]*\n^N(.*?)\n^T(.*?)\n^\^\n^!Type:(Bank|CCard)\s*\n`

		// Create the category output file
		categoryFile, err := os.Create(outputFile)
		if err != nil {
			fmt.Println("Error creating category file:", err)
			//return err
		} else {
			fmt.Println("Created catergory output file.")
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
		if loc == nil {
			fmt.Printf("No Category block found.\n")
			//return nil
		} else {
			// Debugging output
			fmt.Printf("Category block found at position: %d\n", loc[1])
		}

		// Find the position of the next Type block
		restOfText := inputContent[loc[1]:]
		nextTypePattern := `(?mi)^\s*!Type:.*$`
		nextTypeRe := regexp.MustCompile(nextTypePattern)
		nextLoc := nextTypeRe.FindStringIndex(restOfText)
		fmt.Printf("Next type found at:%d\n", nextLoc[1])
		var endPos int
		if nextLoc != nil {
			// Found another Type line.
			endPos = loc[1] + nextLoc[0]
		} else {
			// No other Type found
			endPos = len(inputContent)
		}

		// Extract the text between the Type lines
		textBetweenTypes := inputContent[loc[1]:endPos]

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
				categories = append(categories, fmt.Sprintf("\"%s\"", category))
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
			regex, _ := regexp.Compile(transactionRegexString)

			// Find all matches in the content.
			transactions := regex.FindAllStringSubmatch(textBetweenTypes, -1)
			fmt.Printf("%d categories extracted from account: %s\n\n", len(transactions), accountName)

			// Loop through matches and add categories to the array
			for _, t := range transactions {
				// Check if there is a captured group and extract the content.
				if len(t) > 1 {
					var category string = ""
					rawcategory := strings.TrimSpace(t[20])
					category, _ = splitCategoryAndTag(rawcategory)

					// If the category is not empty, add it to the list
					if category != "" {
						// Remove double quotes
						category = strings.ReplaceAll(category, "\"", "")
						// Add category to the list
						categories = append(categories, fmt.Sprintf("\"%s\"", category))
					}
				}
			}
		}

		// Sort and dedupe category list
		outputCategoryList := sortAndDedupStrings(categories)
		// Write categories to the file
		for _, item := range outputCategoryList {
			_, err := categoryFile.WriteString(item + "\n")
			if err != nil {
				fmt.Printf("Error Writing to category file:\n")
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
	categoriesCmd.Flags().StringVarP(&outputFile, "outputFile", "o", "categories.csv", "Output file for category names")
	categoriesCmd.Flags().StringVarP(&outputFormat, "outputFormat", "f", "CSV", "Output format (CSV, JSON, etc.). Currently only CSV is supported.")
}

func splitCategoryAndTag(originalCategoryValue string) (category string, tag string) {
	// If the category has a tag, split it out
	if strings.Contains(originalCategoryValue, "/") {
		// split the category and tag into separate strings and return the category
		category = strings.Split(originalCategoryValue, "/")[0]
		tag = strings.Split(originalCategoryValue, "/")[1]
	} else {
		// catgeory is the raw category
		category = originalCategoryValue
		tag = ""
	}

	return category, tag
}
