package game

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type Tile int

const TileTypeCount = 34

var honorNames = []string{"E", "S", "W", "N", "Z", "F", "B"}

func BuildWall() []Tile {
	wall := make([]Tile, 0, TileTypeCount*4)
	for tile := 0; tile < TileTypeCount; tile++ {
		for copy := 0; copy < 4; copy++ {
			wall = append(wall, Tile(tile))
		}
	}
	return wall
}

func SortTiles(tiles []Tile) {
	sort.Slice(tiles, func(i, j int) bool {
		return tiles[i] < tiles[j]
	})
}

func TileCounts(tiles []Tile) [TileTypeCount]int {
	var counts [TileTypeCount]int
	for _, tile := range tiles {
		if tile >= 0 && int(tile) < TileTypeCount {
			counts[tile]++
		}
	}
	return counts
}

func (t Tile) String() string {
	if t >= 0 && t < 9 {
		return fmt.Sprintf("%dm", int(t)+1)
	}
	if t >= 9 && t < 18 {
		return fmt.Sprintf("%dp", int(t)-8)
	}
	if t >= 18 && t < 27 {
		return fmt.Sprintf("%ds", int(t)-17)
	}
	if t >= 27 && t < 34 {
		return honorNames[int(t)-27]
	}
	return "?"
}

func (t Tile) IsSuit() bool {
	return t >= 0 && t < 27
}

func (t Tile) Rank() int {
	if !t.IsSuit() {
		return 0
	}
	return int(t)%9 + 1
}

func ParseTile(text string) (Tile, bool) {
	token := strings.TrimSpace(strings.ToLower(text))
	if len(token) < 1 {
		return 0, false
	}
	if len(token) >= 2 {
		rank, err := strconv.Atoi(token[:1])
		if err == nil && rank >= 1 && rank <= 9 {
			switch token[1:] {
			case "m":
				return Tile(rank - 1), true
			case "p":
				return Tile(8 + rank), true
			case "s":
				return Tile(17 + rank), true
			}
		}
	}
	switch strings.ToUpper(token) {
	case "E":
		return Tile(27), true
	case "S":
		return Tile(28), true
	case "W":
		return Tile(29), true
	case "N":
		return Tile(30), true
	case "Z":
		return Tile(31), true
	case "F":
		return Tile(32), true
	case "B":
		return Tile(33), true
	default:
		return 0, false
	}
}

func FormatTiles(tiles []Tile) string {
	if len(tiles) == 0 {
		return "-"
	}
	parts := make([]string, len(tiles))
	for i, tile := range tiles {
		parts[i] = tile.String()
	}
	return strings.Join(parts, " ")
}
