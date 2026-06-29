package online

import (
	"bytes"
	"encoding/json"
	"testing"

	"mahjong/internal/game"
	"mahjong/internal/protocol"
)

func TestDualModeOnlineReplayPrivacyAcceptance(t *testing.T) {
	for _, mode := range []game.RuleMode{game.ModeMCR, game.ModeRiichi} {
		t.Run(string(mode), func(t *testing.T) {
			server := NewServer()
			url, closeServer := startTestServer(t, server)
			defer closeServer()
			client := dialTestClient(t, url)

			sendMessage(t, client, protocol.Message{
				Type:       protocol.MsgCreateRoom,
				Name:       "host",
				Mode:       mode,
				RuleConfig: game.DefaultRuleConfig(mode),
			})
			created := readUntil(t, client, protocol.MsgRoomCreated)
			sendMessage(t, client, protocol.Message{Type: protocol.MsgReady})
			live := readUntilStartedRoomState(t, client)
			assertLiveReplayPrivacy(t, live)
			configureAcceptanceWinningRoom(t, server, created.RoomCode, mode)

			sendMessage(t, client, protocol.Message{
				Type:    protocol.MsgPlayCommand,
				Command: game.GameCommand{Kind: game.CommandWin},
			})
			final := readUntil(t, client, protocol.MsgGameSnapshot)
			assertLiveReplayPrivacy(t, final)
			_ = readUntil(t, client, protocol.MsgReplayReady)
			data := readUntil(t, client, protocol.MsgReplayData)
			if data.Replay == nil {
				t.Fatal("completed replay payload missing")
			}
			if err := game.ValidateReplay(*data.Replay); err != nil {
				t.Fatal(err)
			}
			for seat, player := range data.Replay.Frames[len(data.Replay.Frames)-1].Match.Round.Players {
				if len(player.Hand) == 0 {
					t.Fatalf("completed replay hid seat %d", seat)
				}
			}
			if len(data.Replay.Commands) == 0 || len(data.Replay.Frames) < 2 || data.Replay.ShuffleProof.WallHash == "" {
				t.Fatalf("incomplete replay evidence: %#v", data.Replay)
			}
			if mode == game.ModeMCR && len(data.Replay.MCRSettlements) == 0 {
				t.Fatal("MCR settlement missing")
			}
			if mode == game.ModeRiichi && len(data.Replay.RiichiSettlements) == 0 {
				t.Fatal("Riichi settlement missing")
			}
			encoded, err := json.Marshal(data.Replay)
			if err != nil {
				t.Fatal(err)
			}
			for _, forbidden := range []string{"reconnect_token", "ws://", "127.0.0.1", ".mahjong-session"} {
				if bytes.Contains(encoded, []byte(forbidden)) {
					t.Fatalf("replay contains %q: %s", forbidden, encoded)
				}
			}

			client.Close()
			reconnected := dialTestClient(t, url)
			defer reconnected.Close()
			sendMessage(t, reconnected, protocol.Message{
				Type:           protocol.MsgReconnect,
				PlayerID:       created.PlayerID,
				ReconnectToken: created.ReconnectToken,
			})
			available := readUntil(t, reconnected, protocol.MsgReconnected)
			if available.ReplayID != data.ReplayID || available.Replay != nil {
				t.Fatalf("reconnected=%#v", available)
			}
			sendMessage(t, reconnected, protocol.Message{Type: protocol.MsgRequestReplay})
			requested := readUntil(t, reconnected, protocol.MsgReplayData)
			want, _ := json.Marshal(data.Replay)
			got, _ := json.Marshal(requested.Replay)
			if !bytes.Equal(want, got) {
				t.Fatal("reconnected replay differs from completion payload")
			}
		})
	}
}

func assertLiveReplayPrivacy(t *testing.T, message protocol.Message) {
	t.Helper()
	if message.Replay != nil || message.ReplayID != "" || message.Snapshot.Seed != 0 {
		t.Fatalf("live message leaked replay or seed: %#v", message)
	}
	for seat, player := range message.Snapshot.Players {
		if seat != 0 && len(player.Hand) != 0 {
			t.Fatalf("live message leaked seat %d hand", seat)
		}
	}
	if message.Snapshot.Riichi != nil && len(message.Snapshot.Riichi.UraIndicators) != 0 {
		t.Fatalf("live message leaked ura indicators: %#v", message.Snapshot.Riichi.UraIndicators)
	}
}

func configureAcceptanceWinningRoom(t *testing.T, server *Server, roomCode string, mode game.RuleMode) {
	t.Helper()
	server.mu.Lock()
	defer server.mu.Unlock()
	room := server.rooms[roomCode]
	if room == nil {
		t.Fatalf("room %s not found", roomCode)
	}
	match := room.match
	round := match.Round
	if mode == game.ModeMCR {
		match.RoundNumber = 16
		round.HandNumber = 16
	} else {
		match.RoundNumber = 8
		round.HandNumber = 8
		match.Dealer = 1
		round.Dealer = 1
	}
	round.Current = 0
	round.Phase = game.PhaseAwaitingDiscard
	round.PendingClaim = nil
	round.Over = false
	hand := []string{
		"2m", "3m", "4m",
		"3m", "4m", "5m",
		"4p", "5p", "6p",
		"6s", "7s", "8s",
		"5p", "5p",
	}
	if mode == game.ModeRiichi {
		hand = []string{
			"2m", "3m", "4m",
			"1p", "2p", "3p",
			"4p", "5p", "6p",
			"7s", "8s", "9s",
			"5s", "5s",
		}
	}
	round.Players[0].Hand = acceptanceOnlineTiles(t, hand...)
	draw := "5p"
	if mode == game.ModeRiichi {
		draw = "4m"
	}
	round.RecordEvent(game.EventDraw, 0, acceptanceOnlineTiles(t, draw)[0], "")
	hasWin := false
	for _, action := range round.SnapshotFor("0").LegalActions {
		if action.Kind == game.CommandWin {
			hasWin = true
			break
		}
	}
	if !hasWin {
		t.Fatalf("%s fixture has no win action: %#v", mode, round.SnapshotFor("0").LegalActions)
	}
}

func acceptanceOnlineTiles(t *testing.T, values ...string) []game.Tile {
	t.Helper()
	tiles := make([]game.Tile, len(values))
	for index, value := range values {
		tile, ok := game.ParseTile(value)
		if !ok {
			t.Fatalf("invalid tile %q", value)
		}
		tiles[index] = tile
	}
	return tiles
}
