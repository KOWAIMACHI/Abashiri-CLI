/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"abashiri-cli/storage"
	"database/sql"
	"log"

	"github.com/spf13/cobra"
)

type DeleteArgs struct {
	domain string
}

var deleteArgs = &DeleteArgs{}

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "The \"delete\" commands delete registered domain and associated subdomains from the records",
	Long: `The "delete" command allows you to remove a domain and its associated subdomains from the stored records.

Example usage:
  $ abashiri delete -d example.com`,
	Run: func(cmd *cobra.Command, args []string) {
		db := cmd.Context().Value("db").(*sql.DB)
		ds := storage.NewDomainStorage(db)

		domain, err := cmd.Flags().GetString("domain")
		if err != nil {
			log.Fatal(err)
		}

		err = ds.DeleteDomains(cmd.Context(), domain)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("[+] deleted %s domain from DB", domain)
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.PersistentFlags().StringVarP(&deleteArgs.domain, "domain", "d", "", "root domain")
}
