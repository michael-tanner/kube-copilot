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
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("Welcome to kube-copilot 😎")
		cmd.Println("Available commands:")
		cmd.Println("  help   - Display this help message")
		cmd.Println("  hello  - A sample command to demonstrate functionality")
		cmd.Println("\nIf no command is provided, plain text input will start an AI chat session.")
	},
}

func init() {
	rootCmd.AddCommand(helpCmd)
}
