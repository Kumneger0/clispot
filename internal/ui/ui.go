package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/prop"
	"github.com/kumneger0/clispot/internal/spotify"
	"github.com/kumneger0/clispot/internal/types"
	"github.com/kumneger0/clispot/internal/youtube"
)

type FocusedOn string

const (
	SideView             FocusedOn = "SIDE_VIEW"
	MainView             FocusedOn = "MAIN_VIEW"
	Player               FocusedOn = "PLAYER"
	SearchBar            FocusedOn = "SEARCH_BAR"
	QueueList            FocusedOn = "QUEUE_LIST"
	SearchResult         FocusedOn = "SEARCH_RESULT"
	SearchResultTrack    FocusedOn = "SEARCH_RESULT_TRACK"
	SearchResultArtist   FocusedOn = "SEARCH_RESULT_ARTIST"
	SearchResultPlaylist FocusedOn = "SEARCH_RESULT_PLAYLIST"
)

type SpotifySearchResult struct {
	Tracks, Artists, Albums, Playlists list.Model
}

type Model struct {
	Playlist                                list.Model
	UserTokenInfo                           *types.UserTokenInfo
	SelectedPlayListItems                   list.Model
	LyricsView                              viewport.Model
	FocusedOn                               FocusedOn
	PlayerProcess                           *youtube.Player
	SelectedTrack, NextTrack, PreviousTrack *types.PlaylistTrackObject
	PlayedSeconds                           float64
	Height                                  int
	Width                                   int
	Search                                  textinput.Model
	MusicQueueList                          list.Model
	DBusConn                                *Instance
	IsSearchLoading                         bool
	SearchResult                            *SpotifySearchResult
}

type Instance struct {
	Props *prop.Properties
	Conn  *dbus.Conn
}

func (m Model) Init() tea.Cmd {
	var cmd tea.Cmd
	if m.UserTokenInfo != nil {
		cmd = func() tea.Msg {
			followedArtist, err := spotify.GetFollowedArtist(m.UserTokenInfo.AccessToken)
			if err != nil {
				return nil
			}
			return followedArtist
		}
	}
	return cmd
}

func (m Model) View() string {
	m.Playlist.Title = "Playlist"
	m.SelectedPlayListItems.Title = "Tracks"
	m.MusicQueueList.Title = "Queue"
	removeListDefaults(&m.Playlist)
	removeListDefaults(&m.SelectedPlayListItems)
	removeListDefaults(&m.MusicQueueList)

	dimensions := calculateLayoutDimensions(&m)
	updateListDimensions(&m, dimensions)

	playlistView := getListStyle(&m, dimensions.sidebarWidth, dimensions.contentHeight, SideView).Render(m.Playlist.View())

	searchBar := renderSearchBar(&m, dimensions.mainWidth)
	var mainView string
	if m.IsSearchLoading {
		mainView = getMainStyle(dimensions.mainWidth, dimensions.contentHeight, &m).Render(lipgloss.JoinVertical(lipgloss.Top, searchBar, "loading...."))
	} else if m.SearchResult != nil {
		trackView := getListStyle(&m, dimensions.sidebarWidth, dimensions.contentHeight, SearchResultTrack).Render(m.SearchResult.Tracks.View())
		artistView := getListStyle(&m, dimensions.sidebarWidth, dimensions.contentHeight, SearchResultArtist).Render(m.SearchResult.Artists.View())
		playlistView := getListStyle(&m, dimensions.sidebarWidth, dimensions.contentHeight, SearchResultPlaylist).Render(m.SearchResult.Playlists.View())
		searchResultView := lipgloss.JoinVertical(lipgloss.Top, searchBar, lipgloss.JoinVertical(lipgloss.Top, "Search Result", lipgloss.JoinHorizontal(lipgloss.Top, trackView, artistView, playlistView)))
		mainView = getMainStyle(dimensions.mainWidth, dimensions.contentHeight, &m).Render(searchResultView)
	} else {
		mainView = getMainStyle(dimensions.mainWidth, dimensions.contentHeight, &m).
			Render(lipgloss.JoinVertical(lipgloss.Top, searchBar, m.SelectedPlayListItems.View()))
	}

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
		playedSeconds := int(m.PlayedSeconds)
		currentPosition := time.Second * time.Duration(playedSeconds)
		total := time.Duration(m.SelectedTrack.Track.DurationMS) * time.Millisecond
		playingView = renderNowPlaying(m.SelectedTrack.Track.Name, artistName, currentPosition, total)
	}

	controls := renderPlayerControls()
	playingCombined := strings.TrimSpace(playingView) + "\n\n" + controls

	playing := getPlayerStyles(&m, dimensions).Foreground(lipgloss.Color("21")).Render(playingCombined)

	queueList := getListStyle(&m, dimensions.contentHeight, dimensions.sidebarWidth, QueueList).Render(m.MusicQueueList.View())

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
	sidebarWidth := m.Width * 20 / 100
	inputHeight := min(max(m.Height*6/100, 3), 8)

	mainCenterArea := (m.Width - (sidebarWidth * 2))

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

func removeListDefaults(listToRemoveDefaults *list.Model) {
	if listToRemoveDefaults != nil {
		listToRemoveDefaults.SetShowFilter(false)
		listToRemoveDefaults.SetShowPagination(false)
		listToRemoveDefaults.SetShowHelp(false)
		listToRemoveDefaults.SetShowStatusBar(false)
	}
}
