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
	GameOverIndex int
	SelectedIndex int
	UnicodeTiles  bool
	Game          *game.Game
	HandHitBoxes  []TileHitBox
	StatusLine    string
}

func NewModel() Model {
	return Model{Screen: ScreenMenu, UnicodeTiles: true}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.Screen {
		case ScreenMenu:
			return updateMenu(m, msg)
		case ScreenHelp:
			return updateHelp(m, msg)
		case ScreenTable:
			return updateTable(m, msg)
		case ScreenGameOver:
			return updateGameOver(m, msg)
		default:
			return m, nil
		}
	case tea.MouseMsg:
		if m.Screen == ScreenTable {
			return updateTableMouse(m, msg)
		}
		return m, nil
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
