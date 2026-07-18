package ui

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/prop"
	musicpb "github.com/kumneger0/clispot/gen"
	"github.com/kumneger0/clispot/internal/types"
	"github.com/kumneger0/clispot/internal/youtube"
	"go.dalton.dog/bubbleup"
)

type FocusedOn string

const (
	SideView     FocusedOn = "SIDE_VIEW"
	MainView     FocusedOn = "MAIN_VIEW"
	Player       FocusedOn = "PLAYER"
	SearchBar    FocusedOn = "SEARCH_BAR"
	QueueList    FocusedOn = "QUEUE_LIST"
	SearchResult FocusedOn = "SEARCH_RESULT"
)

type MainViewMode string

const (
	SearchResultMode MainViewMode = "SEARCH_RESULT_MODE"
	//currently im showing the search result in main area which is the center one
	//let's say the user searches for a song or playlist and sees the result and he chose the first result
	//at this time the previous are gone b/c i was sharing  this main new to show items in playlist and the search result
	// so by adding this MainViewMode we can switch b/c modes so that we keep the result in memory
	// meaning we can switch b/n search result and normal mode
	NormalMode   MainViewMode = "NORMAL_MODE"
	LyricsMode   MainViewMode = "LYRICS_MODE"
	HomePageMode MainViewMode = "HOME_PAGE_MODE"
)

type HomePageViewMode int

const (
	HomePageSectionView HomePageViewMode = iota
	HomePageContentView
)

type SpotifySearchResult struct {
	Tracks, Artists, Albums, Playlists list.Model
}

type SelectedTrack struct {
	isLiked bool
	Track   *types.PlaylistTrackObject
}

type MusicQueueList struct {
	list.Model
	PaginationInfo *types.PaginationInfo
}

type Model struct {
	BreadcrumbItems       []types.Breadcrumb
	SideBarList           list.Model
	Alert                 bubbleup.AlertModel
	SelectedPlayListItems list.Model
	LyricsView            viewport.Model
	FocusedOn             FocusedOn
	MainViewMode
	PlayerProcess       *types.Player
	SelectedTrack       *SelectedTrack
	PlayedSeconds       float64
	Height              int
	Width               int
	LibraryWidth        int
	MainViewWidth       int
	PlayerSectionHeight int
	Search              textinput.Model
	MusicQueueList      *MusicQueueList
	YtMusicClient       musicpb.MusicServiceClient
	DBusConn            *Instance
	//actually i need this b/c if user searches and selects playlist or artist
	//at that time when he selects artist or playlist the search were hidden from mainView
	//so that if search again we can show the previous result by comparing the query
	// TODO: find a better way than this looks very ugly
	SearchQuery                              string
	IsSearchLoading, IsLyricsServerInstalled bool
	// SearchResult                             *SpotifySearchResult
	SearchResult     list.Model
	PaginationInfo   *types.PaginationInfo
	IsOnPagination   bool
	CoreDepsPath     *youtube.CoreDepsPath
	HomePageData     *musicpb.GetHomePageResponse
	HomePageList     list.Model
	HomePageViewMode HomePageViewMode
}

type Instance struct {
	Props *prop.Properties
	Conn  *dbus.Conn
}

type SafeModel struct {
	Mu sync.RWMutex
	*Model
}

func (m Model) Init() tea.Cmd {
	cmd := func() tea.Msg {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		followedArtist, err := m.YtMusicClient.GetFollowedArtists(ctx, &musicpb.GetFollowedArtistsRequest{})
		if err != nil {
			return nil
		}
		return followedArtist
	}
	homePageFeed := func() tea.Msg {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		homePage, err := m.YtMusicClient.GetHomePage(ctx, &musicpb.GetHomePageRequest{})
		if err != nil {
			slog.Error(err.Error())
			return types.HomePageResponseMsg{
				Response: nil,
				Err:      err,
			}
		}
		return types.HomePageResponseMsg{
			Response: homePage,
			Err:      nil,
		}
	}
	return tea.Batch(cmd, m.Alert.Init(), homePageFeed)
}

func (m Model) View() string {
	m.SideBarList.Title = "Youtube Music tui"
	m.SelectedPlayListItems.Title = "Tracks"
	m.MusicQueueList.Model.Title = "Queue"
	removeListDefaults(&m.SideBarList)
	removeListDefaults(&m.SelectedPlayListItems)
	removeListDefaults(&m.SearchResult)
	m.SearchResult.SetShowTitle(false)
	if m.MusicQueueList != nil {
		removeListDefaults(&m.MusicQueueList.Model)
	}
	dimensions := calculateLayoutDimensions(&m)
	sideBarView := getStyle(&m, dimensions.sidebarWidth, dimensions.contentHeight, SideView).Render(m.SideBarList.View())
	searchBar := renderSearchBar(&m, dimensions.mainWidth)
	var mainView string
	if m.IsSearchLoading {
		loadingText := dimmerStyle.Render("  ⟳ Loading...")
		mainView = getStyle(&m, dimensions.contentHeight, dimensions.mainWidth, MainView).Render(
			lipgloss.JoinVertical(lipgloss.Top, searchBar, loadingText),
		)
	} else if m.MainViewMode == SearchResultMode {
		searchView := getStyle(&m, dimensions.contentHeight-10, dimensions.mainWidth-10, SearchResult).Render(m.SearchResult.View())
		resultHeader := titleStyle.Render("  Search Results")
		searchResultView := lipgloss.JoinVertical(lipgloss.Top,
			searchBar,
			resultHeader,
			lipgloss.JoinHorizontal(lipgloss.Top, searchView),
		)
		mainView = getStyle(&m, dimensions.contentHeight, dimensions.mainWidth, MainView).Render(searchResultView)
	} else if m.MainViewMode == LyricsMode {
		mainView = getStyle(&m, dimensions.contentHeight, dimensions.mainWidth, MainView).Render(
			lipgloss.JoinVertical(lipgloss.Top, searchBar, m.LyricsView.View()),
		)
	} else if m.MainViewMode == HomePageMode {
		homePageContent := renderHomePage(&m)
		mainView = getStyle(&m, dimensions.contentHeight, dimensions.mainWidth, MainView).Render(
			lipgloss.JoinVertical(lipgloss.Top, searchBar, homePageContent),
		)
	} else {
		mainView = getStyle(&m, dimensions.contentHeight, dimensions.mainWidth, MainView).
			Render(lipgloss.JoinVertical(lipgloss.Top, searchBar, m.SelectedPlayListItems.View()))
	}

	var playingView string

	if m.SelectedTrack != nil && m.SelectedTrack.Track != nil {
		playedSeconds := int(m.PlayedSeconds)
		currentPosition := time.Second * time.Duration(playedSeconds)
		total := time.Duration(m.SelectedTrack.Track.Track.DurationMS) * time.Millisecond
		playingView = renderNowPlaying(&m, currentPosition, total)
	}

	controls := renderPlayerControls(m.IsLyricsServerInstalled)
	playingCombined := strings.TrimSpace(playingView) + "\n" + controls

	playing := getPlayerStyles(&m, dimensions).
		Foreground(playerFg).
		Render(playingCombined)

	queueList := getStyle(&m, dimensions.contentHeight, dimensions.sidebarWidth, QueueList).Render(m.MusicQueueList.View())

	combinedView := lipgloss.JoinVertical(lipgloss.Top,
		lipgloss.JoinHorizontal(lipgloss.Top, sideBarView, mainView, queueList),
		playing,
	)
	return m.Alert.Render(combinedView)
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
	sidebarWidth := m.Width * 22 / 100
	inputHeight := min(max(m.Height*10/100, 2), 3)
	mainCenterArea := (m.Width - (sidebarWidth * 2))

	return layoutDimensions{
		sidebarWidth:  sidebarWidth,
		mainWidth:     mainCenterArea,
		contentHeight: m.Height * 90 / 100,
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
