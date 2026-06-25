package bot

import (
	"context"
	"testing"

	"mahjong/internal/game"
)

func TestHeuristicBotWinsCompleteHand(t *testing.T) {
	snapshot := snapshotWithHand(t,
		"1m", "2m", "3m",
		"4m", "5m", "6m",
		"2p", "3p", "4p",
		"7s", "7s", "7s",
		"E", "E",
	)

	command := NewHeuristicBot().Decide(context.Background(), snapshot, "0")

	if command.Kind != game.CommandWin {
		t.Fatalf("command = %#v, want win", command)
	}
}

func TestHeuristicBotDeclaresKongWhenAvailable(t *testing.T) {
	snapshot := snapshotWithHand(t, "1m", "1m", "1m", "1m", "2m", "3m")

	command := NewHeuristicBot().Decide(context.Background(), snapshot, "0")

	if command.Kind != game.CommandKong || command.Tile != "1m" {
		t.Fatalf("command = %#v, want kong 1m", command)
	}
}

func TestHeuristicBotDiscardsLegalTile(t *testing.T) {
	snapshot := snapshotWithHand(t, "1m", "2m", "3m", "E", "F")

	command := NewHeuristicBot().Decide(context.Background(), snapshot, "0")

	if command.Kind != game.CommandDiscard {
		t.Fatalf("command = %#v, want discard", command)
	}
	if command.TileIndex < 0 || command.TileIndex >= len(snapshot.Players[0].Hand) {
		t.Fatalf("tile index = %d, want legal hand index", command.TileIndex)
	}
}

func TestRiichiHeuristicBotChoosesOnlyFromLegalActions(t *testing.T) {
	snapshot := snapshotWithHand(t,
		"1m", "2m", "3m",
		"4m", "5m", "6m",
		"2p", "3p", "4p",
		"7s", "7s", "7s",
		"E", "E",
	)
	snapshot.LegalActions = []game.LegalAction{{Kind: game.CommandDiscard, TileIndex: 3}}

	command := NewHeuristicBot().Decide(context.Background(), snapshot, "0")

	if command.Kind != game.CommandDiscard || command.TileIndex != 3 {
		t.Fatalf("command = %#v, want only legal discard index 3", command)
	}
}

func TestHeuristicBotRejectsUnknownPlayer(t *testing.T) {
	snapshot := snapshotWithHand(t, "1m", "2m", "3m")

	command := NewHeuristicBot().Decide(context.Background(), snapshot, "missing")

	if command.Kind != game.CommandQuit {
		t.Fatalf("command = %#v, want safe quit for unknown player", command)
	}
}

func TestHeuristicBotAcceptsActiveDiscardWin(t *testing.T) {
	snapshot := snapshotWithHand(t,
		"1m", "2m",
		"4m", "5m", "6m",
		"2p", "3p", "4p",
		"7s", "7s", "7s",
		"E", "E",
	)
	snapshot.Phase = game.PhaseAwaitingClaim
	snapshot.Current = 0
	snapshot.PendingClaim = &game.PendingClaim{
		Discarder: 3,
		Tile:      mustBotTiles(t, "3m")[0],
		Options:   []game.ClaimOption{{Kind: game.ClaimWin, Player: 0}},
	}

	command := NewHeuristicBot().Decide(context.Background(), snapshot, "0")

	if command.Kind != game.CommandClaimWin {
		t.Fatalf("command = %#v, want discard win", command)
	}
}

func TestHeuristicBotReturnsLegalActivePongCommand(t *testing.T) {
	snapshot := snapshotWithHand(t, "3m", "3m", "1p", "2p", "4p", "5p", "7p", "8p", "1s", "2s", "4s", "5s", "N")
	snapshot.Phase = game.PhaseAwaitingClaim
	snapshot.Current = 0
	snapshot.PendingClaim = &game.PendingClaim{
		Discarder: 3,
		Tile:      mustBotTiles(t, "3m")[0],
		Options:   []game.ClaimOption{{Kind: game.ClaimPong, Player: 0, Consumed: mustBotTiles(t, "3m", "3m")}},
	}

	command := NewHeuristicBot().Decide(context.Background(), snapshot, "0")

	if command.Kind != game.CommandPong && command.Kind != game.CommandPass {
		t.Fatalf("command = %#v, want legal pong or pass", command)
	}
}

func TestHeuristicBotReturnsLegalActiveChowCommand(t *testing.T) {
	snapshot := snapshotWithHand(t, "1m", "2m", "2m", "4m", "1p", "2p", "4p", "5p", "7p", "1s", "2s", "4s", "N")
	snapshot.Phase = game.PhaseAwaitingClaim
	snapshot.Current = 0
	snapshot.PendingClaim = &game.PendingClaim{
		Discarder: 3,
		Tile:      mustBotTiles(t, "3m")[0],
		Options: []game.ClaimOption{
			{Kind: game.ClaimChow, Player: 0, Consumed: mustBotTiles(t, "1m", "2m")},
			{Kind: game.ClaimChow, Player: 0, Consumed: mustBotTiles(t, "2m", "4m")},
		},
	}

	command := NewHeuristicBot().Decide(context.Background(), snapshot, "0")

	if command.Kind == game.CommandPass {
		return
	}
	if command.Kind != game.CommandChow || command.TileIndex < 0 || command.TileIndex >= 2 {
		t.Fatalf("command = %#v, want legal chow option or pass", command)
	}
}

func snapshotWithHand(t *testing.T, texts ...string) game.GameSnapshot {
	t.Helper()
	g := game.NewGame(5)
	g.Players[0].Hand = mustBotTiles(t, texts...)
	return g.Snapshot()
}

func mustBotTiles(t *testing.T, texts ...string) []game.Tile {
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
