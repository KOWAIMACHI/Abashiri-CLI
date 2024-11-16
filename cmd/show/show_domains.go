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
	Short: "The \"domain\" sub commands display the collected domain",
	Long: `The "domain" sub command display the list of domains that have been collected during the scanning process.

Example usage:
  $ abashiri show domain -d example.com`,
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
