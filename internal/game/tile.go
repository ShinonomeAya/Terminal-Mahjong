package game

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type Tile int

const (
	BaseTileTypeCount = 34
	TileTypeCount     = BaseTileTypeCount
)

const (
	FlowerPlum Tile = BaseTileTypeCount + iota
	FlowerOrchid
	FlowerChrysanthemum
	FlowerBamboo
	FlowerSpring
	FlowerSummer
	FlowerAutumn
	FlowerWinter
	RedFiveMan
	RedFivePin
	RedFiveSou
)

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

func BuildMCRWall() []Tile {
	wall := BuildWall()
	for tile := FlowerPlum; tile <= FlowerWinter; tile++ {
		wall = append(wall, tile)
	}
	return wall
}

func SortTiles(tiles []Tile) {
	sort.Slice(tiles, func(i, j int) bool {
		left, right := tiles[i].Base(), tiles[j].Base()
		if left != right {
			return left < right
		}
		return tiles[i] < tiles[j]
	})
}

func TileCounts(tiles []Tile) [TileTypeCount]int {
	var counts [TileTypeCount]int
	for _, tile := range tiles {
		base := tile.Base()
		if base >= 0 && int(base) < TileTypeCount {
			counts[base]++
		}
	}
	return counts
}

func (t Tile) String() string {
	switch t {
	case RedFiveMan:
		return "0m"
	case RedFivePin:
		return "0p"
	case RedFiveSou:
		return "0s"
	}
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
	switch t {
	case FlowerPlum:
		return "P1"
	case FlowerOrchid:
		return "P2"
	case FlowerChrysanthemum:
		return "P3"
	case FlowerBamboo:
		return "P4"
	case FlowerSpring:
		return "S1"
	case FlowerSummer:
		return "S2"
	case FlowerAutumn:
		return "S3"
	case FlowerWinter:
		return "S4"
	}
	return "?"
}

func (t Tile) IsSuit() bool {
	base := t.Base()
	return base >= 0 && base < 27
}

func (t Tile) Rank() int {
	if !t.IsSuit() {
		return 0
	}
	return int(t.Base())%9 + 1
}

func (t Tile) IsFlower() bool {
	return t >= FlowerPlum && t <= FlowerWinter
}

func (t Tile) IsRed() bool {
	return t >= RedFiveMan && t <= RedFiveSou
}

func (t Tile) Base() Tile {
	switch t {
	case RedFiveMan:
		return 4
	case RedFivePin:
		return 13
	case RedFiveSou:
		return 22
	default:
		return t
	}
}

func ParseTile(text string) (Tile, bool) {
	token := strings.TrimSpace(strings.ToLower(text))
	if len(token) < 1 {
		return 0, false
	}
	if len(token) >= 2 {
		rank, err := strconv.Atoi(token[:1])
		if err == nil && rank == 0 {
			switch token[1:] {
			case "m":
				return RedFiveMan, true
			case "p":
				return RedFivePin, true
			case "s":
				return RedFiveSou, true
			}
		}
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
	case "P1":
		return FlowerPlum, true
	case "P2":
		return FlowerOrchid, true
	case "P3":
		return FlowerChrysanthemum, true
	case "P4":
		return FlowerBamboo, true
	case "S1":
		return FlowerSpring, true
	case "S2":
		return FlowerSummer, true
	case "S3":
		return FlowerAutumn, true
	case "S4":
		return FlowerWinter, true
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
