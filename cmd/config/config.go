/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package config

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
var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Modify the gsync config file creating and setting contexts.",
}

func init() {
	logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	gcf.Directory = ".gsync"
	gcf.Name = "config.yaml"
}
