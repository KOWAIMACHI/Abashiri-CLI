/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "abashiri",
	Short: "",
	Long: `. . 
し  < ABASHIRI-CLI!!!
 ▽`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		dir, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		db, err := sql.Open("sqlite3", fmt.Sprintf("%s/.abashiri/abashiri.db", dir))
		if err != nil {
			return err
		}

		if err := db.Ping(); err != nil {
			return err
		}

		// FIXME: かなり怪しい使い方
		cmd.SetContext(context.WithValue(cmd.Context(), "db", db))
		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		db := cmd.Context().Value("db").(*sql.DB)
		db.Close()
	},
}

func Execute() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
	rootCmd.Execute()
}
