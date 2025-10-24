package ui

import (
	"fmt"
	"io"
	"os"

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
	UserTokenInfo         *types.UserTokenInfo
	SelectedPlayListItems list.Model
	FocusedOn             FocusedOn
	PlayerProcess         *os.Process
	SelectedTrackID       *string
	Height                int
	Width                 int
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width - 4
		m.Height = msg.Height - 4
		return m, nil
	case tea.KeyMsg:
		model, cmd := m.handleKeyPress(msg)
		m = model
		cmds = append(cmds, cmd)
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
			//TODO: show some kind of error message
			return m, nil
		}
		trackName := selectedMusic.Track.Name
		albumName := selectedMusic.Track.Album.Name
		var artistNames []string
		for _, artist := range selectedMusic.Track.Artists {
			artistNames = append(artistNames, artist.Name)
		}

		playerProcess := m.PlayerProcess

		if playerProcess != nil {
			_ = playerProcess.Kill()
			//TODO:Show error message
		}

		process, err := youtube.SearchAndDownloadMusic(trackName, albumName, artistNames)
		if err != nil {
			//TODO: implement some kind of way to show the error message
			return m, nil
		}

		m.PlayerProcess = process
		m.SelectedTrackID = &selectedMusic.Track.ID
		return m, nil
	}
	selectedItem, ok := m.Playlist.SelectedItem().(types.SpotifyPlaylist)
	if !ok {
		return m, nil
	}
	playlistItems, err := spotify.GetPlaylistItems(selectedItem.ID, m.UserTokenInfo.AccessToken)

	if err != nil {
		//TODO:  log error using slog
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
	m.Playlist.Title = "Playlist"
	m.SelectedPlayListItems.Title = "Tracks"

	m.Playlist.SetShowFilter(false)
	m.Playlist.SetShowPagination(false)
	m.Playlist.SetShowHelp(false)
	m.Playlist.SetShowStatusBar(false)

	dimensions := calculateLayoutDimensions(&m)
	updateListDimensions(&m, dimensions)

	m.SelectedPlayListItems.SetShowFilter(false)
	m.SelectedPlayListItems.SetShowPagination(false)
	m.SelectedPlayListItems.SetShowHelp(false)
	m.SelectedPlayListItems.SetShowStatusBar(false)
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
	}
}

type layoutDimensions struct {
	sidebarWidth  int
	mainWidth     int
	contentHeight int
	inputHeight   int
}

func calculateLayoutDimensions(m *Model) layoutDimensions {
	sidebarWidth := m.Width * 30 / 100
	return layoutDimensions{
		sidebarWidth:  sidebarWidth,
		mainWidth:     (m.Width - sidebarWidth) * 90 / 100,
		contentHeight: m.Height * 90 / 100,
		inputHeight:   m.Height - (m.Height * 90 / 100),
	}
}

func updateListDimensions(m *Model, d layoutDimensions) {
	listHeight := d.contentHeight - 4
	m.Playlist.SetHeight(listHeight)
	m.Playlist.SetWidth(d.sidebarWidth)
}
