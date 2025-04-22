/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/michael-tanner/kube-copilot/api/config"
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
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := ""
		if len(args) > 1 {
			value = args[1]
		}

		// Define aliases for config keys
		aliases := map[string]string{
			"key":            "OPENAI_API_KEY",
			"openai_api_key": "OPENAI_API_KEY",
			"ns":             "namespace",
			"namespace":      "namespace",
		}

		// Use alias if present
		if realKey, ok := aliases[key]; ok {
			key = realKey
		}

		viper.Set(key, value)
		configPath := fmt.Sprintf("./%s/%s.%s", config.ContextDir, config.ContextFile, config.ContextType)
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
		cmd.Printf("Set '%s' to '%s' in context.\n", key, value)
	},
}

func init() {
	rootCmd.AddCommand(setCmd)
}
