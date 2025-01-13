/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package version

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/alex067/gsync/internal/pkg/version"
	"github.com/spf13/cobra"
)

var (
	logger *slog.Logger
)

// configCmd represents the config command
var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the current client version.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version: %s\n", version.Version)
	},
}

func init() {
	logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)
}
