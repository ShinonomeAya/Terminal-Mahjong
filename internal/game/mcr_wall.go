package game

import "fmt"

func (rules *MCRRuleSet) BuildWall() []Tile {
	return BuildMCRWall()
}

func (rules *MCRRuleSet) Deal(round *Game) error {
	const tilesPerPlayer = 13
	for dealt := 0; dealt < tilesPerPlayer; dealt++ {
		for player := range round.Players {
			if _, ok := rules.draw(round, player, DrawNormal, false); !ok {
				return fmt.Errorf("MCR wall exhausted during initial deal")
			}
		}
	}
	return nil
}

func (rules *MCRRuleSet) Draw(round *Game, player int, source DrawSource) (Tile, bool) {
	return rules.draw(round, player, source, true)
}

func (rules *MCRRuleSet) draw(round *Game, player int, source DrawSource, recordNormal bool) (Tile, bool) {
	tile, ok := takeWallTile(round, source)
	if !ok {
		finishMCRWallExhaustion(round, player)
		return -1, false
	}
	for tile.IsFlower() {
		round.Players[player].Flowers = append(round.Players[player].Flowers, tile)
		round.RecordEvent(EventFlower, player, tile, "")
		tile, ok = takeWallTile(round, DrawReplacement)
		if !ok {
			finishMCRWallExhaustion(round, player)
			return -1, false
		}
		source = DrawReplacement
	}
	round.Players[player].AddTile(tile)
	if source == DrawReplacement {
		round.RecordEvent(EventReplacementDraw, player, tile, "")
	} else if recordNormal {
		round.RecordEvent(EventDraw, player, tile, "")
	}
	return tile, true
}

func takeWallTile(round *Game, source DrawSource) (Tile, bool) {
	if len(round.Wall) == 0 {
		return -1, false
	}
	if source == DrawReplacement {
		index := len(round.Wall) - 1
		tile := round.Wall[index]
		round.Wall = round.Wall[:index]
		return tile, true
	}
	tile := round.Wall[0]
	round.Wall = round.Wall[1:]
	return tile, true
}

func finishMCRWallExhaustion(round *Game, player int) {
	round.Over = true
	round.Phase = PhaseRoundOver
	round.Reason = "draw: replacement wall exhausted"
	round.RecordEvent(EventWallExhausted, player, -1, round.Reason)
}
