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

func startClientTestServer() (string, func()) {
	server := online.NewServer()
	httpServer := httptest.NewServer(server)
	return "ws" + strings.TrimPrefix(httpServer.URL, "http") + "/ws", httpServer.Close
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
