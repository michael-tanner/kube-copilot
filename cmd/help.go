/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// helpCmd represents the help command
var helpCmd = &cobra.Command{
	Use:   "help",
	Short: "Show help for kube-copilot commands",
	Long: `kube-copilot: AI client for Kubernetes

Interact with your Kubernetes cluster using natural language or commands.

Available commands:
  help        - Show this help message
  set         - Set configuration parameters (e.g., namespace, OpenAI key)
  status      - Show the current kube-copilot context status
  ns          - List all namespaces in the current Kubernetes cluster (aliases: namespace, namespaces)

If no command is provided, plain text input will start an AI chat session.
`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("kube-copilot: AI client for Kubernetes")
		cmd.Println()
		cmd.Println("Available commands:")
		cmd.Println("  help        - Show this help message")
		cmd.Println("  set         - Set configuration parameters (e.g., namespace, OpenAI key)")
		cmd.Println("  status      - Show the current kube-copilot context status")
		cmd.Println("  ns          - List all namespaces in the current Kubernetes cluster (aliases: namespace, namespaces)")
		cmd.Println()
		cmd.Println("If no command is provided, plain text input will start an AI chat session.")
	},
}

func init() {
	rootCmd.AddCommand(helpCmd)
}
