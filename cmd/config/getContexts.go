/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package config

import (
	"log/slog"
	"os"

	"github.com/Consensys/gsync/internal/pkg/prompt"
	"github.com/spf13/cobra"
)

// getContextsCmd represents the getContexts command
var getContextsCmd = &cobra.Command{
	Use:   "get-contexts",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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
