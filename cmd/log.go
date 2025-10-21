package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func clispotLog() *cobra.Command {
	return &cobra.Command{
		Use:          "log",
		Short:        "show log files use it for error reporting",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			logs, err := os.ReadFile("/tmp/clispot.log")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading log file: %v", err)
				os.Exit(1)
			}
			fmt.Print(string(logs))
		},
	}
}
