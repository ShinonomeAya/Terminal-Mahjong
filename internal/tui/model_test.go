package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewModelStartsAtMenu(t *testing.T) {
	model := NewModel()
	if model.Screen != ScreenMenu {
		t.Fatalf("screen = %v, want menu", model.Screen)
	}
	if model.MenuIndex != 0 {
		t.Fatalf("menu index = %d, want 0", model.MenuIndex)
	}
}

func TestMenuDownMovesSelection(t *testing.T) {
	model := NewModel()
	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	updated := next.(Model)
	if updated.MenuIndex != 1 {
		t.Fatalf("menu index = %d, want 1", updated.MenuIndex)
	}
}

func TestMenuEnterStartsSoloGame(t *testing.T) {
	model := NewModel()
	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := next.(Model)
	if updated.Screen != ScreenTable {
		t.Fatalf("screen = %v, want table", updated.Screen)
	}
	if updated.Game == nil {
		t.Fatal("expected game to be created")
	}
}

func TestMenuViewContainsOptions(t *testing.T) {
	view := NewModel().View()
	for _, text := range []string{"TERMINAL MAHJONG", "Solo Game", "How to Play", "Quit"} {
		if !strings.Contains(view, text) {
			t.Fatalf("view missing %q:\n%s", text, view)
		}
	}
}

func TestMenuViewUsesReadableSections(t *testing.T) {
	view := NewModel().View()
	for _, text := range []string{"Menu", "Controls", "Up/Down choose"} {
		if !strings.Contains(view, text) {
			t.Fatalf("menu missing %q:\n%s", text, view)
		}
	}
}

func TestHelpViewContainsKeyboardAndMouseControls(t *testing.T) {
	model := NewModel()
	model.Screen = ScreenHelp

	view := model.View()

	for _, text := range []string{"Keyboard", "Mouse", "Enter/Space", "Second click"} {
		if !strings.Contains(view, text) {
			t.Fatalf("help missing %q:\n%s", text, view)
		}
	}
}

func TestTableRightMovesSelectedTile(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyRight})
	updated := next.(Model)

	if updated.SelectedIndex != 1 {
		t.Fatalf("selected index = %d, want 1", updated.SelectedIndex)
	}
}

func TestTableLeftAtFirstTileStaysAtFirstTile(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyLeft})
	updated := next.(Model)

	if updated.SelectedIndex != 0 {
		t.Fatalf("selected index = %d, want 0", updated.SelectedIndex)
	}
}

func TestTableEnterDiscardsSelectedTile(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable
	model.SelectedIndex = 0
	startEvents := len(model.Game.Events)

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := next.(Model)

	if len(updated.Game.Players[0].Discards) != 1 {
		t.Fatalf("discards = %d, want 1", len(updated.Game.Players[0].Discards))
	}
	if len(updated.Game.Events) <= startEvents+1 {
		t.Fatalf("events = %d, want AI turns after human discard", len(updated.Game.Events))
	}
	if !updated.Game.Over && updated.Game.Current != 0 {
		t.Fatalf("current = %d, want human turn after AI advance", updated.Game.Current)
	}
}

func TestMouseClickSelectsTile(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable
	model.HandHitBoxes = handHitBoxes(len(model.Game.Players[0].Hand), 2, 10)

	next, _ := model.Update(tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
		X:      model.HandHitBoxes[2].X1,
		Y:      model.HandHitBoxes[2].Y,
	})
	updated := next.(Model)

	if updated.SelectedIndex != 2 {
		t.Fatalf("selected index = %d, want 2", updated.SelectedIndex)
	}
	if !strings.Contains(updated.StatusLine, "Mouse selected [03]") {
		t.Fatalf("status line = %q, want mouse selection feedback", updated.StatusLine)
	}
}

func TestSecondMouseClickDiscardsSelectedTile(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable
	model.SelectedIndex = 2
	model.HandHitBoxes = handHitBoxes(len(model.Game.Players[0].Hand), 2, 10)

	next, _ := model.Update(tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
		X:      model.HandHitBoxes[2].X1,
		Y:      model.HandHitBoxes[2].Y,
	})
	updated := next.(Model)

	if len(updated.Game.Players[0].Discards) != 1 {
		t.Fatalf("discards = %d, want 1", len(updated.Game.Players[0].Discards))
	}
	if !strings.Contains(updated.StatusLine, "Discarded [03]") {
		t.Fatalf("status line = %q, want discard feedback", updated.StatusLine)
	}
}

func TestKeyboardSelectionUpdatesStatusLine(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyRight})
	updated := next.(Model)

	if !strings.Contains(updated.StatusLine, "Selected [02]") {
		t.Fatalf("status line = %q, want keyboard selection feedback", updated.StatusLine)
	}
}

func TestTableViewIncludesStatusLine(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable
	model.StatusLine = "Mouse selected [04] 🀊 (4m)"

	view := model.View()

	if !strings.Contains(view, "Status: Mouse selected [04]") {
		t.Fatalf("view missing status line:\n%s", view)
	}
}

func TestGameOverEnterMainMenuReturnsToMenu(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenGameOver
	model.GameOverIndex = 1

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := next.(Model)

	if updated.Screen != ScreenMenu {
		t.Fatalf("screen = %v, want menu", updated.Screen)
	}
}

func TestGameOverEnterRestartStartsNewGame(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenGameOver
	model.GameOverIndex = 0

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := next.(Model)

	if updated.Screen != ScreenTable {
		t.Fatalf("screen = %v, want table", updated.Screen)
	}
	if updated.Game == nil || len(updated.Game.Events) == 0 {
		t.Fatal("expected restarted game with initial draw event")
	}
}

func TestGameOverViewContainsChoicesAndControls(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenGameOver

	view := model.View()

	for _, text := range []string{"GAME OVER", "Restart", "Main Menu", "Quit", "Controls"} {
		if !strings.Contains(view, text) {
			t.Fatalf("game over missing %q:\n%s", text, view)
		}
	}
}
