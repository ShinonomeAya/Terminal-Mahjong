package online

import (
	"bytes"
	"encoding/json"
	"testing"

	"mahjong/internal/game"
	"mahjong/internal/protocol"
)

func TestDualModePrivacyAndReconnectMatrix(t *testing.T) {
	tests := []struct {
		name   string
		mode   game.RuleMode
		config game.RuleConfig
	}{
		{name: "mcr", mode: game.ModeMCR, config: game.DefaultRuleConfig(game.ModeMCR)},
		{name: "riichi", mode: game.ModeRiichi, config: game.DefaultRuleConfig(game.ModeRiichi)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := NewServer()
			url, closeServer := startTestServer(t, server)
			defer closeServer()

			host := dialTestClient(t, url)
			sendMessage(t, host, protocol.Message{Type: protocol.MsgCreateRoom, Name: "host", Mode: test.mode, RuleConfig: test.config})
			created := readUntil(t, host, protocol.MsgRoomCreated)
			guest := dialTestClient(t, url)
			defer guest.Close()
			sendMessage(t, guest, protocol.Message{Type: protocol.MsgJoinRoom, RoomCode: created.RoomCode, Name: "guest"})
			_ = readUntil(t, guest, protocol.MsgRoomJoined)

			if test.mode == game.ModeMCR {
				server.mu.Lock()
				server.rooms[created.RoomCode].match.Round.Players[0].Flowers = []game.Tile{game.FlowerPlum}
				server.mu.Unlock()
			}

			sendMessage(t, host, protocol.Message{Type: protocol.MsgReady})
			sendMessage(t, guest, protocol.Message{Type: protocol.MsgReady})
			hostStarted := readUntilStartedRoomState(t, host)
			guestStarted := readUntilStartedRoomState(t, guest)
			assertPrivateSnapshot(t, hostStarted.Match.Round, 0)
			assertPrivateSnapshot(t, guestStarted.Match.Round, 1)
			assertDualModeSpecificPrivacy(t, test.mode, hostStarted.Match.Round, guestStarted.Match.Round)

			discard, ok := firstDiscardAction(hostStarted.Match.Round.LegalActions)
			if !ok {
				t.Fatalf("host legal actions missing discard: %#v", hostStarted.Match.Round.LegalActions)
			}
			sendMessage(t, host, protocol.Message{
				Type:    protocol.MsgPlayCommand,
				Command: game.GameCommand{Kind: game.CommandDiscard, TileIndex: discard.TileIndex},
			})
			hostUpdate := readUntil(t, host, protocol.MsgGameSnapshot)
			guestUpdate := readUntil(t, guest, protocol.MsgGameSnapshot)
			assertPrivateSnapshot(t, hostUpdate.Match.Round, 0)
			assertPrivateSnapshot(t, guestUpdate.Match.Round, 1)
			assertDualModeSpecificPrivacy(t, test.mode, hostUpdate.Match.Round, guestUpdate.Match.Round)
			want, err := json.Marshal(hostUpdate.Match)
			if err != nil {
				t.Fatal(err)
			}
			host.Close()

			reconnectedClient := dialTestClient(t, url)
			defer reconnectedClient.Close()
			sendMessage(t, reconnectedClient, protocol.Message{
				Type:           protocol.MsgReconnect,
				PlayerID:       created.PlayerID,
				ReconnectToken: created.ReconnectToken,
			})
			reconnected := readUntil(t, reconnectedClient, protocol.MsgReconnected)
			got, err := json.Marshal(reconnected.Match)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, want) {
				t.Fatalf("reconnected private match differs\nbefore=%s\nafter=%s", want, got)
			}
		})
	}
}

func assertDualModeSpecificPrivacy(t *testing.T, mode game.RuleMode, first game.GameSnapshot, second game.GameSnapshot) {
	t.Helper()
	switch mode {
	case game.ModeMCR:
		for _, snapshot := range []game.GameSnapshot{first, second} {
			if len(snapshot.Players[0].Flowers) != 1 || snapshot.Players[0].Flowers[0] != game.FlowerPlum {
				t.Fatalf("MCR public flowers missing: %#v", snapshot.Players[0].Flowers)
			}
		}
	case game.ModeRiichi:
		for _, snapshot := range []game.GameSnapshot{first, second} {
			if snapshot.Riichi == nil || snapshot.Riichi.DeadWallCount != 14 || len(snapshot.Riichi.DoraIndicators) == 0 {
				t.Fatalf("Riichi public state missing: %#v", snapshot.Riichi)
			}
			if len(snapshot.Riichi.UraIndicators) != 0 {
				t.Fatalf("Riichi ura indicators leaked: %#v", snapshot.Riichi.UraIndicators)
			}
		}
	}
}
