/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

type RootArgs struct {
	verbose bool
}

var rootArgs = &RootArgs{}

var rootCmd = &cobra.Command{
	Use:   "abashiri-cli",
	Short: "",
	Long: `. . 
し  < ABASHIRI-CLI!!!
 ▽`,
}

func Execute() {
	rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&rootArgs.verbose, "verbose", "v", false, "Enable verbose output")
}
