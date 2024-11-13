/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"abashiri-cli/cmd/show"

	"github.com/spf13/cobra"
)

// showCmd represents the show command
var showCmd = &cobra.Command{
	Use:   "show",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
}

func init() {
	rootCmd.AddCommand(showCmd)
	showCmd.AddCommand(show.ShowDomainsCmd)
	showCmd.AddCommand(show.ShowURLsCmd)

	show.ShowDomainsCmd.PersistentFlags().StringVarP(&scanArgs.domain, "domain", "d", "", "root domain")
	show.ShowDomainsCmd.MarkPersistentFlagRequired("domain")

	show.ShowURLsCmd.PersistentFlags().StringVarP(&scanArgs.domain, "domain", "d", "", "root domain")
	show.ShowURLsCmd.MarkPersistentFlagRequired("domain")
}
