/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"abashiri-cli/core/discovery"
	"abashiri-cli/storage"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
)

type ScanArgs struct {
	domain string
	mode   string
}

var scanArgs = &ScanArgs{}

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "The \"scan\" command performs comprehensive subdomain and URL discovery for a given domain",
	Long: `The "scan" command performs comprehensive subdomain and URL discovery for a given domain

Subdomain enumeration
  Passive Scanning: Collects subdomains from publicly available sources without direct interaction.
  Active Scanning: Actively probes for live subdomains through direct requests.

URL enumeration
  Passive Scanning: Collects urls from publicly available sources without direct interaction.
  Active Scanning: Actively probes for live subdomains through direct requests.

Example usage:
  $ abashiri scan -d example.com
  $ abashiri scan -d example.com -m passive
`,

	Run: func(cmd *cobra.Command, args []string) {
		domain, _ := cmd.Flags().GetString("domain")
		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			log.Fatalf("Failed to parse verbose flag: %v", err)
		}
		mode, _ := cmd.Flags().GetString("mode")

		if domain == "" {
			cmd.Usage()
			os.Exit(1)
		}
		fmt.Printf("[+] Scanning domain: %s\n", domain)

		db := cmd.Context().Value("db").(*sql.DB)

		option := &discovery.Option{
			Verbose: verbose,
		}

		es := discovery.NewEumerationService(
			discovery.NewDomainEnumerationService(storage.NewDomainStorage(db), option),
			discovery.NewURLEumerationService(storage.NewURLStorage(db)),
			option)

		err = es.StartScan(cmd.Context(), domain, mode)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)

	scanCmd.PersistentFlags().StringVarP(&scanArgs.domain, "domain", "d", "", "root domain")
	scanCmd.MarkPersistentFlagRequired("domain")
	scanCmd.PersistentFlags().StringVarP(&scanArgs.mode, "mode", "m", "passive", "passive/active")
}
