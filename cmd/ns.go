/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/michael-tanner/kube-copilot/internal/api"
	"github.com/spf13/cobra"
)

// nsCmd represents the namespace listing command
var nsCmd = &cobra.Command{
	Use:     "ns",
	Aliases: []string{"namespace", "namespaces"},
	Short:   "List all namespaces in the current Kubernetes cluster",
	Long:    `Lists all namespaces from the current Kubernetes cluster using client-go.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("Listing namespaces in the current Kubernetes cluster...")

		service := api.NewService()
		namespaces, err := service.GetKubeNamespaces()
		if err != nil {
			cmd.Println("Failed to list namespaces:", err)
			return
		}

		cmd.Println("Namespaces in the current cluster:")
		for _, ns := range namespaces {
			cmd.Println(" -", ns)
		}
		cmd.Println("Namespace listing complete.")
	},
}

func init() {
	rootCmd.AddCommand(nsCmd)
}
