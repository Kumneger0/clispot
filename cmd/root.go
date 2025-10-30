package cmd

import (
	"fmt"

	"log"
	"log/slog"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/kumneger0/clispot/internal/mpris"
	"github.com/kumneger0/clispot/internal/spotify"
	"github.com/kumneger0/clispot/internal/types"
	"github.com/kumneger0/clispot/internal/ui"
)

var (
	Program *tea.Program
)

func newRootCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clispot",
		Short: "spotify music player",
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := spotify.ReadUserCredentials()

			if err != nil {
				slog.Error(err.Error())
				spotify.Authenticate()
			}

			if token.ExpiresAt < time.Now().Unix() && token.RefreshToken != "" {
				token, err = spotify.RefreshToken(token.RefreshToken)
				if err != nil {
					slog.Error(err.Error())
					log.Fatal(err)
				}
			}

			featuredPlaylist, err := spotify.GetFeaturedPlaylist(token.AccessToken)

			playListToRender := func() []types.SpotifyPlaylist {
				if err == nil && featuredPlaylist != nil {
					return featuredPlaylist.Playlists.Items
				}
				userPlayList, err := spotify.GetUserPlaylists(token.AccessToken)
				if err != nil || userPlayList == nil {
					slog.Error(err.Error())
					fmt.Fprintln(os.Stdout, err)
					return []types.SpotifyPlaylist{}
				}
				return userPlayList.Items
			}()

			var items []list.Item
			for _, item := range playListToRender {
				items = append(items, item)
			}

			playlists := list.New(items, ui.CustomDelegate{}, 10, 20)
			playlistItems := list.New([]list.Item{}, ui.CustomDelegate{}, 10, 20)

			input := textinput.New()
			input.Placeholder = "Search tracks, artists, albums..."
			input.Prompt = "> "
			input.CharLimit = 256

			musicQueueList := list.New([]list.Item{}, ui.CustomDelegate{}, 10, 20)

			ins, messageChan, err := mpris.GetDbusInstance()

			if err != nil {
				slog.Error(err.Error())
			}

			defer ins.Conn.Close()

			model := ui.Model{
				Playlist:              playlists,
				UserTokenInfo:         token,
				SelectedPlayListItems: playlistItems,
				FocusedOn:             ui.SideView,
				Search:                input,
				MusicQueueList:        musicQueueList,
				DBusConn:              ins,
			}

			Program := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())

			go func() {
				for v := range *messageChan {
					Program.Send(v)
				}
			}()

			_, err = Program.Run()
			if err != nil {
				slog.Error(err.Error())
				log.Fatal(err)
			}
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
