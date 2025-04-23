/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get configuration parameters",
	Long: `Retrieve configuration parameters for kube-copilot.

Usage:
  get <key>
  get <key.subkey> (for nested parameters)
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		keys := strings.Split(key, ".")
		value := viper.Get(keys[0])

		// Traverse nested keys if applicable
		for _, subkey := range keys[1:] {
			if nestedMap, ok := value.(map[string]interface{}); ok {
				value = nestedMap[subkey]
			} else {
				value = nil
				break
			}
		}

		if value == nil {
			cmd.Printf("Key '%s' not found.\n", key)
		} else {
			cmd.Printf("Value for '%s': %v\n", key, value)
		}
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
