/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package config

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/Consensys/gsync/internal/pkg/gcontext"
	"github.com/spf13/cobra"
)

// createContextCmd represents the createContext command
var createContextCmd = &cobra.Command{
	Use:   "create-context",
	Short: "Create a new gsync context",
	Long:  `Create a new gsync context to allow for fast switching between different grafana tenants and configs`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("Creating new context")

		absConfigPath, absConfigFilePath, err := gcf.GetAbsolutePath()
		if err != nil {
			logger.Error(err.Error())
			os.Exit(1)
		}

		// Search and/or create config diretory
		if _, err := os.Stat(absConfigPath); err != nil {
			if os.IsNotExist(err) {
				logger.Info("Creating new local config directory in user home")
				os.MkdirAll(absConfigPath, 0755)
				//isContextDirectoryExist = false
			} else {
				logger.Error("Failed searching for local config directory", slog.String("error", err.Error()))
				os.Exit(1)
			}
		}

		var newContext gcontext.GContext

		fmt.Print("Grafana URL (Required): ")
		fmt.Scanln(&newContext.Url)
		fmt.Print("Context Name (Required): ")
		fmt.Scanln(&newContext.Name)
		fmt.Print("Dashboards Path (Absolute): ")
		fmt.Scanln(&newContext.Context.Dashboards.Path)
		fmt.Print("Grafana Tenant (Required): ")
		fmt.Scanln(&newContext.Context.Dashboards.GrafanaTenant)
		fmt.Print("Grafana Auth Token (Required): ")
		fmt.Scanln(&newContext.Authentication.Grafana.Token)
		fmt.Print("Grafana Gsync Folder Uid (Optional): ")
		fmt.Scanln(&newContext.Context.Dashboards.GrafanResources.FolderUid)

		err = configContext.CreateNewContext(newContext, gcf)
		if err != nil {
			logger.Error("Failed to create new context", slog.String("error", err.Error()))
			os.Exit(1)
		}

		logger.Info("Stored new context: ", newContext.Name, absConfigFilePath)
		logger.Info("Successfully created context")
	},
}

func init() {
	ConfigCmd.AddCommand(createContextCmd)
}
