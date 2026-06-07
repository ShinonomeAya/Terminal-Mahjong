package tui

import tea "github.com/charmbracelet/bubbletea"

func updateTable(m Model, key tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.Game == nil {
		return m, nil
	}
	handLen := len(m.Game.Players[0].Hand)
	switch key.Type {
	case tea.KeyLeft:
		if handLen > 0 {
			m.SelectedIndex = (m.SelectedIndex + handLen - 1) % handLen
		}
	case tea.KeyRight:
		if handLen > 0 {
			m.SelectedIndex = (m.SelectedIndex + 1) % handLen
		}
	case tea.KeyEnter:
		return discardSelected(m)
	}
	switch key.String() {
	case " ":
		return discardSelected(m)
	case "q":
		m.Game.Quit("quit")
		m.Screen = ScreenGameOver
	}
	return m, nil
}

func discardSelected(m Model) (tea.Model, tea.Cmd) {
	if m.Game == nil || len(m.Game.Players[0].Hand) == 0 {
		return m, nil
	}
	if m.SelectedIndex >= len(m.Game.Players[0].Hand) {
		m.SelectedIndex = len(m.Game.Players[0].Hand) - 1
	}
	if _, err := m.Game.HumanDiscardSelected(m.SelectedIndex); err != nil {
		return m, nil
	}
	m.Game.AdvanceAIUntilHumanTurn()
	if len(m.Game.Players[0].Hand) == 0 {
		m.SelectedIndex = 0
	} else if m.SelectedIndex >= len(m.Game.Players[0].Hand) {
		m.SelectedIndex = len(m.Game.Players[0].Hand) - 1
	}
	if m.Game.Over {
		m.Screen = ScreenGameOver
	}
	return m, nil
}

func updateTableMouse(m Model, msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if msg.Action != tea.MouseActionPress || msg.Button != tea.MouseButtonLeft || m.Game == nil {
		return m, nil
	}
	boxes := currentHandHitBoxes(m)
	index, ok := tileIndexAt(boxes, msg.X, msg.Y)
	if !ok {
		return m, nil
	}
	if index == m.SelectedIndex {
		return discardSelected(m)
	}
	m.SelectedIndex = index
	return m, nil
}
