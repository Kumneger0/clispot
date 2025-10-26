package ui

import (
	"log/slog"

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
You can turn me on with just a touch, baby
I look around and
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
	case tea.WindowSizeMsg:
		m.Width = msg.Width - 4
		m.Height = msg.Height - 4

		dims := calculateLayoutDimensions(&m)
		m.LyricsView.Width = dims.mainWidth
		m.LyricsView.Height = dims.contentHeight

		m.LyricsView.SetContent(testLyrics)

		return m, nil
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
	case "q", "ctrl+c":
		if m.PlayerProcess != nil {
			err := m.PlayerProcess.Kill()
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

func (m Model) handleEnterKey() (Model, tea.Cmd) {
	if m.FocusedOn == MainView {
		selectedMusic, ok := m.SelectedPlayListItems.SelectedItem().(types.PlaylistTrackObject)
		if !ok {
			//TODO: find a way to show error message for the user
			slog.Error("failed to cast SelectedPlayListItems to PlaylistTrackObject")
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
			err := playerProcess.Kill()
			if err != nil {
				slog.Error(err.Error())
			}
			//TODO:Show error message
		}

		process, err := youtube.SearchAndDownloadMusic(trackName, albumName, artistNames)
		if err != nil {
			slog.Error(err.Error())
			//TODO: implement some kind of way to show the error message
			return m, nil
		}

		m.PlayerProcess = process
		m.SelectedTrack = &selectedMusic
		return m, nil
	}
	selectedItem, ok := m.Playlist.SelectedItem().(types.SpotifyPlaylist)
	if !ok {
		slog.Error("failed to cast Playlist to SpotifyPlaylist")
		return m, nil
	}
	playlistItems, err := spotify.GetPlaylistItems(selectedItem.ID, m.UserTokenInfo.AccessToken)

	if err != nil {
		slog.Error(err.Error())
		return m, nil
	}

	var playListItemSongs []list.Item

	for _, item := range playlistItems.Items {
		playListItemSongs = append(playListItemSongs, item)
	}

	cmd := tea.Batch(m.SelectedPlayListItems.SetItems(playListItemSongs), m.MusicQueueList.SetItems(playListItemSongs))
	return m, cmd
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
