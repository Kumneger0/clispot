package cmd

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"log"
	"log/slog"
	"os"
	"time"

	logSetup "github.com/kumneger0/clispot/internal/logger"

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

func newRootCmd(version string, spotifyClientID string, spotifyClientSecret string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clispot",
		Short: "spotify music player",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRoot(cmd, spotifyClientID, spotifyClientSecret)
		},
	}

	cmd.AddCommand(newVersionCmd(version))
	cmd.AddCommand(clispotLog())
	cmd.AddCommand(ManCmd(cmd))
	return cmd
}

func runRoot(cmd *cobra.Command, spotifyClientID, spotifyClientSecret string) error {
	debugDir, err := cmd.Flags().GetString("debug-dir")

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if err := os.MkdirAll(debugDir, 0755); err != nil {
		fmt.Printf("failed to create debug directory '%s': %v\n", debugDir, err)
		os.Exit(1)
	}

	fileInfo, err := os.Stat(debugDir)
	if err != nil {
		fmt.Printf("failed to stat debug directory '%s': %v\n", debugDir, err)
		os.Exit(1)
	}

	if !fileInfo.IsDir() {
		fmt.Printf("the debug path '%v' is not a directory\n", debugDir)
		os.Exit(1)
	}

	logger := logSetup.Init(debugDir)
	defer logger.Close()

	err = doAllDepsInstalled()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
		return nil
	}
	token, err := spotify.ReadUserCredentials()

	if err != nil {
		slog.Error(err.Error())
		token, _ = spotify.Authenticate(spotifyClientID, spotifyClientSecret)
	}

	if token == nil {
		slog.Error("failed to get user token")
		fmt.Println("we have failed to get your access token please open up an issue on our github page")
		os.Exit(1)
	}
	if token.ExpiresAt < time.Now().Unix() && token.RefreshToken != "" {
		token, err = spotify.RefreshToken(token.RefreshToken, spotifyClientID, spotifyClientSecret)
		if err != nil {
			slog.Error(err.Error())
			userHomeDir, _ := os.UserHomeDir()
			clispotLogDir := filepath.Join(userHomeDir, ".clispot")
			fmt.Printf("we have failed to refresh ur token could you delete clispot dir by using rm -rf %v", clispotLogDir)
			os.Exit(1)
		}
	}

	featuredPlaylist, err := spotify.GetFeaturedPlaylist(token.AccessToken)
	playListToRender := func() []types.Playlist {
		if err == nil && featuredPlaylist != nil {
			return featuredPlaylist.Playlists.Items
		}
		userPlayList, err := spotify.GetUserPlaylists(token.AccessToken)
		if err != nil || userPlayList == nil {
			slog.Error(err.Error())
			fmt.Fprintln(os.Stdout, err)
			return []types.Playlist{}
		}
		return userPlayList.Items
	}()

	ins, messageChan, err := mpris.GetDbusInstance()

	if err != nil {
		slog.Error(err.Error())
	}

	var items []list.Item
	for _, item := range playListToRender {
		items = append(items, item)
	}

	model := ui.Model{
		GetUserToken: func() *types.UserTokenInfo {
			token, err := validateToken(token, spotifyClientID, spotifyClientSecret)
			if err != nil {
				slog.Error(err.Error())
				return nil
			}
			return token
		},
		FocusedOn:    ui.SideView,
		DBusConn:     ins,
		MainViewMode: ui.NormalMode,
		UserArguments: &ui.UserArguments{
			DebugPath: debugDir,
		},
	}

	playlists := list.New(items, ui.CustomDelegate{Model: &model}, 10, 20)
	playlistItems := list.New([]list.Item{}, ui.CustomDelegate{Model: &model}, 10, 20)

	input := textinput.New()
	input.Placeholder = "Search tracks, artists, albums..."
	input.Prompt = "> "
	input.CharLimit = 256

	model.Search = input
	musicQueueList := list.New([]list.Item{}, ui.CustomDelegate{Model: &model}, 10, 20)

	model.Playlist = playlists
	model.SelectedPlayListItems = playlistItems
	model.MusicQueueList = musicQueueList

	Program := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())

	go func() {
		if messageChan == nil {
			return
		}
		for v := range *messageChan {
			Program.Send(v)
		}
	}()

	_, err = Program.Run()
	if err != nil {
		slog.Error(err.Error())
		log.Fatal(err)
	}

	if ins != nil {
		ins.Conn.Close()
	}
	return nil
}

func validateToken(token *types.UserTokenInfo, spotifyClientID, spotifyClientSecret string) (*types.UserTokenInfo, error) {
	if token.ExpiresAt > time.Now().Unix() {
		return token, nil
	}
	if token.RefreshToken != "" {
		token, err := spotify.RefreshToken(spotifyClientID, spotifyClientSecret, token.RefreshToken)
		if err != nil {
			slog.Error(err.Error())
			return nil, err
		}
		return token, nil
	}
	//this means something went wrong re-authenticate
	token, err := spotify.Authenticate(spotifyClientID, spotifyClientSecret)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func doAllDepsInstalled() error {
	toolNames := []string{"yt-dlp", "ffmpeg"}
	var error error
	for _, toolName := range toolNames {
		_, err := exec.LookPath(toolName)
		if err != nil {
			error = fmt.Errorf("failed to find %v in the path have u installed it", toolName)
			break
		}
	}
	return error
}

func Execute(version string, spotifyClientID string, spotifyClientSecret string) error {
	cmd := newRootCmd(version, spotifyClientID, spotifyClientSecret)
	userHomeDir, _ := os.UserHomeDir()
	defaultDebugDir := filepath.Join(userHomeDir, ".clispot", "logs")
	cmd.Flags().StringP("debug-dir", "d", defaultDebugDir, "a path to store app logs")
	if err := cmd.Execute(); err != nil {
		return fmt.Errorf("error executing root command: %w", err)
	}
	return nil
}
