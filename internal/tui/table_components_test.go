package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

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

func TestWideTableSeatShowsPublicModeStateAndActiveTurn(t *testing.T) {
	match, err := game.NewMatch(31, game.NewRiichiRuleSet(game.DefaultRuleConfig(game.ModeRiichi).Riichi))
	if err != nil {
		t.Fatal(err)
	}
	snapshot := match.Round.SnapshotFor("0")
	snapshot.Current = 1
	snapshot.Players[1].Melds = []game.Meld{{Kind: game.MeldPong, Tiles: mustUITiles(t, "1m", "1m", "1m")}}
	snapshot.Riichi.Declarations[1] = game.RiichiAccepted
	model := NewModel()
	model.Online = true
	model.OnlineStarted = true
	model.OnlineSeat = 0
	model.OnlineSnapshot = snapshot
	model.OnlineMatch = match.SnapshotFor("0")
	model.OnlineMatch.Points = [4]int{25000, 24000, 26000, 25000}
	model.Width = 140

	view := renderWideTable(model, tableStateFor(model))

	for _, want := range []string{"24000", "副露:1", "立直", ">"} {
		if !strings.Contains(view, want) {
			t.Fatalf("wide seat missing %q:\n%s", want, view)
		}
	}
}

func TestDiscardRiverUsesStableSixTileRowsAndLatestMarker(t *testing.T) {
	model := NewModel()
	tiles := mustUITiles(t, "1m", "2m", "3m", "4m", "5m", "6m", "7m")

	river := renderDiscardRiver(model, tiles)
	lines := strings.Split(river, "\n")

	if len(lines) != 2 {
		t.Fatalf("river rows = %d, want 2:\n%s", len(lines), river)
	}
	if !strings.Contains(lines[1], "◆") {
		t.Fatalf("latest discard marker missing:\n%s", river)
	}
}

func TestTacticalFallbackMovesBelowAtMediumWidth(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGameWithRules(game.ModeRiichi, game.DefaultRuleConfig(game.ModeRiichi))
	model.Screen = ScreenTable
	model.Width = 90

	view := renderTable(model)

	if !strings.Contains(view, "战术分析") || lineIndexContaining(view, "战术分析") <= lineIndexContaining(view, "手牌托盘") {
		t.Fatalf("medium tactical rail should appear below table:\n%s", view)
	}
}

func TestTacticalFallbackHidesWhenMediumTerminalIsTooShort(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGameWithRules(game.ModeRiichi, game.DefaultRuleConfig(game.ModeRiichi))
	model.Screen = ScreenTable
	model.Width = 80
	model.Height = 42

	view := renderTable(model)

	if strings.Contains(view, "战术分析") {
		t.Fatalf("medium tactical rail should be hidden when it exceeds terminal height:\n%s", view)
	}
	if lines := strings.Count(strings.TrimRight(view, "\n"), "\n") + 1; lines > model.Height {
		t.Fatalf("medium table rendered %d lines, want at most %d:\n%s", lines, model.Height, view)
	}
}

func TestTabTogglesTacticalRailAtCompactWidth(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGameWithRules(game.ModeRiichi, game.DefaultRuleConfig(game.ModeRiichi))
	model.Screen = ScreenTable
	model.Width = 64
	if strings.Contains(renderTable(model), "战术分析") {
		t.Fatal("compact tactical rail should be hidden by default")
	}

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	updated := next.(Model)

	if !updated.ShowTactical || !strings.Contains(renderTable(updated), "战术分析") {
		t.Fatalf("Tab did not reveal compact tactical rail:\n%s", renderTable(updated))
	}
}
