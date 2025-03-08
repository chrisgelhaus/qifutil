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

// tagsCmd represents the tags command
var tagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "Extract tags from a QIF file",
	Long:  `Extract tags from a QIF file.`,
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
		var tags []string
		var transactionRegexString string = `D(?<month>\d{1,2})\/(\s?(?<day>\d{1,2}))'(?<year>\d{2})[\r\n]+(U(?<amount1>.*?)[\r\n]+)(T(?<amount2>.*?)[\r\n]+)(C(?<cleared>.*?)[\r\n]+)((N(?<number>.*?)[\r\n]+)?)(P(?<payee>.*?)[\r\n]+)((M(?<memo>.*?)[\r\n]+)?)(L(?<category>.*?)[\r\n]+)`
		var tagRecordRegex string = `(?m)(^N(.*)\n^(D(.*)\n^)?\^\n)`
		var tagBlockHeaderRegex string = `(?m)^!Type:Tag\n`
		var accountBlockHeaderRegex string = `(?m)^!Account[^\n]*\n^N(.*?)\n^T(.*?)\n^\^\n^!Type:(Bank|CCard)\s*\n`

		// Create the tag output file
		tagFile, err := os.Create(outputFile)
		if err != nil {
			fmt.Println("Error creating tag file:", err)
		} else {
			fmt.Println("Created tag output file.")
		}
		defer tagFile.Close()

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

		// Find the position of the Tag Block
		tagTypeRe, err := regexp.Compile(tagBlockHeaderRegex)
		if err != nil {
			fmt.Println("Error compiling regular expression: ", err)
		}
		loc := tagTypeRe.FindStringIndex(inputContent)
		if loc == nil {
			fmt.Printf("No Tag block found.\n")
		} else {
			// Debugging output
			fmt.Printf("Tag block found at position: %d\n", loc[1])
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
		regex, _ := regexp.Compile(tagRecordRegex)

		// Find all matches in the content.
		matches := regex.FindAllStringSubmatch(textBetweenTypes, -1)
		fmt.Printf("%d entries extracted from the tag block.\n", len(matches))

		// Extract patterns to array from the Tag block.
		for _, t := range matches {
			// Check if there is a captured group and extract the content.
			if len(t) > 1 {
				// Ensure there is a captured group.
				tag := strings.TrimSpace(t[2])
				if tag != "" {
					tags = append(tags, fmt.Sprintf("\"%s\"", tag))
				}
			}
		}

		// Gather categories from the Accounts
		// Compile the regex
		regex, _ = regexp.Compile(accountBlockHeaderRegex)

		accountBlocks := regex.FindAllStringSubmatchIndex(inputContent, -1)
		if len(accountBlocks) == 0 {
			fmt.Println("No matches found.")
		}

		// loop over each account block and pull out tags
		for _, accountBlock := range accountBlocks {
			// Find the next Type Block
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
			fmt.Printf("%d tags extracted from account: %s\n", len(transactions), accountName)

			// Loop through matches and add categories to the array
			for _, t := range transactions {
				if len(t) > 1 {
					_, tag := splitCategoryAndTag(t[20])
					tag = strings.TrimSpace(tag)
					if tag != "" {
						// Remove double quotes
						tag = strings.ReplaceAll(tag, "\"", "")
						// Add tag to the list
						tags = append(tags, fmt.Sprintf("\"%s\"", tag))
					}
				}
			}
		}

		// Sort and dedupe tag list
		outputTagList := sortAndDedupStrings(tags)
		// Write tags to the file
		for _, item := range outputTagList {
			_, err := tagFile.WriteString(item + "\n")
			if err != nil {
				fmt.Printf("Error Writing to tag file:\n")
			}
		}

		fmt.Println("Extracted Tags: ", len(outputTagList))
	},
}

func init() {
	exportCmd.AddCommand(tagsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// tagsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// tagsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	tagsCmd.Flags().StringVarP(&inputFile, "inputFile", "i", "", "Input QIF file")
	tagsCmd.Flags().StringVarP(&outputFile, "outputFile", "o", "tags.csv", "Output file for tag names")
	tagsCmd.Flags().StringVarP(&outputFormat, "outputFormat", "f", "CSV", "Output format (CSV, JSON, etc.). Currently only CSV is supported.")
}
