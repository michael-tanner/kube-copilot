/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"strings"

	"github.com/michael-tanner/kube-copilot/internal/api"

	"github.com/spf13/cobra"
)

var service = api.NewService()

// promptCmd represents the prompt command
var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "Send a prompt to the AI chat session",
	Long:  "Send a prompt to the AI chat session. You can also invoke the CLI with any text and it will be treated as a prompt.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			prompt := strings.Join(args, " ")
			cmd.Println("Prompt:", prompt)
			cmd.Println("Sending prompt to AI chat session...")
			resp, err := service.SendPrompt(prompt)
			if err != nil {
				cmd.Println("Error sending prompt:", err)
				return
			}
			cmd.Println(resp.InputPrompt)
		} else {
			cmd.Println("Error: No prompt provided.")
			cmd.SilenceUsage = true
			// Optionally, set a non-zero exit code:
			// os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(promptCmd)
}
