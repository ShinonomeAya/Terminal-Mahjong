package tui

import (
	"strings"
	"testing"

	"mahjong/internal/game"
)

func TestSharedTableStateUsesRecipientSnapshot(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGameWithRules(game.ModeRiichi, game.DefaultRuleConfig(game.ModeRiichi))

	state := tableStateFor(model)

	if state.Mode != game.ModeRiichi || state.ViewerSeat != 0 || state.Online {
		t.Fatalf("local table state = %#v", state)
	}
	if len(state.Snapshot.Players) != 4 || len(state.Snapshot.Players[0].Hand) == 0 {
		t.Fatalf("local recipient snapshot = %#v", state.Snapshot.Players)
	}
}

func TestWideTableSkeletonContainsApprovedRegions(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGameWithRules(game.ModeRiichi, game.DefaultRuleConfig(game.ModeRiichi))
	model.Screen = ScreenTable
	model.Width = 140

	view := renderWideTable(model, tableStateFor(model))

	for _, want := range []string{"日麻", "上家", "对家", "下家", "牌桌", "手牌", "操作", "战术分析"} {
		if !strings.Contains(view, want) {
			t.Fatalf("wide table missing %q:\n%s", want, view)
		}
	}
	if strings.Count(view, "战术分析") != 1 {
		t.Fatalf("wide table tactical rail count = %d:\n%s", strings.Count(view, "战术分析"), view)
	}
}
