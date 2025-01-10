/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package config

import (
	"log/slog"
	"os"

	"github.com/Consensys/gsync/internal/pkg/gcontext"
	"github.com/spf13/cobra"
)

var (
	gcf           gcontext.GConfigFile
	logger        *slog.Logger
	configContext gcontext.GConfigContext
)

// configCmd represents the config command
var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "A brief description of your command",
	Long:  ``,
}

func init() {
	logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	gcf.Directory = ".gsync"
	gcf.Name = "config.yaml"
}
