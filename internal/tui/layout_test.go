package tui

import (
	"strings"
	"testing"
)

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
		if len([]rune(line)) > 96 {
			t.Fatalf("line too wide (%d runes):\n%s", len([]rune(line)), line)
		}
	}
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
