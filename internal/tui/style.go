package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"mahjong/internal/game"
)

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

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#D6CCC2")).
			Padding(0, 1)

	tileFaceStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#111827")).
			Background(lipgloss.Color("#F8FAFC")).
			Bold(true).
			Padding(0, 1)
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

func stylePanel(title string, body string) string {
	content := strings.TrimRight(title+"\n"+body, "\n")
	return panelStyle.Render(content)
}

func stylePanelWidth(title string, body string, width int) string {
	content := strings.TrimRight(title+"\n"+body, "\n")
	return panelStyle.Width(width).Render(content)
}

func styleTileFace(label string, selected bool) string {
	style := tileFaceStyle
	if selected {
		style = style.Background(lipgloss.Color("#FDE68A"))
	}
	return style.Render(label)
}

func styleMahjongTile(tile game.Tile, label string, selected bool) string {
	style := tileFaceStyle.Foreground(tileColor(tile))
	if selected {
		style = style.Background(lipgloss.Color("#FDE68A"))
	}
	return style.Render(label)
}

func tileColor(tile game.Tile) lipgloss.Color {
	switch {
	case tile >= 0 && tile < 9:
		return lipgloss.Color("#DC2626")
	case tile >= 9 && tile < 18:
		return lipgloss.Color("#2563EB")
	case tile >= 18 && tile < 27:
		return lipgloss.Color("#16A34A")
	default:
		return lipgloss.Color("#7C3AED")
	}
}
