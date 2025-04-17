/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the current kube-copilot context status",
	Long:  `Display various status items for kube-copilot, such as configuration and environment status.`,
	Run: func(cmd *cobra.Command, args []string) {
		openaiKey := viper.GetString("OPENAI_API_KEY")
		if openaiKey != "" {
			cmd.Println("OPENAI_API_KEY: set ✅")
		} else {
			cmd.Println("OPENAI_API_KEY: not set ❌")
		}
		namespace := viper.GetString("namespace")
		cmd.Printf("namespace: %s\n", namespace)
		// Add more status items here as needed
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
