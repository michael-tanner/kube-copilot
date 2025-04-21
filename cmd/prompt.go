/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"strings"

	"github.com/michael-tanner/kube-copilot/internal/api"
	"github.com/spf13/cobra"
)

// promptCmd represents the prompt command
var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "Send a prompt to the AI assistant",
	Long: `Send a text prompt to the OpenAI assistant and display the response.
This is the main way to interact with the AI capabilities of kube-copilot.`,
	Args: cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Println("Error: No prompt provided.")
			cmd.Println("Usage: kube-copilot prompt <your question or prompt>")
			return
		}

		prompt := strings.Join(args, " ")
		fmt.Println("Prompt:", prompt)
		fmt.Println("Sending prompt to AI chat session...")
		fmt.Println()

		service := api.NewService()
		response, err := service.SendPrompt(prompt)
		if err != nil {
			cmd.Printf("Error: %v\n", err)
			return
		}

		if response != nil && response.Content != "" {
			fmt.Println(response.Content)
		} else {
			cmd.Println("Received empty response from AI assistant.")
		}
	},
}

func init() {
	rootCmd.AddCommand(promptCmd)
}
