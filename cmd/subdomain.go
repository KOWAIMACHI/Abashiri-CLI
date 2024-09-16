/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"abashiri-cli/cmd/subdomain"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
)

var subdomainCmd = &cobra.Command{
	Use:   "subdomain",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
}

func init() {
	rootCmd.AddCommand(subdomainCmd)
	subdomainCmd.AddCommand(subdomain.GetCmd)
	subdomainCmd.AddCommand(subdomain.ScanCmd)
}
