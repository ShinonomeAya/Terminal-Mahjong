package tui

import (
	"strings"
	"testing"

	"mahjong/internal/game"
)

func TestTacticalViewUsesOnlyViewerHandAndBoundsEvents(t *testing.T) {
	snapshot := game.NewGame(17).SnapshotFor("0")
	snapshot.Players[0].Hand = mustUITiles(t, "1m", "2m", "3m", "4m", "5m", "6m", "7p", "8p", "9p", "2s", "2s", "3s", "4s")
	snapshot.Players[0].HandCount = len(snapshot.Players[0].Hand)
	snapshot.Players[1].Hand = nil
	for index := 0; index < 8; index++ {
		snapshot.Events = append(snapshot.Events, game.GameEvent{Kind: game.EventDiscard, Player: index % 4, Tile: game.Tile(index)})
	}
	model := NewModel()
	model.Online = true
	model.OnlineSeat = 0
	model.OnlineSnapshot = snapshot

	view := tacticalViewFor(model, tableStateFor(model))

	if view.Shanten != 0 || game.FormatTiles(view.Effective) != "2s 5s" {
		t.Fatalf("tactical view = %#v", view)
	}
	if len(view.Events) != 5 {
		t.Fatalf("events = %d, want 5", len(view.Events))
	}
}

func TestTacticalRailRendersLocalizedAnalysis(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGameWithRules(game.ModeRiichi, game.DefaultRuleConfig(game.ModeRiichi))
	model.Game.Players[0].Hand = mustUITiles(t, "1m", "2m", "3m", "4m", "5m", "6m", "7p", "8p", "9p", "2s", "2s", "3s", "4s")

	rail := renderTacticalRail(model, tacticalViewFor(model, tableStateFor(model)))

	effectiveTile := game.TileLabel(mustUITiles(t, "2s")[0], true)
	for _, want := range []string{"战术分析", "向听：0", "有效牌", effectiveTile, "最近事件"} {
		if !strings.Contains(rail, want) {
			t.Fatalf("tactical rail missing %q:\n%s", want, rail)
		}
	}
}
