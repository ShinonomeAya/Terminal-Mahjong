package game

import "testing"

func TestBuildWallHasFourCopiesOfEachTile(t *testing.T) {
	wall := BuildWall()
	if len(wall) != 136 {
		t.Fatalf("wall length = %d, want 136", len(wall))
	}
	counts := TileCounts(wall)
	for tile, count := range counts {
		if count != 4 {
			t.Fatalf("tile %d count = %d, want 4", tile, count)
		}
	}
}

func TestParseTile(t *testing.T) {
	cases := map[string]Tile{
		"1m": 0,
		"9m": 8,
		"1p": 9,
		"9p": 17,
		"1s": 18,
		"9s": 26,
		"E":  27,
		"B":  33,
	}
	for text, want := range cases {
		got, ok := ParseTile(text)
		if !ok || got != want {
			t.Fatalf("ParseTile(%q) = %v, %v; want %v, true", text, got, ok, want)
		}
	}
}

func TestRedFivesNormalizeToBaseTiles(t *testing.T) {
	tests := []struct {
		text string
		red  Tile
		base Tile
	}{
		{text: "0m", red: RedFiveMan, base: 4},
		{text: "0p", red: RedFivePin, base: 13},
		{text: "0s", red: RedFiveSou, base: 22},
	}
	for _, test := range tests {
		parsed, ok := ParseTile(test.text)
		if !ok || parsed != test.red || parsed.String() != test.text || parsed.Base() != test.base || !parsed.IsRed() || !parsed.IsSuit() || parsed.Rank() != 5 {
			t.Fatalf("red tile %s parsed=%v ok=%t base=%v rank=%d", test.text, parsed, ok, parsed.Base(), parsed.Rank())
		}
	}
	counts := TileCounts([]Tile{RedFiveMan, 4, RedFivePin, 13, RedFiveSou, 22})
	if counts[4] != 2 || counts[13] != 2 || counts[22] != 2 {
		t.Fatalf("normalized counts = %v/%v/%v", counts[4], counts[13], counts[22])
	}
}

func TestRedFivesWorkInSharedHandAlgorithms(t *testing.T) {
	hand := mustTiles(t, "1m", "2m", "3m", "4m", "0m", "6m", "2p", "3p", "4p", "7s", "8s", "9s", "E", "E")
	if !CanWin(hand) {
		t.Fatalf("red-five hand should win: %v", hand)
	}
	SortTiles(hand)
	if got := FormatTiles(hand); got != "1m 2m 3m 4m 0m 6m 2p 3p 4p 7s 8s 9s E E" {
		t.Fatalf("sorted red-five hand = %s", got)
	}
}

func TestBuildMCRWallContains144Tiles(t *testing.T) {
	wall := BuildMCRWall()
	if len(wall) != 144 {
		t.Fatalf("wall length = %d, want 144", len(wall))
	}
	for tile := Tile(0); tile < BaseTileTypeCount; tile++ {
		if got := countTileInWall(wall, tile); got != 4 {
			t.Fatalf("tile %s count = %d, want 4", tile, got)
		}
	}
	for tile := FlowerPlum; tile <= FlowerWinter; tile++ {
		if got := countTileInWall(wall, tile); got != 1 {
			t.Fatalf("flower %s count = %d, want 1", tile, got)
		}
		if !tile.IsFlower() {
			t.Fatalf("%s should be a flower", tile)
		}
		if parsed, ok := ParseTile(tile.String()); !ok || parsed != tile {
			t.Fatalf("flower %s did not round trip: %v/%v", tile, parsed, ok)
		}
	}
}

func countTileInWall(wall []Tile, wanted Tile) int {
	count := 0
	for _, tile := range wall {
		if tile == wanted {
			count++
		}
	}
	return count
}
