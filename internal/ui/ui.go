package ui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kumneger0/clispot/internal/spotify"
	"github.com/kumneger0/clispot/internal/types"
	"github.com/kumneger0/clispot/internal/youtube"
)

type FocusedOn string

const (
	SideView FocusedOn = "SIDE_VIEW"
	MainView FocusedOn = "MAIN_VIEW"
)

type Model struct {
	Playlist              list.Model
	UserTokenInfo         types.UserTokenInfo
	SelectedPlayListItems list.Model
	FocusedOn             FocusedOn
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		model, cmd := m.handleKeyPress(msg)
		m = model
		cmds = append(cmds, cmd)
	case types.PlayMusicMsg:
		//TODO: implement music playing
	default:
		//TODO: do something here if no key matched
	}

	if m.FocusedOn == SideView {
		model, cmd := m.Playlist.Update(msg)
		m.Playlist = model
		cmds = append(cmds, cmd)
	}

	if m.FocusedOn == MainView {
		model, cmd := m.SelectedPlayListItems.Update(msg)
		m.Playlist = model
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "tab":
		if m.FocusedOn == MainView {
			m.FocusedOn = SideView
		}
		if m.FocusedOn == SideView {
			m.FocusedOn = MainView
		}
	case "enter":
		return m.handleEnterKey()
	}

	return m, nil
}

func (m Model) handleEnterKey() (Model, tea.Cmd) {
	if m.FocusedOn == MainView {
		// TODO: this means the user clicked on song so
		// find the song on youtube and start streaming
		selectedMusic, ok := m.SelectedPlayListItems.SelectedItem().(types.PlaylistTrackObject)
		if !ok {
			fmt.Println("uff we have fucked up")
			return m, nil
		}
		trackName := selectedMusic.Track.Name
		albumName := selectedMusic.Track.Album.Name
		var artistNames []string
		for _, artist := range selectedMusic.Track.Artists {
			artistNames = append(artistNames, artist.Name)
		}

		musicPath, err := youtube.SearchAndDownloadMusic(trackName, albumName, artistNames)
		if err != nil {
			//TODO: implement some kind of way to show the error message
			return m, nil
		}

		cmd := func() tea.Msg {
			return types.PlayMusicMsg{
				MusicPath: musicPath,
			}
		}
		return m, cmd
	}
	selectedItem, ok := m.Playlist.SelectedItem().(types.SpotifyPlaylist)
	if !ok {
		return m, nil
	}
	playlistItems, err := spotify.GetPlaylistItems(selectedItem.ID, m.UserTokenInfo.AccessToken)

	if err != nil {
		//TODO:  log error using slog
		fmt.Println("err", err)
		return m, nil
	}

	var playListItemSongs []list.Item

	for _, item := range playlistItems.Items {
		playListItemSongs = append(playListItemSongs, item)
	}

	cmd := m.SelectedPlayListItems.SetItems(playListItemSongs)
	return m, cmd
}

func (m Model) View() string {
	playlistView := m.Playlist.View()
	selectedPlaylistMusicView := m.SelectedPlayListItems.View()
	combinedView := lipgloss.JoinHorizontal(lipgloss.Top, playlistView, selectedPlaylistMusicView)
	return combinedView
}

type CustomDelegate struct {
	list.DefaultDelegate
	*Model
}

func (d CustomDelegate) Height() int {
	return 1
}

func (d CustomDelegate) Spacing() int {
	return 0
}

func (d CustomDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}
func (d CustomDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	var title string
	switch item := item.(type) {
	case types.SpotifyPlaylist, types.PlaylistTrackObject:
		title = item.FilterValue()
		str := lipgloss.NewStyle().Width(50).Render(title)
		if index == m.Index() {
			fmt.Fprint(w, selectedStyle.Render(" "+str+" "))
		} else {
			fmt.Fprint(w, normalStyle.Render(" "+str+" "))
		}
	default:
		fmt.Println("we fucked up")
	}
}
