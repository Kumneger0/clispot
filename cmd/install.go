package cmd

import (
	"context"

	"github.com/kumneger0/clispot/internal/install"
	"github.com/spf13/cobra"
)

func installDeps() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "install",
		Short:        "install missing dependencies",
		SilenceUsage: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			isFFmpegOnly, err := cmd.Flags().GetBool("ffmpeg-only")
			if err != nil {
				return err
			}
			isYtDlpOnly, err := cmd.Flags().GetBool("yt-dlp-only")
			if err != nil {
				return err
			}

			isBoth := isFFmpegOnly && isYtDlpOnly

			if isBoth {
				//TODO: implement a function to download both of them
			}

			if isYtDlpOnly {
				_, err := install.YtDlp(context.TODO())
				if err != nil {
					return err
				}
			}

			if isFFmpegOnly {
				//TODO: implement a function to download yt-dlp
			}

			return nil
		},
	}

	cmd.Flags().Bool("ffmpeg-only", false, "installs only ffmpeg, if it already exists and it is not the latest version it install the latest version if it is already the latest version it doesn't do anything this includes ffprobe too")
	cmd.Flags().Bool("yt-dlp-only", false, "installs only yt-dlp, if it already exists and it is not the latest version it install the latest version if it is already the latest version it doesn't do anything")
	return cmd
}
