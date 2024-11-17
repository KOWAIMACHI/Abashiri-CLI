package show

import (
	"abashiri-cli/storage"
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
)

var ShowDomainsCmd = &cobra.Command{
	Use:   "domain",
	Short: "The \"domain\" sub commands display the collected domain",
	Long: `The "domain" sub command display the list of domains that have been collected during the scanning process.

Example usage:
  $ abashiri show domain -d example.com
  $ abashiri show domain --root/-r`,
	Run: func(cmd *cobra.Command, args []string) {
		db := cmd.Context().Value("db").(*sql.DB)
		ds := storage.NewDomainStorage(db)

		isRoot, err := cmd.Flags().GetBool("root")
		if err != nil {
			log.Fatalf("Failed to parse root flag: %v", err)
		}

		if isRoot {
			showRootDomains(cmd.Context(), ds)
			return
		}

		domain, err := cmd.Flags().GetString("domain")
		if err != nil {
			log.Fatal(err)
		}

		if domain == "" {
			cmd.Help()
			return
		}

		if err := showSubDomains(cmd.Context(), ds, domain); err != nil {
			log.Fatal(err)
		}
	},
}

func showRootDomains(ctx context.Context, ds storage.DomainStorage) error {
	domains, err := ds.GetRootDomains(ctx)
	if err != nil {
		return err
	}

	for _, domain := range domains {
		fmt.Println(domain)
	}
	return nil
}

func showSubDomains(ctx context.Context, ds storage.DomainStorage, domain string) error {
	domains, err := ds.GetSubDomainsByDomain(ctx, domain)
	if err != nil {
		return err
	}

	for _, domain := range domains {
		fmt.Println(domain)
	}
	return nil
}
