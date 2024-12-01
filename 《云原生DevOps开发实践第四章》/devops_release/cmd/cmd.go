package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"devops_release/cmd/server"
)

var rootCmd = &cobra.Command{
	Use:               "nighting-release",
	Short:             "-v",
	SilenceUsage:      true,
	DisableAutoGenTag: true,
	Long:              `nr`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("至少需要一个参数")
		}
		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		usageStr := `nighting-release, 可以使用 -h 查看命令`
		fmt.Println(usageStr)
	},
}

func init() {
	rootCmd.AddCommand(server.StartCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}
