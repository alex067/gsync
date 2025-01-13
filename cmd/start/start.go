/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package start

import (
	"log/slog"
	"os"

	"github.com/alex067/gsync/internal/pkg/gcontext"
	"github.com/spf13/cobra"
)

var (
	gcf           gcontext.GConfigFile
	logger        *slog.Logger
	configContext gcontext.GConfigContext
)

// configCmd represents the config command
var StartCmd = &cobra.Command{
	Use:   "start",
	Short: "A brief description of your command",
	Long:  ``,
}

func init() {
	logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	gcf.Directory = ".gsync"
	gcf.Name = "config.yaml"

	StartCmd.AddCommand(dashboardCmd)
}
