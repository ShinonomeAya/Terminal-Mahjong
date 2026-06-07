package game

import "strings"

var tileGlyphs = [TileTypeCount]string{
	"🀇", "🀈", "🀉", "🀊", "🀋", "🀌", "🀍", "🀎", "🀏",
	"🀙", "🀚", "🀛", "🀜", "🀝", "🀞", "🀟", "🀠", "🀡",
	"🀐", "🀑", "🀒", "🀓", "🀔", "🀕", "🀖", "🀗", "🀘",
	"🀀", "🀁", "🀂", "🀃", "🀄", "🀅", "🀆",
}

func TileGlyph(tile Tile) string {
	if tile < 0 || int(tile) >= TileTypeCount {
		return "?"
	}
	return tileGlyphs[int(tile)]
}

func TileLabel(tile Tile, unicode bool) string {
	if unicode {
		return TileGlyph(tile)
	}
	return tile.String()
}

func FormatTileLabels(tiles []Tile, unicode bool) string {
	if len(tiles) == 0 {
		return "-"
	}
	parts := make([]string, len(tiles))
	for i, tile := range tiles {
		parts[i] = TileLabel(tile, unicode)
	}
	return strings.Join(parts, " ")
}
