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

// tagsCmd represents the tags command
var tagsOutputFile string

type tagList struct {
	XMLName xml.Name `xml:"tags"`
	Tags    []string `xml:"tag"`
}

var tagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "Extract tags from a QIF file",
	Long:  `Extract tags from a QIF file.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// A failure past flag parsing is about the data, not how the command
		// was called, so cobra should print it without the usage text.
		cmd.SilenceUsage = true

		if inputFile == "" {
			return fmt.Errorf("missing required flag --inputFile")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// A failure past flag parsing is about the data, not how the command
		// was called, so cobra should print it without the usage text.
		cmd.SilenceUsage = true

		var tags []string

		// Build output file path using outputPath if provided
		outputFilePath := tagsOutputFile
		if outputPath != "" {
			outputFilePath = filepath.Join(outputPath, tagsOutputFile)
		}

		// Create the tag output file
		tagFile, err := os.Create(outputFilePath)
		if err != nil {
			fmt.Println("Error creating tag file:", err)
		} else {
			fmt.Println("Created tag output file.")
		}
		defer tagFile.Close()

		// Load input file
		inputBytes, err := os.ReadFile(inputFile)
		if err != nil {
			// Carrying on would scan empty content and write an empty list
			// while reporting success.
			return fmt.Errorf("reading %q: %w", inputFile, err)
		} else {
			fmt.Printf("Input file opened. Length: %d\n", len(inputBytes))
		}
		inputContent := string(inputBytes)

		// Standardize Line Endings to simplify Regex
		inputContent = strings.ReplaceAll(inputContent, "\r\n", "\n")

		// Find the position of the Tag Block
		tagTypeRe := tagBlockPattern
		if err != nil {
			fmt.Println("Error compiling regular expression: ", err)
		}
		loc := tagTypeRe.FindStringIndex(inputContent)
		if loc == nil {
			fmt.Printf("No Tag block found.\n")
		} else {
			fmt.Printf("Tag block found at position: %d\n", loc[1])
		}

		// Find the position of the next Type block
		var tagBlockEnd int
		if loc != nil {
			tagBlockEnd = loc[1]
		}
		restOfText := inputContent[tagBlockEnd:]
		nextTypePattern := `(?mi)^\s*!Type:.*$`
		nextTypeRe := regexp.MustCompile(nextTypePattern)
		nextLoc := nextTypeRe.FindStringIndex(restOfText)
		var endPos int
		if nextLoc != nil {
			fmt.Printf("Next type found at:%d\n", nextLoc[1])
			// Found another Type line.
			endPos = tagBlockEnd + nextLoc[0]
		} else {
			fmt.Printf("No next type block found.\n")
			// No other Type found
			endPos = len(inputContent)
		}

		// Extract the text between the Type lines
		textBetweenTypes := inputContent[tagBlockEnd:endPos]

		// Use the existing pattern to match entries
		regex := tagRecordPattern

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
					tags = append(tags, tag)
				}
			}
		}

		// Gather categories from the Accounts
		// Compile the regex
		regex = accountBlockPattern

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

			fmt.Printf("%d tags extracted from account: %s\n", len(records), accountName)

			// A split line carries a category, and so a tag, of its own.
			for _, fields := range records {
				raw := []string{fields.Category}
				for _, split := range fields.Splits {
					raw = append(raw, split.Category)
				}

				for _, value := range raw {
					if _, tag := utils.SplitCategoryAndTag(value); tag != "" {
						tags = append(tags, strings.ReplaceAll(tag, "\"", ""))
					}
				}
			}
		}

		// Sort and dedupe tag list
		outputTagList := sortAndDedupStrings(tags)
		// Write tags to the file
		switch strings.ToUpper(outputFormat) {
		case "JSON":
			jsonData, err := json.MarshalIndent(outputTagList, "", "  ")
			if err != nil {
				return fmt.Errorf("encoding JSON: %w", err)
			}
			tagFile.Write(jsonData)
		case "XML":
			xmlData, err := xml.MarshalIndent(tagList{Tags: outputTagList}, "", "  ")
			if err != nil {
				return fmt.Errorf("encoding XML: %w", err)
			}
			tagFile.Write([]byte(xml.Header))
			tagFile.Write(xmlData)
		default:
			for _, item := range outputTagList {
				_, err := tagFile.WriteString(fmt.Sprintf("\"%s\"\n", item))
				if err != nil {
					fmt.Printf("Error Writing to tag file:\n")
				}
			}
		}

		fmt.Println("Extracted Tags: ", len(outputTagList))

		return nil
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
	tagsCmd.Flags().StringVarP(&tagsOutputFile, "outputFile", "o", "tags.csv", "Output file for tag names")
	tagsCmd.Flags().StringVarP(&outputFormat, "outputFormat", "f", "CSV", "Output format (CSV, JSON, XML).")
}
