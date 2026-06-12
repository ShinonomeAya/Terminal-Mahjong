package tui

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"mahjong/internal/game"
	"mahjong/internal/online"
	"mahjong/internal/protocol"
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
	for _, text := range []string{"TERMINAL MAHJONG", "Solo Game", "Create Online Room", "Reconnect Online", "How to Play", "Quit"} {
		if !strings.Contains(view, text) {
			t.Fatalf("view missing %q:\n%s", text, view)
		}
	}
}

func TestOnlineConnectedMessageShowsSnapshotTable(t *testing.T) {
	model := NewModel()
	snapshot := game.NewGame(7).Snapshot()

	next, _ := model.Update(onlineConnectedMsg{
		Message: protocol.Message{
			Type:     protocol.MsgRoomCreated,
			RoomCode: "000123",
			PlayerID: "player-1",
			Seat:     0,
			Snapshot: snapshot,
		},
	})
	updated := next.(Model)

	if updated.Screen != ScreenTable {
		t.Fatalf("screen = %v, want table", updated.Screen)
	}
	if !updated.Online {
		t.Fatal("expected online mode")
	}
	if updated.OnlineSeat != 0 || updated.OnlineRoomCode != "000123" {
		t.Fatalf("online metadata seat=%d room=%q", updated.OnlineSeat, updated.OnlineRoomCode)
	}
	if updated.NetworkStatus != NetworkYourTurn {
		t.Fatalf("network status = %q, want your turn", updated.NetworkStatus)
	}
	view := updated.View()
	for _, text := range []string{"Room:000123", "Network: your turn", "Hand Tray"} {
		if !strings.Contains(view, text) {
			t.Fatalf("online view missing %q:\n%s", text, view)
		}
	}
}

func TestOnlineTableEnterSendsDiscardAndAppliesSnapshot(t *testing.T) {
	server := online.NewServer()
	httpServer := httptest.NewServer(server)
	defer httpServer.Close()
	serverURL := "ws" + strings.TrimPrefix(httpServer.URL, "http") + "/ws"

	client := online.NewClient(serverURL, "first")
	defer client.Close()
	created, err := client.CreateRoom(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	model := applyOnlineConnected(NewModel(), onlineConnectedMsg{Message: created, Client: client})
	startEvents := len(model.OnlineSnapshot.Events)

	next, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := next.(Model)
	if cmd == nil {
		t.Fatal("expected online discard command")
	}
	if !strings.Contains(updated.StatusLine, "Discarding [01]") {
		t.Fatalf("status line = %q", updated.StatusLine)
	}

	msg := cmd()
	next, _ = updated.Update(msg)
	updated = next.(Model)

	if len(updated.OnlineSnapshot.Events) <= startEvents {
		t.Fatalf("events = %d, want more than %d", len(updated.OnlineSnapshot.Events), startEvents)
	}
	if updated.OnlineSnapshot.Current != 1 {
		t.Fatalf("current = %d, want 1", updated.OnlineSnapshot.Current)
	}
	if updated.NetworkStatus != NetworkWaiting {
		t.Fatalf("network status = %q, want waiting", updated.NetworkStatus)
	}
}

func TestOnlineTableEnterWithoutSnapshotDoesNotSendCommand(t *testing.T) {
	model := NewModel()
	model.Online = true
	model.Screen = ScreenTable
	model.OnlineClient = nil

	next, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := next.(Model)

	if cmd != nil {
		t.Fatal("expected no command")
	}
	if updated.StatusLine == "" {
		t.Fatal("expected status feedback")
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

func TestWindowSizeMessageUpdatesModelDimensions(t *testing.T) {
	model := NewModel()

	next, _ := model.Update(tea.WindowSizeMsg{Width: 72, Height: 24})
	updated := next.(Model)

	if updated.Width != 72 || updated.Height != 24 {
		t.Fatalf("size = %dx%d, want 72x24", updated.Width, updated.Height)
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

func TestKeyboardDiscardShowsLastActionFeedback(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable
	model.SelectedIndex = 0

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := next.(Model)

	view := updated.View()

	if !strings.Contains(view, "Last Action: Discarded [01]") {
		t.Fatalf("view missing discard feedback:\n%s", view)
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

func TestSecondMouseClickShowsLastActionFeedback(t *testing.T) {
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

	view := updated.View()

	if !strings.Contains(view, "Last Action: Discarded [03]") {
		t.Fatalf("view missing mouse discard feedback:\n%s", view)
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
