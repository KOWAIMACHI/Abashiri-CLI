package show

import (
	"abashiri-cli/storage"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
)

var ShowDomainsCmd = &cobra.Command{
	Use:   "domain",
	Short: "List all enumerated subdomains for a given domain",
	Long: `This command retrieves and lists all the subdomains associated with a specified domain stored in an SQLite database.
It is part of a tool for managing collected domains and URLs, providing an easy way to view all subdomains that have been previously stored.`,
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
		defer db.Close()
		ds := storage.NewDomainStorage(db)
		ctx := context.Background()
		domains, err := ds.GetSubDomainsByDomain(ctx, domain)
		if err != nil {
			log.Fatal(err)
		}

		for _, subdomain := range domains {
			fmt.Println(subdomain)
		}
	},
}
