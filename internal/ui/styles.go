package ui

import (
	"slices"

	"github.com/charmbracelet/lipgloss"
)

var (
	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#7D56F4")).
			Bold(true)
)

func getBorderColor(isFocused bool) lipgloss.Color {
	if isFocused {
		return lipgloss.Color("#7D56F4")
	}
	return lipgloss.Color("#44475A")
}

func getItemBorder(isSelected bool) lipgloss.Border {
	if isSelected {
		return lipgloss.DoubleBorder()
	}
	return lipgloss.NormalBorder()
}

func getPlayerStyles(m *Model, dims layoutDimensions) lipgloss.Style {
	//we have two sidebars one for users library and one for music queue and have a main area which the center one
	// so the player section should take full available width so we need to calculate this way
	// i added 2 b/c there was a tiny space remaining at the right side, adding 2 fixes that issue
	width := dims.mainWidth + (dims.sidebarWidth*2 + 2)

	inputStyle := getStyle(m, dims.inputHeight, width, Player)
	return inputStyle
}

func getStyle(m *Model, height, width int, focusedOn FocusedOn) lipgloss.Style {
	partsToLimitWidth := []FocusedOn{SearchResultArtist, SearchResultPlaylist, SearchResultTrack, MainView}
	style := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(0).
		Border(getItemBorder(m.FocusedOn == focusedOn)).
		BorderForeground(getBorderColor(m.FocusedOn == focusedOn))

	if focusedOn != Player {
		style = style.MaxHeight(height)
	}

	if slices.Contains(partsToLimitWidth, focusedOn) {
		style = style.MaxWidth(width)
	}
	return style
}
