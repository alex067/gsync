/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log/slog"
	"os"

	"github.com/Consensys/gsync/cmd/clear"
	"github.com/Consensys/gsync/cmd/config"
	"github.com/Consensys/gsync/cmd/start"
	"github.com/spf13/cobra"
)

var logger *slog.Logger

// rootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "gsync",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
		slog.SetDefault(logger)
	}
	RootCmd.AddCommand(config.ConfigCmd)
	RootCmd.AddCommand(start.StartCmd)
	RootCmd.AddCommand(clear.ClearCmd)
}
