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

	"github.com/spf13/cobra"
)

// showLinksCmd represents the showLinks command
var ShowLinksCmd = &cobra.Command{
	Use:   "links",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		domain, _ := cmd.Flags().GetString("domain")
		db, err := sql.Open("sqlite3", "./abashiri.db")
		if err != nil {
			log.Fatal(err)
		}
		ds := storage.NewDomainStorage(db)
		ls := storage.NewLinkStorage(db)
		ctx := context.Background()

		// domainにchildがいれば、再起的に表示したい
		// 今は、とりあえずrootドメインから取れる状態
		domains, err := ds.GetSubDomains(ctx, domain)
		if err != nil {
			log.Fatal(err)
		}
		for _, subdomain := range domains {
			links, err := ls.GetLinks(ctx, subdomain)
			if err != nil {
				log.Fatal(err)
			}
			if links == nil {
				continue
			}

			fmt.Printf("\nLinks of %s\n", subdomain)
			for _, link := range links {
				fmt.Println(link)
			}
			fmt.Printf("\n")
		}
	},
}
