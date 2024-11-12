package subdomain

import (
	"abashiri-cli/storage"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
)

var GetCmd = &cobra.Command{
	Use:   "get",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		domain, _ := cmd.Flags().GetString("domain")
		db, err := sql.Open("sqlite3", "./abashiri.db")
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
		ds := storage.NewDomainStorage(db)
		domains, err := ds.GetSubDomains(domain)
		if err != nil {
			log.Fatal(err)
		}

		for _, subdomain := range domains {
			fmt.Println(subdomain)
		}
	},
}
