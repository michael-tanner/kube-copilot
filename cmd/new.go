/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/michael-tanner/kube-copilot/api"

	"github.com/spf13/cobra"
)

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Start a new conversation thread",
	Long: `Resets the current conversation thread with the AI assistant,
starting a fresh conversation without the context of previous interactions.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := api.UpdateCurrentThreadID("")
		if err != nil {
			cmd.Println("Error resetting conversation thread:", err)
			return
		}
		cmd.Println("Started a new conversation thread. Previous context has been cleared.")
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
}
