package ui

import (
	"log/slog"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/godbus/dbus/v5"
	"github.com/kumneger0/clispot/internal/command"
	"github.com/kumneger0/clispot/internal/lyrics"
	"github.com/kumneger0/clispot/internal/notification"
	"github.com/kumneger0/clispot/internal/spotify"
	"github.com/kumneger0/clispot/internal/types"
	"github.com/kumneger0/clispot/internal/youtube"
	"go.dalton.dog/bubbleup"
)

var testLyrics = `working hard to get the lyrics for you`

type MusicMetadata struct {
	artistName string
	title      string
	length     int64
}

func getMusicMetadata(music MusicMetadata) map[string]interface{} {
	var metadata = map[string]interface{}{
		"mpris:trackid": "/org/mpris/MediaPlayer2/" + music.title,
		"mpris:length":  music.length,
		"xesam:title":   music.title,
		"xesam:artist":  music.artistName,
	}
	return metadata
}

func (m Model) getSearchResultModel(searchResponse *types.SearchResponse) (Model, tea.Cmd) {
	var tracks []list.Item
	for _, value := range searchResponse.Tracks.Items {
		track := types.PlaylistTrackObject{
			AddedAt:        "",
			AddedBy:        nil,
			IsLocal:        false,
			Track:          value,
			IsItFromQueue:  false,
			IsItFromSearch: true,
		}
		tracks = append(tracks, track)
	}
	var artist []list.Item
	for _, value := range searchResponse.Artists.Items {
		value.IsItFromSearch = true
		artist = append(artist, value)
	}
	var playlist []list.Item
	for _, value := range searchResponse.Playlists.Items {
		value.IsItFromSearch = true
		playlist = append(playlist, value)
	}

	if m.SearchResult != nil {
		m.SearchResult.Tracks.SetItems(tracks)
		m.SearchResult.Artists.SetItems(artist)
		m.SearchResult.Playlists.SetItems(playlist)
		return m, nil
	}

	searchResult := SpotifySearchResult{
		Tracks:    list.New(tracks, CustomDelegate{Model: &m}, 10, 20),
		Artists:   list.New(artist, CustomDelegate{Model: &m}, 10, 20),
		Playlists: list.New(playlist, CustomDelegate{Model: &m}, 10, 20),
	}
	m.SearchResult = &searchResult
	return m, nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case types.CheckUserSavedTrackResponseMsg:
		if msg.Err != nil {
			slog.Error(msg.Err.Error())
			return m, nil
		}
		m.SelectedTrack.isLiked = msg.Saved
		return m, nil
	case types.SearchingMsg:
		m.IsSearchLoading = true
	case *lyrics.Response:
		m.LyricsView.SetContent(*msg.Lyrics)
		m.IsSearchLoading = false
		m.MainViewMode = LyricsMode
	case lyrics.Response:
		m.LyricsView.SetContent(*msg.Lyrics)
		m.IsSearchLoading = false
		m.MainViewMode = LyricsMode
	case types.SpotifySearchResultMsg:
		var alertCmd tea.Cmd
		if msg.Err != nil {
			alertCmd = m.Alert.NewAlertCmd(bubbleup.ErrorKey, msg.Err.Error())
			cmds = append(cmds, alertCmd)
		}
		if msg.Result != nil {
			m.FocusedOn = SearchResult
			m.MainViewMode = SearchResultMode
			model, cmd := m.getSearchResultModel(msg.Result)
			m = model
			removeListDefaults(&m.SearchResult.Artists)
			removeListDefaults(&m.SearchResult.Playlists)
			removeListDefaults(&m.SearchResult.Tracks)

			if m.SearchResult != nil {
				m.SearchResult.Tracks.Title = "Tracks"
				m.SearchResult.Artists.Title = "artist"
				m.SearchResult.Playlists.Title = "playlist"
			}
			cmds = append(cmds, cmd)
			m.IsSearchLoading = false
		}
	case *types.UserFollowedArtistResponse:
		playlist := m.Playlist.Items()
		for _, artist := range msg.Artists.Items {
			playlist = append(playlist, artist)
		}
		cmd := m.Playlist.SetItems(playlist)
		cmds = append(cmds, cmd)
	case types.DBusMessage:
		model, cmd := m.handleDbusMessage(msg.MessageType, cmds)
		m = model
		cmds = append(cmds, cmd)
	case types.LikeUnlikeTrackMsg:
		if msg.TrackID == m.SelectedTrack.Track.Track.ID {
			m.SelectedTrack.isLiked = msg.Like
		}
	case youtube.ScanFuncArgs:
		var alertCmd tea.Cmd
		if msg.LogType == youtube.WARNING {
			alertCmd = m.Alert.NewAlertCmd(bubbleup.WarnKey, msg.Line)
		}
		if msg.LogType == youtube.ERROR {
			alertCmd = m.Alert.NewAlertCmd(bubbleup.ErrorKey, msg.Line)
			notificationTitle := "YtDlp Error"
			notificationMessage := msg.Line
			notification.Notify(notificationTitle, notificationMessage)
		}
		cmds = append(cmds, alertCmd)
	case types.UpdatePlayedSeconds:
		if m.PlayerProcess == nil {
			return m, nil
		}
		if !m.PlayerProcess.OtoPlayer.IsPlaying() {
			return m, nil
		}

		if m.SelectedTrack == nil || m.SelectedTrack.Track == nil {
			return m, nil
		}

		if msg.TrackID != m.SelectedTrack.Track.Track.ID {
			return m, nil
		}

		m.PlayedSeconds = m.PlayerProcess.ByteCounterReader.CurrentSeconds()

		totalDurationInSeconds := m.SelectedTrack.Track.Track.DurationMS / 1000
		if (float64(totalDurationInSeconds) - (m.PlayedSeconds)) < 4 {
			m.PlayedSeconds = 0
			model, cmd := m.handleMusicChange(true, false)
			m = model
			cmds = append(cmds, cmd)
		}

		trackID := m.SelectedTrack.Track.Track.ID

		cmd := tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
			return types.UpdatePlayedSeconds{
				TrackID: trackID,
			}
		})
		cmds = append(cmds, cmd)

	case tea.WindowSizeMsg:
		m.Width = msg.Width - 4
		m.Height = msg.Height - 4
		dims := calculateLayoutDimensions(&m)
		m.LibraryWidth = dims.sidebarWidth
		m.MainViewWidth = dims.mainWidth
		m.PlayerSectionHeight = dims.inputHeight
		m.LyricsView.Width = dims.mainWidth
		m.LyricsView.Height = dims.contentHeight
		m.LyricsView.SetContent(testLyrics)
		return m, nil
	case types.UpdatePlaylistMsg:
		if msg.Playlist != nil {
			var playListItemSongs []list.Item
			for _, item := range msg.Playlist {
				if msg.ShouldAppendQueue {
					item.IsItFromQueue = true
				}
				playListItemSongs = append(playListItemSongs, *item)
			}
			m.MainViewMode = NormalMode
			m.IsSearchLoading = false
			var currentItems []list.Item
			if msg.ShouldAppendQueue && m.MusicQueueList != nil {
				currentItems = m.MusicQueueList.Items()
			} else {
				currentItems = m.SelectedPlayListItems.Items()
			}
			if msg.ShouldAppend {
				playListItemSongs = append(currentItems, playListItemSongs...)
				m.IsOnPagination = false
			}

			var cmd tea.Cmd
			if msg.ShouldAppendQueue {
				if m.MusicQueueList != nil {
					return m, nil
				}
				cmd = m.MusicQueueList.SetItems(playListItemSongs)
			} else {
				cmd = m.SelectedPlayListItems.SetItems(playListItemSongs)
			}
			if msg.PaginationInfo != nil {
				m.PaginationInfo = msg.PaginationInfo
			} else {
				m.PaginationInfo = nil
			}
			cmds = append(cmds, cmd)
		}
		if msg.Err != nil {
			alertCmd := m.Alert.NewAlertCmd(bubbleup.ErrorKey, msg.Err.Error())
			cmds = append(cmds, alertCmd)
		}
	case tea.KeyMsg:
		model, cmd := m.handleKeyPress(msg)
		m = model
		cmds = append(cmds, cmd)
	case tea.MouseMsg:
		x := msg.X
		y := msg.Y
		if x > m.LibraryWidth && x <= (m.LibraryWidth+m.MainViewWidth) && y <= (m.Height-m.PlayerSectionHeight) {
			if m.MainViewMode == LyricsMode && m.FocusedOn != SearchBar {
				lyricsModel, cmd := m.LyricsView.Update(msg)
				m.LyricsView = lyricsModel
				cmds = append(cmds, cmd)
			}
		}

	default:
		//TODO: do something here if no key matched
	}
	model, cmd := updateFocusedComponent(&m, msg, &cmds)
	m = model
	outAlert, outCmd := m.Alert.Update(msg)
	cmds = append(cmds, outCmd, cmd)
	m.Alert = outAlert.(bubbleup.AlertModel)
	return m, tea.Batch(cmds...)
}

func (m Model) handleDbusMessage(msg types.MessageType, cmds []tea.Cmd) (Model, tea.Cmd) {
	switch msg {
	case types.NextTrack:
		model, cmd := m.handleMusicChange(true, true)
		m = model
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	case types.PreviousTrack:
		model, cmd := m.handleMusicChange(false, true)
		m = model
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	case types.PlayPause:
		model, cmd := m.HandleMusicPausePlay()
		m = model
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

func (m Model) handlePagination(listModel *list.Model, ShouldAppendQueue bool, currentIndex *int) (Model, tea.Cmd) {
	if listModel == nil {
		return m, nil
	}
	if currentIndex == nil {
		index := listModel.GlobalIndex()
		currentIndex = &index
	}
	totalItems := listModel.Items()
	if *currentIndex+5 >= len(totalItems) && m.PaginationInfo != nil && m.PaginationInfo.Next != "" {
		userToken := m.GetUserToken()
		if userToken == nil {
			slog.Error("failed to get user access token")
		} else {
			if m.IsOnPagination {
				return m, nil
			}
			m.IsOnPagination = true
			var paginationInfo *types.PaginationInfo
			if m.FocusedOn == QueueList && m.MusicQueueList != nil && m.MusicQueueList.PaginationInfo != nil {
				paginationInfo = m.MusicQueueList.PaginationInfo
			} else {
				paginationInfo = m.PaginationInfo
			}
			return m, getNextPageItems(&m, paginationInfo, ShouldAppendQueue)
		}
	}
	return m, nil
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "down", "j":
		if m.FocusedOn != MainView && m.FocusedOn != QueueList {
			return m, nil
		}
		listModel := getListItemForMusicToChoose(&m, m.FocusedOn)
		return m.handlePagination(listModel, m.FocusedOn == QueueList, nil)
	case "ctrl+k":
		m.FocusedOn = SearchBar
		return m, m.Search.Focus()
	case "a":
		return m.addMusicToQueue()
	case "r":
		if m.FocusedOn == QueueList && m.MusicQueueList != nil {
			if len(m.MusicQueueList.Model.Items()) > 0 {
				m.MusicQueueList.Model.RemoveItem(m.MusicQueueList.GlobalIndex())
			}
		}
	case "ctrl+l":
		if m.MainViewMode == LyricsMode {
			m.MainViewMode = NormalMode
			return m, nil
		}
		return m.getMusicLyrics(m.SelectedTrack)
	case "l":
		if m.SelectedTrack != nil && m.SelectedTrack.Track != nil {
			userToken := m.GetUserToken()
			if userToken == nil {
				slog.Error("m.GetUserToken is return nil ")
				return m, nil
			}
			cmd := func() tea.Msg {
				var shouldRemove bool
				if m.SelectedTrack.isLiked {
					shouldRemove = true
				} else {
					shouldRemove = false
				}
				err := m.SpotifyClient.SaveRemoveTrackForCurrentUser(userToken.AccessToken, []string{m.SelectedTrack.Track.Track.ID}, shouldRemove)
				if err != nil {
					slog.Error(err.Error())
				}
				likeUnlikeTrackMsg := types.LikeUnlikeTrackMsg{
					TrackID: m.SelectedTrack.Track.Track.ID,
					Like:    !shouldRemove,
					Err:     err,
				}
				return likeUnlikeTrackMsg
			}
			return m, cmd
		}
	case " ":
		if m.FocusedOn != Player {
			return m, nil
		}
		return m.HandleMusicPausePlay()
	case "b":
		if m.FocusedOn != Player {
			return m, nil
		}
		return m.handleMusicChange(false, true)
	case "n":
		if m.FocusedOn != Player {
			return m, nil
		}
		return m.handleMusicChange(true, true)
	case "q", "ctrl+c":
		if m.FocusedOn == SearchBar {
			return m, nil
		}
		if m.PlayerProcess != nil {
			err := m.PlayerProcess.Close(true)
			if err != nil {
				slog.Error(err.Error())
			}
		}
		if m.LyricsServerProcess != nil {
			err := command.KillProcess(m.LyricsServerProcess)
			if err != nil {
				slog.Error(err.Error())
			}
		}
		return m, tea.Quit
	case "tab":
		return changeFocusMode(&m, false)
	case "shift+tab":
		return changeFocusMode(&m, true)
	case "enter":
		return m.handleEnterKey()
	}
	return m, nil
}

func getNextPageItems(m *Model, paginationInfo *types.PaginationInfo, ShouldAppendQueue bool) tea.Cmd {
	userToken := m.GetUserToken()
	if userToken == nil {
		slog.Error("failed to get user access token")
		return nil
	}
	switch paginationInfo.NextPageURLType {
	case types.NextPageURLTypePlaylistTracks:
		return func() tea.Msg {
			playlistItems, err := m.SpotifyClient.GetPlaylistItems(userToken.AccessToken, paginationInfo.NextItemID, &paginationInfo.Next)
			if err != nil {
				return types.UpdatePlaylistMsg{
					Playlist: nil,
					Err:      err,
				}
			}

			if playlistItems.Next != "" {
				paginationInfo.Next = playlistItems.Next
			}

			return types.UpdatePlaylistMsg{
				Playlist:          playlistItems.Items,
				Err:               nil,
				ShouldAppend:      true,
				PaginationInfo:    paginationInfo,
				ShouldAppendQueue: ShouldAppendQueue,
			}
		}
	case types.NextPageURLTypeUserSavedItems:
		return func() tea.Msg {
			userSavedTracks, err := m.SpotifyClient.GetUserSavedTracks(userToken.AccessToken, &paginationInfo.Next)
			if err != nil {
				return types.UpdatePlaylistMsg{
					Playlist: nil,
					Err:      err,
				}
			}
			var playlistItems []*types.PlaylistTrackObject
			for _, item := range userSavedTracks.Items {
				playlistItems = append(playlistItems, &types.PlaylistTrackObject{
					AddedAt: "",
					AddedBy: nil,
					IsLocal: false,
					Track:   item.Track,
				})
			}
			var paginationInfo *types.PaginationInfo
			if userSavedTracks.Next != "" {
				paginationInfo = &types.PaginationInfo{
					Next:            userSavedTracks.Next,
					NextPageURLType: types.NextPageURLTypeUserSavedItems,
					NextItemID:      "",
				}
			}

			return types.UpdatePlaylistMsg{
				Playlist:          playlistItems,
				Err:               nil,
				ShouldAppend:      true,
				PaginationInfo:    paginationInfo,
				ShouldAppendQueue: ShouldAppendQueue,
			}
		}
	}
	return nil
}
func (m Model) getMusicLyrics(track *SelectedTrack) (Model, tea.Cmd) {
	var alertCmd tea.Cmd
	if !m.IsLyricsServerInstalled {
		alertCmd = m.Alert.NewAlertCmd(bubbleup.ErrorKey, "Lyrics server is not installed")
		return m, alertCmd
	}
	if m.FocusedOn != Player {
		return m, nil
	}
	if track == nil || track.Track == nil {
		alertCmd = m.Alert.NewAlertCmd(bubbleup.ErrorKey, "Failed to get lyrics for this track")
		return m, alertCmd
	}
	trackName := track.Track.Track.Name
	var artistNames []string
	for _, artist := range track.Track.Track.Artists {
		artistNames = append(artistNames, artist.Name)
	}
	albumName := track.Track.Track.Album.Name

	lyricsCmd := func() tea.Msg {
		isServerRunning, err := lyrics.IsLyricsServerRunning()
		if err != nil {
			slog.Error(err.Error())
		}

		if !isServerRunning {
			process, err := lyrics.StartLyricsServer()
			if err != nil {
				slog.Error(err.Error())
				return nil
			}
			m.LyricsServerProcess = process
		}
		lyricsResponse, err := lyrics.GetMusicLyrics(trackName, artistNames, albumName)
		if err != nil {
			slog.Error(err.Error())
			i := rand.Intn(len(lyrics.LyricsErrors))
			errMessage := lyrics.LyricsErrors[i]
			return &lyrics.Response{
				Match:  nil,
				Lyrics: &errMessage,
			}
		}
		return lyricsResponse
	}
	loadingCmd := SendLoadingCmd()
	return m, tea.Batch(loadingCmd, lyricsCmd, alertCmd)
}

func (m Model) handleMusicChange(isForward, isSkip bool) (Model, tea.Cmd) {
	if m.MusicQueueList == nil {
		return m, nil
	}

	if len(m.MusicQueueList.Model.Items()) <= 0 {
		return m, nil
	}

	var currentlySelectedMusicIndex int
	for index, track := range m.MusicQueueList.Model.Items() {
		if track.(types.PlaylistTrackObject).Track.ID == m.SelectedTrack.Track.Track.ID {
			currentlySelectedMusicIndex = index
			break
		}
	}

	if currentlySelectedMusicIndex == 0 && !isForward {
		return m, nil
	}

	var nextTrackIndex int
	if isForward && len(m.MusicQueueList.Model.Items()) == (currentlySelectedMusicIndex+1) {
		nextTrackIndex = 0
	} else if isForward {
		nextTrackIndex = currentlySelectedMusicIndex + 1
	} else {
		nextTrackIndex = currentlySelectedMusicIndex - 1
	}

	musicToPlay, ok := m.MusicQueueList.Model.Items()[nextTrackIndex].(types.PlaylistTrackObject)
	if !ok {
		slog.Error("failed to cast SelectedPlayListItems to PlaylistTrackObject")
		return m, nil
	}
	m.MusicQueueList.Model.Select(nextTrackIndex)
	var paginationCmd tea.Cmd
	var model Model
	if isForward {
		if m.MusicQueueList != nil {
			model, paginationCmd = m.handlePagination(&m.MusicQueueList.Model, true, &nextTrackIndex)
			m = model
		}
	}
	model, cmd := m.PlaySelectedMusic(musicToPlay, isSkip)
	m = model
	return m, tea.Batch(cmd, paginationCmd)
}

func (m Model) addMusicToQueue() (Model, tea.Cmd) {
	var itemToAdd list.Item
	var currentlyPlayingTrackID string
	if m.FocusedOn == MainView && m.MainViewMode == NormalMode {
		itemToAdd = m.SelectedPlayListItems.SelectedItem()
	} else if m.FocusedOn == SearchResultTrack && m.MainViewMode == SearchResultMode {
		itemToAdd = m.SearchResult.Tracks.SelectedItem()
	} else {
		return m, nil
	}

	if m.SelectedTrack != nil && m.SelectedTrack.Track != nil {
		currentlyPlayingTrackID = m.SelectedTrack.Track.Track.ID
	}

	var musicQueue = m.MusicQueueList.Items()

	if len(musicQueue) == 0 {
		return m, m.MusicQueueList.SetItems([]list.Item{itemToAdd})
	}

	item, ok := itemToAdd.(types.PlaylistTrackObject)
	if !ok {
		slog.Error("failed to cast itemToAdd to PlaylistTrackObject")
		return m, nil
	}

	item.IsItFromQueue = true
	itemToAdd = item

	var currentlyPlayingTrackIndex int
	for index, item := range m.MusicQueueList.Items() {
		playlistTrackObject, ok := item.(types.PlaylistTrackObject)
		if !ok {
			continue
		}
		if playlistTrackObject.Track.ID == currentlyPlayingTrackID {
			currentlyPlayingTrackIndex = index
		}
	}

	var itemsAfterCurrentlyPlayingTrack = m.MusicQueueList.Items()[currentlyPlayingTrackIndex+1:]
	var itemsBeforeCurrentlyPlayingTrack = m.MusicQueueList.Items()[:currentlyPlayingTrackIndex+1]
	cmd := m.MusicQueueList.SetItems(append(itemsBeforeCurrentlyPlayingTrack, append([]list.Item{itemToAdd}, itemsAfterCurrentlyPlayingTrack...)...))
	return m, cmd
}

func (m Model) HandleMusicPausePlay() (Model, tea.Cmd) {
	if m.PlayerProcess == nil {
		return m, nil
	}
	if m.PlayerProcess.OtoPlayer == nil {
		return m, nil
	}
	if m.PlayerProcess.OtoPlayer.IsPlaying() {
		m.PlayerProcess.OtoPlayer.Pause()

		if m.DBusConn != nil {
			dbusErr := m.DBusConn.Props.Set("org.mpris.MediaPlayer2.Player",
				"PlaybackStatus",
				dbus.MakeVariant("Paused"),
			)
			if dbusErr != nil {
				slog.Error(dbusErr.Error())
			}
		}

		return m, nil
	}

	if m.DBusConn != nil {
		dbusErr := m.DBusConn.Props.Set("org.mpris.MediaPlayer2.Player",
			"PlaybackStatus",
			dbus.MakeVariant("Playing"),
		)

		if dbusErr != nil {
			slog.Error(dbusErr.Error())
		}
	}

	m.PlayerProcess.OtoPlayer.Play()
	return m, nil
}

func getListItemForMusicToChoose(m *Model, focusedOn FocusedOn) *list.Model {
	if focusedOn == MainView {
		return &m.SelectedPlayListItems
	}
	if focusedOn == QueueList && m.MusicQueueList != nil {
		return &m.MusicQueueList.Model
	}
	return nil
}

func (m Model) handleEnterKey() (Model, tea.Cmd) {
	if m.FocusedOn == MainView || m.FocusedOn == QueueList {
		listItemToChooseMusicFrom := getListItemForMusicToChoose(&m, m.FocusedOn)
		selectedMusic, ok := listItemToChooseMusicFrom.SelectedItem().(types.PlaylistTrackObject)
		if !ok {
			alertCmd := m.Alert.NewAlertCmd(bubbleup.ErrorKey, "failed to cast SelectedPlayListItems to PlaylistTrackObject")
			return m, alertCmd
		}

		var items []list.Item
		for _, item := range m.SelectedPlayListItems.Items() {
			playlistItem := types.PlaylistTrackObject{
				Track:         item.(types.PlaylistTrackObject).Track,
				IsItFromQueue: true,
			}
			items = append(items, playlistItem)
		}
		if m.MusicQueueList == nil {
			return m, nil
		}
		m.MusicQueueList.Model.SetItems(items)
		m.MusicQueueList.Model.Select(m.MusicQueueList.GlobalIndex())
		return m.PlaySelectedMusic(selectedMusic, false)
	}
	userToken := m.GetUserToken()

	if userToken == nil {
		slog.Error("nil user token")
		return m, nil
	}

	if m.FocusedOn == SearchBar {
		query := m.Search.Value()
		if query == m.SearchQuery && m.SearchResult != nil {
			m.MainViewMode = SearchResultMode
			return m, nil
		}

		loadingCmd := SendLoadingCmd()
		searchingCmd := func() tea.Msg {
			searchResult, err := m.SpotifyClient.GetSearchResults(userToken.AccessToken, query)
			return types.SpotifySearchResultMsg{
				Result: searchResult,
				Err:    err,
			}
		}
		m.SearchQuery = query
		return m, tea.Batch(loadingCmd, searchingCmd)
	}

	if m.FocusedOn == SideView {
		if userToken == nil {
			slog.Error("failed to get user access token")
			return m, nil
		}

		m.PaginationInfo = nil
		loadingCmd := SendLoadingCmd()
		switch selectedItem := m.Playlist.SelectedItem().(type) {
		case types.Playlist:
			cmd := m.getPlaylistItems(userToken.AccessToken, selectedItem.ID)
			return m, tea.Batch(loadingCmd, cmd)
		case types.Artist:
			cmd := m.getArtistTracks(userToken.AccessToken, selectedItem.ID)
			return m, tea.Batch(loadingCmd, cmd)
		case spotify.UserSavedTracksListItem:
			cmd := m.getUserSavedTracks(userToken.AccessToken)
			return m, tea.Batch(loadingCmd, cmd)
		}
	}
	if m.FocusedOn == SearchResultArtist {
		selectedItem, ok := m.SearchResult.Artists.SelectedItem().(types.Artist)
		if !ok {
			slog.Error("failed to cast the selected item to types.Artist")
			return m, nil
		}
		loadingCmd := SendLoadingCmd()
		cmd := m.getArtistTracks(userToken.AccessToken, selectedItem.ID)
		m.MainViewMode = NormalMode
		m.FocusedOn = MainView
		updateDelegate(&m)
		return m, tea.Batch(cmd, loadingCmd)
	}
	if m.FocusedOn == SearchResultPlaylist {
		selectedItem, ok := m.SearchResult.Playlists.SelectedItem().(types.Playlist)
		if !ok {
			slog.Error("failed to cast the selected item to types.Artist")
			return m, nil
		}

		loadingCmd := SendLoadingCmd()
		cmd := m.getPlaylistItems(userToken.AccessToken, selectedItem.ID)
		m.MainViewMode = NormalMode
		m.FocusedOn = MainView
		updateDelegate(&m)
		return m, tea.Batch(cmd, loadingCmd)
	}
	if m.FocusedOn == SearchResultTrack {
		selectedMusic, ok := m.SearchResult.Tracks.SelectedItem().(types.PlaylistTrackObject)
		if !ok {
			slog.Error("failed to cast the selected item to types.Artist")
			return m, nil
		}
		return m.PlaySelectedMusic(selectedMusic, true)
	}
	return m, nil
}

func (m Model) getArtistTracks(accessToken, artistID string) tea.Cmd {
	return func() tea.Msg {
		artistSongs, err := m.SpotifyClient.GetArtistsTopTrack(accessToken, artistID)
		if err != nil {
			slog.Error(err.Error())
			return types.UpdatePlaylistMsg{
				Playlist: nil,
				Err:      err,
			}
		}
		var tracks []*types.PlaylistTrackObject
		for _, track := range artistSongs.Tracks {
			tracks = append(tracks, &types.PlaylistTrackObject{
				AddedAt: "",
				AddedBy: nil,
				IsLocal: false,
				Track:   track,
			})
		}
		return types.UpdatePlaylistMsg{
			Playlist: tracks,
			Err:      err,
		}
	}
}

func (m Model) getUserSavedTracks(accessToken string) tea.Cmd {
	return func() tea.Msg {
		savedTracks, err := m.SpotifyClient.GetUserSavedTracks(accessToken, nil)
		if err != nil {
			slog.Error(err.Error())
			return types.UpdatePlaylistMsg{
				Playlist: nil,
				Err:      err,
			}
		}
		var tracks []*types.PlaylistTrackObject
		for _, track := range savedTracks.Items {
			tracks = append(tracks, &types.PlaylistTrackObject{
				AddedAt: "",
				AddedBy: nil,
				IsLocal: false,
				Track:   track.Track,
			})
		}
		var paginationInfo *types.PaginationInfo
		if savedTracks.Next != "" {
			paginationInfo = &types.PaginationInfo{
				Next:            savedTracks.Next,
				NextPageURLType: types.NextPageURLTypeUserSavedItems,
			}
		}
		return types.UpdatePlaylistMsg{
			Playlist:       tracks,
			Err:            err,
			PaginationInfo: paginationInfo,
			ShouldAppend:   false,
		}
	}
}

func (m Model) getPlaylistItems(accessToken, playlistID string) tea.Cmd {
	return func() tea.Msg {
		playlistItems, err := m.SpotifyClient.GetPlaylistItems(accessToken, playlistID, nil)
		if err != nil {
			slog.Error(err.Error())
			return types.UpdatePlaylistMsg{
				Playlist: nil,
				Err:      err,
			}
		}
		var paginationInfo *types.PaginationInfo
		if playlistItems.Next != "" {
			paginationInfo = &types.PaginationInfo{
				Next:            playlistItems.Next,
				NextPageURLType: types.NextPageURLTypePlaylistTracks,
				NextItemID:      playlistID,
			}
		}
		return types.UpdatePlaylistMsg{
			Playlist:       playlistItems.Items,
			Err:            err,
			PaginationInfo: paginationInfo,
			ShouldAppend:   false,
		}
	}
}

func (m Model) PlaySelectedMusic(selectedMusic types.PlaylistTrackObject, isSkip bool) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	trackName := selectedMusic.Track.Name
	albumName := selectedMusic.Track.Album.Name
	var artistNames []string
	for _, artist := range selectedMusic.Track.Artists {
		artistNames = append(artistNames, artist.Name)
	}

	playerProcess := m.PlayerProcess

	if playerProcess != nil {
		err := playerProcess.Close(isSkip)
		if err != nil {
			slog.Error(err.Error())
			alertCmd := m.Alert.NewAlertCmd(bubbleup.ErrorKey, err.Error())
			return m, alertCmd
		}
	}
	durationSec := selectedMusic.Track.DurationMS / 1000
	process, err := youtube.SearchAndDownloadMusic(trackName, albumName, artistNames, selectedMusic.Track.ID, m.PlayerProcess == nil, m.YtDlpErrWriter, durationSec, *m.CoreDepsPath)
	if err != nil {
		slog.Error(err.Error())
		alertCmd := m.Alert.NewAlertCmd(bubbleup.ErrorKey, err.Error())
		return m, alertCmd
	}

	cmd := func() tea.Msg {
		if process != nil {
			return types.UpdatePlayedSeconds{
				TrackID: selectedMusic.Track.ID,
			}
		}
		return nil
	}

	cmds = append(cmds, cmd)

	metadata := getMusicMetadata(MusicMetadata{
		artistName: strings.Join(artistNames, ","),
		length:     int64(selectedMusic.Track.DurationMS),
		title:      selectedMusic.Track.Name,
	})

	if m.DBusConn != nil {
		dbusErr := m.DBusConn.Props.Set(
			"org.mpris.MediaPlayer2.Player",
			"Metadata",
			dbus.MakeVariant(metadata),
		)

		if dbusErr != nil {
			slog.Error(dbusErr.Error())
		}

		dbusErr = m.DBusConn.Props.Set("org.mpris.MediaPlayer2.Player",
			"PlaybackStatus",
			dbus.MakeVariant("Playing"),
		)

		if dbusErr != nil {
			slog.Error(dbusErr.Error())
		}
	}
	likedCmd := func() tea.Msg {
		userToken := m.GetUserToken()
		if userToken == nil {
			return nil
		}
		response, err := m.SpotifyClient.CheckUserSavedTrack(userToken.AccessToken, selectedMusic.Track.ID)
		if err != nil {
			return types.CheckUserSavedTrackResponseMsg{
				Saved: false,
				Err:   err,
			}
		}
		var isSaved bool
		if len(response) > 0 {
			isSaved = response[0]
		} else {
			isSaved = false
		}
		return types.CheckUserSavedTrackResponseMsg{
			Saved: isSaved,
			Err:   err,
		}
	}

	cmds = append(cmds, likedCmd)

	m.PlayerProcess = process
	m.SelectedTrack = &SelectedTrack{
		isLiked: false,
		Track:   &selectedMusic,
	}

	if m.MainViewMode == LyricsMode {
		model, cmd := m.getMusicLyrics(m.SelectedTrack)
		m = model
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

func changeFocusMode(m *Model, shift bool) (Model, tea.Cmd) {
	var next, prev FocusedOn
	switch m.FocusedOn {
	case SideView:
		next, prev = MainView, Player
	case MainView:
		if m.MainViewMode == SearchResultMode {
			next = SearchResultTrack
		} else {
			next = QueueList
		}
		prev = SideView
	case SearchResultTrack:
		next, prev = SearchResultArtist, SideView
	case SearchResultArtist:
		next, prev = SearchResultPlaylist, SearchResultTrack
	case SearchResultPlaylist:
		next, prev = QueueList, SearchResultArtist
	case QueueList:
		if m.MainViewMode == SearchResultMode {
			prev = SearchResultPlaylist
		} else {
			prev = MainView
		}
		next = Player
	case Player:
		next, prev = SideView, QueueList
	default:
		if shift {
			items := m.SelectedPlayListItems.Items()
			if len(items) > 0 {
				m.FocusedOn = MainView
				m.SelectedPlayListItems.Select(len(items) - 1)
			} else {
				m.FocusedOn = SideView
			}
			return *m, nil
		}
		m.FocusedOn = SideView
		return *m, nil
	}

	if shift {
		m.FocusedOn = prev
	} else {
		m.FocusedOn = next
	}

	updateDelegate(m)
	return *m, nil
}

func updateDelegate(m *Model) {
	if m == nil {
		return
	}
	m.SelectedPlayListItems.SetDelegate(CustomDelegate{Model: m})
	if m.MusicQueueList != nil {
		m.MusicQueueList.SetDelegate(CustomDelegate{Model: m})
	}
	m.Playlist.SetDelegate(CustomDelegate{Model: m})

	if m.SearchResult != nil {
		m.SearchResult.Tracks.SetDelegate(CustomDelegate{Model: m})
		m.SearchResult.Artists.SetDelegate(CustomDelegate{Model: m})
		m.SearchResult.Playlists.SetDelegate(CustomDelegate{Model: m})
	}
}

func updateFocusedComponent(m *Model, msg tea.Msg, cmdsFromParent *[]tea.Cmd) (Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds = *cmdsFromParent
	cmds = append(cmds, cmd)
	switch m.FocusedOn {
	case SearchBar:
		m.Search.Focus()
		m.Search, cmd = m.Search.Update(msg)
		cmds = append(cmds, cmd)
	case SideView:
		m.Search.Blur()
		m.Playlist, cmd = m.Playlist.Update(msg)
		cmds = append(cmds, cmd)
	case QueueList:
		var cmd tea.Cmd
		if m.MusicQueueList != nil {
			m.MusicQueueList.Model, cmd = m.MusicQueueList.Model.Update(msg)
		}
		cmds = append(cmds, cmd)
	case MainView:
		switch m.MainViewMode {
		case NormalMode:
			m.SelectedPlayListItems, cmd = m.SelectedPlayListItems.Update(msg)
			cmds = append(cmds, cmd)
		}
	case SearchResultPlaylist:
		if m.SearchResult != nil {
			m.SearchResult.Playlists, cmd = m.SearchResult.Playlists.Update(msg)
			cmds = append(cmds, cmd)
		}
	case SearchResultArtist:
		if m.SearchResult != nil {
			m.SearchResult.Artists, cmd = m.SearchResult.Artists.Update(msg)
			cmds = append(cmds, cmd)
		}
	case SearchResultTrack:
		if m.SearchResult != nil {
			m.SearchResult.Tracks, cmd = m.SearchResult.Tracks.Update(msg)
			cmds = append(cmds, cmd)
		}
	default:
	}
	return *m, tea.Batch(cmds...)
}

func SendLoadingCmd() tea.Cmd {
	return func() tea.Msg {
		return types.SearchingMsg{}
	}
}
