/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package show

import (
	"abashiri-cli/storage"
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var EXCLUDE_EXT = map[string]bool{
	".mp3":  true,
	".gif":  true,
	".css":  true,
	".js":   true,
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".svg":  true,
	".bmp":  true,
	".webp": true,
	".ico":  true,
	".pdf":  true,
	".ttf":  true,
}

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
		domains, err := ds.GetSubDomainsByParent(cmd.Context(), domain)
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
			var currentURL *url.URL
			for i, u := range urls {
				u, err := url.Parse(u)
				if i == 0 {
					currentURL = u
				}
				if err != nil {
					log.Fatal(err)
				}
				if EXCLUDE_EXT[strings.ToLower(filepath.Ext(u.Path))] {
					continue
				}

				if currentURL.Host == u.Host && currentURL.Path == u.Path {
					continue
				}
				fmt.Println(u)
				currentURL = u
			}
			fmt.Printf("\n")
		}
	},
}
