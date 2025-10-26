package ui

import (
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

func getSideBarStyles(sidebarWidth int, contentHeight int, m *Model) lipgloss.Style {
	sideBarStyle := lipgloss.NewStyle().
		Width(sidebarWidth).
		Height(contentHeight).
		Padding(0).
		Border(getItemBorder(m.FocusedOn == SideView)).
		BorderForeground(getBorderColor(m.FocusedOn == SideView)).
		MaxHeight(contentHeight)
	return sideBarStyle
}

func getMainStyle(mainWidth int, contentHeight int, m *Model) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(mainWidth).
		Height(contentHeight).
		Padding(0).
		Border(getItemBorder(m.FocusedOn == MainView)).
		BorderForeground(getBorderColor(m.FocusedOn == MainView)).
		MaxHeight(contentHeight).
		MaxWidth(mainWidth)
}

func getPlayerStyles(m *Model, inputHeight int) lipgloss.Style {
	inputStyle := lipgloss.NewStyle().Width(m.Width).
		Height(inputHeight).
		Padding(0).
		Border(getItemBorder(m.FocusedOn == Player)).
		BorderForeground(getBorderColor(m.FocusedOn == Player))
	return inputStyle
}

func getQueueListStyle(m *Model, height, width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(0).
		Border(getItemBorder(m.FocusedOn == QueueList)).
		BorderForeground(getBorderColor(m.FocusedOn == QueueList)).
		MaxHeight(height)
}
