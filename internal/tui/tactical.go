package tui

import (
	"fmt"
	"strings"

	"mahjong/internal/game"
)

type tacticalView struct {
	Shanten      int
	Effective    []game.Tile
	Improvements []game.TileImprovement
	Legal        []game.CommandKind
	Events       []game.GameEvent
	ModeStatus   string
}

func tacticalViewFor(m Model, state tableViewState) tacticalView {
	if state.ViewerSeat < 0 || state.ViewerSeat >= len(state.Snapshot.Players) {
		return tacticalView{Shanten: 0}
	}
	hand := append([]game.Tile(nil), state.Snapshot.Players[state.ViewerSeat].Hand...)
	view := tacticalView{Shanten: game.ShantenStandard(hand)}
	if len(hand)%3 == 1 {
		view.Effective = game.EffectiveTiles(hand)
	} else if len(hand)%3 == 2 {
		view.Improvements = game.ImprovementTiles(hand)
		if len(view.Improvements) > 0 {
			view.Effective = append([]game.Tile(nil), view.Improvements[0].Effective...)
		}
	}
	seen := make(map[game.CommandKind]bool)
	for _, action := range state.Snapshot.LegalActions {
		if !seen[action.Kind] {
			seen[action.Kind] = true
			view.Legal = append(view.Legal, action.Kind)
		}
	}
	start := len(state.Snapshot.Events) - 5
	if start < 0 {
		start = 0
	}
	view.Events = append([]game.GameEvent(nil), state.Snapshot.Events[start:]...)
	view.ModeStatus = tacticalModeStatus(m, state)
	return view
}

func renderTacticalRail(m Model, view tacticalView) string {
	effective := game.FormatTileLabels(view.Effective, m.UnicodeTiles)
	if effective == "" {
		effective = "-"
	}
	improvements := tacticalImprovementLines(m, view.Improvements, 3)
	if len(improvements) == 0 {
		improvements = []string{"-"}
	}
	events := make([]string, 0, len(view.Events))
	for _, event := range view.Events {
		events = append(events, formatEventLine(m, event, m.UnicodeTiles))
	}
	if len(events) == 0 {
		events = append(events, "-")
	}
	if m.chinese() {
		body := fmt.Sprintf("向听：%d\n有效牌\n%s\n改良牌\n%s\n\n规则状态\n%s\n\n最近事件\n%s",
			view.Shanten,
			effective,
			strings.Join(improvements, "\n"),
			emptyTacticalText(view.ModeStatus),
			strings.Join(events, "\n"),
		)
		return stylePanelWidth("战术分析", body, tacticalRailWidth)
	}
	body := fmt.Sprintf("Shanten: %d\nEffective\n%s\nImprovements\n%s\n\nRule Status\n%s\n\nRecent Events\n%s",
		view.Shanten,
		effective,
		strings.Join(improvements, "\n"),
		emptyTacticalText(view.ModeStatus),
		strings.Join(events, "\n"),
	)
	return stylePanelWidth("Tactical Analysis", body, tacticalRailWidth)
}

func tacticalImprovementLines(m Model, improvements []game.TileImprovement, limit int) []string {
	if len(improvements) < limit {
		limit = len(improvements)
	}
	lines := make([]string, 0, limit)
	for _, improvement := range improvements[:limit] {
		discard := game.TileLabel(improvement.Discard, m.UnicodeTiles)
		effective := game.FormatTileLabels(improvement.Effective, m.UnicodeTiles)
		lines = append(lines, discard+" > "+effective)
	}
	return lines
}

func tacticalModeStatus(m Model, state tableViewState) string {
	if state.Snapshot.Riichi != nil {
		if m.chinese() {
			return fmt.Sprintf("宝牌:%s  本场:%d", game.FormatTileLabels(state.Snapshot.Riichi.DoraIndicators, m.UnicodeTiles), state.Snapshot.Riichi.Honba)
		}
		return fmt.Sprintf("Dora:%s  Honba:%d", game.FormatTileLabels(state.Snapshot.Riichi.DoraIndicators, m.UnicodeTiles), state.Snapshot.Riichi.Honba)
	}
	if state.Mode == game.ModeMCR {
		flowers := len(state.Snapshot.Players[state.ViewerSeat].Flowers)
		if m.chinese() {
			return fmt.Sprintf("花牌:%d", flowers)
		}
		return fmt.Sprintf("Flowers:%d", flowers)
	}
	return ruleModeName(m, state.Mode)
}

func emptyTacticalText(text string) string {
	if text == "" {
		return "-"
	}
	return text
}
