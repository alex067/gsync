/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package config

import (
	"log/slog"
	"os"

	"github.com/alex067/gsync/internal/pkg/prompt"
	"github.com/spf13/cobra"
)

// getContextsCmd represents the getContexts command
var getContextsCmd = &cobra.Command{
	Use:   "get-contexts",
	Short: "Gets the current configured contexts set in the gysnc file.",
	Run: func(cmd *cobra.Command, args []string) {
		err := configContext.ReadConfigFile(gcf)
		if err != nil {
			logger.Error(
				"Failed to read context file",
				slog.String("error", err.Error()),
			)
			os.Exit(1)
		}

		var mSelector prompt.MultiSelector
		mSelector.RunGetContextDisplay(configContext.CurrentContext, configContext.Contexts)
	},
}

func init() {
	ConfigCmd.AddCommand(getContextsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getContextsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getContextsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
