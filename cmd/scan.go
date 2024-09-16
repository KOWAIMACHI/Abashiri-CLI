/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"abashiri-cli/scan"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
)

var domain string
var verbose bool

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if domain == "" {
			fmt.Println("Error: --domain flag is required")
			cmd.Usage()
			os.Exit(1)
		}
		fmt.Printf("[+] Scanning domain: %s\n", domain)
		db, err := sql.Open("sqlite3", "./abashiri.db")
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
		ds := scan.NewDomainEnumerationService(db,
			&scan.Option{
				Verbose: verbose,
			},
		)
		ds.StartScan(domain)
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
	scanCmd.PersistentFlags().StringVar(&domain, "domain", "", "root domain")
	scanCmd.MarkPersistentFlagRequired("domain")
	scanCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
}
