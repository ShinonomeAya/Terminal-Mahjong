package tui

import (
	"testing"

	"mahjong/internal/game"
)

func mustUITiles(t *testing.T, texts ...string) []game.Tile {
	t.Helper()
	tiles := make([]game.Tile, 0, len(texts))
	for _, text := range texts {
		tile, ok := game.ParseTile(text)
		if !ok {
			t.Fatalf("ParseTile(%q) failed", text)
		}
		tiles = append(tiles, tile)
	}
	game.SortTiles(tiles)
	return tiles
}
