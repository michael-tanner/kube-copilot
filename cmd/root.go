/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/michael-tanner/kube-copilot/api/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kube-copilot",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.ArbitraryArgs, // Allow arbitrary arguments for catch-all
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Println("No command provided. Starting AI chat session...")
			// Placeholder for AI chat session logic
			cmd.Println("AI chat session started. Type your queries below:")
			return
		}
		// Forward all unknown input to the prompt command
		promptCmd.Run(cmd, args)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// GetRootCmd returns the root command instance
func GetRootCmd() *cobra.Command {
	return rootCmd
}

func init() {
	// Ensure the context directory exists
	if _, err := os.Stat(config.ContextDir); os.IsNotExist(err) {
		err := os.MkdirAll(config.ContextDir, 0755)
		if err != nil {
			fmt.Println("Unable to create context directory:", err)
			os.Exit(1)
		}
	}

	// Central Viper config setup
	config.SetupViperConfig()
	_ = viper.ReadInConfig() // Ignore error if config does not exist

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
