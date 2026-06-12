package online

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"mahjong/internal/game"
	"mahjong/internal/protocol"
)

func TestClientCreatesRoomAndPersistsSession(t *testing.T) {
	server := NewServer()
	url, closeServer := startTestServer(t, server)
	defer closeServer()

	client := NewClient(url+"/ws", "first")
	defer client.Close()

	created, err := client.CreateRoom(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if created.RoomCode == "" || created.PlayerID == "" || created.ReconnectToken == "" {
		t.Fatalf("created = %#v", created)
	}

	sessionPath := filepath.Join(t.TempDir(), "session.json")
	if err := SaveClientSession(sessionPath, client.Session()); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadClientSession(sessionPath)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.PlayerID != created.PlayerID || loaded.ReconnectToken != created.ReconnectToken || loaded.RoomCode != created.RoomCode {
		t.Fatalf("loaded = %#v created = %#v", loaded, created)
	}
}

func TestClientReconnectsFromSavedSession(t *testing.T) {
	server := NewServer()
	url, closeServer := startTestServer(t, server)
	defer closeServer()

	first := NewClient(url+"/ws", "first")
	created, err := first.CreateRoom(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	first.Close()

	reconnecting := NewClient(url+"/ws", "first")
	defer reconnecting.Close()
	reconnected, err := reconnecting.Reconnect(context.Background(), ClientSession{
		ServerURL:      url + "/ws",
		Name:           "first",
		PlayerID:       created.PlayerID,
		ReconnectToken: created.ReconnectToken,
		RoomCode:       created.RoomCode,
	})
	if err != nil {
		t.Fatal(err)
	}
	if reconnected.Type != protocol.MsgReconnected || reconnected.PlayerID != created.PlayerID || reconnected.RoomCode != created.RoomCode {
		t.Fatalf("reconnected = %#v created = %#v", reconnected, created)
	}
	if reconnected.Snapshot.WallCount == 0 {
		t.Fatalf("missing snapshot after reconnect: %#v", reconnected)
	}
}

func TestClientJoinsRoom(t *testing.T) {
	server := NewServer()
	url, closeServer := startTestServer(t, server)
	defer closeServer()

	first := NewClient(url+"/ws", "first")
	defer first.Close()
	created, err := first.CreateRoom(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	second := NewClient(url+"/ws", "second")
	defer second.Close()
	joined, err := second.JoinRoom(context.Background(), created.RoomCode)
	if err != nil {
		t.Fatal(err)
	}
	if joined.Type != protocol.MsgRoomJoined || joined.RoomCode != created.RoomCode || joined.PlayerID == "" {
		t.Fatalf("joined = %#v", joined)
	}
}

func TestClientSendsCommandAndReceivesBroadcast(t *testing.T) {
	server := NewServer()
	url, closeServer := startTestServer(t, server)
	defer closeServer()

	first := NewClient(url+"/ws", "first")
	defer first.Close()
	created, err := first.CreateRoom(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	second := NewClient(url+"/ws", "second")
	defer second.Close()
	if _, err := second.JoinRoom(context.Background(), created.RoomCode); err != nil {
		t.Fatal(err)
	}

	if err := first.SendCommand(context.Background(), game.GameCommand{Kind: game.CommandDiscard, TileIndex: 0}); err != nil {
		t.Fatal(err)
	}
	update, err := second.ReadUntil(context.Background(), 2*time.Second, protocol.MsgGameSnapshot)
	if err != nil {
		t.Fatal(err)
	}
	if update.Snapshot.Current != 1 || len(update.Snapshot.Events) == 0 {
		t.Fatalf("update = %#v", update)
	}
}
