package ui

import (
	"log/slog"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kumneger0/clispot/internal/spotify"
	"github.com/kumneger0/clispot/internal/types"
	"github.com/kumneger0/clispot/internal/youtube"
)

var testLyrics = `
Yeah
I've been tryna call
I've been on my own for long enough
Maybe you can show me how to love, maybe
I'm goin' through withdrawals
You don't even have to do too much
You can turn me
Sin City's cold and empty (oh)
No one's around to judge me (oh)
I can't see clearly when you're gone
I said, ooh, I'm blinded by the lights
No, I can't sleep until I feel your touch
I said, ooh, I'm drowning in the night
Oh, when I'm like this, you're the one I trust
(Hey, hey, hey)
I'm running out of time
'Cause I can see the sun light up the sky
So I hit the road in overdrive, baby, oh
The city's cold and empty (oh)
No one's around to judge me (oh)
I can't see clearly when you're gone
I said, ooh, I'm blinded by the lights
No, I can't sleep until I feel your touch
I said, ooh, I'm drowning in the night
Oh, when I'm like this, you're the one I trust
I'm just walking by to let you know (by to let you know)
I could never say it on the phone (say it on the phone)
Will never let you go this time (ooh)
I said, ooh, I'm blinded by the lights
No, I can't sleep until I feel your touch
(Hey, hey, hey)
(Hey, hey, hey)
I said, ooh, I'm blinded by the lights
No, I can't sleep until I feel your touch`

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case types.UpdatePlayedSeconds:
		m.PlayedSeconds = msg.CurrentSeconds
		cmd := func() tea.Cmd {
			return tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
				if m.PlayerProcess != nil {
					currentSeconds := m.PlayerProcess.ByteCounterReader.CurrentSeconds()
					return types.UpdatePlayedSeconds{
						CurrentSeconds: currentSeconds,
					}
				}
				return nil
			})
		}
		var totalDurationInSeconds int

		if m.SelectedTrack != nil {
			totalDurationInSeconds = m.SelectedTrack.Track.DurationMs / 1000
		}
		diff := float64(totalDurationInSeconds) - (m.PlayedSeconds)

		if diff < 2 {
			model, cmd := m.handleMusicChange(true)
			m = model
			cmds = append(cmds, cmd)
		}

		cmds = append(cmds, cmd())
	case tea.WindowSizeMsg:
		m.Width = msg.Width - 4
		m.Height = msg.Height - 4

		dims := calculateLayoutDimensions(&m)
		m.LyricsView.Width = dims.mainWidth
		m.LyricsView.Height = dims.contentHeight

		m.LyricsView.SetContent(testLyrics)
		return m, nil
	case types.UpdatePlaylistMsg:
		if msg.Playlist != nil {
			var playListItemSongs []list.Item
			for _, item := range msg.Playlist {
				playListItemSongs = append(playListItemSongs, *item)
			}
			m.SelectedTrack = nil
			cmd := m.SelectedPlayListItems.SetItems(playListItemSongs)
			cmds = append(cmds, cmd)
		}
		if msg.Err != nil {
			//TODO: find nice way to show error messages
		}
	case tea.KeyMsg:
		model, cmd := m.handleKeyPress(msg)
		m = model
		cmds = append(cmds, cmd)
	default:
		//TODO: do something here if no key matched
	}
	return updateFocusedComponent(&m, msg, &cmds)
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+k":
		m.FocusedOn = SearchBar
		return m, m.Search.Focus()
	case " ":
		return m.handleMusicPausePlay()
	case "b":
		return m.handleMusicChange(false)
	case "n":
		return m.handleMusicChange(true)
	case "q", "ctrl+c":
		if m.PlayerProcess != nil {
			err := m.PlayerProcess.Close()
			if err != nil {
				slog.Error(err.Error())
			}
		}
		return m, tea.Quit
	case "tab":
		return changeFocusMode(&m, false)
	case "enter":
		return m.handleEnterKey()
	}
	return m, nil
}

func (m Model) handleMusicChange(isForward bool) (Model, tea.Cmd) {
	if m.FocusedOn != Player || len(m.MusicQueueList.Items()) <= 0 {
		return m, nil
	}
	currentlySelectedMusicIndex := m.MusicQueueList.GlobalIndex()
	if currentlySelectedMusicIndex == 0 && !isForward {
		return m, nil
	}

	var nextTrackIndex int
	if isForward && len(m.MusicQueueList.Items()) == (currentlySelectedMusicIndex+1) {
		nextTrackIndex = 0
	} else if isForward {
		nextTrackIndex = currentlySelectedMusicIndex + 1
	} else {
		nextTrackIndex = currentlySelectedMusicIndex - 1
	}

	musicToPlay, ok := m.MusicQueueList.Items()[nextTrackIndex].(types.PlaylistTrackObject)
	if !ok {
		slog.Error("failed to cast SelectedPlayListItems to PlaylistTrackObject")
		return m, nil
	}
	m.MusicQueueList.Select(nextTrackIndex)
	return m.PlaySelectedMusic(musicToPlay)
}

func (m Model) handleMusicPausePlay() (Model, tea.Cmd) {
	if m.FocusedOn != Player || m.PlayerProcess == nil {
		return m, nil
	}
	if m.PlayerProcess.OtoPlayer.IsPlaying() {
		m.PlayerProcess.OtoPlayer.Pause()
		return m, nil
	}
	m.PlayerProcess.OtoPlayer.Play()
	return m, nil
}

func getListItemForMusicToChoose(m *Model, focusedOn FocusedOn) *list.Model {
	if focusedOn == MainView {
		return &m.SelectedPlayListItems
	}
	if focusedOn == QueueList {
		return &m.MusicQueueList
	}

	return nil
}

func (m Model) handleEnterKey() (Model, tea.Cmd) {
	if m.FocusedOn == MainView || m.FocusedOn == QueueList {
		listItemToChooseMusicFrom := getListItemForMusicToChoose(&m, m.FocusedOn)
		selectedMusic, ok := listItemToChooseMusicFrom.SelectedItem().(types.PlaylistTrackObject)
		if !ok {
			//TODO: find a way to show error message for the user
			slog.Error("failed to cast SelectedPlayListItems to PlaylistTrackObject")
			return m, nil
		}

		m.MusicQueueList.SetItems(m.SelectedPlayListItems.Items())
		m.MusicQueueList.Select(m.SelectedPlayListItems.GlobalIndex())
		return m.PlaySelectedMusic(selectedMusic)
	}
	selectedItem, ok := m.Playlist.SelectedItem().(types.SpotifyPlaylist)
	if !ok {
		slog.Error("failed to cast Playlist to SpotifyPlaylist")
		return m, nil
	}

	cmd := func() tea.Msg {
		playlistItems, err := spotify.GetPlaylistItems(selectedItem.ID, m.UserTokenInfo.AccessToken)
		return types.UpdatePlaylistMsg{
			Playlist: playlistItems.Items,
			Err:      err,
		}
	}
	return m, cmd
}

func (m Model) PlaySelectedMusic(selectedMusic types.PlaylistTrackObject) (Model, tea.Cmd) {
	trackName := selectedMusic.Track.Name
	albumName := selectedMusic.Track.Album.Name
	var artistNames []string
	for _, artist := range selectedMusic.Track.Artists {
		artistNames = append(artistNames, artist.Name)
	}

	playerProcess := m.PlayerProcess

	if playerProcess != nil {
		err := playerProcess.Close()
		if err != nil {
			slog.Error(err.Error())
		}
		//TODO:Show error message
	}
	process, err := youtube.SearchAndDownloadMusic(trackName, albumName, artistNames, m.PlayerProcess == nil)
	if err != nil {
		slog.Error(err.Error())
		//TODO: implement some kind of way to show the error message
		return m, nil
	}

	cmd := func() tea.Cmd {
		return tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
			if process != nil {
				currentSeconds := process.ByteCounterReader.CurrentSeconds()
				return types.UpdatePlayedSeconds{
					CurrentSeconds: currentSeconds,
				}
			}
			return nil
		})
	}

	m.PlayerProcess = process
	m.SelectedTrack = &selectedMusic
	return m, cmd()
}

func changeFocusMode(m *Model, shift bool) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	currentlyFocusedOn := m.FocusedOn
	switch currentlyFocusedOn {
	case SideView:
		if shift {
			m.FocusedOn = Player
		} else {
			m.FocusedOn = MainView
			chatListLastIndex := len(m.SelectedPlayListItems.Items()) - 1
			m.SelectedPlayListItems.Select(chatListLastIndex)
		}
	case MainView:
		if shift {
			m.FocusedOn = SideView
		} else {
			m.FocusedOn = QueueList
		}
	case QueueList:
		if shift {
			m.FocusedOn = MainView
		} else {
			m.FocusedOn = Player
		}
	default:
		if shift {
			m.FocusedOn = MainView
			chatListLastIndex := len(m.SelectedPlayListItems.Items()) - 1
			m.SelectedPlayListItems.Select(chatListLastIndex)
		} else {
			m.FocusedOn = SideView
		}
	}
	return *m, tea.Batch(cmds...)
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
		m.MusicQueueList, cmd = m.MusicQueueList.Update(msg)
		cmds = append(cmds, cmd)
	default:
		m.SelectedPlayListItems, cmd = m.SelectedPlayListItems.Update(msg)
		cmds = append(cmds, cmd)
		lyrics, cmd := m.LyricsView.Update(msg)
		m.LyricsView = lyrics
		cmds = append(cmds, cmd)
	}
	return *m, tea.Batch(cmds...)
}
