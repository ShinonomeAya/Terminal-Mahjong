package tui

import (
	"fmt"

	"mahjong/internal/game"

	tea "github.com/charmbracelet/bubbletea"
)

func updateTable(m Model, key tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.Online {
		return updateOnlineTable(m, key)
	}
	if m.Game == nil {
		return m, nil
	}
	handLen := len(m.Game.Players[0].Hand)
	switch key.Type {
	case tea.KeyLeft:
		if m.SelectedIndex > 0 {
			m.SelectedIndex--
			m.StatusLine = selectionStatus("Selected", m.SelectedIndex, m.Game.Players[0].Hand[m.SelectedIndex], m.UnicodeTiles)
		}
	case tea.KeyRight:
		if handLen > 0 && m.SelectedIndex < handLen-1 {
			m.SelectedIndex++
			m.StatusLine = selectionStatus("Selected", m.SelectedIndex, m.Game.Players[0].Hand[m.SelectedIndex], m.UnicodeTiles)
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

func updateOnlineTable(m Model, key tea.KeyMsg) (tea.Model, tea.Cmd) {
	hand := onlineHand(m)
	handLen := len(hand)
	switch key.Type {
	case tea.KeyLeft:
		if m.SelectedIndex > 0 {
			m.SelectedIndex--
			m.StatusLine = selectionStatus("Selected", m.SelectedIndex, hand[m.SelectedIndex], m.UnicodeTiles)
		}
	case tea.KeyRight:
		if handLen > 0 && m.SelectedIndex < handLen-1 {
			m.SelectedIndex++
			m.StatusLine = selectionStatus("Selected", m.SelectedIndex, hand[m.SelectedIndex], m.UnicodeTiles)
		}
	case tea.KeyEnter:
		return discardOnlineSelected(m)
	}
	switch key.String() {
	case " ":
		return discardOnlineSelected(m)
	case "r":
		return readyOnline(m)
	case "h":
		return winOnline(m)
	case "k":
		return kongOnline(m)
	case "q":
		if m.OnlineClient != nil {
			m.OnlineClient.Close()
		}
		m.Screen = ScreenMenu
		m.Online = false
		m.OnlineClient = nil
		m.NetworkStatus = NetworkLocal
	}
	return m, nil
}

func discardOnlineSelected(m Model) (tea.Model, tea.Cmd) {
	hand := onlineHand(m)
	if !m.OnlineStarted {
		m.StatusLine = "Waiting for players to ready"
		return m, nil
	}
	if m.OnlineClient == nil || len(hand) == 0 {
		m.StatusLine = "Online room is not ready"
		return m, nil
	}
	if m.OnlineSnapshot.Current != m.OnlineSeat {
		m.StatusLine = "Waiting for your turn"
		return m, nil
	}
	if m.SelectedIndex >= len(hand) {
		m.SelectedIndex = len(hand) - 1
	}
	discardIndex := m.SelectedIndex
	discardTile := hand[discardIndex]
	m.StatusLine = selectionStatus("Discarding", discardIndex, discardTile, m.UnicodeTiles)
	return m, sendOnlineDiscardCmd(m.OnlineClient, discardIndex)
}

func winOnline(m Model) (tea.Model, tea.Cmd) {
	if !m.OnlineStarted {
		m.StatusLine = "Waiting for players to ready"
		return m, nil
	}
	if m.OnlineClient == nil {
		m.StatusLine = "Online room is not ready"
		return m, nil
	}
	if m.OnlineSnapshot.Current != m.OnlineSeat {
		m.StatusLine = "Waiting for your turn"
		return m, nil
	}
	if !canOnlineWin(m) {
		m.StatusLine = "Win is not available"
		return m, nil
	}
	m.StatusLine = "Winning"
	return m, sendOnlineWinCmd(m.OnlineClient)
}

func kongOnline(m Model) (tea.Model, tea.Cmd) {
	if !m.OnlineStarted {
		m.StatusLine = "Waiting for players to ready"
		return m, nil
	}
	if m.OnlineClient == nil {
		m.StatusLine = "Online room is not ready"
		return m, nil
	}
	if m.OnlineSnapshot.Current != m.OnlineSeat {
		m.StatusLine = "Waiting for your turn"
		return m, nil
	}
	tile, ok := onlineKongTile(m)
	if !ok {
		m.StatusLine = "Kong is not available"
		return m, nil
	}
	m.StatusLine = fmt.Sprintf("Kong %s", game.TileLabel(tile, m.UnicodeTiles))
	return m, sendOnlineKongCmd(m.OnlineClient, tile.String())
}

func readyOnline(m Model) (tea.Model, tea.Cmd) {
	if m.OnlineClient == nil {
		m.StatusLine = "Online room is not ready"
		return m, nil
	}
	m.StatusLine = "Ready sent"
	return m, sendOnlineReadyCmd(m.OnlineClient)
}

func discardSelected(m Model) (tea.Model, tea.Cmd) {
	if m.Game == nil || len(m.Game.Players[0].Hand) == 0 {
		return m, nil
	}
	if m.SelectedIndex >= len(m.Game.Players[0].Hand) {
		m.SelectedIndex = len(m.Game.Players[0].Hand) - 1
	}
	discardIndex := m.SelectedIndex
	discardTile := m.Game.Players[0].Hand[discardIndex]
	if _, err := m.Game.HumanDiscardSelected(m.SelectedIndex); err != nil {
		return m, nil
	}
	m.StatusLine = selectionStatus("Discarded", discardIndex, discardTile, m.UnicodeTiles)
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
	if msg.Action != tea.MouseActionPress || msg.Button != tea.MouseButtonLeft {
		return m, nil
	}
	if m.Online {
		return updateOnlineTableMouse(m, msg)
	}
	if m.Game == nil {
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
	m.StatusLine = selectionStatus("Mouse selected", index, m.Game.Players[0].Hand[index], m.UnicodeTiles)
	return m, nil
}

func updateOnlineTableMouse(m Model, msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	hand := onlineHand(m)
	if len(hand) == 0 {
		return m, nil
	}
	boxes := currentHandHitBoxes(m)
	index, ok := tileIndexAt(boxes, msg.X, msg.Y)
	if !ok {
		return m, nil
	}
	if index == m.SelectedIndex {
		return discardOnlineSelected(m)
	}
	m.SelectedIndex = index
	m.StatusLine = selectionStatus("Mouse selected", index, hand[index], m.UnicodeTiles)
	return m, nil
}

func selectionStatus(action string, index int, tile game.Tile, unicode bool) string {
	return fmt.Sprintf("%s [%02d] %s (%s)", action, index+1, game.TileLabel(tile, unicode), tile.String())
}

func updateGameOver(m Model, key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.Type {
	case tea.KeyDown:
		if m.GameOverIndex < len(gameOverItems)-1 {
			m.GameOverIndex++
		}
	case tea.KeyUp:
		if m.GameOverIndex > 0 {
			m.GameOverIndex--
		}
	case tea.KeyEnter:
		if m.Online {
			return updateOnlineGameOver(m)
		}
		switch m.GameOverIndex {
		case 0:
			m.Game = newStartedGame()
			m.Screen = ScreenTable
			m.SelectedIndex = 0
			m.GameOverIndex = 0
		case 1:
			m.Game = nil
			m.Screen = ScreenMenu
			m.SelectedIndex = 0
			m.GameOverIndex = 0
		case 2:
			return m, tea.Quit
		}
	}
	switch key.String() {
	case "r":
		if m.Online {
			return m, nil
		}
		m.Game = newStartedGame()
		m.Screen = ScreenTable
		m.SelectedIndex = 0
		m.GameOverIndex = 0
	case "m":
		return clearOnlineStateForMenu(m), nil
	case "q":
		return m, tea.Quit
	}
	return m, nil
}

func updateOnlineGameOver(m Model) (tea.Model, tea.Cmd) {
	switch m.GameOverIndex {
	case 0:
		return m, nil
	case 1:
		return clearOnlineStateForMenu(m), nil
	case 2:
		if m.OnlineClient != nil {
			m.OnlineClient.Close()
		}
		return m, tea.Quit
	default:
		return m, nil
	}
}

func clearOnlineStateForMenu(m Model) Model {
	if m.OnlineClient != nil {
		m.OnlineClient.Close()
	}
	m.Game = nil
	m.Online = false
	m.OnlineClient = nil
	m.OnlineSnapshot = game.GameSnapshot{}
	m.OnlinePlayerID = ""
	m.OnlineRoomCode = ""
	m.OnlineSeat = 0
	m.OnlineEvents = nil
	m.OnlineReadySeats = nil
	m.OnlineOccupiedSeats = nil
	m.OnlineStarted = false
	m.Screen = ScreenMenu
	m.SelectedIndex = 0
	m.GameOverIndex = 0
	m.StatusLine = ""
	m.NetworkStatus = NetworkLocal
	return m
}
