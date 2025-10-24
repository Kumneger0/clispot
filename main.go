package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
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
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	token, err := spotify.ReadUserCredentials()

	if err != nil {
		spotify.Authenticate()
	}

	if token.ExpiresAt < time.Now().Unix() && token.RefreshToken != "" {
		token, err = spotify.RefreshToken(token.RefreshToken)
		if err != nil {
			//TODO:show the error for the user
			os.Exit(1)
		}
	}

	featuredPlaylist, err := spotify.GetFeaturedPlaylist(token.AccessToken)

	playListToRender := func() []types.SpotifyPlaylist {
		if err == nil && featuredPlaylist != nil {
			return featuredPlaylist.Playlists.Items
		}
		userPlayList, err := spotify.GetUserPlaylists(token.AccessToken)
		if err != nil || userPlayList == nil {
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

	model := ui.Model{
		Playlist:              playlists,
		UserTokenInfo:         token,
		SelectedPlayListItems: list.New([]list.Item{}, ui.CustomDelegate{}, 10, 20),
		FocusedOn:             ui.SideView,
	}

	Program := tea.NewProgram(model, tea.WithAltScreen())

	_, err = Program.Run()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
