package cmd

import (
	"context"
	"log"
	"os"

	"github.com/kumneger0/clispot/internal/command"
	"github.com/kumneger0/clispot/internal/install"
	"github.com/spf13/cobra"
)

func installDeps() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "install",
		Short:        "install missing dependencies",
		SilenceUsage: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			ffmpegInstallCmd, err := install.FFmpegInstallCommand()
			if err != nil {
				log.Fatal(err)
			}
			for _, step := range ffmpegInstallCmd {
				installCmd, err := command.ExecCommand(ctx, step.Command, step.Args...)
				if err != nil {
					log.Fatal(err)
				}
				installCmd.Stdout = os.Stdout
				installCmd.Stdin = os.Stdin
				installCmd.Stderr = os.Stderr
				if err := installCmd.Run(); err != nil {
					log.Fatal(err)
				}
			}

			return nil
		},
	}
	return cmd
}
