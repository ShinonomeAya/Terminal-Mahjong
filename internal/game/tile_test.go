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
