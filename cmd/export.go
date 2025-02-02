/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"abashiri-cli/storage"
	"database/sql"
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

//go:embed export.tmpl
var htmlTemplate []byte

type ExportArgs struct {
	domain string
}

var exportArgs = &ExportArgs{}

var EXCLUDE_EXT = map[string]bool{
	".mp3":   true,
	".gif":   true,
	".css":   true,
	".js":    true,
	".jpg":   true,
	".jpeg":  true,
	".png":   true,
	".svg":   true,
	".bmp":   true,
	".webp":  true,
	".ico":   true,
	".pdf":   true,
	".ttf":   true,
	".woff":  true,
	".woff2": true,
}

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export scanned URLs as an HTML file",
	Long:  "Exports all URLs found under a specified domain and its subdomains into an HTML file.",
	Run: func(cmd *cobra.Command, args []string) {
		db := cmd.Context().Value("db").(*sql.DB)
		ds := storage.NewDomainStorage(db)
		ls := storage.NewURLStorage(db)

		domain, _ := cmd.Flags().GetString("domain")

		domains, err := ds.GetSubDomainsByParent(cmd.Context(), domain)
		if err != nil {
			log.Fatal(err)
		}

		outputFile := path.Base(domain + ".html")

		var data []struct {
			Domain string
			URLs   []string
		}

		for _, subdomain := range domains {
			urls, err := ls.GetURLs(cmd.Context(), subdomain)
			if err != nil {
				log.Fatal(err)
			}
			if urls == nil {
				continue
			}
			var filteredURLs []string
			var currentURL *url.URL
			for i, u := range urls {
				uu, err := url.Parse(u)
				if i == 0 {
					currentURL = uu
				}
				if err != nil {
					log.Fatal(err)
				}
				if EXCLUDE_EXT[strings.ToLower(filepath.Ext(uu.Path))] {
					continue
				}

				if currentURL.Host == uu.Host && currentURL.Path == uu.Path {
					continue
				}
				filteredURLs = append(filteredURLs, u)
				currentURL = uu
			}
			data = append(data, struct {
				Domain string
				URLs   []string
			}{Domain: subdomain, URLs: filteredURLs})
		}

		file, err := os.Create(outputFile)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		tmpl := template.Must(template.New("export").Parse(string(htmlTemplate)))
		if err := tmpl.Execute(file, data); err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Exported URLs to %s\n", outputFile)
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
	exportCmd.PersistentFlags().StringVarP(&exportArgs.domain, "domain", "d", "", "root domain")
	exportCmd.MarkPersistentFlagRequired("domain")
}
