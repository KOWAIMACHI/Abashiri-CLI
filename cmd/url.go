/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"abashiri-cli/core/discovery"
	"abashiri-cli/storage"
	"database/sql"
	"log"
	"os"

	"github.com/spf13/cobra"
)

type URLScanArgs struct {
	includeSubDomains bool
}

var urlScanArgs = &URLScanArgs{}

// urlCmd represents the url command
var urlCmd = &cobra.Command{
	Use:   "url",
	Short: "The \"url\" command performs comprehensive url discovery for a given domain",
	Long: `The "url" command performs comprehensive URL discovery for a given domain

[!] URL enumeration
  Passive Scanning: Collects urls from publicly available sources without direct interaction.
  Active Scanning: Actively probes for live subdomains through direct requests.

[!] Example usage:
  $ abashiri url -d example.com
  $ abashiri url -d example.com -s
`,

	Run: func(cmd *cobra.Command, args []string) {
		domain, err := cmd.Flags().GetString("domain")
		if err != nil {
			log.Fatalf("Failed to parse --domain/-d: %v", err)
		}

		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			log.Fatalf("Failed to parse --verbose/-v: %v", err)
		}

		is, err := cmd.Flags().GetBool("include-subdomains")
		if err != nil {
			log.Fatalf("Failed to parse --include-subdomains: %v", err)
		}

		if domain == "" {
			cmd.Usage()
			os.Exit(1)
		}

		db := cmd.Context().Value("db").(*sql.DB)

		option := &discovery.Option{
			SubDomainScan: false,
			URLScan:       true,
			Verbose:       verbose,
			URLOption: discovery.URLOption{
				IncudeSubDomain: is,
			},
		}

		es := discovery.NewEumerationService(
			discovery.NewDomainEnumerationService(storage.NewDomainStorage(db), option),
			discovery.NewURLEumerationService(storage.NewURLStorage(db)),
			option)

		err = es.StartScan(cmd.Context(), domain)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(urlCmd)

	urlCmd.PersistentFlags().BoolVarP(&scanArgs.verbose, "verbose", "v", false, "Enable verbose output")
	urlCmd.PersistentFlags().StringVarP(&scanArgs.domain, "domain", "d", "", "domain")
	urlCmd.MarkPersistentFlagRequired("domain")
	urlCmd.PersistentFlags().BoolVarP(&urlScanArgs.includeSubDomains, "include-subdomains", "s", true, "Scan also subdomains")
}
