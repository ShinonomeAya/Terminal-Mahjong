package online

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"mahjong/internal/game"
	"mahjong/internal/protocol"
)

func TestMCRReconnectRestoresCanonicalPrivateMatchSnapshot(t *testing.T) {
	server := NewServer()
	url, closeServer := startTestServer(t, server)
	defer closeServer()
	client := dialTestClient(t, url)
	sendMessage(t, client, protocol.Message{Type: protocol.MsgCreateRoom, Mode: game.ModeMCR, RuleConfig: game.DefaultRuleConfig(game.ModeMCR)})
	created := readUntil(t, client, protocol.MsgRoomCreated)
	want, err := json.Marshal(created.Match)
	if err != nil {
		t.Fatal(err)
	}
	client.Close()

	reconnectedClient := dialTestClient(t, url)
	defer reconnectedClient.Close()
	sendMessage(t, reconnectedClient, protocol.Message{Type: protocol.MsgReconnect, PlayerID: created.PlayerID, ReconnectToken: created.ReconnectToken})
	reconnected := readUntil(t, reconnectedClient, protocol.MsgReconnected)
	got, err := json.Marshal(reconnected.Match)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("reconnected MCR match differs\nbefore=%s\nafter=%s", want, got)
	}
}

func TestWebSocketServerCreatesAndJoinsRoom(t *testing.T) {
	server := NewServer()
	url, closeServer := startTestServer(t, server)
	defer closeServer()

	first := dialTestClient(t, url)
	defer first.Close()
	sendMessage(t, first, protocol.Message{Type: protocol.MsgCreateRoom, Name: "first"})
	created := readUntil(t, first, protocol.MsgRoomCreated)
	if created.RoomCode == "" || created.PlayerID == "" || created.ReconnectToken == "" {
		t.Fatalf("created = %#v", created)
	}
	if created.Seat != 0 {
		t.Fatalf("created seat = %d, want 0", created.Seat)
	}

	second := dialTestClient(t, url)
	defer second.Close()
	sendMessage(t, second, protocol.Message{Type: protocol.MsgJoinRoom, RoomCode: created.RoomCode, Name: "second"})
	joined := readUntil(t, second, protocol.MsgRoomJoined)
	if joined.RoomCode != created.RoomCode || joined.PlayerID == "" {
		t.Fatalf("joined = %#v", joined)
	}
	if joined.Seat != 1 {
		t.Fatalf("joined seat = %d, want 1", joined.Seat)
	}
	if len(joined.Snapshot.Players) != 4 {
		t.Fatalf("snapshot players = %d", len(joined.Snapshot.Players))
	}
}

func TestWebSocketServerSerializesConcurrentWritesPerConnection(t *testing.T) {
	server := NewServer()
	url, closeServer := startTestServer(t, server)
	defer closeServer()

	host := dialTestClient(t, url)
	defer host.Close()
	sendMessage(t, host, protocol.Message{Type: protocol.MsgCreateRoom, Name: "host"})
	created := readUntil(t, host, protocol.MsgRoomCreated)

	guest := dialTestClient(t, url)
	defer guest.Close()
	sendMessage(t, guest, protocol.Message{Type: protocol.MsgJoinRoom, RoomCode: created.RoomCode, Name: "guest"})
	_ = readUntil(t, guest, protocol.MsgRoomJoined)
	_ = readUntil(t, host, protocol.MsgRoomState)

	const messageCount = 50
	readMessages := func(conn *websocket.Conn, count int) error {
		if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
			return err
		}
		for range count {
			var message protocol.Message
			if err := conn.ReadJSON(&message); err != nil {
				return err
			}
		}
		return nil
	}

	readErrors := make(chan error, 2)
	go func() { readErrors <- readMessages(host, messageCount*2) }()
	go func() { readErrors <- readMessages(guest, messageCount) }()

	writeErrors := make(chan error, 2)
	go func() {
		for range messageCount {
			if err := host.WriteJSON(protocol.Message{Type: protocol.MsgListRooms}); err != nil {
				writeErrors <- err
				return
			}
		}
		writeErrors <- nil
	}()
	go func() {
		for range messageCount {
			if err := guest.WriteJSON(protocol.Message{Type: protocol.MsgReady}); err != nil {
				writeErrors <- err
				return
			}
		}
		writeErrors <- nil
	}()

	for range 2 {
		if err := <-writeErrors; err != nil {
			t.Fatal(err)
		}
	}
	for range 2 {
		if err := <-readErrors; err != nil {
			t.Fatal(err)
		}
	}
}

func TestWebSocketServerSendsRecipientPrivateMatchSnapshots(t *testing.T) {
	server := NewServer()
	url, closeServer := startTestServer(t, server)
	defer closeServer()
	config := game.DefaultRuleConfig(game.ModeMCR)

	first := dialTestClient(t, url)
	defer first.Close()
	sendMessage(t, first, protocol.Message{Type: protocol.MsgCreateRoom, Name: "first", Mode: game.ModeMCR, RuleConfig: config})
	created := readUntil(t, first, protocol.MsgRoomCreated)
	assertPrivateSnapshot(t, created.Snapshot, 0)
	assertPrivateSnapshot(t, created.Match.Round, 0)
	if created.Match.Mode != game.ModeMCR || created.Match.RuleConfig != config {
		t.Fatalf("created match = %#v", created.Match)
	}
	server.mu.Lock()
	room := server.rooms[created.RoomCode]
	totalTiles := len(room.match.Round.Wall)
	for _, player := range room.match.Round.Players {
		totalTiles += len(player.Hand) + len(player.Flowers)
		for _, meld := range player.Melds {
			totalTiles += len(meld.Tiles)
		}
	}
	server.mu.Unlock()
	if totalTiles != 144 {
		t.Fatalf("MCR room tile total = %d, want 144", totalTiles)
	}

	second := dialTestClient(t, url)
	defer second.Close()
	sendMessage(t, second, protocol.Message{Type: protocol.MsgJoinRoom, RoomCode: created.RoomCode, Name: "second"})
	joined := readUntil(t, second, protocol.MsgRoomJoined)
	assertPrivateSnapshot(t, joined.Snapshot, 1)
	assertPrivateSnapshot(t, joined.Match.Round, 1)
	if joined.Match.Mode != game.ModeMCR {
		t.Fatalf("joined match mode = %q", joined.Match.Mode)
	}

	state := readUntil(t, first, protocol.MsgRoomState)
	assertPrivateSnapshot(t, state.Snapshot, 0)
	if state.Match.Mode != game.ModeMCR {
		t.Fatalf("room state match mode = %q", state.Match.Mode)
	}
}

func TestRiichiRoomUsesRiichiRuleSetAndReconnectsCanonicalPrivateSnapshot(t *testing.T) {
	server := NewServer()
	url, closeServer := startTestServer(t, server)
	defer closeServer()
	config := game.DefaultRuleConfig(game.ModeRiichi)

	client := dialTestClient(t, url)
	sendMessage(t, client, protocol.Message{Type: protocol.MsgCreateRoom, Name: "east", Mode: game.ModeRiichi, RuleConfig: config})
	created := readUntil(t, client, protocol.MsgRoomCreated)
	if created.Match.Mode != game.ModeRiichi || created.Match.RuleConfig != config {
		t.Fatalf("created match = %#v", created.Match)
	}
	if created.Match.Round.Riichi == nil || created.Match.Round.Riichi.DeadWallCount != 14 || len(created.Match.Round.Riichi.DoraIndicators) != 1 {
		t.Fatalf("created riichi snapshot = %#v", created.Match.Round.Riichi)
	}
	server.mu.Lock()
	room := server.rooms[created.RoomCode]
	hasRiichiState := room != nil && room.match.Round.Riichi != nil && len(room.match.Round.Riichi.DeadWall) == 14
	server.mu.Unlock()
	if !hasRiichiState {
		t.Fatal("room did not use authoritative RiichiRuleSet")
	}
	want, err := json.Marshal(created.Match)
	if err != nil {
		t.Fatal(err)
	}
	client.Close()

	reconnectedClient := dialTestClient(t, url)
	defer reconnectedClient.Close()
	sendMessage(t, reconnectedClient, protocol.Message{Type: protocol.MsgReconnect, PlayerID: created.PlayerID, ReconnectToken: created.ReconnectToken})
	reconnected := readUntil(t, reconnectedClient, protocol.MsgReconnected)
	got, err := json.Marshal(reconnected.Match)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("reconnected Riichi match differs\nbefore=%s\nafter=%s", want, got)
	}
}

func TestRiichiWebSocketReadyDiscardAndReconnectSmoke(t *testing.T) {
	server := NewServer()
	url, closeServer := startTestServer(t, server)
	defer closeServer()
	config := game.DefaultRuleConfig(game.ModeRiichi)

	host := dialTestClient(t, url)
	sendMessage(t, host, protocol.Message{Type: protocol.MsgCreateRoom, Name: "east", Mode: game.ModeRiichi, RuleConfig: config})
	created := readUntil(t, host, protocol.MsgRoomCreated)
	clients := []*websocket.Conn{host}
	defer func() {
		for _, client := range clients {
			client.Close()
		}
	}()

	for seat := 1; seat < 4; seat++ {
		client := dialTestClient(t, url)
		clients = append(clients, client)
		sendMessage(t, client, protocol.Message{Type: protocol.MsgJoinRoom, RoomCode: created.RoomCode, Name: "guest"})
		joined := readUntil(t, client, protocol.MsgRoomJoined)
		if joined.Seat != seat || joined.Match.Mode != game.ModeRiichi {
			t.Fatalf("joined = %#v", joined)
		}
	}
	for _, client := range clients {
		sendMessage(t, client, protocol.Message{Type: protocol.MsgReady})
	}
	started := readUntilStartedRoomState(t, host)
	assertPrivateSnapshot(t, started.Match.Round, 0)
	if started.Match.Round.Riichi == nil || len(started.Match.Round.Riichi.UraIndicators) != 0 {
		t.Fatalf("started riichi private snapshot = %#v", started.Match.Round.Riichi)
	}
	discard, ok := firstDiscardAction(started.Match.Round.LegalActions)
	if !ok {
		t.Fatalf("started legal actions missing discard: %#v", started.Match.Round.LegalActions)
	}

	sendMessage(t, host, protocol.Message{Type: protocol.MsgPlayCommand, Command: game.GameCommand{Kind: game.CommandDiscard, TileIndex: discard.TileIndex}})
	update := readUntil(t, host, protocol.MsgGameSnapshot)
	assertPrivateSnapshot(t, update.Match.Round, 0)
	if update.Match.Round.Riichi == nil || len(update.Match.Round.Riichi.UraIndicators) != 0 {
		t.Fatalf("post-discard riichi private snapshot = %#v", update.Match.Round.Riichi)
	}
	want, err := json.Marshal(update.Match)
	if err != nil {
		t.Fatal(err)
	}
	host.Close()

	reconnectedClient := dialTestClient(t, url)
	clients[0] = reconnectedClient
	sendMessage(t, reconnectedClient, protocol.Message{Type: protocol.MsgReconnect, PlayerID: created.PlayerID, ReconnectToken: created.ReconnectToken})
	reconnected := readUntil(t, reconnectedClient, protocol.MsgReconnected)
	got, err := json.Marshal(reconnected.Match)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("reconnected Riichi smoke match differs\nbefore=%s\nafter=%s", want, got)
	}
}

func TestWebSocketServerRejectsConflictingJoinMode(t *testing.T) {
	server := NewServer()
	url, closeServer := startTestServer(t, server)
	defer closeServer()

	host := dialTestClient(t, url)
	defer host.Close()
	sendMessage(t, host, protocol.Message{
		Type:       protocol.MsgCreateRoom,
		Mode:       game.ModeMCR,
		RuleConfig: game.DefaultRuleConfig(game.ModeMCR),
	})
	created := readUntil(t, host, protocol.MsgRoomCreated)

	joining := dialTestClient(t, url)
	defer joining.Close()
	sendMessage(t, joining, protocol.Message{
		Type:       protocol.MsgJoinRoom,
		RoomCode:   created.RoomCode,
		Mode:       game.ModeRiichi,
		RuleConfig: game.DefaultRuleConfig(game.ModeRiichi),
	})
	errMsg := readUntil(t, joining, protocol.MsgError)
	if !strings.Contains(errMsg.Error, "rule mode") {
		t.Fatalf("error = %#v, want rule mode conflict", errMsg)
	}
}

func TestWebSocketServerListsWaitingRooms(t *testing.T) {
	server := NewServer()
	url, closeServer := startTestServer(t, server)
	defer closeServer()

	first := dialTestClient(t, url)
	defer first.Close()
	sendMessage(t, first, protocol.Message{Type: protocol.MsgCreateRoom, Name: "first"})
	created := readUntil(t, first, protocol.MsgRoomCreated)

	lister := dialTestClient(t, url)
	defer lister.Close()
	sendMessage(t, lister, protocol.Message{Type: protocol.MsgListRooms})
	list := readUntil(t, lister, protocol.MsgRoomList)

	if len(list.Rooms) != 1 {
		t.Fatalf("rooms = %#v, want one waiting room", list.Rooms)
	}
	room := list.Rooms[0]
	if room.Code != created.RoomCode || room.Occupied != 1 || room.Ready != 0 || room.Started {
		t.Fatalf("room = %#v, created = %#v", room, created)
	}
	if room.Mode != game.ModeCompatibility || room.RuleConfig != (game.RuleConfig{}) {
		t.Fatalf("room rules = %q/%#v", room.Mode, room.RuleConfig)
	}
}

func TestWebSocketServerBroadcastsRoomStateWhenPlayerJoins(t *testing.T) {
	server := NewServer()
	url, closeServer := startTestServer(t, server)
	defer closeServer()

	first := dialTestClient(t, url)
	defer first.Close()
	sendMessage(t, first, protocol.Message{Type: protocol.MsgCreateRoom})
	created := readUntil(t, first, protocol.MsgRoomCreated)

	second := dialTestClient(t, url)
	defer second.Close()
	sendMessage(t, second, protocol.Message{Type: protocol.MsgJoinRoom, RoomCode: created.RoomCode})
	_ = readUntil(t, second, protocol.MsgRoomJoined)

	state := readUntil(t, first, protocol.MsgRoomState)
	if state.RoomCode != created.RoomCode {
		t.Fatalf("room code = %q, want %q", state.RoomCode, created.RoomCode)
	}
	if len(state.OccupiedSeats) != 2 || state.OccupiedSeats[0] != 0 || state.OccupiedSeats[1] != 1 {
		t.Fatalf("occupied seats = %#v, want [0 1]", state.OccupiedSeats)
	}
	if state.Started {
		t.Fatalf("started = true, want false before players ready")
	}
}

func TestWebSocketServerRejectsJoinAfterRoomStarted(t *testing.T) {
	server := NewServer()
	url, closeServer := startTestServer(t, server)
	defer closeServer()

	first := dialTestClient(t, url)
	defer first.Close()
	sendMessage(t, first, protocol.Message{Type: protocol.MsgCreateRoom})
	created := readUntil(t, first, protocol.MsgRoomCreated)
	sendMessage(t, first, protocol.Message{Type: protocol.MsgReady})
	started := readUntil(t, first, protocol.MsgRoomState)
	if !started.Started {
		t.Fatalf("started = false, want true")
	}

	late := dialTestClient(t, url)
	defer late.Close()
	sendMessage(t, late, protocol.Message{Type: protocol.MsgJoinRoom, RoomCode: created.RoomCode})
	errMsg := readUntil(t, late, protocol.MsgError)
	if !strings.Contains(errMsg.Error, "already started") {
		t.Fatalf("error = %#v, want already started", errMsg)
	}
}

func TestWebSocketServerRejectsNonTurnCommand(t *testing.T) {
	server := NewServer()
	url, closeServer := startTestServer(t, server)
	defer closeServer()

	first := dialTestClient(t, url)
	defer first.Close()
	sendMessage(t, first, protocol.Message{Type: protocol.MsgCreateRoom})
	created := readUntil(t, first, protocol.MsgRoomCreated)

	second := dialTestClient(t, url)
	defer second.Close()
	sendMessage(t, second, protocol.Message{Type: protocol.MsgJoinRoom, RoomCode: created.RoomCode})
	_ = readUntil(t, second, protocol.MsgRoomJoined)

	sendMessage(t, first, protocol.Message{Type: protocol.MsgReady})
	_ = readUntil(t, first, protocol.MsgRoomState)
	sendMessage(t, second, protocol.Message{Type: protocol.MsgReady})
	_ = readUntil(t, second, protocol.MsgRoomState)

	sendMessage(t, second, protocol.Message{Type: protocol.MsgPlayCommand, Command: game.GameCommand{Kind: game.CommandDiscard, TileIndex: 0}})
	errMsg := readUntil(t, second, protocol.MsgError)
	if !strings.Contains(errMsg.Error, "not the current player") {
		t.Fatalf("error = %#v", errMsg)
	}
}

func TestWebSocketServerBroadcastsAcceptedCommand(t *testing.T) {
	server := NewServer()
	url, closeServer := startTestServer(t, server)
	defer closeServer()

	first := dialTestClient(t, url)
	defer first.Close()
	sendMessage(t, first, protocol.Message{Type: protocol.MsgCreateRoom})
	created := readUntil(t, first, protocol.MsgRoomCreated)

	second := dialTestClient(t, url)
	defer second.Close()
	sendMessage(t, second, protocol.Message{Type: protocol.MsgJoinRoom, RoomCode: created.RoomCode})
	_ = readUntil(t, second, protocol.MsgRoomJoined)

	sendMessage(t, first, protocol.Message{Type: protocol.MsgReady})
	_ = readUntil(t, first, protocol.MsgRoomState)
	sendMessage(t, second, protocol.Message{Type: protocol.MsgReady})
	_ = readUntil(t, second, protocol.MsgRoomState)

	sendMessage(t, first, protocol.Message{Type: protocol.MsgPlayCommand, Command: game.GameCommand{Kind: game.CommandDiscard, TileIndex: 0}})
	firstUpdate := readUntil(t, first, protocol.MsgGameSnapshot)
	update := readUntil(t, second, protocol.MsgGameSnapshot)
	assertPrivateSnapshot(t, firstUpdate.Snapshot, 0)
	assertPrivateSnapshot(t, update.Snapshot, 1)
	assertPrivateSnapshot(t, firstUpdate.Result.Snapshot, 0)
	assertPrivateSnapshot(t, update.Result.Snapshot, 1)
	if (update.Snapshot.Current != 0 && update.Snapshot.Current != 1) || len(update.Snapshot.Events) == 0 {
		t.Fatalf("broadcast update = %#v", update)
	}
	wantHand := 14
	if update.Snapshot.Phase == game.PhaseAwaitingClaim {
		wantHand = 13
	}
	currentHand := len(update.Snapshot.Players[update.Snapshot.Current].Hand)
	if update.Snapshot.Current != 1 {
		wantHand = 0
	}
	if currentHand != wantHand {
		t.Fatalf("visible current hand = %d, want %d for viewer 1/current %d in phase %s", currentHand, wantHand, update.Snapshot.Current, update.Snapshot.Phase)
	}
}

func TestWebSocketServerAdvancesBotsForUnoccupiedSeats(t *testing.T) {
	server := NewServer()
	url, closeServer := startTestServer(t, server)
	defer closeServer()

	first := dialTestClient(t, url)
	defer first.Close()
	sendMessage(t, first, protocol.Message{Type: protocol.MsgCreateRoom})
	_ = readUntil(t, first, protocol.MsgRoomCreated)

	sendMessage(t, first, protocol.Message{Type: protocol.MsgReady})
	started := readUntil(t, first, protocol.MsgRoomState)
	if !started.Started {
		t.Fatalf("started = false, want true for single occupied ready room")
	}
	startEvents := len(started.Snapshot.Events)

	sendMessage(t, first, protocol.Message{Type: protocol.MsgPlayCommand, Command: game.GameCommand{Kind: game.CommandDiscard, TileIndex: 0}})
	update := readUntil(t, first, protocol.MsgGameSnapshot)
	if update.Snapshot.Over {
		return
	}
	if update.Snapshot.Current != 0 {
		t.Fatalf("current = %d, want human seat after bot turns", update.Snapshot.Current)
	}
	wantHand := 14
	if update.Snapshot.Phase == game.PhaseAwaitingClaim {
		wantHand = 13
	}
	if len(update.Snapshot.Players[0].Hand) != wantHand {
		t.Fatalf("human hand = %d, want %d in phase %s", len(update.Snapshot.Players[0].Hand), wantHand, update.Snapshot.Phase)
	}
	if len(update.Snapshot.Events) <= startEvents+1 {
		t.Fatalf("events = %d, want bot actions after human discard", len(update.Snapshot.Events))
	}
}

func TestWebSocketServerBroadcastsReadyRoomState(t *testing.T) {
	server := NewServer()
	url, closeServer := startTestServer(t, server)
	defer closeServer()

	first := dialTestClient(t, url)
	defer first.Close()
	sendMessage(t, first, protocol.Message{Type: protocol.MsgCreateRoom})
	created := readUntil(t, first, protocol.MsgRoomCreated)

	second := dialTestClient(t, url)
	defer second.Close()
	sendMessage(t, second, protocol.Message{Type: protocol.MsgJoinRoom, RoomCode: created.RoomCode})
	_ = readUntil(t, second, protocol.MsgRoomJoined)
	_ = readUntil(t, first, protocol.MsgRoomState)

	sendMessage(t, first, protocol.Message{Type: protocol.MsgReady})
	state := readUntil(t, second, protocol.MsgRoomState)
	if state.RoomCode != created.RoomCode {
		t.Fatalf("room code = %q, want %q", state.RoomCode, created.RoomCode)
	}
	if len(state.ReadySeats) != 1 || state.ReadySeats[0] != 0 {
		t.Fatalf("ready seats = %#v, want [0]", state.ReadySeats)
	}
	if state.Started {
		t.Fatalf("started = true, want false until all occupied seats are ready")
	}

	sendMessage(t, second, protocol.Message{Type: protocol.MsgReady})
	_ = readUntil(t, first, protocol.MsgRoomState)
	started := readUntil(t, first, protocol.MsgRoomState)
	if !started.Started {
		t.Fatalf("started = false, want true after occupied seats ready: %#v", started)
	}
	if len(started.ReadySeats) != 2 {
		t.Fatalf("ready seats = %#v, want two ready seats", started.ReadySeats)
	}
	if len(started.Snapshot.Players[0].Hand) != 14 {
		t.Fatalf("current hand = %d, want 14 after room start draw", len(started.Snapshot.Players[0].Hand))
	}
}

func TestWebSocketServerReconnectsWithToken(t *testing.T) {
	server := NewServer()
	url, closeServer := startTestServer(t, server)
	defer closeServer()

	first := dialTestClient(t, url)
	sendMessage(t, first, protocol.Message{Type: protocol.MsgCreateRoom})
	created := readUntil(t, first, protocol.MsgRoomCreated)
	first.Close()

	reconnectedClient := dialTestClient(t, url)
	defer reconnectedClient.Close()
	sendMessage(t, reconnectedClient, protocol.Message{
		Type:           protocol.MsgReconnect,
		PlayerID:       created.PlayerID,
		ReconnectToken: created.ReconnectToken,
	})
	reconnected := readUntil(t, reconnectedClient, protocol.MsgReconnected)
	if reconnected.PlayerID != created.PlayerID || reconnected.RoomCode != created.RoomCode {
		t.Fatalf("reconnected = %#v created = %#v", reconnected, created)
	}
	assertPrivateSnapshot(t, reconnected.Snapshot, 0)
	if reconnected.Match.Mode != game.ModeCompatibility {
		t.Fatalf("missing match after reconnect: %#v", reconnected.Match)
	}
}

func TestWebSocketServerBroadcastsAndReconnectsPendingHumanClaim(t *testing.T) {
	server := NewServer()
	url, closeServer := startTestServer(t, server)
	defer closeServer()

	first := dialTestClient(t, url)
	defer first.Close()
	sendMessage(t, first, protocol.Message{Type: protocol.MsgCreateRoom, Name: "first"})
	created := readUntil(t, first, protocol.MsgRoomCreated)

	second := dialTestClient(t, url)
	sendMessage(t, second, protocol.Message{Type: protocol.MsgJoinRoom, RoomCode: created.RoomCode, Name: "second"})
	joined := readUntil(t, second, protocol.MsgRoomJoined)
	_ = readUntil(t, first, protocol.MsgRoomState)

	sendMessage(t, first, protocol.Message{Type: protocol.MsgReady})
	_ = readUntil(t, first, protocol.MsgRoomState)
	sendMessage(t, second, protocol.Message{Type: protocol.MsgReady})
	_ = readUntil(t, second, protocol.MsgRoomState)
	_ = readUntil(t, first, protocol.MsgRoomState)

	setPendingPongFixture(t, server, created.RoomCode)
	sendMessage(t, first, protocol.Message{Type: protocol.MsgPlayCommand, Command: game.GameCommand{Kind: game.CommandDiscard, TileIndex: 0}})
	update := readUntil(t, second, protocol.MsgGameSnapshot)
	if update.Snapshot.Phase != game.PhaseAwaitingClaim || update.Snapshot.Current != 1 || update.Snapshot.PendingClaim == nil {
		t.Fatalf("pending claim snapshot = %#v", update.Snapshot)
	}

	sendMessage(t, first, protocol.Message{Type: protocol.MsgPlayCommand, Command: game.GameCommand{Kind: game.CommandPass}})
	errMsg := readUntil(t, first, protocol.MsgError)
	if !strings.Contains(errMsg.Error, "not the current player") {
		t.Fatalf("wrong-seat error = %#v", errMsg)
	}

	second.Close()
	reconnectedClient := dialTestClient(t, url)
	defer reconnectedClient.Close()
	sendMessage(t, reconnectedClient, protocol.Message{
		Type:           protocol.MsgReconnect,
		PlayerID:       joined.PlayerID,
		ReconnectToken: joined.ReconnectToken,
	})
	reconnected := readUntil(t, reconnectedClient, protocol.MsgReconnected)
	if reconnected.Snapshot.Phase != game.PhaseAwaitingClaim || reconnected.Snapshot.PendingClaim == nil {
		t.Fatalf("reconnected pending claim = %#v", reconnected.Snapshot)
	}

	sendMessage(t, reconnectedClient, protocol.Message{Type: protocol.MsgPlayCommand, Command: game.GameCommand{Kind: game.CommandPass}})
	resolved := readUntil(t, reconnectedClient, protocol.MsgGameSnapshot)
	if resolved.Snapshot.Phase != game.PhaseAwaitingDiscard || resolved.Snapshot.PendingClaim != nil || resolved.Snapshot.Current != 1 {
		t.Fatalf("resolved snapshot = %#v", resolved.Snapshot)
	}
}

func setPendingPongFixture(t *testing.T, server *Server, roomCode string) {
	t.Helper()
	server.mu.Lock()
	defer server.mu.Unlock()
	room := server.rooms[roomCode]
	if room == nil {
		t.Fatalf("room %s not found", roomCode)
	}
	round := room.match.Round
	round.Current = 0
	round.Phase = game.PhaseAwaitingDiscard
	round.PendingClaim = nil
	round.Over = false
	round.Events = nil
	round.Players[0].Hand = mustOnlineTiles(t, "3m", "1p", "2p", "4p", "5p", "7p", "8p", "1s", "2s", "4s", "5s", "7s", "N", "N")
	round.Players[0].Discards = nil
	round.Players[1].Hand = mustOnlineTiles(t, "3m", "3m", "1p", "2p", "4p", "5p", "7p", "8p", "1s", "2s", "4s", "5s", "N")
	round.Players[1].Melds = nil
	round.Players[2].Hand = mustOnlineTiles(t, "1p", "2p", "4p", "5p", "7p", "8p", "1s", "2s", "4s", "5s", "7s", "8s", "N")
	round.Players[3].Hand = mustOnlineTiles(t, "1p", "2p", "4p", "5p", "7p", "8p", "1s", "2s", "4s", "5s", "7s", "8s", "S")
}

func mustOnlineTiles(t *testing.T, texts ...string) []game.Tile {
	t.Helper()
	tiles := make([]game.Tile, 0, len(texts))
	for _, text := range texts {
		tile, ok := game.ParseTile(text)
		if !ok {
			t.Fatalf("bad tile: %s", text)
		}
		tiles = append(tiles, tile)
	}
	game.SortTiles(tiles)
	return tiles
}

func assertPrivateSnapshot(t *testing.T, snapshot game.GameSnapshot, seat int) {
	t.Helper()
	if snapshot.Seed != 0 || snapshot.ShuffleProof.Seed != 0 {
		t.Fatalf("live snapshot leaked seed: %#v", snapshot)
	}
	for index, player := range snapshot.Players {
		if player.HandCount == 0 {
			t.Fatalf("player %d missing hand count: %#v", index, player)
		}
		if index == seat && len(player.Hand) != player.HandCount {
			t.Fatalf("seat %d hand/count = %d/%d", seat, len(player.Hand), player.HandCount)
		}
		if index != seat && player.Hand != nil {
			t.Fatalf("seat %d received opponent %d hand: %v", seat, index, player.Hand)
		}
	}
}

func TestWebSocketServerRejectsReconnectAfterConfiguredWindow(t *testing.T) {
	server := NewServerWithOptions(ServerOptions{ReconnectWindow: time.Nanosecond})
	url, closeServer := startTestServer(t, server)
	defer closeServer()

	first := dialTestClient(t, url)
	sendMessage(t, first, protocol.Message{Type: protocol.MsgCreateRoom})
	created := readUntil(t, first, protocol.MsgRoomCreated)
	first.Close()

	time.Sleep(2 * time.Millisecond)

	reconnecting := dialTestClient(t, url)
	defer reconnecting.Close()
	sendMessage(t, reconnecting, protocol.Message{
		Type:           protocol.MsgReconnect,
		PlayerID:       created.PlayerID,
		ReconnectToken: created.ReconnectToken,
	})
	errMsg := readUntil(t, reconnecting, protocol.MsgError)
	if !strings.Contains(errMsg.Error, "reconnect window expired") {
		t.Fatalf("error = %#v", errMsg)
	}
}

func TestWebSocketServerDoesNotListExpiredIdleRooms(t *testing.T) {
	server := NewServerWithOptions(ServerOptions{RoomIdleTTL: time.Nanosecond})
	url, closeServer := startTestServer(t, server)
	defer closeServer()

	host := dialTestClient(t, url)
	sendMessage(t, host, protocol.Message{Type: protocol.MsgCreateRoom})
	_ = readUntil(t, host, protocol.MsgRoomCreated)
	host.Close()

	time.Sleep(2 * time.Millisecond)

	lister := dialTestClient(t, url)
	defer lister.Close()
	sendMessage(t, lister, protocol.Message{Type: protocol.MsgListRooms})
	list := readUntil(t, lister, protocol.MsgRoomList)
	if len(list.Rooms) != 0 {
		t.Fatalf("rooms = %#v, want expired room hidden", list.Rooms)
	}
}

func startTestServer(t *testing.T, server *Server) (string, func()) {
	t.Helper()
	httpServer := httptest.NewServer(server)
	return "ws" + strings.TrimPrefix(httpServer.URL, "http"), httpServer.Close
}

func dialTestClient(t *testing.T, url string) *websocket.Conn {
	t.Helper()
	conn, _, err := websocket.DefaultDialer.Dial(url+"/ws", nil)
	if err != nil {
		t.Fatal(err)
	}
	return conn
}

func sendMessage(t *testing.T, conn *websocket.Conn, message protocol.Message) {
	t.Helper()
	if err := conn.WriteJSON(message); err != nil {
		t.Fatal(err)
	}
}

func readUntil(t *testing.T, conn *websocket.Conn, messageType protocol.MessageType) protocol.Message {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	if err := conn.SetReadDeadline(deadline); err != nil {
		t.Fatal(err)
	}
	for {
		var message protocol.Message
		err := conn.ReadJSON(&message)
		if err != nil {
			t.Fatalf("did not receive %s: %v", messageType, err)
		}
		if message.Type == messageType {
			return message
		}
	}
}

func readUntilStartedRoomState(t *testing.T, conn *websocket.Conn) protocol.Message {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	if err := conn.SetReadDeadline(deadline); err != nil {
		t.Fatal(err)
	}
	for {
		var message protocol.Message
		err := conn.ReadJSON(&message)
		if err != nil {
			t.Fatalf("did not receive started room_state: %v", err)
		}
		if message.Type == protocol.MsgRoomState && message.Started {
			return message
		}
	}
}

func firstDiscardAction(actions []game.LegalAction) (game.LegalAction, bool) {
	for _, action := range actions {
		if action.Kind == game.CommandDiscard {
			return action, true
		}
	}
	return game.LegalAction{}, false
}

func TestReplayPrivacyBeforeCompletion(t *testing.T) {
	match, err := game.NewMatch(140014, game.NewRiichiRuleSet(game.DefaultRuleConfig(game.ModeRiichi).Riichi))
	if err != nil {
		t.Fatal(err)
	}
	room := &room{code: "140014", match: match}
	session := &session{playerID: "player", seat: 0}

	message := stateMessageForSession(room, session, protocol.Message{Type: protocol.MsgGameSnapshot})

	if message.Replay != nil || message.ReplayID != "" {
		t.Fatalf("live message leaked replay: %#v", message)
	}
	if message.Snapshot.Seed != 0 || message.Snapshot.Players[1].Hand != nil {
		t.Fatalf("live privacy regressed: %#v", message.Snapshot)
	}
}

func TestReplayDeliveryAfterCompletedMatch(t *testing.T) {
	server := NewServer()
	url, closeServer := startTestServer(t, server)
	defer closeServer()
	client := dialTestClient(t, url)
	defer client.Close()

	sendMessage(t, client, protocol.Message{Type: protocol.MsgCreateRoom, Name: "host"})
	created := readUntil(t, client, protocol.MsgRoomCreated)
	if created.Replay != nil || created.ReplayID != "" {
		t.Fatalf("create leaked replay: %#v", created)
	}
	sendMessage(t, client, protocol.Message{Type: protocol.MsgReady})
	_ = readUntilStartedRoomState(t, client)
	configureCompatibilityWinningRoom(t, server, created.RoomCode)

	sendMessage(t, client, protocol.Message{
		Type:    protocol.MsgPlayCommand,
		Command: game.GameCommand{Kind: game.CommandWin},
	})
	snapshot := readUntil(t, client, protocol.MsgGameSnapshot)
	if !snapshot.Match.Complete {
		t.Fatalf("final match snapshot = %#v", snapshot.Match)
	}
	ready := readUntil(t, client, protocol.MsgReplayReady)
	data := readUntil(t, client, protocol.MsgReplayData)

	if ready.ReplayID == "" || ready.ReplayID != data.ReplayID || data.Replay == nil {
		t.Fatalf("ready=%#v data=%#v", ready, data)
	}
	if err := game.ValidateReplay(*data.Replay); err != nil {
		t.Fatal(err)
	}
	for seat, player := range data.Replay.Frames[len(data.Replay.Frames)-1].Match.Round.Players {
		if len(player.Hand) == 0 {
			t.Fatalf("completed replay hid seat %d hand", seat)
		}
	}
	encoded, err := json.Marshal(data.Replay)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(encoded, []byte("reconnect_token")) || bytes.Contains(encoded, []byte("ws://")) {
		t.Fatalf("replay contains connection data: %s", encoded)
	}
}

func TestReplayReconnectRequestsCanonicalCompletedPayload(t *testing.T) {
	server := NewServer()
	url, closeServer := startTestServer(t, server)
	defer closeServer()
	client := dialTestClient(t, url)

	sendMessage(t, client, protocol.Message{Type: protocol.MsgCreateRoom, Name: "host"})
	created := readUntil(t, client, protocol.MsgRoomCreated)
	sendMessage(t, client, protocol.Message{Type: protocol.MsgReady})
	_ = readUntilStartedRoomState(t, client)
	configureCompatibilityWinningRoom(t, server, created.RoomCode)
	sendMessage(t, client, protocol.Message{Type: protocol.MsgPlayCommand, Command: game.GameCommand{Kind: game.CommandWin}})
	_ = readUntil(t, client, protocol.MsgGameSnapshot)
	_ = readUntil(t, client, protocol.MsgReplayReady)
	original := readUntil(t, client, protocol.MsgReplayData)
	client.Close()

	reconnectedClient := dialTestClient(t, url)
	defer reconnectedClient.Close()
	sendMessage(t, reconnectedClient, protocol.Message{
		Type:           protocol.MsgReconnect,
		PlayerID:       created.PlayerID,
		ReconnectToken: created.ReconnectToken,
	})
	reconnected := readUntil(t, reconnectedClient, protocol.MsgReconnected)
	if reconnected.ReplayID != original.ReplayID || reconnected.Replay != nil {
		t.Fatalf("reconnected = %#v", reconnected)
	}
	sendMessage(t, reconnectedClient, protocol.Message{Type: protocol.MsgRequestReplay})
	requested := readUntil(t, reconnectedClient, protocol.MsgReplayData)
	want, _ := json.Marshal(original.Replay)
	got, _ := json.Marshal(requested.Replay)
	if !bytes.Equal(want, got) {
		t.Fatalf("requested replay differs\nwant=%s\ngot=%s", want, got)
	}
}

func TestReplayRequestRejectsIncompleteAndUnjoinedClients(t *testing.T) {
	server := NewServer()
	url, closeServer := startTestServer(t, server)
	defer closeServer()

	unjoined := dialTestClient(t, url)
	defer unjoined.Close()
	sendMessage(t, unjoined, protocol.Message{Type: protocol.MsgRequestReplay})
	if message := readUntil(t, unjoined, protocol.MsgError); !strings.Contains(message.Error, "not joined") {
		t.Fatalf("unjoined error = %#v", message)
	}

	joined := dialTestClient(t, url)
	defer joined.Close()
	sendMessage(t, joined, protocol.Message{Type: protocol.MsgCreateRoom})
	_ = readUntil(t, joined, protocol.MsgRoomCreated)
	sendMessage(t, joined, protocol.Message{Type: protocol.MsgRequestReplay})
	if message := readUntil(t, joined, protocol.MsgError); !strings.Contains(message.Error, "not available") {
		t.Fatalf("incomplete error = %#v", message)
	}
}

func configureCompatibilityWinningRoom(t *testing.T, server *Server, roomCode string) {
	t.Helper()
	server.mu.Lock()
	defer server.mu.Unlock()
	room := server.rooms[roomCode]
	if room == nil {
		t.Fatalf("room %s not found", roomCode)
	}
	round := room.match.Round
	round.Current = 0
	round.Phase = game.PhaseAwaitingDiscard
	round.PendingClaim = nil
	round.Over = false
	round.Players[0].Hand = mustOnlineTiles(t,
		"1m", "2m", "3m",
		"4m", "5m", "6m",
		"2p", "3p", "4p",
		"7s", "7s", "7s",
		"E", "E",
	)
}
