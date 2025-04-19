/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"strings"
	"github.com/michael-tanner/kube-copilot/internal/api"
	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the current kube-copilot context status",
	Long:  `Display various status items for kube-copilot, such as configuration and environment status.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("-----")
		cmd.Println("----- Kube Copilot CLI Status")
		cmd.Println("-----")
		service := api.NewService()
		status, err := service.CheckStatus()
		if err != nil {
			cmd.Printf("Error checking status: %v\n", err)
			return
		}
		if status.OpenaiApiKeyIsSet {
			cmd.Println("OPENAI_API_KEY: set ✅")
		} else {
			cmd.Println("OPENAI_API_KEY: not set ❌")
		}
		cmd.Printf("Current Namespace: %s\n", status.CurrentNamespace)
		cmd.Println("-----")
		cmd.Printf("Kube Cluster Name: %s\n", status.KubeClusterName)
		cmd.Printf("Namespaces: %s\n", strings.Join(status.KubeNamespaces, ", "))
		cmd.Println("-----")
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
