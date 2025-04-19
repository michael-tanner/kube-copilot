/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set configuration parameters",
	Long: `Set configuration parameters for kube-copilot.

Usage:
  set ns <namespace>
  set namespace <namespace>
  set key <value>
  set openai_api_key <value>
`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := args[1]
		switch key {
		case "ns", "namespace":
			viper.Set("namespace", value)
			configPath := fmt.Sprintf("./%s/%s.%s", contextDir, contextFile, contextType)
			if err := viper.WriteConfigAs(configPath); err != nil {
				// If file does not exist, create it
				if _, ok := err.(viper.ConfigFileNotFoundError); ok {
					if err := viper.SafeWriteConfigAs(configPath); err != nil {
						cmd.Println("Failed to create config file:", err)
						return
					}
				} else {
					cmd.Println("Failed to write config:", err)
					return
				}
			}
			cmd.Printf("Namespace set to '%s'.\n", value)
		case "key", "openai_api_key":
			viper.Set("OPENAI_API_KEY", value)
			configPath := fmt.Sprintf("./%s/%s.%s", contextDir, contextFile, contextType)
			if err := viper.WriteConfigAs(configPath); err != nil {
				// If file does not exist, create it
				if _, ok := err.(viper.ConfigFileNotFoundError); ok {
					if err := viper.SafeWriteConfigAs(configPath); err != nil {
						cmd.Println("Failed to create config file:", err)
						return
					}
				} else {
					cmd.Println("Failed to write config:", err)
					return
				}
			}
			cmd.Println("OPENAI_API_KEY saved to context.")
		default:
			cmd.Printf("Unknown key: %s\n", key)
		}
	},
}

func init() {
	rootCmd.AddCommand(setCmd)
}
