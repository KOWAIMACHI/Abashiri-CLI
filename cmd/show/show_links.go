/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package show

import (
	"abashiri-cli/storage"
	"database/sql"
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var ShowURLsCmd = &cobra.Command{
	Use:   "url",
	Short: "The \"url\" sub commands display the collected url",
	Long: `The "url" sub command display the list of urls that have been collected during the scanning process.

Example usage:
  $ abashiri show url -d example.com`,
	Run: func(cmd *cobra.Command, args []string) {

		db := cmd.Context().Value("db").(*sql.DB)
		ds := storage.NewDomainStorage(db)
		ls := storage.NewURLStorage(db)

		// domainにchildがいれば、再起的に表示したい
		// 今は、とりあえずrootドメインから取れる状態
		domain, _ := cmd.Flags().GetString("domain")
		domains, err := ds.GetSubDomainsByDomain(cmd.Context(), domain)
		if err != nil {
			log.Fatal(err)
		}
		for _, subdomain := range domains {
			urls, err := ls.GetURLs(cmd.Context(), subdomain)
			if err != nil {
				log.Fatal(err)
			}
			if urls == nil {
				continue
			}

			fmt.Printf("\nURLs of %s\n", subdomain)
			for _, url := range urls {
				fmt.Println(url)
			}
			fmt.Printf("\n")
		}
	},
}
