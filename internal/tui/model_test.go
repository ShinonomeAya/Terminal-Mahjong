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
