package ui

import (
	"fmt"
	"io"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kumneger0/clispot/internal/spotify"
	"github.com/kumneger0/clispot/internal/types"
)

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
	var isSelected bool
	switch item := item.(type) {
	case types.PlaylistTrackObject:
		title = item.FilterValue()
		if d.Model != nil {
			isSelected = (d.Model.FocusedOn == MainView || d.Model.FocusedOn == QueueList) && m.Index() == index
		}
	case types.Artist:
		title = item.FilterValue()
		if d.Model != nil {
			isSelected = d.Model.FocusedOn == SideView && m.Index() == index
		}
	case types.Playlist:
		title = item.FilterValue()
		if d.Model != nil {
			isSelected = d.Model.FocusedOn == SideView && m.Index() == index
		}
	case spotify.UserSavedTracksListItem:
		title = item.FilterValue()
	default:
	}
	str := lipgloss.NewStyle().Render(title)
	if isSelected || index == m.Index() {
		fmt.Fprint(w, selectedStyle.Render(" "+str+" "))
	} else {
		fmt.Fprint(w, normalStyle.Render(" "+str+" "))
	}
}

func renderSearchBar(m *Model, width int) string {
	if width < 20 {
		width = 20
	}
	m.Search.Width = width - 6

	box := lipgloss.NewStyle().
		Width(width).
		Padding(0, 1).
		Margin(0).
		Border(lipgloss.HiddenBorder()).
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("255"))

	var content string
	if m.Search.Value() == "" && !m.Search.Focused() {
		content = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Search tracks, artists, albums...")
	} else {
		content = strings.TrimRight(m.Search.View(), "\n")
	}
	return strings.TrimRight(box.Render(content), "\n")
}

func renderNowPlaying(selectedTrack *SelectedTrack, currentPosition, TotalDuration time.Duration) string {
	if selectedTrack == nil {
		return ""
	}
	var stringBuilder strings.Builder
	var artistNames []string
	var artists = selectedTrack.Track.Track.Artists
	for _, artist := range artists {
		artistNames = append(artistNames, artist.Name)
	}
	artistName := strings.Join(artistNames, ",")
	stringBuilder.WriteString(artistName)

	trackName := selectedTrack.Track.Track.Name

	var likedIndicator string
	if selectedTrack.isLiked {
		likedIndicator = "❤️"
	} else {
		likedIndicator = "💔"
	}

	barWidth := 40
	var progressFloat float64
	if TotalDuration == 0 {
		progressFloat = 1.0
	} else {
		progressFloat = float64(currentPosition.Abs()) / float64(TotalDuration.Abs()) * float64(barWidth)
	}
	progress := max(min(int(math.Max(progressFloat, 1)), barWidth), 0)

	left := strings.Repeat("▰", progress)
	rightCount := max(barWidth-progress, 0)
	right := strings.Repeat("▱", rightCount)

	return fmt.Sprintf("▶ %s — %s %s / %s\n%s %s\n",
		trackName,
		artistName,
		formatTime(currentPosition),
		formatTime(TotalDuration),
		left+right,
		likedIndicator,
	)
}

func renderPlayerControls(isLyricsServerInstalled bool) string {
	btn := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		Foreground(lipgloss.Color("230")).
		Background(lipgloss.Color("237"))

	key := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("250"))
	label := lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
	var controls []string
	prevBtn := btn.Render(key.Render("⏮") + " " + label.Render("Prev\n[b]"))
	playBtn := btn.Render(key.Render("⏯") + " " + label.Render("Play/Pause\n[space]"))
	nextBtn := btn.Render(key.Render("⏭") + " " + label.Render("Next\n[n]"))
	quitBtn := btn.Render(key.Render("✖") + " " + label.Render("Quit\n[q]"))
	controls = append(controls, prevBtn, playBtn, nextBtn, quitBtn)
	if isLyricsServerInstalled {
		lyricsIcon := "📝"
		lyricsBtn := btn.Render(key.Render(lyricsIcon) + " " + label.Render("Lyrics\n[ctrl+l]"))
		controls = append(controls, lyricsBtn)
	}
	row := lipgloss.JoinHorizontal(lipgloss.Top, controls...)
	return row
}
