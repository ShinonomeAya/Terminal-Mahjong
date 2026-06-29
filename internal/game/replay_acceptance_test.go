package game_test

import (
	"encoding/json"
	"testing"
	"time"

	"mahjong/internal/game"
	"mahjong/internal/replay"
)

func TestDualModeReplayAcceptance(t *testing.T) {
	for _, mode := range []game.RuleMode{game.ModeMCR, game.ModeRiichi} {
		t.Run(string(mode), func(t *testing.T) {
			match := completedAcceptanceMatch(t, mode)
			createdAt := time.Unix(140021, 0).UTC()
			file, err := match.CompletedReplay("acceptance", createdAt, acceptanceParticipants())
			if err != nil {
				t.Fatal(err)
			}
			path, err := replay.Save(t.TempDir(), file)
			if err != nil {
				t.Fatal(err)
			}
			loaded, err := replay.Load(path)
			if err != nil {
				t.Fatal(err)
			}
			wantJSON, _ := json.Marshal(file)
			gotJSON, _ := json.Marshal(loaded)
			if string(gotJSON) != string(wantJSON) {
				t.Fatal("saved replay differs from authoritative export")
			}
			if loaded.FinalStandings != match.Points {
				t.Fatalf("standings=%v source=%v", loaded.FinalStandings, match.Points)
			}
			if mode == game.ModeMCR && len(loaded.MCRSettlements) != len(match.MCRSettlements) {
				t.Fatalf("MCR settlements=%d source=%d", len(loaded.MCRSettlements), len(match.MCRSettlements))
			}
			if mode == game.ModeRiichi && len(loaded.RiichiSettlements) != len(match.RiichiSettlements) {
				t.Fatalf("Riichi settlements=%d source=%d", len(loaded.RiichiSettlements), len(match.RiichiSettlements))
			}
			for index, frame := range loaded.Frames {
				if frame.Index != index {
					t.Fatalf("frame %d has index %d", index, frame.Index)
				}
			}
			var commands []game.GameCommand
			for _, frame := range loaded.Frames {
				if frame.Command != nil {
					commands = append(commands, *frame.Command)
				}
			}
			if len(commands) != len(loaded.Commands) {
				t.Fatalf("frame commands=%d commands=%d", len(commands), len(loaded.Commands))
			}
			resealed, err := match.CompletedReplay("acceptance", createdAt, acceptanceParticipants())
			if err != nil {
				t.Fatal(err)
			}
			if resealed.Checksum != loaded.Checksum || resealed.ReplayID != loaded.ReplayID {
				t.Fatalf("unstable replay identity checksum=%q/%q id=%q/%q", resealed.Checksum, loaded.Checksum, resealed.ReplayID, loaded.ReplayID)
			}
		})
	}
}

func completedAcceptanceMatch(t *testing.T, mode game.RuleMode) *game.Match {
	t.Helper()
	config := game.DefaultRuleConfig(mode)
	var rules game.RuleSet
	switch mode {
	case game.ModeMCR:
		rules = game.NewMCRRuleSet(config.MCR)
	case game.ModeRiichi:
		rules = game.NewRiichiRuleSet(config.Riichi)
	default:
		t.Fatalf("unsupported mode %q", mode)
	}
	match, err := game.NewMatch(140021, rules)
	if err != nil {
		t.Fatal(err)
	}
	if mode == game.ModeMCR {
		match.RoundNumber = 16
		match.Round.HandNumber = 16
	} else {
		match.RoundNumber = 8
		match.Round.HandNumber = 8
		match.Dealer = 1
		match.Round.Dealer = 1
	}
	match.Round.Current = 0
	match.Round.Phase = game.PhaseAwaitingDiscard
	match.Round.PendingClaim = nil
	match.Round.Over = false
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
	match.Round.Players[0].Hand = acceptanceTiles(t, hand...)
	draw := "5p"
	if mode == game.ModeRiichi {
		draw = "4m"
	}
	match.Round.RecordEvent(game.EventDraw, 0, acceptanceTiles(t, draw)[0], "")
	result := match.ApplyCommand(game.GameCommand{PlayerID: "0", Kind: game.CommandWin})
	if !result.OK || !match.Complete {
		t.Fatalf("result=%#v complete=%t", result, match.Complete)
	}
	return match
}

func acceptanceParticipants() []game.ReplayParticipant {
	return []game.ReplayParticipant{
		{Seat: 0, ID: "0", Name: "You"},
		{Seat: 1, ID: "1", Name: "AI-1"},
		{Seat: 2, ID: "2", Name: "AI-2"},
		{Seat: 3, ID: "3", Name: "AI-3"},
	}
}

func acceptanceTiles(t *testing.T, values ...string) []game.Tile {
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
