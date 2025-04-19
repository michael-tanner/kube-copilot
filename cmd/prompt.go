/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// promptCmd represents the prompt command
var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "Send a prompt to the AI chat session",
	Long:  "Send a prompt to the AI chat session. You can also invoke the CLI with any text and it will be treated as a prompt.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			cmd.Println(joinArgs(args))
		} else {
			cmd.Println("")
		}
	},
}

// joinArgs joins the arguments with spaces.
func joinArgs(args []string) string {
	result := ""
	for i, arg := range args {
		if i > 0 {
			result += " "
		}
		result += arg
	}
	return result
}

func init() {
	rootCmd.AddCommand(promptCmd)
}
