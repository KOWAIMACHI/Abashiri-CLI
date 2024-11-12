/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"abashiri-cli/cmd/subdomain"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
)

type SubDomainArgs struct {
	domain string
	mode   string
}

var subDomainArgs = &SubDomainArgs{}

var subdomainCmd = &cobra.Command{
	Use:   "subdomain",
	Short: "Collect subdomains for a specified domain",
	Long: `This command is used to collect subdomains for a given domain by performing passive or active scanning.
It leverages various discovery techniques to find subdomains and stores the results in an SQLite database for later retrieval.
This tool is useful for gathering a comprehensive list of subdomains associated with a domain for security testing and analysis.`,
}

func init() {
	rootCmd.AddCommand(subdomainCmd)
	subdomainCmd.AddCommand(subdomain.GetCmd)
	subdomainCmd.AddCommand(subdomain.ScanCmd)

	subdomainCmd.PersistentFlags().StringVar(&subDomainArgs.domain, "domain", "", "root domain")
	subdomainCmd.MarkPersistentFlagRequired("domain")
	subdomainCmd.PersistentFlags().StringVar(&subDomainArgs.mode, "mode", "passive", "passive/active")
}
