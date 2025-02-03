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

// subdomainCmd represents the subdomain command
var subdomainCmd = &cobra.Command{
	Use:   "subdomain",
	Short: "The \"subdomain\" command performs comprehensive subdomain discovery for a given domain",
	Long: `The "subdomain" command performs comprehensive subdomain discovery for a given domain

[!] Subdomain enumeration
  Passive Scanning: Collects subdomains from publicly available sources without direct interaction.
  Active Scanning: Actively probes for live subdomains through direct requests.

[!] Example usage:
  $ abashiri subdomain -d example.com
  $ abashiri subdomain -d example.com -m active -v 
`,

	Run: func(cmd *cobra.Command, args []string) {
		domain, err := cmd.Flags().GetString("domain")
		if err != nil {
			log.Fatalf("Failed to parse --domain/-d: %v", err)
		}
		mode, err := cmd.Flags().GetString("mode")
		if err != nil {
			log.Fatalf("Failed to parse --mode/-d: %v", err)
		}
		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			log.Fatalf("Failed to parse --verbose/-v: %v", err)
		}

		if domain == "" {
			cmd.Usage()
			os.Exit(1)
		}

		db := cmd.Context().Value("db").(*sql.DB)

		option := &discovery.Option{
			SubDomainScan: true,
			URLScan:       true,
			Verbose:       verbose,
			SubDomainOption: discovery.SubDomainOption{
				Mode: mode,
			},
			URLOption: discovery.URLOption{
				IncudeSubDomain: true,
			},
		}

		es := discovery.NewEumerationService(storage.NewStorageService(db), option)

		err = es.StartScan(cmd.Context(), domain)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(subdomainCmd)

	subdomainCmd.PersistentFlags().BoolVarP(&scanArgs.verbose, "verbose", "v", false, "Enable verbose output")
	subdomainCmd.PersistentFlags().StringVarP(&scanArgs.domain, "domain", "d", "", "root domain")
	subdomainCmd.MarkPersistentFlagRequired("domain")
	subdomainCmd.PersistentFlags().StringVarP(&scanArgs.mode, "mode", "m", "passive", "passive/active")
}
