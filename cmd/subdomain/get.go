package subdomain

import (
	"abashiri-cli/scan"
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
)

var GetCmd = &cobra.Command{
	Use:   "get",
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
		defer db.Close()
		ds := scan.NewDomainEnumerationService(db,
			&scan.Option{},
		)
		ds.GetSubDomains(domain)
	},
}