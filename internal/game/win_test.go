package game

import "testing"

func mustTiles(t *testing.T, texts ...string) []Tile {
	t.Helper()
	tiles := make([]Tile, 0, len(texts))
	for _, text := range texts {
		tile, ok := ParseTile(text)
		if !ok {
			t.Fatalf("bad tile in test: %s", text)
		}
		tiles = append(tiles, tile)
	}
	return tiles
}

func TestCanWinWithSequencesTripletAndPair(t *testing.T) {
	hand := mustTiles(t,
		"1m", "2m", "3m",
		"4m", "5m", "6m",
		"2p", "3p", "4p",
		"7s", "7s", "7s",
		"E", "E",
	)
	if !CanWin(hand) {
		t.Fatal("expected hand to win")
	}
}

func TestCanWinWithAllTriplets(t *testing.T) {
	hand := mustTiles(t,
		"1m", "1m", "1m",
		"9m", "9m", "9m",
		"2p", "2p", "2p",
		"F", "F", "F",
		"B", "B",
	)
	if !CanWin(hand) {
		t.Fatal("expected all-triplet hand to win")
	}
}

func TestCanWinRejectsIncompleteShape(t *testing.T) {
	hand := mustTiles(t,
		"1m", "2m", "3m",
		"4m", "5m", "6m",
		"2p", "3p", "4p",
		"7s", "8s", "9s",
		"E", "F",
	)
	if CanWin(hand) {
		t.Fatal("expected hand not to win")
	}
}
