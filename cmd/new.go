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
	Short: "Start a new ai chat thread",
	Long: `Resets the current ai chat thread with the AI assistant,
starting a fresh conversation without the context of previous interactions.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := api.UpdateCurrentThreadID("")
		if err != nil {
			cmd.Println("Error resetting chat ai thread:", err)
			return
		}
		cmd.Println("Started a new ai thread. Previous thread id has been removed.")
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
}
