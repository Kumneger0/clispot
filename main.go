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
	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/prop"
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

func newProp(value interface{}, cb func(*prop.Change) *dbus.Error) *prop.Prop {
	return &prop.Prop{
		Value:    value,
		Writable: true,
		Emit:     prop.EmitTrue,
		Callback: cb,
	}
}

var player = map[string]*prop.Prop{
	"PlaybackStatus": newProp("paused", nil),
	"Rate":           newProp(1.0, nil),
	"Metadata":       newProp(map[string]interface{}{}, nil),
	"Volume":         newProp(float64(100), nil),
	"Position":       newProp(int64(0), nil),
	"MinimumRate":    newProp(1.0, nil),
	"MaximumRate":    newProp(1.0, nil),
	"CanGoNext":      newProp(true, nil),
	"CanGoPrevious":  newProp(true, nil),
	"CanPlay":        newProp(true, nil),
	"CanPause":       newProp(true, nil),
	"CanSeek":        newProp(false, nil),
	"CanControl":     newProp(true, nil),
}

var mediaPlayer2 = map[string]*prop.Prop{
	"CanQuit":             newProp(false, nil),
	"CanRaise":            newProp(false, nil),
	"HasTrackList":        newProp(false, nil),
	"Identity":            newProp("clispot", nil),
	"SupportedUriSchemes": newProp([]string{}, nil),
	"SupportedMimeTypes":  newProp([]string{}, nil),
}

type MediaPlayer2 struct {
	Props       *prop.Properties
	messageChan *chan types.DBusMessage
}

func (m *MediaPlayer2) Pause() *dbus.Error {
	return nil
}
func (m *MediaPlayer2) Next() *dbus.Error {
	*m.messageChan <- types.DBusMessage{
		MessageType: types.NextTrack,
	}
	return nil
}

func (m *MediaPlayer2) Previous() *dbus.Error {
	*m.messageChan <- types.DBusMessage{
		MessageType: types.PreviousTrack,
	}
	return nil
}

func (m *MediaPlayer2) Play() *dbus.Error {
	return nil
}

func (m *MediaPlayer2) PlayPause() *dbus.Error {
	*m.messageChan <- types.DBusMessage{
		MessageType: types.PlayPause,
	}
	return nil
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
		fmt.Println("we are here")
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

	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	reply, err := conn.RequestName("org.mpris.MediaPlayer2.clispot", dbus.NameFlagReplaceExisting)
	if err != nil {
		log.Fatalln(err)
	}
	if reply != dbus.RequestNameReplyPrimaryOwner {
		log.Fatalln("Name already taken")
	}

	ins := &ui.Instance{
		Conn: conn,
	}

	ins.Props, err = prop.Export(
		conn,
		"/org/mpris/MediaPlayer2",
		map[string]map[string]*prop.Prop{
			"org.mpris.MediaPlayer2":        mediaPlayer2,
			"org.mpris.MediaPlayer2.Player": player,
		},
	)
	if err != nil {
		log.Fatalln("prop.Export error:", err)
	}

	messageChan := make(chan types.DBusMessage, 10)

	mp2 := &MediaPlayer2{
		Props:       ins.Props,
		messageChan: &messageChan,
	}

	if err := conn.Export(mp2, "/org/mpris/MediaPlayer2", "org.mpris.MediaPlayer2.Player"); err != nil {
		log.Fatalln("conn.Export error:", err)
	}

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
		for v := range messageChan {
			Program.Send(v)
		}
	}()

	_, err = Program.Run()
	if err != nil {
		slog.Error(err.Error())
		log.Fatal(err)
	}
}
