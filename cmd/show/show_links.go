/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package show

import (
	"abashiri-cli/storage"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var ShowURLsCmd = &cobra.Command{
	Use:   "url",
	Short: "The \"url\" sub commands display the collected url",
	Long: `The "url" sub command display the list of urls that have been collected during the scanning process.

Example usage:
  $ abashiri show url -d example.com`,
	Run: func(cmd *cobra.Command, args []string) {
		domain, _ := cmd.Flags().GetString("domain")

		dir, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}

		db, err := sql.Open("sqlite3", fmt.Sprintf("%s/.abashiri/abashiri.db", dir))
		if err != nil {
			log.Fatal(err)
		}
		ds := storage.NewDomainStorage(db)
		ls := storage.NewURLStorage(db)
		ctx := context.Background()

		// domainにchildがいれば、再起的に表示したい
		// 今は、とりあえずrootドメインから取れる状態
		domains, err := ds.GetSubDomainsByDomain(ctx, domain)
		if err != nil {
			log.Fatal(err)
		}
		for _, subdomain := range domains {
			urls, err := ls.GetURLs(ctx, subdomain)
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
