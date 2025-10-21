package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var (
	Program *tea.Program
)

func newRootCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clispot",
		Short: "clispot a cli based telegram client",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("this is a root command")
			return nil
		},
	}

	cmd.AddCommand(newVersionCmd(version))
	cmd.AddCommand(clispotLog())
	cmd.AddCommand(ManCmd(cmd))
	return cmd
}

func Execute(version string) error {
	if err := newRootCmd(version).Execute(); err != nil {
		return fmt.Errorf("error executing root command: %w", err)
	}
	return nil
}
