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
	Short: "The \"show\" commands display the collected domains or urls",
	Long: `The "show" command display the list of domains or urls that have been collected during the scanning process.

Example usage:
  $ abashiri show domain -all
  $ abashiri show domain -d example.com
  $ abashiri show url -d example.com`,
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
