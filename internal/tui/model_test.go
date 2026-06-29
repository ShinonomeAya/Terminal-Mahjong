package tui

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gorilla/websocket"

	"mahjong/internal/game"
	"mahjong/internal/online"
	"mahjong/internal/protocol"
	"mahjong/internal/replay"
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

func TestMenuCanStartLocalRiichiWithRedFivesDisabled(t *testing.T) {
	model := NewModel()
	model.SelectedMode = game.ModeRiichi
	model.SelectedRiichiRedFives = 0

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := next.(Model)

	if updated.Game == nil || updated.Game.Mode != game.ModeRiichi || updated.Game.RuleConfig.Riichi.RedFives != 0 {
		t.Fatalf("local riichi game = %#v", updated.Game)
	}
}

func TestMenuCanStartLocalMCR(t *testing.T) {
	model := NewModel()
	model.SelectedMode = game.ModeMCR

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := next.(Model)

	if updated.Game == nil || updated.Game.Mode != game.ModeMCR {
		t.Fatalf("local MCR game = %#v", updated.Game)
	}
}

func TestMenuViewContainsOptions(t *testing.T) {
	view := NewModel().View()
	for _, text := range []string{"终端麻将", "单机对局", "创建联网房间", "浏览联网房间", "加入联网房间", "断线重连", "规则：日麻", "红五：三张", "玩法说明", "回放", "语言：中文", "退出"} {
		if !strings.Contains(view, text) {
			t.Fatalf("view missing %q:\n%s", text, view)
		}
	}
}

func TestMenuLanguageToggleShowsEnglishMenu(t *testing.T) {
	model := NewModel()
	model.MenuIndex = 9

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := next.(Model)

	view := updated.View()
	for _, text := range []string{"TERMINAL MAHJONG", "Solo Game", "Rules: Riichi", "Red fives: 3", "Replays", "Language: English", "Controls"} {
		if !strings.Contains(view, text) {
			t.Fatalf("english menu missing %q:\n%s", text, view)
		}
	}
}

func TestMenuTogglesRuleModeAndRedFives(t *testing.T) {
	model := NewModel()
	model.MenuIndex = 5

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := next.(Model)
	if updated.SelectedMode != game.ModeMCR {
		t.Fatalf("selected mode = %q, want MCR", updated.SelectedMode)
	}

	updated.MenuIndex = 6
	next, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated = next.(Model)
	if updated.SelectedRiichiRedFives != 0 {
		t.Fatalf("red fives = %d, want disabled", updated.SelectedRiichiRedFives)
	}
}

func TestMenuEnterJoinOnlineShowsRoomCodeInput(t *testing.T) {
	model := NewModel()
	model.MenuIndex = 3

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := next.(Model)

	if updated.Screen != ScreenJoinOnline {
		t.Fatalf("screen = %v, want join online", updated.Screen)
	}
	view := updated.View()
	for _, text := range []string{"加入联网房间", "房间号", "回车加入"} {
		if !strings.Contains(view, text) {
			t.Fatalf("join screen missing %q:\n%s", text, view)
		}
	}
}

func TestJoinOnlineInputEditsRoomCode(t *testing.T) {
	model := NewModel()
	model.Screen = ScreenJoinOnline

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1")})
	next, _ = next.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("2")})
	next, _ = next.(Model).Update(tea.KeyMsg{Type: tea.KeyBackspace})
	updated := next.(Model)

	if updated.JoinRoomCode != "1" {
		t.Fatalf("join room code = %q, want 1", updated.JoinRoomCode)
	}
}

func TestJoinOnlineEnterJoinsRoomAndShowsTable(t *testing.T) {
	server := online.NewServer()
	httpServer := httptest.NewServer(server)
	defer httpServer.Close()
	serverURL := "ws" + strings.TrimPrefix(httpServer.URL, "http") + "/ws"

	host := online.NewClient(serverURL, "host")
	defer host.Close()
	created, err := host.CreateRoom(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	model := NewModel()
	model.Screen = ScreenJoinOnline
	model.OnlineServerURL = serverURL
	model.OnlineSession = t.TempDir() + "/session.json"
	model.JoinRoomCode = created.RoomCode

	next, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := next.(Model)
	if cmd == nil {
		t.Fatal("expected join command")
	}
	if !strings.Contains(updated.StatusLine, "Joining room") {
		t.Fatalf("status line = %q", updated.StatusLine)
	}

	msg := cmd()
	next, _ = updated.Update(msg)
	updated = next.(Model)

	if updated.Screen != ScreenTable || !updated.Online {
		t.Fatalf("screen=%v online=%v, want online table", updated.Screen, updated.Online)
	}
	if updated.OnlineRoomCode != created.RoomCode || updated.OnlineSeat != 1 {
		t.Fatalf("room=%q seat=%d", updated.OnlineRoomCode, updated.OnlineSeat)
	}
}

func TestCreateOnlineRoomUsesSelectedRiichiOptions(t *testing.T) {
	server := online.NewServer()
	httpServer := httptest.NewServer(server)
	defer httpServer.Close()
	serverURL := "ws" + strings.TrimPrefix(httpServer.URL, "http") + "/ws"

	model := NewModel()
	model.OnlineServerURL = serverURL
	model.OnlineSession = t.TempDir() + "/session.json"
	model.SelectedMode = game.ModeRiichi
	model.SelectedRiichiRedFives = 0

	msg := createOnlineRoomCmd(model)()
	connected, ok := msg.(onlineConnectedMsg)
	if !ok {
		t.Fatalf("message = %#v, want onlineConnectedMsg", msg)
	}
	if connected.Message.Mode != game.ModeRiichi || connected.Message.RuleConfig.Riichi.RedFives != 0 {
		t.Fatalf("created rules = %q/%#v", connected.Message.Mode, connected.Message.RuleConfig)
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
			Started:  true,
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
	for _, text := range []string{"房间:000123", "网络：轮到你", "手牌托盘"} {
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
	if err := client.SendReady(context.Background()); err != nil {
		t.Fatal(err)
	}
	state, err := client.ReadUntil(context.Background(), 2*time.Second, protocol.MsgRoomState)
	if err != nil {
		t.Fatal(err)
	}
	created.Started = state.Started
	created.ReadySeats = state.ReadySeats
	created.OccupiedSeats = state.OccupiedSeats
	model := applyOnlineConnected(NewModel(), onlineConnectedMsg{Message: created, Client: client})
	startEvents := len(model.OnlineSnapshot.Events)

	next, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := next.(Model)
	if cmd == nil {
		t.Fatal("expected online discard command")
	}
	if !strings.Contains(updated.StatusLine, "正在打出 [01]") {
		t.Fatalf("status line = %q", updated.StatusLine)
	}

	msg := cmd()
	next, _ = updated.Update(msg)
	updated = next.(Model)
	msg = waitOnlineSnapshot(client, nil)()
	next, _ = updated.Update(msg)
	updated = next.(Model)

	if len(updated.OnlineSnapshot.Events) <= startEvents {
		t.Fatalf("events = %d, want more than %d", len(updated.OnlineSnapshot.Events), startEvents)
	}
	if updated.OnlineSnapshot.Current != 0 {
		t.Fatalf("current = %d, want 0 after bot turns", updated.OnlineSnapshot.Current)
	}
	wantHand := 14
	if updated.OnlineSnapshot.Phase == game.PhaseAwaitingClaim {
		wantHand = 13
	}
	if len(updated.OnlineSnapshot.Players[0].Hand) != wantHand {
		t.Fatalf("human hand = %d, want %d in phase %s", len(updated.OnlineSnapshot.Players[0].Hand), wantHand, updated.OnlineSnapshot.Phase)
	}
	if updated.NetworkStatus != NetworkYourTurn {
		t.Fatalf("network status = %q, want your turn", updated.NetworkStatus)
	}
}

func TestOnlineMouseClickSelectsTile(t *testing.T) {
	model := NewModel()
	snapshot := game.NewGame(11).Snapshot()
	model.Online = true
	model.Screen = ScreenTable
	model.OnlineStarted = true
	model.OnlineSeat = 0
	model.OnlineSnapshot = snapshot
	model.HandHitBoxes = handHitBoxes(len(snapshot.Players[0].Hand), 2, 10)

	next, cmd := model.Update(tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
		X:      model.HandHitBoxes[2].X1,
		Y:      model.HandHitBoxes[2].Y,
	})
	updated := next.(Model)

	if cmd != nil {
		t.Fatal("expected no command on first online mouse selection")
	}
	if updated.SelectedIndex != 2 {
		t.Fatalf("selected index = %d, want 2", updated.SelectedIndex)
	}
	if !strings.Contains(updated.StatusLine, "鼠标选中 [03]") {
		t.Fatalf("status line = %q", updated.StatusLine)
	}
}

func TestOnlineSecondMouseClickSendsDiscardAndAppliesSnapshot(t *testing.T) {
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
	if err := client.SendReady(context.Background()); err != nil {
		t.Fatal(err)
	}
	state, err := client.ReadUntil(context.Background(), 2*time.Second, protocol.MsgRoomState)
	if err != nil {
		t.Fatal(err)
	}
	created.Started = state.Started
	created.ReadySeats = state.ReadySeats
	created.OccupiedSeats = state.OccupiedSeats
	model := applyOnlineConnected(NewModel(), onlineConnectedMsg{Message: created, Client: client})
	model.SelectedIndex = 2
	model.HandHitBoxes = handHitBoxes(len(model.OnlineSnapshot.Players[0].Hand), 2, 10)
	startEvents := len(model.OnlineSnapshot.Events)

	next, cmd := model.Update(tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
		X:      model.HandHitBoxes[2].X1,
		Y:      model.HandHitBoxes[2].Y,
	})
	updated := next.(Model)
	if cmd == nil {
		t.Fatal("expected online mouse discard command")
	}
	if !strings.Contains(updated.StatusLine, "正在打出 [03]") {
		t.Fatalf("status line = %q", updated.StatusLine)
	}

	msg := cmd()
	next, _ = updated.Update(msg)
	updated = next.(Model)
	msg = waitOnlineSnapshot(client, nil)()
	next, _ = updated.Update(msg)
	updated = next.(Model)

	if len(updated.OnlineSnapshot.Events) <= startEvents {
		t.Fatalf("events = %d, want more than %d", len(updated.OnlineSnapshot.Events), startEvents)
	}
	if updated.OnlineSnapshot.Current != 0 {
		t.Fatalf("current = %d, want 0 after bot turns", updated.OnlineSnapshot.Current)
	}
	wantHand := 14
	if updated.OnlineSnapshot.Phase == game.PhaseAwaitingClaim {
		wantHand = 13
	}
	if len(updated.OnlineSnapshot.Players[0].Hand) != wantHand {
		t.Fatalf("human hand = %d, want %d in phase %s", len(updated.OnlineSnapshot.Players[0].Hand), wantHand, updated.OnlineSnapshot.Phase)
	}
}

func TestOnlineTableReadySendsReadyAndShowsRoomState(t *testing.T) {
	server := online.NewServer()
	httpServer := httptest.NewServer(server)
	defer httpServer.Close()
	serverURL := "ws" + strings.TrimPrefix(httpServer.URL, "http") + "/ws"

	first := online.NewClient(serverURL, "first")
	defer first.Close()
	created, err := first.CreateRoom(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	second := online.NewClient(serverURL, "second")
	defer second.Close()
	if _, err := second.JoinRoom(context.Background(), created.RoomCode); err != nil {
		t.Fatal(err)
	}
	model := applyOnlineConnected(NewModel(), onlineConnectedMsg{Message: created, Client: first})
	msg := waitOnlineSnapshot(first, nil)()
	next, _ := model.Update(msg)
	model = next.(Model)

	next, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	updated := next.(Model)
	if cmd == nil {
		t.Fatal("expected ready command")
	}
	if !strings.Contains(updated.StatusLine, "Ready") || !strings.Contains(localizeStatusLine(updated, updated.StatusLine), "已准备") {
		t.Fatalf("status line = %q", updated.StatusLine)
	}

	msg = cmd()
	next, _ = updated.Update(msg)
	updated = next.(Model)
	msg = waitOnlineSnapshot(first, nil)()
	next, _ = updated.Update(msg)
	updated = next.(Model)

	if len(updated.OnlineReadySeats) != 1 || updated.OnlineReadySeats[0] != 0 {
		t.Fatalf("ready seats = %#v", updated.OnlineReadySeats)
	}
	view := updated.View()
	for _, text := range []string{"准备：1/2", "按 R 准备", "等待玩家"} {
		if !strings.Contains(view, text) {
			t.Fatalf("view missing %q:\n%s", text, view)
		}
	}
}

func TestOnlineRoomsMessageShowsRoomList(t *testing.T) {
	model := NewModel()

	next, _ := model.Update(onlineRoomsMsg{Rooms: []protocol.RoomSummary{
		{Code: "000123", Occupied: 1, Ready: 0, Wall: 67},
	}})
	updated := next.(Model)

	if updated.Screen != ScreenOnlineRooms {
		t.Fatalf("screen = %v, want online rooms", updated.Screen)
	}
	view := updated.View()
	for _, want := range []string{"联网房间", "000123", "玩家:1", "回车加入"} {
		if !strings.Contains(view, want) {
			t.Fatalf("view missing %q:\n%s", want, view)
		}
	}
}

func TestOnlineRoomsEnterJoinsSelectedRoom(t *testing.T) {
	server := online.NewServer()
	httpServer := httptest.NewServer(server)
	defer httpServer.Close()
	serverURL := "ws" + strings.TrimPrefix(httpServer.URL, "http") + "/ws"

	host := online.NewClient(serverURL, "host")
	defer host.Close()
	created, err := host.CreateRoom(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	model := NewModel()
	model.Screen = ScreenOnlineRooms
	model.OnlineServerURL = serverURL
	model.OnlineSession = t.TempDir() + "/session.json"
	model.OnlineRooms = []protocol.RoomSummary{{Code: created.RoomCode, Occupied: 1, Ready: 0, Wall: 67}}

	next, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := next.(Model)
	if cmd == nil {
		t.Fatal("expected join command")
	}
	if updated.JoinRoomCode != created.RoomCode {
		t.Fatalf("join room code = %q, want %q", updated.JoinRoomCode, created.RoomCode)
	}

	msg := cmd()
	next, _ = updated.Update(msg)
	updated = next.(Model)
	if !updated.Online || updated.OnlineRoomCode != created.RoomCode {
		t.Fatalf("online=%v room=%q want room %q", updated.Online, updated.OnlineRoomCode, created.RoomCode)
	}
}

func TestOnlineTableDiscardBeforeStartedShowsWaiting(t *testing.T) {
	model := NewModel()
	model.Online = true
	model.Screen = ScreenTable
	model.OnlineClient = nil
	model.OnlineSnapshot = game.NewGame(9).Snapshot()

	next, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := next.(Model)

	if cmd != nil {
		t.Fatal("expected no discard command before room starts")
	}
	if !strings.Contains(updated.StatusLine, "Waiting for players") {
		t.Fatalf("status line = %q", updated.StatusLine)
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

func TestOnlineTableWinKeySendsWinCommand(t *testing.T) {
	serverURL, commands, closeServer := startCommandCaptureServer(t)
	defer closeServer()

	model := NewModel()
	model.Online = true
	model.Screen = ScreenTable
	model.OnlineClient = online.NewClient(serverURL, "first")
	defer model.OnlineClient.Close()
	model.OnlineStarted = true
	model.OnlineSeat = 0
	model.OnlineSnapshot = game.NewGame(13).Snapshot()
	model.OnlineSnapshot.Current = 0
	model.OnlineSnapshot.Players[0].Hand = tilesForTUI(t,
		"1m", "2m", "3m",
		"4m", "5m", "6m",
		"2p", "3p", "4p",
		"7s", "7s", "7s",
		"E", "E",
	)
	model.OnlineSnapshot.LegalActions = []game.LegalAction{{Kind: game.CommandWin}}

	next, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	updated := next.(Model)
	if cmd == nil {
		t.Fatal("expected win command")
	}
	if !strings.Contains(updated.StatusLine, "Winning") {
		t.Fatalf("status line = %q", updated.StatusLine)
	}
	_ = cmd()

	message := readCapturedCommand(t, commands)
	if message.Type != protocol.MsgPlayCommand || message.Command.Kind != game.CommandWin {
		t.Fatalf("message = %#v", message)
	}
}

func TestOnlineTableKongKeySendsKongCommand(t *testing.T) {
	serverURL, commands, closeServer := startCommandCaptureServer(t)
	defer closeServer()

	model := NewModel()
	model.Online = true
	model.Screen = ScreenTable
	model.OnlineClient = online.NewClient(serverURL, "first")
	defer model.OnlineClient.Close()
	model.OnlineStarted = true
	model.OnlineSeat = 0
	model.OnlineSnapshot = game.NewGame(13).Snapshot()
	model.OnlineSnapshot.Current = 0
	model.OnlineSnapshot.Players[0].Hand = tilesForTUI(t,
		"1m", "1m", "1m", "1m",
		"2m", "3m", "4m",
		"2p", "3p", "4p",
		"7s", "8s", "9s", "E",
	)
	model.OnlineSnapshot.LegalActions = []game.LegalAction{{Kind: game.CommandKong, Tile: "1m"}}

	next, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	updated := next.(Model)
	if cmd == nil {
		t.Fatal("expected kong command")
	}
	if !strings.Contains(updated.StatusLine, "Kong") {
		t.Fatalf("status line = %q", updated.StatusLine)
	}
	_ = cmd()

	message := readCapturedCommand(t, commands)
	if message.Type != protocol.MsgPlayCommand || message.Command.Kind != game.CommandKong || message.Command.Tile != "1m" {
		t.Fatalf("message = %#v", message)
	}
}

func TestOnlineActionBarShowsReadyWinAndKong(t *testing.T) {
	model := NewModel()
	model.Online = true
	model.Screen = ScreenTable
	model.OnlineStarted = true
	model.OnlineSeat = 0
	model.OnlineSnapshot = game.NewGame(13).Snapshot()
	model.OnlineSnapshot.Current = 0
	model.OnlineSnapshot.Players[0].Hand = tilesForTUI(t,
		"1m", "1m", "1m", "1m",
		"2m", "3m", "4m",
		"2p", "3p", "4p",
		"7s", "7s", "7s", "E",
	)
	model.OnlineSnapshot.LegalActions = []game.LegalAction{{Kind: game.CommandWin}, {Kind: game.CommandKong, Tile: "1m"}}

	view := model.View()
	for _, text := range []string{"[H] 胡", "[K] 杠"} {
		if !strings.Contains(view, text) {
			t.Fatalf("online action bar missing %q:\n%s", text, view)
		}
	}
}

func TestOnlineGameOverSnapshotShowsResultScreen(t *testing.T) {
	model := NewModel()
	snapshot := game.NewGame(13).Snapshot()
	snapshot.Over = true
	snapshot.Winner = 0
	snapshot.Reason = "self-draw"

	next, _ := model.Update(onlineSnapshotMsg{
		Message: protocol.Message{
			Type:     protocol.MsgGameSnapshot,
			RoomCode: "000777",
			Started:  true,
			Snapshot: snapshot,
		},
	})
	updated := next.(Model)

	if updated.Screen != ScreenGameOver {
		t.Fatalf("screen = %v, want game over", updated.Screen)
	}
	view := updated.View()
	for _, text := range []string{"对局结束", "房间：000777", "结果：self-draw", "赢家：座位 1", "返回菜单"} {
		if !strings.Contains(view, text) {
			t.Fatalf("online game over missing %q:\n%s", text, view)
		}
	}
}

func TestOnlineServerErrorShowsStatusAndKeepsNetworkState(t *testing.T) {
	model := NewModel()
	model.Online = true
	model.Screen = ScreenTable
	model.NetworkStatus = NetworkWaiting

	next, cmd := model.Update(onlineSnapshotMsg{
		Message: protocol.Message{Type: protocol.MsgError, Error: "not the current player"},
	})
	updated := next.(Model)

	if cmd != nil {
		t.Fatal("expected no wait command without an online client")
	}
	if !strings.Contains(updated.StatusLine, "not the current player") {
		t.Fatalf("status line = %q", updated.StatusLine)
	}
	if updated.NetworkStatus != NetworkWaiting {
		t.Fatalf("network status = %q, want waiting", updated.NetworkStatus)
	}
}

func TestOnlineReconnectAttemptMessageUpdatesNetworkStatus(t *testing.T) {
	model := NewModel()
	model.Online = true
	model.Screen = ScreenTable

	next, cmd := model.Update(onlineReconnectAttemptMsg{Attempt: 2, Max: 5})
	updated := next.(Model)

	if cmd != nil {
		t.Fatal("expected no command for direct reconnect message without event channel")
	}
	if updated.NetworkStatus != NetworkReconnecting {
		t.Fatalf("network status = %q, want reconnecting", updated.NetworkStatus)
	}
	if updated.ReconnectAttempt != 2 || updated.ReconnectMax != 5 {
		t.Fatalf("attempt = %d/%d, want 2/5", updated.ReconnectAttempt, updated.ReconnectMax)
	}
	if !strings.Contains(updated.View(), "网络：重连中 2/5") {
		t.Fatalf("view missing reconnecting status:\n%s", updated.View())
	}
}

func TestOnlineReconnectSuccessMessageUpdatesNetworkStatus(t *testing.T) {
	model := NewModel()
	model.Online = true
	model.Screen = ScreenTable
	model.NetworkStatus = NetworkReconnecting
	model.ReconnectAttempt = 2
	model.ReconnectMax = 5

	next, cmd := model.Update(onlineReconnectSuccessMsg{})
	updated := next.(Model)

	if cmd != nil {
		t.Fatal("expected no command for direct reconnect success without event channel")
	}
	if updated.NetworkStatus != NetworkReconnected {
		t.Fatalf("network status = %q, want reconnected", updated.NetworkStatus)
	}
	if !strings.Contains(updated.View(), "网络：已重连") {
		t.Fatalf("view missing reconnected status:\n%s", updated.View())
	}
}

func startCommandCaptureServer(t *testing.T) (string, <-chan protocol.Message, func()) {
	t.Helper()
	commands := make(chan protocol.Message, 1)
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Error(err)
			return
		}
		defer conn.Close()
		var message protocol.Message
		if err := conn.ReadJSON(&message); err != nil {
			t.Error(err)
			return
		}
		commands <- message
	}))
	return "ws" + strings.TrimPrefix(server.URL, "http"), commands, server.Close
}

func readCapturedCommand(t *testing.T, commands <-chan protocol.Message) protocol.Message {
	t.Helper()
	select {
	case message := <-commands:
		return message
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for command")
		return protocol.Message{}
	}
}

func tilesForTUI(t *testing.T, texts ...string) []game.Tile {
	t.Helper()
	tiles := make([]game.Tile, len(texts))
	for i, text := range texts {
		tile, ok := game.ParseTile(text)
		if !ok {
			t.Fatalf("bad tile %q", text)
		}
		tiles[i] = tile
	}
	return tiles
}

func TestOnlineGameOverMainMenuClearsOnlineState(t *testing.T) {
	model := NewModel()
	model.Screen = ScreenGameOver
	model.Online = true
	model.OnlineRoomCode = "000777"
	model.OnlineClient = online.NewClient("ws://127.0.0.1:1/ws", "first")
	model.OnlineSnapshot = game.NewGame(13).Snapshot()
	model.OnlineSnapshot.Over = true
	model.GameOverIndex = 1

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := next.(Model)

	if updated.Screen != ScreenMenu {
		t.Fatalf("screen = %v, want menu", updated.Screen)
	}
	if updated.Online || updated.OnlineClient != nil || updated.OnlineRoomCode != "" {
		t.Fatalf("online state not cleared: online=%v client=%v room=%q", updated.Online, updated.OnlineClient, updated.OnlineRoomCode)
	}
}

func TestMenuViewUsesReadableSections(t *testing.T) {
	view := NewModel().View()
	for _, text := range []string{"开始菜单", "操作", "上下选择"} {
		if !strings.Contains(view, text) {
			t.Fatalf("menu missing %q:\n%s", text, view)
		}
	}
}

func TestHelpViewContainsKeyboardAndMouseControls(t *testing.T) {
	model := NewModel()
	model.Screen = ScreenHelp

	view := model.View()

	for _, text := range []string{"键盘", "鼠标", "回车/空格", "再次单击"} {
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

	if !hasDiscardEventSince(updated.Game.Events, startEvents, 0) {
		t.Fatalf("events = %#v, want human discard after event %d", updated.Game.Events, startEvents)
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
	model.Game = game.NewGame(1)
	model.Game.StartHumanTurn()
	model.Screen = ScreenTable
	model.SelectedIndex = 0

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := next.(Model)

	view := updated.View()

	if !strings.Contains(view, "上步：已打出 [01]") {
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
	if !strings.Contains(updated.StatusLine, "鼠标选中 [03]") {
		t.Fatalf("status line = %q, want mouse selection feedback", updated.StatusLine)
	}
}

func TestSecondMouseClickDiscardsSelectedTile(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable
	model.SelectedIndex = 2
	model.HandHitBoxes = handHitBoxes(len(model.Game.Players[0].Hand), 2, 10)
	startEvents := len(model.Game.Events)

	next, _ := model.Update(tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
		X:      model.HandHitBoxes[2].X1,
		Y:      model.HandHitBoxes[2].Y,
	})
	updated := next.(Model)

	if !hasDiscardEventSince(updated.Game.Events, startEvents, 0) {
		t.Fatalf("events = %#v, want human discard after event %d", updated.Game.Events, startEvents)
	}
	if !strings.Contains(updated.StatusLine, "已打出 [03]") {
		t.Fatalf("status line = %q, want discard feedback", updated.StatusLine)
	}
}

func TestSecondMouseClickShowsLastActionFeedback(t *testing.T) {
	model := NewModel()
	model.Game = game.NewGame(1)
	model.Game.StartHumanTurn()
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

	if !strings.Contains(view, "上步：已打出 [03]") {
		t.Fatalf("view missing mouse discard feedback:\n%s", view)
	}
}

func TestKeyboardSelectionUpdatesStatusLine(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyRight})
	updated := next.(Model)

	if !strings.Contains(updated.StatusLine, "选中 [02]") {
		t.Fatalf("status line = %q, want keyboard selection feedback", updated.StatusLine)
	}
}

func TestTableViewIncludesStatusLine(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable
	model.StatusLine = "Mouse selected [04] 🀊 (4m)"

	view := model.View()

	if !strings.Contains(view, "状态：鼠标选中 [04]") {
		t.Fatalf("view missing status line:\n%s", view)
	}
}

func TestLocalClaimPongKeyAppliesMeld(t *testing.T) {
	model := localClaimModel(t, game.ClaimPong,
		game.ClaimOption{Kind: game.ClaimPong, Player: 0, Consumed: mustUITiles(t, "3m", "3m")},
	)

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	updated := next.(Model)

	if len(updated.Game.Players[0].Melds) != 1 || updated.Game.Players[0].Melds[0].Kind != game.MeldPong {
		t.Fatalf("melds = %#v, want pong", updated.Game.Players[0].Melds)
	}
	if updated.Game.Phase != game.PhaseAwaitingDiscard || updated.Game.PendingClaim != nil {
		t.Fatalf("phase/pending = %q/%#v", updated.Game.Phase, updated.Game.PendingClaim)
	}
}

func TestLocalClaimSpacePasses(t *testing.T) {
	model := localClaimModel(t, game.ClaimPong,
		game.ClaimOption{Kind: game.ClaimPong, Player: 0, Consumed: mustUITiles(t, "3m", "3m")},
	)

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeySpace})
	updated := next.(Model)

	if updated.Game.PendingClaim != nil || updated.Game.Phase != game.PhaseAwaitingDiscard {
		t.Fatalf("phase/pending = %q/%#v", updated.Game.Phase, updated.Game.PendingClaim)
	}
	if !strings.Contains(localizeStatusLine(updated, updated.StatusLine), "已过") {
		t.Fatalf("status = %q", updated.StatusLine)
	}
}

func TestClaimResponseDisablesNormalDiscard(t *testing.T) {
	model := localClaimModel(t, game.ClaimPong,
		game.ClaimOption{Kind: game.ClaimPong, Player: 0, Consumed: mustUITiles(t, "3m", "3m")},
	)
	startHand := len(model.Game.Players[0].Hand)

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := next.(Model)

	if len(updated.Game.Players[0].Hand) != startHand || updated.Game.PendingClaim == nil {
		t.Fatalf("enter changed claim state: hand=%d pending=%#v", len(updated.Game.Players[0].Hand), updated.Game.PendingClaim)
	}
}

func TestClaimChowArrowsSelectCombination(t *testing.T) {
	model := localClaimModel(t, game.ClaimChow,
		game.ClaimOption{Kind: game.ClaimChow, Player: 0, Consumed: mustUITiles(t, "1m", "2m")},
		game.ClaimOption{Kind: game.ClaimChow, Player: 0, Consumed: mustUITiles(t, "2m", "4m")},
	)

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyRight})
	updated := next.(Model)

	if updated.ClaimOptionIndex != 1 {
		t.Fatalf("claim option index = %d, want 1", updated.ClaimOptionIndex)
	}
}

func TestOnlineClaimKeyBuildsPongCommand(t *testing.T) {
	model := localClaimModel(t, game.ClaimPong,
		game.ClaimOption{Kind: game.ClaimPong, Player: 0, Consumed: mustUITiles(t, "3m", "3m")},
	)
	model.Online = true
	model.OnlineSeat = 0
	model.OnlineSnapshot = model.Game.Snapshot()
	model.Game = nil

	command, ok := claimCommandForKey(model, "p")

	if !ok || command.Kind != game.CommandPong {
		t.Fatalf("command/ok = %#v/%v, want pong", command, ok)
	}
}

func TestOnlineSnapshotResetsClaimOptionSelection(t *testing.T) {
	model := localClaimModel(t, game.ClaimChow,
		game.ClaimOption{Kind: game.ClaimChow, Player: 0, Consumed: mustUITiles(t, "1m", "2m")},
	)
	model.Online = true
	model.OnlineSeat = 0
	model.OnlineSnapshot = model.Game.Snapshot()
	model.Game = nil
	model.ClaimOptionIndex = 1

	updated := applyOnlineSnapshot(model, protocol.Message{Type: protocol.MsgGameSnapshot, Snapshot: model.OnlineSnapshot})

	if updated.ClaimOptionIndex != 0 {
		t.Fatalf("claim option index = %d, want reset to 0", updated.ClaimOptionIndex)
	}
}

func localClaimModel(t *testing.T, kind game.ClaimKind, options ...game.ClaimOption) Model {
	t.Helper()
	model := NewModel()
	model.Screen = ScreenTable
	model.Game = game.NewGame(5)
	model.Game.Players[0].Hand = mustUITiles(t, "1m", "2m", "3m", "3m", "4m", "5m", "1p", "2p", "4p", "5p", "1s", "2s", "N")
	discard := mustUITiles(t, "3m")[0]
	model.Game.Players[3].Discards = []game.Tile{discard}
	model.Game.Current = 0
	model.Game.Phase = game.PhaseAwaitingClaim
	model.Game.PendingClaim = &game.PendingClaim{Discarder: 3, Tile: discard, Options: options}
	if len(options) == 0 || options[0].Kind != kind {
		t.Fatalf("bad claim fixture: %#v", options)
	}
	return model
}

func hasDiscardEventSince(events []game.GameEvent, start int, player int) bool {
	for _, event := range events[start:] {
		if event.Kind == game.EventDiscard && event.Player == player {
			return true
		}
	}
	return false
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

	for _, text := range []string{"对局结束", "重新开始", "返回菜单", "退出", "操作"} {
		if !strings.Contains(view, text) {
			t.Fatalf("game over missing %q:\n%s", text, view)
		}
	}
}

func TestApplyOnlineSnapshotRetainsMatchSnapshot(t *testing.T) {
	round := game.NewGame(13).SnapshotFor("0")
	match := game.MatchSnapshot{
		Mode:        game.ModeRiichi,
		RuleConfig:  game.DefaultRuleConfig(game.ModeRiichi),
		Points:      [4]int{25000, 25000, 25000, 25000},
		RoundNumber: 1,
		Round:       round,
	}

	updated := applyOnlineSnapshot(NewModel(), protocol.Message{Type: protocol.MsgGameSnapshot, Snapshot: round, Match: match})

	if updated.OnlineMatch.Mode != game.ModeRiichi || updated.OnlineMatch.RuleConfig != match.RuleConfig || updated.OnlineMatch.Points != match.Points {
		t.Fatalf("online match = %#v, want %#v", updated.OnlineMatch, match)
	}
}

func TestMenuStartsLocalMatchCoordinator(t *testing.T) {
	model := NewModel()
	model.SelectedMode = game.ModeRiichi

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := next.(Model)

	if updated.LocalMatch == nil || updated.Game != updated.LocalMatch.Round {
		t.Fatalf("local match/game = %#v/%p", updated.LocalMatch, updated.Game)
	}
	if updated.LocalMatch.Mode != game.ModeRiichi || updated.ReplayDir != "replays" {
		t.Fatalf("mode=%q replay dir=%q", updated.LocalMatch.Mode, updated.ReplayDir)
	}
}

func TestLocalDiscardRecordsReplayCommand(t *testing.T) {
	model := NewModel()
	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = next.(Model)
	before := model.LocalMatch.ReplayCommandCount()

	next, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := next.(Model)

	if updated.LocalMatch.ReplayCommandCount() <= before {
		t.Fatalf("commands=%d want greater than %d", updated.LocalMatch.ReplayCommandCount(), before)
	}
	if updated.Game != updated.LocalMatch.Round {
		t.Fatal("game alias does not track the local match round")
	}
}

func TestLocalQuitMarksMatchAbandonedWithoutSaving(t *testing.T) {
	model := NewModel()
	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = next.(Model)

	next, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	updated := next.(Model)

	if cmd != nil {
		t.Fatal("abandoned match should not schedule replay save")
	}
	if updated.LocalMatch == nil || !updated.LocalMatch.Abandoned || updated.LocalMatch.Complete {
		t.Fatalf("local match = %#v", updated.LocalMatch)
	}
	if updated.Screen != ScreenGameOver {
		t.Fatalf("screen = %v, want game over", updated.Screen)
	}
}

func TestSaveCompletedReplayCmdWritesValidatedFile(t *testing.T) {
	match, err := game.NewMatch(140014, game.NewCompatibilityRuleSet(game.ModeCompatibility, game.RuleConfig{}))
	if err != nil {
		t.Fatal(err)
	}
	match.Round.Players[0].Hand = mustUITiles(t,
		"1m", "2m", "3m",
		"4m", "5m", "6m",
		"2p", "3p", "4p",
		"7s", "7s", "7s",
		"E", "E",
	)
	if result := match.ApplyCommand(game.GameCommand{PlayerID: "0", Kind: game.CommandWin}); !result.OK {
		t.Fatal(result.Error)
	}

	message := saveCompletedReplayCmd(match, t.TempDir())()
	saved, ok := message.(replaySavedMsg)
	if !ok {
		t.Fatalf("message = %#v", message)
	}
	file, err := replay.Load(saved.Path)
	if err != nil {
		t.Fatal(err)
	}
	if file.Mode != game.ModeCompatibility || !file.Complete {
		t.Fatalf("saved replay = %#v", file)
	}
}

func TestReplaySavedMessageUpdatesResultState(t *testing.T) {
	model := NewModel()

	next, cmd := model.Update(replaySavedMsg{Path: "replays/test.json"})
	updated := next.(Model)

	if cmd != nil || updated.LastReplayPath != "replays/test.json" {
		t.Fatalf("path=%q cmd=%v", updated.LastReplayPath, cmd)
	}
}

func TestOnlineReconnectRequestsReplayOnlyOnce(t *testing.T) {
	serverURL, messages, closeServer := startCommandCaptureServer(t)
	defer closeServer()

	client := online.NewClient(serverURL, "first")
	defer client.Close()
	model := NewModel()
	model.OnlineClient = client

	message := protocol.Message{Type: protocol.MsgReconnected, ReplayID: "replay-1"}
	updated, cmd, handled := applyOnlineReplayMessage(model, message)
	if !handled || cmd == nil || updated.ReplayRequestedID != "replay-1" {
		t.Fatalf("handled=%t cmd=%v requested=%q", handled, cmd, updated.ReplayRequestedID)
	}
	if result := cmd(); result != (onlineCommandSentMsg{}) {
		t.Fatalf("request result = %#v", result)
	}
	select {
	case request := <-messages:
		if request.Type != protocol.MsgRequestReplay {
			t.Fatalf("request type = %q", request.Type)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for replay request")
	}

	updated, cmd, handled = applyOnlineReplayMessage(updated, message)
	if !handled || cmd != nil {
		t.Fatalf("duplicate handled=%t cmd=%v", handled, cmd)
	}
}

func TestOnlineReplayDataSavesValidatedFile(t *testing.T) {
	model := NewModel()
	model.ReplayDir = t.TempDir()
	file := tuiReplayFixture(t, "online-replay")

	updated, cmd, handled := applyOnlineReplayMessage(model, protocol.Message{
		Type:     protocol.MsgReplayData,
		ReplayID: file.ReplayID,
		Replay:   &file,
	})
	if !handled || cmd == nil || updated.ReplaySavingID != file.ReplayID {
		t.Fatalf("handled=%t cmd=%v saving=%q", handled, cmd, updated.ReplaySavingID)
	}
	saved, ok := cmd().(replaySavedMsg)
	if !ok {
		t.Fatalf("save result = %#v", cmd())
	}
	next, _ := updated.Update(saved)
	finished := next.(Model)
	if finished.ReplaySavedID != file.ReplayID || finished.LastReplayPath == "" {
		t.Fatalf("saved=%q path=%q", finished.ReplaySavedID, finished.LastReplayPath)
	}
	loaded, err := replay.Load(finished.LastReplayPath)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Checksum != file.Checksum {
		t.Fatalf("checksum = %q, want %q", loaded.Checksum, file.Checksum)
	}
}

func TestLocalMCRRoundTransitionStaysOnTable(t *testing.T) {
	match, err := game.NewMatch(140014, game.NewMCRRuleSet(game.DefaultRuleConfig(game.ModeMCR).MCR))
	if err != nil {
		t.Fatal(err)
	}
	match.Round.Players[0].Hand = mustUITiles(t,
		"1m", "9m", "1p", "9p", "1s", "9s",
		"E", "E", "S", "W", "N", "Z", "F", "B",
	)
	match.Round.RecordEvent(game.EventDraw, 0, mustUITiles(t, "E")[0], "")
	model := NewModel()
	model.Screen = ScreenTable
	model.LocalMatch = match
	model = syncLocalRound(model)

	model, result := applyLocalCommand(model, game.GameCommand{Kind: game.CommandWin})
	if !result.OK {
		t.Fatal(result.Error)
	}
	next, cmd := finishLocalUpdate(model)
	updated := next.(Model)

	if cmd != nil || updated.Screen != ScreenTable || updated.LocalMatch.RoundNumber != 2 {
		t.Fatalf("screen=%v round=%d cmd=%v", updated.Screen, updated.LocalMatch.RoundNumber, cmd)
	}
	if updated.Game != updated.LocalMatch.Round {
		t.Fatal("game alias did not advance to the next MCR round")
	}
}

func TestFinishLocalUpdateSchedulesCompletedReplaySave(t *testing.T) {
	match, err := game.NewMatch(140014, game.NewCompatibilityRuleSet(game.ModeCompatibility, game.RuleConfig{}))
	if err != nil {
		t.Fatal(err)
	}
	match.Round.Players[0].Hand = mustUITiles(t,
		"1m", "2m", "3m",
		"4m", "5m", "6m",
		"2p", "3p", "4p",
		"7s", "7s", "7s",
		"E", "E",
	)
	if result := match.ApplyCommand(game.GameCommand{PlayerID: "0", Kind: game.CommandWin}); !result.OK {
		t.Fatal(result.Error)
	}
	model := NewModel()
	model.LocalMatch = match
	model.ReplayDir = t.TempDir()
	model = syncLocalRound(model)

	next, cmd := finishLocalUpdate(model)
	updated := next.(Model)

	if updated.Screen != ScreenGameOver || cmd == nil {
		t.Fatalf("screen=%v cmd=%v", updated.Screen, cmd)
	}
	if _, ok := cmd().(replaySavedMsg); !ok {
		t.Fatal("completed match did not produce replaySavedMsg")
	}
}

func tuiReplayFixture(t *testing.T, id string) game.ReplayFile {
	t.Helper()
	round := game.NewGame(140015).Snapshot()
	round.Over = true
	match := game.MatchSnapshot{
		Mode:       game.ModeCompatibility,
		RuleConfig: game.RuleConfig{},
		Complete:   true,
		Round:      round,
	}
	file, err := game.SealReplay(game.ReplayFile{
		ApplicationVersion: "test",
		ReplayID:           id,
		CreatedAt:          time.Unix(20, 0).UTC(),
		Mode:               game.ModeCompatibility,
		RuleConfig:         game.RuleConfig{},
		ShuffleProof:       round.ShuffleProof,
		Participants: []game.ReplayParticipant{
			{Seat: 0, ID: "0", Name: "You"},
			{Seat: 1, ID: "1", Name: "AI-1"},
			{Seat: 2, ID: "2", Name: "AI-2"},
			{Seat: 3, ID: "3", Name: "AI-3"},
		},
		Initial:  match,
		Frames:   []game.ReplayFrame{{Index: 0, Match: match}},
		Complete: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	return file
}
