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
	Long:  ``,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&rootArgs.verbose, "verbose", "v", false, "Enable verbose output")
}
