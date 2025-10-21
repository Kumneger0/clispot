package ui

import "github.com/charmbracelet/lipgloss"

var (
	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#7D56F4")).
			Bold(true)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).Padding(0, 1)

	timestampStyle = lipgloss.NewStyle().Height(1).
			Foreground(lipgloss.Color("#999999")).
			Italic(true).
			MarginRight(1).PaddingLeft(4)

	messageStyle = lipgloss.NewStyle().
			PaddingTop(1).
			PaddingBottom(1).
			Foreground(lipgloss.Color("#FFFFFF"))

	replyMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#AAAAAA"))
)
