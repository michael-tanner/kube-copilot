/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	contextDir  = ".kube-copilot-session"
	contextFile = "kc-context"
	contextType = "yaml"
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
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Println("No command provided. Starting AI chat session...")
			// Placeholder for AI chat session logic
			cmd.Println("AI chat session started. Type your queries below:")
			return
		}
		cmd.Printf("Unknown command or input: %s\n", args[0])
		cmd.Println("Use 'help' to see available commands.")
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
	if _, err := os.Stat(contextDir); os.IsNotExist(err) {
		err := os.MkdirAll(contextDir, 0755)
		if err != nil {
			fmt.Println("Unable to create context directory:", err)
			os.Exit(1)
		}
	}

	// Central Viper config setup
	viper.SetConfigName(contextFile)
	viper.SetConfigType(contextType)
	viper.AddConfigPath(contextDir)
	_ = viper.ReadInConfig() // Ignore error if config does not exist

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// Register subcommands here
	rootCmd.AddCommand(addCmd)
	// ...add other commands as needed...
}
