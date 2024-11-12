package subdomain

import (
	"abashiri-cli/core/discovery"
	"abashiri-cli/storage"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
)

var ScanCmd = &cobra.Command{
	Use:   "scan",
	Short: "",
	Long:  ` `,
	Run: func(cmd *cobra.Command, args []string) {
		domain, _ := cmd.Flags().GetString("domain")
		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			log.Fatalf("Failed to parse verbose flag: %v", err)
		}
		mode, _ := cmd.Flags().GetString("mode")

		if domain == "" {
			fmt.Println("domain flag is required")
			cmd.Usage()
			os.Exit(1)
		}
		fmt.Printf("[+] Scanning domain: %s\n", domain)

		db, err := sql.Open("sqlite3", "./abashiri.db")
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		ds := discovery.NewDomainEnumerationService(
			storage.NewDomainStorage(db),
			&discovery.Option{
				Verbose: verbose,
			},
		)
		ctx := context.Background()
		err = ds.StartScan(ctx, domain, mode)
		if err != nil {
			log.Fatal(err)
		}
	},
}
