/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"abashiri-cli/core/discovery"
	"abashiri-cli/storage"
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
)

type ScanArgs struct {
	domain  string
	mode    string
	verbose bool
}

var scanArgs = &ScanArgs{}

var scanCmd = &cobra.Command{
	Use:   "all",
	Short: "The \"all\" command performs comprehensive subdomain and URL discovery for a given domain",
	Long: `The "all" command performs comprehensive subdomain and URL discovery for a given domain

[!] Subdomain enumeration
  Passive Scanning: Collects subdomains from publicly available sources without direct interaction.
  Active Scanning: Actively probes for live subdomains through direct requests.

[!] URL enumeration
  Passive Scanning: Collects urls from publicly available sources without direct interaction.
  Active Scanning: Actively probes for live subdomains through direct requests.

[!] Example usage:
  $ abashiri all -d example.com
  $ abashiri all -d example.com -m passive
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
			Verbose:       verbose,
			URLScan:       true,
			SubDomainScan: true,
			SubDomainOption: discovery.SubDomainOption{
				Mode: mode,
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
	rootCmd.AddCommand(scanCmd)

	scanCmd.PersistentFlags().BoolVarP(&scanArgs.verbose, "verbose", "v", false, "Enable verbose output")
	scanCmd.PersistentFlags().StringVarP(&scanArgs.domain, "domain", "d", "", "root domain")
	scanCmd.MarkPersistentFlagRequired("domain")
	scanCmd.PersistentFlags().StringVarP(&scanArgs.mode, "mode", "m", "passive", "passive/active")
}
