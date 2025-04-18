/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// kubecheckCmd represents the kubecheck command
var kubecheckCmd = &cobra.Command{
	Use:   "kubecheck",
	Short: "List all namespaces in the current Kubernetes cluster",
	Long:  `Lists all namespaces from the current Kubernetes cluster using client-go.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("Starting kubecheck: listing namespaces in the current Kubernetes cluster...")

		var config *rest.Config
		var err error

		// Try in-cluster config first
		config, err = rest.InClusterConfig()
		if err == nil {
			cmd.Println("Using in-cluster Kubernetes configuration.")
		} else {
			// Fallback to kubeconfig file
			kubeconfig := os.Getenv("KUBECONFIG")
			cmd.Printf("Using kubeconfig: %s\n", kubeconfig)
			if kubeconfig == "" {
				home, err := os.UserHomeDir()
				if err != nil {
					cmd.Println("Unable to find home directory:", err)
					return
				}
				kubeconfig = home + "/.kube/config"
				cmd.Printf("Defaulting kubeconfig to: %s\n", kubeconfig)
			}
			cmd.Println("Loading kubeconfig...")
			config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
			if err != nil {
				cmd.Println("Could not find a valid Kubernetes configuration (neither in-cluster nor kubeconfig).")
				cmd.Println("Set up your kubeconfig or run inside a Kubernetes cluster.")
				return
			}
		}

		cmd.Println("Creating Kubernetes client...")
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			cmd.Println("Failed to create Kubernetes client:", err)
			return
		}

		cmd.Println("Querying namespaces from cluster...")
		namespaces, err := clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
		if err != nil {
			cmd.Println("Failed to list namespaces:", err)
			return
		}

		cmd.Println("Namespaces in the current cluster:")
		for _, ns := range namespaces.Items {
			cmd.Println(" -", ns.Name)
		}
		cmd.Println("kubecheck complete.")
	},
}

func init() {
	rootCmd.AddCommand(kubecheckCmd)

}
