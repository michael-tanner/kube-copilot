package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// addCmd represents the add command for setting the OpenAI key
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add configuration parameters",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		keyName := args[0]
		keyValue := args[1]
		normalized := ""
		switch lower := strings.ToLower(keyName); lower {
		case "openai_api_key", "key":
			normalized = "OPENAI_API_KEY"
		default:
			cmd.Printf("Unknown key: %s\n", keyName)
			os.Exit(1)
		}
		viper.Set(normalized, keyValue)
		configPath := fmt.Sprintf("./%s/%s.%s", contextDir, contextFile, contextType)
		if err := viper.WriteConfigAs(configPath); err != nil {
			// If file does not exist, create it
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				if err := viper.SafeWriteConfigAs(configPath); err != nil {
					cmd.Println("Failed to create config file:", err)
					os.Exit(1)
				}
			} else {
				cmd.Println("Failed to write config:", err)
				os.Exit(1)
			}
		}
		cmd.Println("OPENAI_API_KEY saved to context.")
	},
}

// keyCmd represents the subcommand to add an OpenAI key
var keyCmd = &cobra.Command{
	Use:   "key [key]",
	Short: "Add or update the OpenAI API key in context",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		viper.Set("OPENAI_API_KEY", key)
		configPath := fmt.Sprintf("./%s/%s.%s", contextDir, contextFile, contextType)
		if err := viper.WriteConfigAs(configPath); err != nil {
			// If file does not exist, create it
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				if err := viper.SafeWriteConfigAs(configPath); err != nil {
					cmd.Println("Failed to create config file:", err)
					os.Exit(1)
				}
			} else {
				cmd.Println("Failed to write config:", err)
				os.Exit(1)
			}
		}
		cmd.Println("OPENAI_API_KEY saved to context.")
	},
}

func init() {
	addCmd.AddCommand(keyCmd)
}
