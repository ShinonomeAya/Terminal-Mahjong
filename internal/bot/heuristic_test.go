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

func TestHeuristicBotRejectsUnknownPlayer(t *testing.T) {
	snapshot := snapshotWithHand(t, "1m", "2m", "3m")

	command := NewHeuristicBot().Decide(context.Background(), snapshot, "missing")

	if command.Kind != game.CommandQuit {
		t.Fatalf("command = %#v, want safe quit for unknown player", command)
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
