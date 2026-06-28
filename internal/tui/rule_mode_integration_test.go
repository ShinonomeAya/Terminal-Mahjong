package tui

import (
	"context"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"mahjong/internal/game"
	"mahjong/internal/online"
	"mahjong/internal/protocol"
)

func TestRuleModeParityBetweenLocalAndOnlineCreation(t *testing.T) {
	tests := []struct {
		name   string
		mode   game.RuleMode
		config game.RuleConfig
	}{
		{name: "compatibility", mode: game.ModeCompatibility, config: game.RuleConfig{}},
		{name: "mcr", mode: game.ModeMCR, config: game.DefaultRuleConfig(game.ModeMCR)},
		{name: "riichi no red", mode: game.ModeRiichi, config: riichiParityConfig(0)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			local := newStartedGameWithRules(test.mode, test.config)
			onlineSnapshot := startedOnlineRound(t, test.mode, test.config)

			if local.Mode != onlineSnapshot.Match.Mode || local.RuleConfig != onlineSnapshot.Match.RuleConfig {
				t.Fatalf("rules local=%q/%#v online=%q/%#v", local.Mode, local.RuleConfig, onlineSnapshot.Match.Mode, onlineSnapshot.Match.RuleConfig)
			}
			if test.mode == game.ModeMCR {
				if local.Snapshot().WallCount <= 0 || onlineSnapshot.Match.Round.WallCount <= 0 {
					t.Fatalf("MCR wall local=%d online=%d", local.Snapshot().WallCount, onlineSnapshot.Match.Round.WallCount)
				}
			} else if local.Snapshot().WallCount != onlineSnapshot.Match.Round.WallCount {
				t.Fatalf("wall local=%d online=%d", local.Snapshot().WallCount, onlineSnapshot.Match.Round.WallCount)
			}
			if !reflect.DeepEqual(local.Snapshot().LegalActions != nil, onlineSnapshot.Match.Round.LegalActions != nil) {
				t.Fatalf("legal action availability local=%#v online=%#v", local.Snapshot().LegalActions, onlineSnapshot.Match.Round.LegalActions)
			}
		})
	}
}

func riichiParityConfig(redFives int) game.RuleConfig {
	config := game.DefaultRuleConfig(game.ModeRiichi)
	config.Riichi.RedFives = redFives
	return config
}

func startedOnlineRound(t *testing.T, mode game.RuleMode, config game.RuleConfig) protocol.Message {
	t.Helper()
	server := online.NewServer()
	httpServer := httptest.NewServer(server)
	defer httpServer.Close()
	serverURL := "ws" + strings.TrimPrefix(httpServer.URL, "http") + "/ws"
	client := online.NewClient(serverURL, "host")
	defer client.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	created, err := client.CreateRoomWithRules(ctx, mode, config)
	if err != nil {
		t.Fatal(err)
	}
	if err := client.SendReady(ctx); err != nil {
		t.Fatal(err)
	}
	started, err := client.ReadUntil(ctx, 2*time.Second, protocol.MsgRoomState, protocol.MsgError)
	if err != nil {
		t.Fatal(err)
	}
	if !started.Started || started.RoomCode != created.RoomCode {
		t.Fatalf("started message = %#v", started)
	}
	return started
}
