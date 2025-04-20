/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("get called")
	},
}

var spendCmd = &cobra.Command{
	Use:   "spend",
	Short: "Show OpenAI API spend for today, yesterday, and the last 7 days",
	Run: func(cmd *cobra.Command, args []string) {
		// service := api.NewService()
		// spend, err := service.GetOpenAISpend()
		// if err != nil {
		// 	cmd.Println("Error retrieving spend:", err)
		// 	return
		// }
		// cmd.Printf("OpenAI Spend:\n")
		// cmd.Printf("  Today:     $%.2f\n", spend.Today)
		// cmd.Printf("  Yesterday: $%.2f\n", spend.Yesterday)
		// cmd.Printf("  Last 7d:   $%.2f\n", spend.Last7Days)
		cmd.Println("Spend command is not implemented yet.")
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.AddCommand(spendCmd)
}
