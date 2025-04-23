/*
Proxy for kubectl
*/
package cmd

import (
	"github.com/michael-tanner/kube-copilot/api"
	"github.com/spf13/cobra"
)

// kCmd represents the k command
var kCmd = &cobra.Command{
	Use:   "k",
	Short: "Run kubectl-like commands (e.g. get pods, get services)",
	Long:  `Proxy kubectl commands through kube-copilot using the Go client.`,
	Args:  cobra.ArbitraryArgs, // <-- allow arbitrary args and flags
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Println("Usage: k <kubectl-subcommand> [args...]")
			return
		}
		service := api.NewService(&api.CobraOutputWriter{Cmd: cmd})
		resp, err := service.KubectlProxy(args)
		if err != nil {
			cmd.Println("kubectl error:", err)
			return
		}
		for _, line := range resp {
			cmd.Println(line)
		}
	},
}

func init() {
	kCmd.DisableFlagParsing = true
	rootCmd.AddCommand(kCmd)
}
