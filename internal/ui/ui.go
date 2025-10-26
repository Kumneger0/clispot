package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
	"github.com/kumneger0/clispot/internal/types"
)

type FocusedOn string

const (
	SideView  FocusedOn = "SIDE_VIEW"
	MainView  FocusedOn = "MAIN_VIEW"
	Player    FocusedOn = "PLAYER"
	SearchBar FocusedOn = "SEARCH_BAR"
	QueueList FocusedOn = "QUEUE_LIST"
)

type Model struct {
	Playlist                                list.Model
	UserTokenInfo                           *types.UserTokenInfo
	SelectedPlayListItems                   list.Model
	LyricsView                              viewport.Model
	FocusedOn                               FocusedOn
	PlayerProcess                           *os.Process
	SelectedTrack, NextTrack, PreviousTrack *types.PlaylistTrackObject
	Height                                  int
	Width                                   int
	Search                                  textinput.Model
	MusicQueueList                          list.Model
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) View() string {
	m.Playlist.Title = "Playlist"
	m.SelectedPlayListItems.Title = "Tracks"
	m.MusicQueueList.Title = "Queue"
	removeListDefaults(m.Playlist)
	removeListDefaults(m.SelectedPlayListItems)
	removeListDefaults(m.MusicQueueList)

	dimensions := calculateLayoutDimensions(&m)
	updateListDimensions(&m, dimensions)

	playlistView := getSideBarStyles(dimensions.sidebarWidth, dimensions.contentHeight, &m).Render(m.Playlist.View())

	searchBar := renderSearchBar(&m, dimensions.mainWidth)

	mainView := getMainStyle(dimensions.mainWidth, dimensions.contentHeight, &m).Render(lipgloss.JoinVertical(lipgloss.Top, searchBar, m.LyricsView.View()))

	var playingView string

	if m.SelectedTrack != nil {
		var stringBuilder strings.Builder
		stringBuilder.WriteString(m.SelectedTrack.Track.Name)
		stringBuilder.WriteString(" ")
		var artistNames []string
		for _, artist := range m.SelectedTrack.Track.Artists {
			artistNames = append(artistNames, artist.Name)
		}
		artistName := strings.Join(artistNames, ",")
		stringBuilder.WriteString(artistName)
		currentPosition := time.Second * 30
		total := time.Duration(m.SelectedTrack.Track.DurationMs) * time.Millisecond

		playingView = renderNowPlaying(m.SelectedTrack.Track.Name, artistName, currentPosition, total)
	}

	controls := renderPlayerControls()
	playingCombined := strings.TrimSpace(playingView) + "\n\n" + controls

	playing := getPlayerStyles(&m, dimensions.inputHeight).Foreground(lipgloss.Color("21")).Render(playingCombined)

	queueList := getQueueListStyle(&m, dimensions.contentHeight, dimensions.sidebarWidth).Render(m.MusicQueueList.View())

	combinedView := lipgloss.JoinVertical(lipgloss.Top,
		lipgloss.JoinHorizontal(lipgloss.Top, playlistView, mainView, queueList),
		playing,
	)
	return combinedView
}

func formatTime(d time.Duration) string {
	if d < 0 {
		d = 0
	}

	totalSeconds := int(d.Seconds())
	minutes := totalSeconds / 60
	seconds := totalSeconds % 60

	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

type layoutDimensions struct {
	sidebarWidth  int
	mainWidth     int
	contentHeight int
	inputHeight   int
}

func calculateLayoutDimensions(m *Model) layoutDimensions {
	sidebarWidth := m.Width * 15 / 100
	inputHeight := min(max(m.Height*6/100, 3), 8)

	mainCenterArea := (m.Width - (sidebarWidth * 2)) * 95 / 100

	//main area is basically total width minus left sidebar minus right sidebar and

	return layoutDimensions{
		sidebarWidth:  sidebarWidth,
		mainWidth:     mainCenterArea,
		contentHeight: m.Height * 85 / 100,
		inputHeight:   inputHeight,
	}
}

func updateListDimensions(m *Model, d layoutDimensions) {
	listHeight := d.contentHeight - 4
	m.Playlist.SetHeight(listHeight)
	m.Playlist.SetWidth(d.sidebarWidth)
	m.SelectedPlayListItems.SetHeight(listHeight)
	m.SelectedPlayListItems.SetWidth(d.sidebarWidth)
}

func getTerminalWidth() int {
	if !term.IsTerminal(os.Stdin.Fd()) {
		fmt.Println("Not running in a terminal.")
		return 0
	}

	width, _, err := term.GetSize(os.Stdin.Fd())
	if err != nil {
		fmt.Printf("Error getting terminal size: %v\n", err)
		return 0
	}

	return width
}

func removeListDefaults(listToRemoveDefaults list.Model) {
	listToRemoveDefaults.SetShowFilter(false)
	listToRemoveDefaults.SetShowPagination(false)
	listToRemoveDefaults.SetShowHelp(false)
	listToRemoveDefaults.SetShowStatusBar(false)
}
