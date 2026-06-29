package tui

import (
	"fmt"

	"mahjong/internal/game"

	tea "github.com/charmbracelet/bubbletea"
)

func updateTable(m Model, key tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Type == tea.KeyTab {
		m.ShowTactical = !m.ShowTactical
		return m, nil
	}
	if m.Online {
		return updateOnlineTable(m, key)
	}
	if m.Game == nil {
		return m, nil
	}
	if isClaimResponse(m) {
		return updateClaimResponse(m, key)
	}
	handLen := len(m.Game.Players[0].Hand)
	switch key.Type {
	case tea.KeyLeft:
		if m.SelectedIndex > 0 {
			m.SelectedIndex--
			m.StatusLine = selectionStatus(m, "Selected", m.SelectedIndex, m.Game.Players[0].Hand[m.SelectedIndex], m.UnicodeTiles)
		}
	case tea.KeyRight:
		if handLen > 0 && m.SelectedIndex < handLen-1 {
			m.SelectedIndex++
			m.StatusLine = selectionStatus(m, "Selected", m.SelectedIndex, m.Game.Players[0].Hand[m.SelectedIndex], m.UnicodeTiles)
		}
	case tea.KeyEnter:
		return discardSelected(m)
	}
	switch key.String() {
	case " ":
		return discardSelected(m)
	case "q":
		if m.LocalMatch != nil {
			m, _ = applyLocalCommand(m, game.GameCommand{Kind: game.CommandQuit})
		} else {
			m.Game.Quit("quit")
		}
		m.Screen = ScreenGameOver
	case "l":
		return riichiLocal(m)
	}
	return m, nil
}

func updateOnlineTable(m Model, key tea.KeyMsg) (tea.Model, tea.Cmd) {
	if isClaimResponse(m) {
		return updateClaimResponse(m, key)
	}
	hand := onlineHand(m)
	handLen := len(hand)
	switch key.Type {
	case tea.KeyLeft:
		if m.SelectedIndex > 0 {
			m.SelectedIndex--
			m.StatusLine = selectionStatus(m, "Selected", m.SelectedIndex, hand[m.SelectedIndex], m.UnicodeTiles)
		}
	case tea.KeyRight:
		if handLen > 0 && m.SelectedIndex < handLen-1 {
			m.SelectedIndex++
			m.StatusLine = selectionStatus(m, "Selected", m.SelectedIndex, hand[m.SelectedIndex], m.UnicodeTiles)
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
	case "l":
		return riichiOnline(m)
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

func riichiOnline(m Model) (tea.Model, tea.Cmd) {
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
	index, ok := selectedLegalTileIndex(m.OnlineSnapshot.LegalActions, game.CommandRiichi, m.SelectedIndex)
	if !ok {
		m.StatusLine = "Riichi is not available"
		return m, nil
	}
	m.StatusLine = "Riichi"
	return m, sendOnlineGameCommandCmd(m.OnlineClient, game.GameCommand{Kind: game.CommandRiichi, TileIndex: index})
}

func riichiLocal(m Model) (tea.Model, tea.Cmd) {
	if m.Game == nil {
		return m, nil
	}
	index, ok := selectedLegalTileIndex(m.Game.Snapshot().LegalActions, game.CommandRiichi, m.SelectedIndex)
	if !ok {
		m.StatusLine = "Riichi is not available"
		return m, nil
	}
	var result game.CommandResult
	m, result = applyLocalCommand(m, game.GameCommand{Kind: game.CommandRiichi, TileIndex: index})
	if !result.OK {
		m.StatusLine = result.Error
		return m, nil
	}
	m.StatusLine = "Riichi"
	m = advanceLocalAI(m)
	return finishLocalUpdate(m)
}

func selectedLegalTileIndex(actions []game.LegalAction, kind game.CommandKind, selected int) (int, bool) {
	fallback := -1
	for _, action := range actions {
		if action.Kind != kind {
			continue
		}
		if fallback < 0 {
			fallback = action.TileIndex
		}
		if action.TileIndex == selected {
			return selected, true
		}
	}
	if fallback >= 0 {
		return fallback, true
	}
	return 0, false
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
	m.StatusLine = selectionStatus(m, "Discarding", discardIndex, discardTile, m.UnicodeTiles)
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
	if m.LocalMatch != nil {
		var result game.CommandResult
		m, result = applyLocalCommand(m, game.GameCommand{Kind: game.CommandDiscard, TileIndex: m.SelectedIndex})
		if !result.OK {
			return m, nil
		}
	} else {
		if _, err := m.Game.HumanDiscardSelected(m.SelectedIndex); err != nil {
			return m, nil
		}
	}
	m.StatusLine = selectionStatus(m, "Discarded", discardIndex, discardTile, m.UnicodeTiles)
	m = advanceLocalAI(m)
	if len(m.Game.Players[0].Hand) == 0 {
		m.SelectedIndex = 0
	} else if m.SelectedIndex >= len(m.Game.Players[0].Hand) {
		m.SelectedIndex = len(m.Game.Players[0].Hand) - 1
	}
	return finishLocalUpdate(m)
}

func updateTableMouse(m Model, msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if msg.Action != tea.MouseActionPress || msg.Button != tea.MouseButtonLeft {
		return m, nil
	}
	if m.Online {
		return updateOnlineTableMouse(m, msg)
	}
	if isClaimResponse(m) {
		return m, nil
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
	m.StatusLine = selectionStatus(m, "Mouse selected", index, m.Game.Players[0].Hand[index], m.UnicodeTiles)
	return m, nil
}

func updateOnlineTableMouse(m Model, msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if isClaimResponse(m) {
		return m, nil
	}
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
	m.StatusLine = selectionStatus(m, "Mouse selected", index, hand[index], m.UnicodeTiles)
	return m, nil
}

func updateClaimResponse(m Model, key tea.KeyMsg) (tea.Model, tea.Cmd) {
	options := activeClaimOptions(m)
	if len(options) == 0 {
		return m, nil
	}
	if options[0].Kind == game.ClaimChow {
		switch key.Type {
		case tea.KeyLeft:
			if m.ClaimOptionIndex > 0 {
				m.ClaimOptionIndex--
			}
			return m, nil
		case tea.KeyRight:
			if m.ClaimOptionIndex < len(options)-1 {
				m.ClaimOptionIndex++
			}
			return m, nil
		}
	}
	if key.String() == "q" {
		if m.Online {
			return clearOnlineStateForMenu(m), nil
		}
		m.Game.Quit("quit")
		m.Screen = ScreenGameOver
		return m, nil
	}
	command, ok := claimCommandForKey(m, key.String())
	if !ok {
		return m, nil
	}
	if m.Online {
		if m.OnlineClient == nil {
			m.StatusLine = "Online room is not ready"
			return m, nil
		}
		m.StatusLine = claimStatus(command, true)
		return m, sendOnlineGameCommandCmd(m.OnlineClient, command)
	}
	command.PlayerID = "0"
	var result game.CommandResult
	m, result = applyLocalCommand(m, command)
	if !result.OK {
		m.StatusLine = result.Error
		return m, nil
	}
	m.StatusLine = claimStatus(command, false)
	m.ClaimOptionIndex = 0
	m = advanceLocalAI(m)
	return finishLocalUpdate(m)
}

func applyLocalCommand(m Model, command game.GameCommand) (Model, game.CommandResult) {
	command.PlayerID = "0"
	if m.LocalMatch != nil {
		result := m.LocalMatch.ApplyCommand(command)
		return syncLocalRound(m), result
	}
	return m, m.Game.ApplyCommand(command)
}

func advanceLocalAI(m Model) Model {
	if m.LocalMatch != nil {
		m.LocalMatch.AdvanceAIUntilHumanTurn()
		return syncLocalRound(m)
	}
	if m.Game != nil {
		m.Game.AdvanceAIUntilHumanTurn()
	}
	return m
}

func finishLocalUpdate(m Model) (tea.Model, tea.Cmd) {
	if m.LocalMatch != nil {
		if m.LocalMatch.Complete {
			m.Screen = ScreenGameOver
			return m, saveCompletedReplayCmd(m.LocalMatch, m.ReplayDir)
		}
		return m, nil
	}
	if m.Game != nil && m.Game.Over {
		m.Screen = ScreenGameOver
	}
	return m, nil
}

func claimCommandForKey(m Model, key string) (game.GameCommand, bool) {
	options := activeClaimOptions(m)
	if len(options) == 0 {
		return game.GameCommand{}, false
	}
	command := game.GameCommand{Kind: game.CommandPass}
	switch key {
	case " ", "esc":
		return command, true
	case "h":
		if options[0].Kind == game.ClaimWin {
			command.Kind = game.CommandClaimWin
			return command, true
		}
	case "p":
		if options[0].Kind == game.ClaimPong {
			command.Kind = game.CommandPong
			return command, true
		}
	case "c":
		if options[0].Kind == game.ClaimChow && m.ClaimOptionIndex >= 0 && m.ClaimOptionIndex < len(options) {
			command.Kind = game.CommandChow
			command.TileIndex = m.ClaimOptionIndex
			return command, true
		}
	}
	return game.GameCommand{}, false
}

func isClaimResponse(m Model) bool {
	if m.Online {
		return m.OnlineSnapshot.Phase == game.PhaseAwaitingClaim && m.OnlineSnapshot.Current == m.OnlineSeat && m.OnlineSnapshot.PendingClaim != nil
	}
	return m.Game != nil && m.Game.Phase == game.PhaseAwaitingClaim && m.Game.Current == 0 && m.Game.PendingClaim != nil
}

func activeClaimOptions(m Model) []game.ClaimOption {
	var pending *game.PendingClaim
	if m.Online {
		pending = m.OnlineSnapshot.PendingClaim
	} else if m.Game != nil {
		pending = m.Game.PendingClaim
	}
	if pending == nil || pending.Active < 0 || pending.Active >= len(pending.Options) {
		return nil
	}
	first := pending.Options[pending.Active]
	end := pending.Active + 1
	for end < len(pending.Options) {
		option := pending.Options[end]
		if option.Player != first.Player || option.Kind != first.Kind {
			break
		}
		end++
	}
	return pending.Options[pending.Active:end]
}

func claimStatus(command game.GameCommand, sending bool) string {
	prefix := "Claimed "
	if sending {
		prefix = "Claiming "
	}
	switch command.Kind {
	case game.CommandPass:
		if sending {
			return "Passing claim"
		}
		return "Passed claim"
	case game.CommandClaimWin:
		return prefix + "win"
	case game.CommandPong:
		return prefix + "pong"
	case game.CommandChow:
		return prefix + "chow"
	default:
		return ""
	}
}

func selectionStatus(m Model, action string, index int, tile game.Tile, unicode bool) string {
	return fmt.Sprintf("%s [%02d] %s (%s)", actionVerb(m, action), index+1, game.TileLabel(tile, unicode), tile.String())
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
			m = restartLocalMatch(m)
			m.Screen = ScreenTable
			m.SelectedIndex = 0
			m.GameOverIndex = 0
		case 1:
			m.Game = nil
			m.LocalMatch = nil
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
		m = restartLocalMatch(m)
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
	m.LocalMatch = nil
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
