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

	if !strings.Contains(view, "selected") && !strings.Contains(view, "▲") {
		t.Fatalf("view missing selected marker:\n%s", view)
	}
}
