package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
	logSetup "github.com/kumneger0/clispot/internal/logger"
	"github.com/kumneger0/clispot/internal/spotify"
	"github.com/kumneger0/clispot/internal/types"
	"github.com/kumneger0/clispot/internal/ui"
)

var version = ""

type Token struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	ExpiresAt    int64  `json:"expires_at,omitempty"`
}

func main() {
	logger := logSetup.Init()
	defer logger.Close()

	slog.Info("starting the application")

	err := godotenv.Load()
	if err != nil {
		slog.Info(err.Error())
		log.Fatal("Error loading .env file")
	}

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

	model := ui.Model{
		Playlist:              playlists,
		UserTokenInfo:         token,
		SelectedPlayListItems: playlistItems,
		FocusedOn:             ui.SideView,
		Search:                input,
		MusicQueueList:        musicQueueList,
	}

	Program := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())

	_, err = Program.Run()
	if err != nil {
		slog.Error(err.Error())
		log.Fatal(err)
	}
}
