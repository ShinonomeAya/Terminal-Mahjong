package tui

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#D9F99D"))

	sectionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#93C5FD"))

	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#111827")).
			Background(lipgloss.Color("#FDE68A"))

	mutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF"))

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FBBF24"))
)

func visibleWidth(text string) int {
	return lipgloss.Width(text)
}

func styleTitle(text string) string {
	return titleStyle.Render(text)
}

func styleSectionTitle(text string) string {
	return sectionStyle.Render(text)
}

func styleSelectedTile(text string) string {
	return selectedStyle.Render(text)
}

func styleMuted(text string) string {
	return mutedStyle.Render(text)
}

func styleStatus(text string) string {
	return statusStyle.Render(text)
}
