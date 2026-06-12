package online

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"mahjong/internal/game"
	"mahjong/internal/protocol"
)

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
	update := readUntil(t, second, protocol.MsgGameSnapshot)
	if update.Snapshot.Current != 1 || len(update.Snapshot.Events) == 0 {
		t.Fatalf("broadcast update = %#v", update)
	}
	if len(update.Snapshot.Players[1].Hand) != 14 {
		t.Fatalf("next player hand = %d, want 14 after turn draw", len(update.Snapshot.Players[1].Hand))
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
	if len(update.Snapshot.Players[0].Hand) != 14 {
		t.Fatalf("human hand = %d, want 14 after bot turns return control", len(update.Snapshot.Players[0].Hand))
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
	if reconnected.Snapshot.Seed == 0 {
		t.Fatalf("missing snapshot after reconnect: %#v", reconnected)
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
