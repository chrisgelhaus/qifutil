/*
Copyright © 2025 Chris Gelhaus <chrisgelhaus@live.com>
*/
package cmd

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

// accountsCmd represents the accounts command
var accountsCmd = &cobra.Command{
	Use:   "accounts",
	Short: "Extract account names from a QIF file",
	Long:  `Extract account names from a QIF file.`,
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

		var accountNames []string
		var accountBlockHeaderRegex string = `(?m)^!Account[^\n]*\n^N(.*?)\n^T(.*?)\n^\^\n^!Type:(Bank|CCard)\s*\n`

		// Create the account output file
		accountFile, err := os.Create(outputFile)
		if err != nil {
			fmt.Println("Error creating account file:", err)
		} else {
			fmt.Println("Created account output file.")
		}
		defer accountFile.Close()

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

		// Gather the Accounts
		// Compile the regex
		regex, err := regexp.Compile(accountBlockHeaderRegex)
		if err != nil {
			fmt.Println("Error collecting accounts:", err)
		}
		accountBlocks := regex.FindAllStringSubmatchIndex(inputContent, -1)
		if len(accountBlocks) == 0 {
			fmt.Println("No matches found.")
		}

		// loop over each account block and pull out payees
		for _, accountBlock := range accountBlocks {
			accountName := inputContent[accountBlock[2]:accountBlock[3]]
			accountName = strings.TrimSpace(accountName)
			// Remove double quotes
			accountName = strings.ReplaceAll(accountName, "\"", "")
			accountNames = append(accountNames, fmt.Sprintf("\"%s\"", accountName))
		}

		// Sort and dedupe payee list
		outputAccountList := sortAndDedupStrings(accountNames)
		// Write payees to the file
		for _, item := range outputAccountList {
			_, err := accountFile.WriteString(item + "\n")
			if err != nil {
				fmt.Printf("Error Writing to account file:\n")
			}
		}

		fmt.Println("Extracted Account: ", len(outputAccountList))

	},
	PostRun: func(cmd *cobra.Command, args []string) {
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
	},
}

func init() {
	exportCmd.AddCommand(accountsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// accountsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// accountsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	accountsCmd.Flags().StringVarP(&inputFile, "inputFile", "i", "", "Input QIF file")
	accountsCmd.Flags().StringVarP(&outputFile, "outputFile", "o", "accounts.csv", "Output file for account names")
	accountsCmd.Flags().StringVarP(&outputFormat, "outputFormat", "f", "CSV", "Output format (CSV, JSON, etc.). Currently only CSV is supported.")

}

// sortAndDedupStrings sorts a slice of strings in ascending order,
// removes duplicate entries, and eliminates any blank strings.
//
// Parameters:
//
//	arr []string - The input slice of strings to be sorted and deduplicated.
//
// Returns:
//
//	[]string - A new slice of strings that is sorted, deduplicated, and free of blank strings.
func sortAndDedupStrings(arr []string) []string {
	sort.Strings(arr)

	n := len(arr)
	if n == 0 {
		return arr
	}

	// Deduplication
	deduped := []string{arr[0]}
	for i := 1; i < n; i++ {
		if arr[i] != arr[i-1] {
			deduped = append(deduped, arr[i])
		}
	}

	// Remove Blanks
	var result []string
	for _, str := range deduped {
		if strings.TrimSpace(str) != "" {
			result = append(result, str)
		}
	}
	return result
}
