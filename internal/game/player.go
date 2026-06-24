package game

import "fmt"

type MeldKind string

const (
	MeldChow MeldKind = "chow"
	MeldPong MeldKind = "pong"
	MeldKong MeldKind = "kong"
)

type Meld struct {
	Kind  MeldKind
	Tiles []Tile
}

type Player struct {
	Name     string
	Human    bool
	Hand     []Tile
	Flowers  []Tile
	Melds    []Meld
	Discards []Tile
}

func NewPlayers() []Player {
	return []Player{
		{Name: "You", Human: true},
		{Name: "AI-1"},
		{Name: "AI-2"},
		{Name: "AI-3"},
	}
}

func (p *Player) AddTile(tile Tile) {
	p.Hand = append(p.Hand, tile)
	SortTiles(p.Hand)
}

func (p *Player) RemoveTile(tile Tile) bool {
	for i, handTile := range p.Hand {
		if handTile == tile {
			p.Hand = append(p.Hand[:i], p.Hand[i+1:]...)
			return true
		}
	}
	for i, handTile := range p.Hand {
		if handTile.Base() == tile.Base() {
			p.Hand = append(p.Hand[:i], p.Hand[i+1:]...)
			return true
		}
	}
	return false
}

func (p *Player) RemoveAt(index int) (Tile, error) {
	if index < 0 || index >= len(p.Hand) {
		return 0, fmt.Errorf("hand index %d is out of range", index+1)
	}
	tile := p.Hand[index]
	p.Hand = append(p.Hand[:index], p.Hand[index+1:]...)
	return tile, nil
}

func (p *Player) Count(tile Tile) int {
	count := 0
	for _, handTile := range p.Hand {
		if handTile.Base() == tile.Base() {
			count++
		}
	}
	return count
}

func (p *Player) AddMeld(kind MeldKind, tiles []Tile) {
	p.Melds = append(p.Melds, Meld{Kind: kind, Tiles: append([]Tile(nil), tiles...)})
}

func (p *Player) MeldSummary() string {
	if len(p.Melds) == 0 {
		return "-"
	}
	parts := make([]string, len(p.Melds))
	for i, meld := range p.Melds {
		parts[i] = fmt.Sprintf("%s(%s)", meld.Kind, FormatTiles(meld.Tiles))
	}
	return join(parts, " | ")
}

func join(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for _, part := range parts[1:] {
		result += sep + part
	}
	return result
}
