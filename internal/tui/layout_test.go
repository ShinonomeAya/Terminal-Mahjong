package tui

import (
	"strings"
	"testing"

	"mahjong/internal/game"
)

func TestTileCellRendersUnselectedUnicodeTile(t *testing.T) {
	tile := mustUITiles(t, "2m")[0]

	cell := renderTileCell(1, tile, false, true)

	if strings.Contains(cell, "[02]") || !strings.Contains(cell, "🀈") {
		t.Fatalf("cell = %q, want unicode glyph without visible index", cell)
	}
}

func TestTileCellRendersSelectedUnicodeTile(t *testing.T) {
	tile := mustUITiles(t, "2m")[0]

	cell := renderTileCell(1, tile, true, true)

	if strings.Contains(cell, "[02]") || !strings.Contains(cell, "▶") || !strings.Contains(cell, "🀈") || !strings.Contains(cell, "◀") {
		t.Fatalf("selected cell = %q, want selected markers", cell)
	}
}

func TestTileCellRendersFallbackNotation(t *testing.T) {
	tile := mustUITiles(t, "2m")[0]

	cell := renderTileCell(1, tile, false, false)

	if !strings.Contains(cell, "2m") {
		t.Fatalf("fallback cell = %q, want 2m", cell)
	}
}

func TestTileCellStaysWithinWidthBudget(t *testing.T) {
	tile := mustUITiles(t, "2m")[0]

	cell := renderTileCell(1, tile, true, true)

	if got := runeWidth(cell); got > handCellW {
		t.Fatalf("cell width = %d, want <= %d: %q", got, handCellW, cell)
	}
}

func TestRenderTableAddsChineseLabels(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable

	view := renderTable(model)

	for _, text := range []string{"终端麻将", "对手", "牌桌", "手牌"} {
		if !strings.Contains(view, text) {
			t.Fatalf("view missing Chinese label %q:\n%s", text, view)
		}
	}
}

func TestRenderHandShowsTilesInOneRowWithoutIndices(t *testing.T) {
	hand := mustUITiles(t, "1m", "2m", "3m", "4m", "5m", "6m", "7m", "8m", "9m", "1p", "2p", "3p", "4p", "5p")

	view := renderHand(hand, 1, true)

	if strings.Contains(view, "[01]") || strings.Contains(view, "[14]") {
		t.Fatalf("hand should not show numeric tile prefixes:\n%s", view)
	}
	if strings.Count(view, "🀇") != 1 || !strings.Contains(view, "🀝") {
		t.Fatalf("hand missing unicode mahjong tiles:\n%s", view)
	}
	if strings.Count(view, "Hand:") != 1 {
		t.Fatalf("hand should render as one row:\n%s", view)
	}
}

func TestRenderTableIncludesUnicodeTiles(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Game.Players[0].Hand = mustUITiles(t, "1m", "1p", "1s")
	model.Screen = ScreenTable
	model.UnicodeTiles = true

	view := renderTable(model)

	if !strings.Contains(view, "🀇") || !strings.Contains(view, "🀙") || !strings.Contains(view, "🀐") {
		t.Fatalf("view does not appear to include unicode tiles:\n%s", view)
	}
}

func TestRenderTableIncludesFallbackLabels(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Game.Players[0].Hand = mustUITiles(t, "1m", "2m", "E")
	model.Screen = ScreenTable
	model.UnicodeTiles = false

	view := renderTable(model)

	if !strings.Contains(view, "1m") || !strings.Contains(view, "2m") || !strings.Contains(view, "E") {
		t.Fatalf("view missing fallback labels:\n%s", view)
	}
}

func TestRenderTableIncludesRecentEventsPanel(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Game.Events = nil
	model.Game.RecordEvent(game.EventDraw, 0, mustUITiles(t, "1m")[0], "")
	model.Game.RecordEvent(game.EventDiscard, 0, mustUITiles(t, "1m")[0], "")
	model.Game.RecordEvent(game.EventDraw, 1, mustUITiles(t, "2m")[0], "")
	model.Game.RecordEvent(game.EventDiscard, 1, mustUITiles(t, "2m")[0], "")
	model.Screen = ScreenTable
	model.UnicodeTiles = false

	view := renderTable(model)

	for _, text := range []string{"Recent Events", "01. You draw 1m", "04. AI-1 discard 2m"} {
		if !strings.Contains(view, text) {
			t.Fatalf("view missing recent event text %q:\n%s", text, view)
		}
	}
	if strings.Contains(view, "Last:") {
		t.Fatalf("view still uses one-line Last summary:\n%s", view)
	}
}

func TestRenderTableUsesUnicodeInRecentEvents(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Game.Events = nil
	model.Game.RecordEvent(game.EventDiscard, 0, mustUITiles(t, "1m")[0], "")
	model.Screen = ScreenTable
	model.UnicodeTiles = true

	view := renderTable(model)

	if !strings.Contains(view, "01. You discard 🀇") {
		t.Fatalf("view missing unicode event tile:\n%s", view)
	}
}

func TestRenderTableIncludesRoundStatusPanel(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Game.Events = nil
	model.Game.RecordEvent(game.EventDraw, 0, mustUITiles(t, "1m")[0], "")
	model.Screen = ScreenTable

	view := renderTable(model)

	for _, text := range []string{"Round Status", "Wall:", "Turn: You", "Events: 1"} {
		if !strings.Contains(view, text) {
			t.Fatalf("view missing round status text %q:\n%s", text, view)
		}
	}
}

func TestRenderTableShowsNetworkStatus(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable
	model.NetworkStatus = NetworkReconnecting
	model.ReconnectAttempt = 2
	model.ReconnectMax = 5

	view := renderTable(model)

	if !strings.Contains(view, "Network: reconnecting 2/5") {
		t.Fatalf("view missing reconnecting status:\n%s", view)
	}
}

func TestRenderTableDefaultsToLocalNetworkStatus(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable

	view := renderTable(model)

	if !strings.Contains(view, "Network: local") {
		t.Fatalf("view missing local network status:\n%s", view)
	}
}

func TestRenderTableLimitsRecentEventsPanel(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Game.Events = nil
	for _, tile := range mustUITiles(t, "1m", "2m", "3m", "4m", "5m") {
		model.Game.RecordEvent(game.EventDiscard, 0, tile, "")
	}
	model.Screen = ScreenTable
	model.UnicodeTiles = false

	view := renderTable(model)

	if strings.Contains(view, "01. You discard 1m") {
		t.Fatalf("view should trim oldest event from compact event panel:\n%s", view)
	}
	if !strings.Contains(view, "02. You discard 2m") || !strings.Contains(view, "05. You discard 5m") {
		t.Fatalf("view missing recent tail events:\n%s", view)
	}
}

func TestRenderTableMarksSelectedTile(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable
	model.SelectedIndex = 2

	view := renderTable(model)

	if !strings.Contains(view, "Focus:") {
		t.Fatalf("view missing focus marker:\n%s", view)
	}
}

func TestRenderTableShowsSelectedTileInHandRow(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Game.Players[0].Hand = mustUITiles(t, "1m", "2m", "3m")
	model.Screen = ScreenTable
	model.SelectedIndex = 1

	view := renderTable(model)

	if strings.Contains(view, "▶ [02]") || !strings.Contains(view, "▶") || !strings.Contains(view, "Focus: [02]") {
		t.Fatalf("view does not clearly show selected tile:\n%s", view)
	}
}

func TestRenderTableShowsHandTrayFocus(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Game.Players[0].Hand = mustUITiles(t, "1m", "2m", "3m")
	model.Screen = ScreenTable
	model.SelectedIndex = 1

	view := renderTable(model)

	for _, text := range []string{"Hand Tray", "Focus: [02]", "▲"} {
		if !strings.Contains(view, text) {
			t.Fatalf("view missing hand tray focus text %q:\n%s", text, view)
		}
	}
}

func TestRenderTableShowsWaitingActionFeedback(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable

	view := renderTable(model)

	if !strings.Contains(view, "Last Action: Waiting for input") {
		t.Fatalf("view missing waiting action feedback:\n%s", view)
	}
}

func TestRenderTableShowsActionBarNearHandTray(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable

	view := renderTable(model)

	for _, text := range []string{"Actions:", "[Enter/Space] Discard", "[Click] Tile", "[H] Win:off", "[K] Kong:off", "[Q] Quit"} {
		if !strings.Contains(view, text) {
			t.Fatalf("view missing action bar text %q:\n%s", text, view)
		}
	}
	if lineIndexContaining(view, "Actions:") <= lineIndexContaining(view, "Hand:") {
		t.Fatalf("action bar should sit below the hand tray:\n%s", view)
	}
}

func TestRenderTableHighlightsReadyWinAction(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Game.Players[0].Hand = mustUITiles(t,
		"1m", "2m", "3m",
		"4m", "5m", "6m",
		"2p", "3p", "4p",
		"7s", "7s", "7s",
		"E", "E",
	)
	model.Screen = ScreenTable

	view := renderTable(model)

	if !strings.Contains(view, "[H] Win:READY") {
		t.Fatalf("view missing ready win action:\n%s", view)
	}
}

func TestRenderTableHighlightsReadyKongAction(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Game.Players[0].Hand = mustUITiles(t, "1m", "1m", "1m", "1m", "2m", "3m")
	model.Screen = ScreenTable

	view := renderTable(model)

	if !strings.Contains(view, "[K] Kong:READY") {
		t.Fatalf("view missing ready kong action:\n%s", view)
	}
}

func TestReadyActionStateKeepsStableVisibleWidth(t *testing.T) {
	action := actionState("[H] Win", true)

	if !strings.Contains(action, "[H] Win:READY") {
		t.Fatalf("ready action missing label: %q", action)
	}
	if got := visibleWidth(action); got != visibleWidth("[H] Win:READY") {
		t.Fatalf("ready action width = %d, want plain width", got)
	}
}

func TestRenderTableUsesCompactActionBarForNarrowWidth(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable
	model.Width = 64

	view := renderTable(model)

	if !strings.Contains(view, "[Enter] Discard") || strings.Contains(view, "[Enter/Space] Discard") {
		t.Fatalf("view should use compact action bar:\n%s", view)
	}
	for _, line := range strings.Split(strings.TrimRight(view, "\n"), "\n") {
		if strings.Contains(line, "Actions:") && visibleWidth(line) > 80 {
			t.Fatalf("compact action bar too wide (%d cells):\n%s", visibleWidth(line), line)
		}
	}
}

func TestRenderTableKeepsReadableLineWidth(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable

	view := renderTable(model)

	for _, line := range strings.Split(strings.TrimRight(view, "\n"), "\n") {
		if visibleWidth(line) > 96 {
			t.Fatalf("line too wide (%d cells):\n%s", visibleWidth(line), line)
		}
	}
}

func TestRenderTableUsesCompactControlsForNarrowWidth(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable
	model.Width = 64

	view := renderTable(model)

	if !strings.Contains(view, "Arrows select | Enter discard | Click tile | Q quit") {
		t.Fatalf("view missing compact controls:\n%s", view)
	}
	for _, line := range strings.Split(strings.TrimRight(view, "\n"), "\n") {
		if visibleWidth(line) > 80 {
			t.Fatalf("compact line too wide (%d cells):\n%s", visibleWidth(line), line)
		}
	}
}

func TestRenderTableUsesClientSections(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable

	view := renderTable(model)

	for _, text := range []string{"TERMINAL MAHJONG", "Opponents", "Table", "You", "Controls"} {
		if !strings.Contains(view, text) {
			t.Fatalf("view missing section %q:\n%s", text, view)
		}
	}
}

func TestRenderTableShowsAllOpponentsByName(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable

	view := renderTable(model)

	for _, text := range []string{"AI-1", "AI-2", "AI-3"} {
		if !strings.Contains(view, text) {
			t.Fatalf("view missing opponent %q:\n%s", text, view)
		}
	}
}

func TestRenderTableArrangesOpponentsAsSeats(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable

	view := renderTable(model)

	for _, text := range []string{"Opposite: AI-2", "Left: AI-1", "Right: AI-3"} {
		if !strings.Contains(view, text) {
			t.Fatalf("view missing seat label %q:\n%s", text, view)
		}
	}
	if !lineContainsAll(view, "Left: AI-1", "Right: AI-3") {
		t.Fatalf("left and right seats should share a table row:\n%s", view)
	}
}

func TestRenderTableUsesReferenceInspiredTabletop(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable
	model.Width = 120

	view := renderTable(model)

	for _, want := range []string{"TERMINAL MAHJONG", "AI-2", "AI-1", "AI-3", "CENTER", "Hand Tray"} {
		if !strings.Contains(view, want) {
			t.Fatalf("view missing %q:\n%s", want, view)
		}
	}
	if lineIndexContaining(view, "CENTER") <= lineIndexContaining(view, "AI-2") {
		t.Fatalf("center should appear below opposite seat:\n%s", view)
	}
	if lineIndexContaining(view, "Hand Tray") <= lineIndexContaining(view, "CENTER") {
		t.Fatalf("hand tray should appear below center table:\n%s", view)
	}
}

func TestRenderTableCentersMainBoardWhenWide(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable
	model.Width = 120

	view := renderTable(model)

	line := firstLineContaining(view, "TERMINAL MAHJONG")
	if !strings.HasPrefix(line, " ") {
		t.Fatalf("wide title should be centered with leading space:\n%s", view)
	}
}

func TestRenderTableKeepsReferenceInspiredWidth(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable
	model.Width = 120

	view := renderTable(model)

	for _, line := range strings.Split(strings.TrimRight(view, "\n"), "\n") {
		if visibleWidth(line) > 120 {
			t.Fatalf("line too wide (%d cells):\n%s", visibleWidth(line), line)
		}
	}
}

func TestRenderTableSplitsFullHandIntoTwoRows(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable

	view := renderTable(model)

	if strings.Count(view, "Hand:") != 1 {
		t.Fatalf("view should keep one stable hand row:\n%s", view)
	}
}

func lineContainsAll(text string, parts ...string) bool {
	for _, line := range strings.Split(text, "\n") {
		matches := true
		for _, part := range parts {
			if !strings.Contains(line, part) {
				matches = false
				break
			}
		}
		if matches {
			return true
		}
	}
	return false
}

func TestHandHitBoxesFindTileIndex(t *testing.T) {
	boxes := handHitBoxes(3, 2, 4)
	index, ok := tileIndexAt(boxes, boxes[1].X1, boxes[1].Y)
	if !ok {
		t.Fatal("expected hit")
	}
	if index != 1 {
		t.Fatalf("index = %d, want 1", index)
	}
}

func TestDefaultHandHitBoxesMatchRenderedHandRows(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	boxes := currentHandHitBoxes(model)

	if boxes[0].Y != handRowY {
		t.Fatalf("first row y = %d, want %d", boxes[0].Y, handRowY)
	}
	if boxes[7].Y != handRowY {
		t.Fatalf("eighth tile y = %d, want same hand row %d", boxes[7].Y, handRowY)
	}
}

func TestDefaultHandHitBoxesMatchRenderedHandStartLine(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable

	view := renderTable(model)

	if got := lineIndexContaining(view, "Hand:"); got != handRowY {
		t.Fatalf("rendered hand row = %d, want hitbox row %d:\n%s", got, handRowY, view)
	}
}

func lineIndexContaining(text string, part string) int {
	for index, line := range strings.Split(strings.TrimRight(text, "\n"), "\n") {
		if strings.Contains(line, part) {
			return index
		}
	}
	return -1
}

func firstLineContaining(text string, part string) string {
	for _, line := range strings.Split(strings.TrimRight(text, "\n"), "\n") {
		if strings.Contains(line, part) {
			return line
		}
	}
	return ""
}
