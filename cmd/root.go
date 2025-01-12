/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log/slog"
	"os"

	"github.com/alex067/gsync/cmd/clear"
	"github.com/alex067/gsync/cmd/config"
	"github.com/alex067/gsync/cmd/start"
	"github.com/alex067/gsync/cmd/version"
	"github.com/spf13/cobra"
)

var logger *slog.Logger

var RootCmd = &cobra.Command{
	Use:   "gsync",
	Short: "gsync syncs Grafana changes back to your local respository.",
}

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
	RootCmd.AddCommand(version.VersionCmd)
}
