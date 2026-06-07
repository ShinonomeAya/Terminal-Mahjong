package tui

import (
	"strings"
	"testing"

	"mahjong/internal/game"
)

func TestTileCellRendersUnselectedUnicodeTile(t *testing.T) {
	tile := mustUITiles(t, "2m")[0]

	cell := renderTileCell(1, tile, false, true)

	if !strings.Contains(cell, "[02]") || !strings.Contains(cell, "🀈") {
		t.Fatalf("cell = %q, want index and unicode glyph", cell)
	}
}

func TestTileCellRendersSelectedUnicodeTile(t *testing.T) {
	tile := mustUITiles(t, "2m")[0]

	cell := renderTileCell(1, tile, true, true)

	if !strings.Contains(cell, "▶ [02]") || !strings.Contains(cell, "◀") {
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

	if !strings.Contains(view, "Selected:") {
		t.Fatalf("view missing selected marker:\n%s", view)
	}
}

func TestRenderTableShowsSelectedTileInHandRow(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Game.Players[0].Hand = mustUITiles(t, "1m", "2m", "3m")
	model.Screen = ScreenTable
	model.SelectedIndex = 1

	view := renderTable(model)

	if !strings.Contains(view, "▶ [02]") || !strings.Contains(view, "Selected: [02]") {
		t.Fatalf("view does not clearly show selected tile:\n%s", view)
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

func TestRenderTableSplitsFullHandIntoTwoRows(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable

	view := renderTable(model)

	if strings.Count(view, "Hand:") != 1 || !strings.Contains(view, "\n      ") {
		t.Fatalf("view does not split hand into stable rows:\n%s", view)
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
	if boxes[7].Y != handRowY+1 {
		t.Fatalf("second row y = %d, want %d", boxes[7].Y, handRowY+1)
	}
}
