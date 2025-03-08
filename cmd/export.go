/*
Copyright © 2025 Chris Gelhaus <chrisgelhaus@live.com>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:               "export",
	Short:             "Export accounts, categories, payees and transactions from a QIF file",
	Long:              `Export accounts, categories, payees and transactions from a QIF file`,
	Aliases:           []string{"ex"},
	PreRun:            func(cmd *cobra.Command, args []string) {},
	Run:               func(cmd *cobra.Command, args []string) {},
	PostRun:           func(cmd *cobra.Command, args []string) {},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {},
}

func init() {
	rootCmd.AddCommand(exportCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// exportCmd.PersistentFlags().StringVarP(&inputFile, "inputFile", "i", "", "")
	// exportCmd.PersistentFlags().StringVarP(&outputFile, "outputFile", "o", "", "")
	// exportCmd.PersistentFlags().StringVarP(&outputFormat, "outputFormat", "f", "CSV", "")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// exportCmd.Flags().BoolVarP(&toggle, "toggle", "t", false, "Help message for toggle")
}
