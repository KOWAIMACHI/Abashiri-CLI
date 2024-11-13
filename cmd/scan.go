/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"abashiri-cli/core/discovery"
	"abashiri-cli/storage"
	"context"
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
	Short: "Scan a domain for subdomains using passive or active methods",
	Long: `This command scans a given domain to discover subdomains using either passive or active techniques.
Passive scanning involves gathering subdomains from publicly available sources.
 - Subfinder
 - Amass

Active scanning performs actual requests to identify live subdomains.
 - DNSBruteforce

The results are then stored in a database for later reference.`,

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

		dir, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		db, err := sql.Open("sqlite3", fmt.Sprintf("%s/.abashiri/abashiri.db", dir))
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		option := &discovery.Option{
			Verbose: verbose,
		}

		es := discovery.NewEumerationService(
			discovery.NewDomainEnumerationService(storage.NewDomainStorage(db), option),
			discovery.NewURLEumerationService(storage.NewURLStorage(db)),
			option)

		ctx := context.Background()
		err = es.StartScan(ctx, domain, mode)
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
