package game

import "testing"

func TestTileGlyphMapsSuits(t *testing.T) {
	cases := map[string]string{
		"1m": "🀇",
		"9m": "🀏",
		"1s": "🀐",
		"9s": "🀘",
		"1p": "🀙",
		"9p": "🀡",
	}
	for text, want := range cases {
		tile, ok := ParseTile(text)
		if !ok {
			t.Fatalf("ParseTile(%q) failed", text)
		}
		if got := TileGlyph(tile); got != want {
			t.Fatalf("TileGlyph(%q) = %q, want %q", text, got, want)
		}
	}
}

func TestTileGlyphMapsHonors(t *testing.T) {
	cases := map[string]string{
		"E": "🀀",
		"S": "🀁",
		"W": "🀂",
		"N": "🀃",
		"Z": "🀄",
		"F": "🀅",
		"B": "🀆",
	}
	for text, want := range cases {
		tile, ok := ParseTile(text)
		if !ok {
			t.Fatalf("ParseTile(%q) failed", text)
		}
		if got := TileGlyph(tile); got != want {
			t.Fatalf("TileGlyph(%q) = %q, want %q", text, got, want)
		}
	}
}

func TestTileLabelFallbackUsesExistingNotation(t *testing.T) {
	tile, ok := ParseTile("5p")
	if !ok {
		t.Fatal("ParseTile failed")
	}
	if got := TileLabel(tile, false); got != "5p" {
		t.Fatalf("TileLabel fallback = %q, want 5p", got)
	}
}

func TestFormatTileLabelsSupportsUnicodeAndFallback(t *testing.T) {
	tiles := mustTiles(t, "1m", "2m", "E")
	if got := FormatTileLabels(tiles, true); got != "🀇 🀈 🀀" {
		t.Fatalf("unicode labels = %q", got)
	}
	if got := FormatTileLabels(tiles, false); got != "1m 2m E" {
		t.Fatalf("fallback labels = %q", got)
	}
}
