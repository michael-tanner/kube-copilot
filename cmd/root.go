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
			fmt.Println("No command provided. Starting AI chat session...")
			// Placeholder for AI chat session logic
			fmt.Println("AI chat session started. Type your queries below:")
			return
		}
		fmt.Printf("Unknown command or input: %s\n", args[0])
		fmt.Println("Use 'help' to see available commands.")
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
	// Set up Viper for context management

	// Ensure the context directory exists
	if _, err := os.Stat(contextDir); os.IsNotExist(err) {
		err := os.MkdirAll(contextDir, 0755)
		if err != nil {
			fmt.Println("Unable to create context directory:", err)
			os.Exit(1)
		}
	}

	viper.SetConfigName(contextFile)
	viper.SetConfigType(contextType)
	viper.AddConfigPath(contextDir)

	// Read in config file if it exists
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using context file:", viper.ConfigFileUsed())
	}

	// Optionally set defaults for context
	// viper.SetDefault("context", map[string]interface{}{})

	// Example: bind a persistent flag to a viper key
	// rootCmd.PersistentFlags().String("context", "", "Set context")
	// viper.BindPFlag("context", rootCmd.PersistentFlags().Lookup("context"))

	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.kube-copilot.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// Register the add command
	rootCmd.AddCommand(addCmd)
}

// addCmd represents the add command for setting the OpenAI key
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add configuration parameters",
}

// oaikeyCmd represents the subcommand to add an OpenAI key
var oaikeyCmd = &cobra.Command{
	Use:   "oaikey [key]",
	Short: "Add or update the OpenAI API key in context",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		viper.Set("OpenAIKey", key)
		configPath := fmt.Sprintf("./%s/%s.%s", contextDir, contextFile, contextType)
		if err := viper.WriteConfigAs(configPath); err != nil {
			// If file does not exist, create it
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				if err := viper.SafeWriteConfigAs(configPath); err != nil {
					fmt.Println("Failed to create config file:", err)
					os.Exit(1)
				}
			} else {
				fmt.Println("Failed to write config:", err)
				os.Exit(1)
			}
		}
		fmt.Println("OpenAIKey saved to context.")
	},
}

func init() {
	addCmd.AddCommand(oaikeyCmd)
}
