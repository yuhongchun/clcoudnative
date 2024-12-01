package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"devops_build/cmd/server"
)

var rootCmd = &cobra.Command{
	Use:          "nighting-build",
	Short:        "nighting build tool",
	Long:         "uniform config file building center",
	SilenceUsage: true,
	Args:         cobra.MinimumNArgs(1),
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	rootCmd.AddCommand(server.ServerCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
