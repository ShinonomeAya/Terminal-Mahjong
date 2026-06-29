package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"mahjong/internal/game"
	"mahjong/internal/online"
	"mahjong/internal/protocol"
	"mahjong/internal/replay"
)

func TestRunClientReadyFlagSendsReady(t *testing.T) {
	serverURL, closeServer := startClientTestServer()
	defer closeServer()

	sessionPath := filepath.Join(t.TempDir(), "session.json")
	var output bytes.Buffer
	if err := run(context.Background(), []string{
		"-server", serverURL,
		"-session", sessionPath,
		"-ready",
	}, &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "type=room_state") {
		t.Fatalf("output missing room_state:\n%s", output.String())
	}
	if !strings.Contains(output.String(), "started=true") {
		t.Fatalf("output missing started=true:\n%s", output.String())
	}
}

func TestRunClientWinFlagSendsWinCommand(t *testing.T) {
	serverURL, commands, closeServer := startClientCommandCaptureServer(t)
	defer closeServer()

	sessionPath := filepath.Join(t.TempDir(), "session.json")
	if err := online.SaveClientSession(sessionPath, online.ClientSession{
		ServerURL:      serverURL,
		Name:           "first",
		PlayerID:       "p1",
		ReconnectToken: "token",
		RoomCode:       "000001",
	}); err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	if err := run(context.Background(), []string{
		"-server", serverURL,
		"-session", sessionPath,
		"-reconnect",
		"-win",
	}, &output); err != nil {
		t.Fatal(err)
	}
	message := readCapturedCommand(t, commands)
	if message.Type != protocol.MsgPlayCommand || message.Command.Kind != game.CommandWin {
		t.Fatalf("message = %#v, want win command", message)
	}
}

func TestRunClientKongFlagSendsKongCommand(t *testing.T) {
	serverURL, commands, closeServer := startClientCommandCaptureServer(t)
	defer closeServer()

	sessionPath := filepath.Join(t.TempDir(), "session.json")
	if err := online.SaveClientSession(sessionPath, online.ClientSession{
		ServerURL:      serverURL,
		Name:           "first",
		PlayerID:       "p1",
		ReconnectToken: "token",
		RoomCode:       "000001",
	}); err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	if err := run(context.Background(), []string{
		"-server", serverURL,
		"-session", sessionPath,
		"-reconnect",
		"-kong", "1m",
	}, &output); err != nil {
		t.Fatal(err)
	}
	message := readCapturedCommand(t, commands)
	if message.Type != protocol.MsgPlayCommand || message.Command.Kind != game.CommandKong || message.Command.Tile != "1m" {
		t.Fatalf("message = %#v, want kong 1m command", message)
	}
}

func TestRunListsRooms(t *testing.T) {
	server := online.NewServer()
	httpServer := httptest.NewServer(server)
	defer httpServer.Close()
	url := "ws" + strings.TrimPrefix(httpServer.URL, "http")

	host := online.NewClient(url+"/ws", "host")
	defer host.Close()
	created, err := host.CreateRoom(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	var out strings.Builder
	err = run(context.Background(), []string{"-server", url + "/ws", "-list"}, &out)
	if err != nil {
		t.Fatal(err)
	}
	text := out.String()
	for _, want := range []string{"rooms=1", created.RoomCode, "occupied=1", "ready=0"} {
		if !strings.Contains(text, want) {
			t.Fatalf("output missing %q:\n%s", want, text)
		}
	}
}

func TestRunClientCreatesRiichiRoomWithFlags(t *testing.T) {
	serverURL, closeServer := startClientTestServer()
	defer closeServer()

	sessionPath := filepath.Join(t.TempDir(), "session.json")
	var output bytes.Buffer
	if err := run(context.Background(), []string{
		"-server", serverURL,
		"-session", sessionPath,
		"-mode", "riichi",
		"-red-fives", "0",
	}, &output); err != nil {
		t.Fatal(err)
	}
	text := output.String()
	for _, want := range []string{"type=room_created", "mode=riichi", "red_fives=0"} {
		if !strings.Contains(text, want) {
			t.Fatalf("output missing %q:\n%s", want, text)
		}
	}
}

func TestWatchOncePrintsRoomState(t *testing.T) {
	serverURL, closeServer := startClientMessageServer(t, protocol.Message{
		Type:     protocol.MsgRoomState,
		RoomCode: "000777",
		Started:  true,
		Snapshot: game.NewGame(13).Snapshot(),
	})
	defer closeServer()

	client := online.NewClient(serverURL, "first")
	defer client.Close()

	var output bytes.Buffer
	if err := watchOnce(context.Background(), client, 1, t.TempDir(), &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "type=room_state") || !strings.Contains(output.String(), "room=000777") {
		t.Fatalf("output missing room_state:\n%s", output.String())
	}
}

func TestWatchOnceSavesReplayData(t *testing.T) {
	file := clientReplayFixture(t, "online-client")
	serverURL, closeServer := startClientMessageServer(t, protocol.Message{
		Type:     protocol.MsgReplayData,
		ReplayID: file.ReplayID,
		Replay:   &file,
	})
	defer closeServer()

	client := online.NewClient(serverURL, "first")
	defer client.Close()
	replayDir := t.TempDir()
	var output bytes.Buffer
	if err := watchOnce(context.Background(), client, 1, replayDir, &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "replay=") {
		t.Fatalf("output missing replay path:\n%s", output.String())
	}
	entries, issues, err := replay.List(replayDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 0 || len(entries) != 1 || entries[0].ReplayID != file.ReplayID {
		t.Fatalf("entries=%#v issues=%#v", entries, issues)
	}
}

func startClientTestServer() (string, func()) {
	server := online.NewServer()
	httpServer := httptest.NewServer(server)
	return "ws" + strings.TrimPrefix(httpServer.URL, "http") + "/ws", httpServer.Close
}

func startClientMessageServer(t *testing.T, message protocol.Message) (string, func()) {
	t.Helper()
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Error(err)
			return
		}
		defer conn.Close()
		if err := conn.WriteJSON(message); err != nil {
			t.Error(err)
		}
	}))
	return "ws" + strings.TrimPrefix(server.URL, "http"), server.Close
}

func startClientCommandCaptureServer(t *testing.T) (string, <-chan protocol.Message, func()) {
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
		if message.Type == protocol.MsgReconnect {
			if err := conn.WriteJSON(protocol.Message{
				Type:           protocol.MsgReconnected,
				PlayerID:       message.PlayerID,
				ReconnectToken: message.ReconnectToken,
				RoomCode:       "000001",
				Snapshot:       game.NewGame(13).Snapshot(),
			}); err != nil {
				t.Error(err)
				return
			}
			if err := conn.ReadJSON(&message); err != nil {
				t.Error(err)
				return
			}
		}
		commands <- message
		_ = conn.WriteJSON(protocol.Message{
			Type:     protocol.MsgGameSnapshot,
			RoomCode: "000001",
			Snapshot: game.NewGame(13).Snapshot(),
		})
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

func clientReplayFixture(t *testing.T, id string) game.ReplayFile {
	t.Helper()
	round := game.NewGame(140016).Snapshot()
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
