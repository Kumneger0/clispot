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

type MainViewMode string

const (
	SearchResultMode MainViewMode = "SearchResultMode"
	//currently im showing the search result in main area which is the center one
	//let's say the user searches for a song or playlist and sees the result and he chose the first result
	//at this time the previous are gone b/c i was sharing  this main new to show items in playlist and the search result
	// so by adding this MainViewMode we can switch b/c modes so that we keep the result in memory
	// meaning we can switch b/n search result and normal mode
	NormalMode MainViewMode = "Normal Mode"
)

type SpotifySearchResult struct {
	Tracks, Artists, Albums, Playlists list.Model
}

type UserArguments struct {
	DebugPath string
}

type Model struct {
	Playlist              list.Model
	SelectedPlayListItems list.Model
	LyricsView            viewport.Model
	FocusedOn             FocusedOn
	MainViewMode
	PlayerProcess                           *youtube.Player
	SelectedTrack, NextTrack, PreviousTrack *types.PlaylistTrackObject
	PlayedSeconds                           float64
	Height                                  int
	Width                                   int
	Search                                  textinput.Model
	MusicQueueList                          list.Model
	DBusConn                                *Instance
	//actually i need this b/c if user searches and selects playlist or artist
	//at that time when he selects artist or playlist the search were hidden from mainView
	//so that if search again we can show the previous result by comparing the query
	// TODO: find a better way than this looks very ugly
	SearchQuery     string
	IsSearchLoading bool
	SearchResult    *SpotifySearchResult
	GetUserToken    func() *types.UserTokenInfo
	*UserArguments
}

type Instance struct {
	Props *prop.Properties
	Conn  *dbus.Conn
}

func (m Model) Init() tea.Cmd {
	var cmd tea.Cmd
	userTokenInfo := m.GetUserToken()
	if userTokenInfo != nil {
		cmd = func() tea.Msg {
			followedArtist, err := spotify.GetFollowedArtist(userTokenInfo.AccessToken)
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

	playlistView := getStyle(&m, dimensions.sidebarWidth, dimensions.contentHeight, SideView).Render(m.Playlist.View())

	searchBar := renderSearchBar(&m, dimensions.mainWidth)
	var mainView string
	if m.IsSearchLoading {
		mainView = getStyle(&m, dimensions.contentHeight, dimensions.mainWidth, MainView).Render(lipgloss.JoinVertical(lipgloss.Top, searchBar, "loading...."))
	} else if m.MainViewMode == SearchResultMode {
		trackView := getStyle(&m, dimensions.sidebarWidth, dimensions.contentHeight, SearchResultTrack).Render(m.SearchResult.Tracks.View())
		artistView := getStyle(&m, dimensions.sidebarWidth, dimensions.contentHeight, SearchResultArtist).Render(m.SearchResult.Artists.View())
		playlistView := getStyle(&m, dimensions.mainWidth/4, dimensions.contentHeight, SearchResultPlaylist).Render(m.SearchResult.Playlists.View())
		searchResultView := lipgloss.JoinVertical(lipgloss.Top, searchBar, lipgloss.JoinVertical(lipgloss.Top, "Search Result", lipgloss.JoinHorizontal(lipgloss.Top, trackView, artistView, playlistView)))
		mainView = getStyle(&m, dimensions.contentHeight, dimensions.mainWidth, MainView).Render(searchResultView)
	} else {
		mainView = getStyle(&m, dimensions.contentHeight, dimensions.mainWidth, MainView).
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

	queueList := getStyle(&m, dimensions.contentHeight, dimensions.sidebarWidth, QueueList).Render(m.MusicQueueList.View())

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

func removeListDefaults(listToRemoveDefaults *list.Model) {
	if listToRemoveDefaults != nil {
		listToRemoveDefaults.SetShowFilter(false)
		listToRemoveDefaults.SetShowPagination(false)
		listToRemoveDefaults.SetShowHelp(false)
		listToRemoveDefaults.SetShowStatusBar(false)
	}
}
