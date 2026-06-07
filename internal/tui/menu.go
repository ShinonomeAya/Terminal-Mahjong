package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

var menuItems = []string{"Solo Game", "How to Play", "Quit"}

func updateMenu(m Model, key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.Type {
	case tea.KeyDown:
		m.MenuIndex = (m.MenuIndex + 1) % len(menuItems)
	case tea.KeyUp:
		m.MenuIndex = (m.MenuIndex + len(menuItems) - 1) % len(menuItems)
	case tea.KeyEnter:
		switch m.MenuIndex {
		case 0:
			m.Game = newStartedGame()
			m.Screen = ScreenTable
		case 1:
			m.Screen = ScreenHelp
		case 2:
			return m, tea.Quit
		}
	}
	return m, nil
}

func updateHelp(m Model, key tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Type == tea.KeyEsc || key.String() == "q" {
		m.Screen = ScreenMenu
	}
	return m, nil
}

func renderMenu(m Model) string {
	var out strings.Builder
	out.WriteString("TERMINAL MAHJONG\n\n")
	out.WriteString("Menu\n")
	for i, item := range menuItems {
		prefix := "  "
		if i == m.MenuIndex {
			prefix = "> "
		}
		out.WriteString(fmt.Sprintf("%s%s\n", prefix, item))
	}
	out.WriteString("\nControls\n")
	out.WriteString("Up/Down choose | Enter confirm | Q quit\n")
	return out.String()
}

func renderHelp() string {
	return strings.Join([]string{
		"TERMINAL MAHJONG HELP",
		"",
		"Keyboard",
		"Left/Right select tile",
		"Enter/Space discard selected tile",
		"Q quit current game",
		"",
		"Mouse",
		"Click selects a hand tile",
		"Second click discards the selected tile",
		"",
		"Controls",
		"Esc returns to menu",
		"",
	}, "\n")
}
