/*
Copyright © 2025 Chris Gelhaus <chrisgelhaus@live.com>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var version = "1.8.1"

var inputFile string

// var outputFile string
var outputFormat string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "qifutil",
	Version: version,
	Short:   "Quicken Interchange Format (QIF) utility",
	Long: `qifutil is a utility for working with Quicken Interchange Format (QIF) files.
It can extract transactions from a QIF file into CSV format for importing into other apps and more.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) {
	// },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Define persistent flags that are shared between commands
	rootCmd.PersistentFlags().StringVar(&inputFile, "inputFile", "", "Path to input QIF file")
	rootCmd.PersistentFlags().StringVar(&outputPath, "outputPath", "", "Path to output directory")
	rootCmd.PersistentFlags().StringVar(&selectedAccounts, "accounts", "", "Comma-separated list of accounts to process")
	rootCmd.PersistentFlags().StringVar(&startDate, "startDate", "", "Start date filter (YYYY-MM-DD)")
	rootCmd.PersistentFlags().StringVar(&endDate, "endDate", "", "End date filter (YYYY-MM-DD)")
}
