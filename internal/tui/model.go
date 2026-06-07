package tui

import (
	"mahjong/internal/game"

	tea "github.com/charmbracelet/bubbletea"
)

type Screen int

const (
	ScreenMenu Screen = iota
	ScreenHelp
	ScreenTable
	ScreenGameOver
)

type Model struct {
	Screen        Screen
	MenuIndex     int
	SelectedIndex int
	UnicodeTiles  bool
	Game          *game.Game
}

func NewModel() Model {
	return Model{Screen: ScreenMenu, UnicodeTiles: true}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch m.Screen {
	case ScreenMenu:
		return updateMenu(m, key)
	case ScreenHelp:
		return updateHelp(m, key)
	default:
		return m, nil
	}
}

func (m Model) View() string {
	switch m.Screen {
	case ScreenMenu:
		return renderMenu(m)
	case ScreenHelp:
		return renderHelp()
	case ScreenTable:
		return renderTable(m)
	case ScreenGameOver:
		return renderGameOver(m)
	default:
		return ""
	}
}
