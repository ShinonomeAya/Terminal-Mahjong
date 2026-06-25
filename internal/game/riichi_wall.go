package game

import "fmt"

const riichiDeadWallSize = 14

func BuildRiichiWall(redFives int) []Tile {
	wall := BuildWall()
	if redFives != 3 {
		return wall
	}
	replacements := map[Tile]Tile{4: RedFiveMan, 13: RedFivePin, 22: RedFiveSou}
	for index, tile := range wall {
		if red, ok := replacements[tile]; ok {
			wall[index] = red
			delete(replacements, tile)
		}
	}
	return wall
}

func (rules *RiichiRuleSet) BuildWall() []Tile {
	return BuildRiichiWall(rules.config.RedFives)
}

func (rules *RiichiRuleSet) Deal(round *Game) error {
	if len(round.Wall) < riichiDeadWallSize+53 {
		return fmt.Errorf("riichi wall has %d tiles, need at least %d", len(round.Wall), riichiDeadWallSize+53)
	}
	deadStart := len(round.Wall) - riichiDeadWallSize
	dead := append([]Tile(nil), round.Wall[deadStart:]...)
	round.Wall = round.Wall[:deadStart]
	round.Riichi = &RiichiRoundState{
		DeadWall:       dead,
		DoraIndicators: []Tile{dead[4]},
		UraIndicators:  []Tile{dead[5]},
	}
	for dealt := 0; dealt < 13; dealt++ {
		for player := range round.Players {
			round.Players[player].AddTile(round.Wall[0])
			round.Wall = round.Wall[1:]
		}
	}
	dealer := round.Dealer
	if dealer < 0 || dealer >= len(round.Players) {
		dealer = 0
	}
	round.Players[dealer].AddTile(round.Wall[0])
	round.Wall = round.Wall[1:]
	round.Current = dealer
	round.Phase = PhaseAwaitingDiscard
	return nil
}

func (rules *RiichiRuleSet) Draw(round *Game, player int, source DrawSource) (Tile, bool) {
	if source == DrawReplacement {
		return rules.drawRinshan(round, player)
	}
	if round.Riichi != nil && player >= 0 && player < len(round.Riichi.TemporaryFuriten) {
		round.Riichi.TemporaryFuriten[player] = false
	}
	if len(round.Wall) == 0 {
		finishRiichiWallExhaustion(round, player)
		return -1, false
	}
	tile := round.Wall[0]
	round.Wall = round.Wall[1:]
	round.Players[player].AddTile(tile)
	round.RecordEvent(EventDraw, player, tile, "")
	return tile, true
}

func (rules *RiichiRuleSet) drawRinshan(round *Game, player int) (Tile, bool) {
	if round.Riichi == nil || round.Riichi.RinshanDraws >= 4 || len(round.Riichi.DeadWall) != riichiDeadWallSize || len(round.Wall) == 0 {
		return -1, false
	}
	index := round.Riichi.RinshanDraws
	tile := round.Riichi.DeadWall[index]
	replacementIndex := len(round.Wall) - 1
	round.Riichi.DeadWall[index] = round.Wall[replacementIndex]
	round.Wall = round.Wall[:replacementIndex]
	round.Riichi.RinshanDraws++
	round.Players[player].AddTile(tile)
	round.RecordEvent(EventReplacementDraw, player, tile, "rinshan")
	return tile, true
}

func revealRiichiKanDora(round *Game) bool {
	if round == nil || round.Riichi == nil || len(round.Riichi.DeadWall) != riichiDeadWallSize || round.Riichi.KanCount >= 4 {
		return false
	}
	round.Riichi.KanCount++
	indicatorIndex := 4 + round.Riichi.KanCount*2
	round.Riichi.DoraIndicators = append(round.Riichi.DoraIndicators, round.Riichi.DeadWall[indicatorIndex])
	round.Riichi.UraIndicators = append(round.Riichi.UraIndicators, round.Riichi.DeadWall[indicatorIndex+1])
	return true
}

func RiichiDoraTile(indicator Tile) Tile {
	base := indicator.Base()
	if base.IsSuit() {
		if base.Rank() == 9 {
			return base - 8
		}
		return base + 1
	}
	if base >= 27 && base <= 30 {
		if base == 30 {
			return 27
		}
		return base + 1
	}
	switch base {
	case 31:
		return 33
	case 33:
		return 32
	case 32:
		return 31
	default:
		return -1
	}
}

func finishRiichiWallExhaustion(round *Game, player int) {
	round.Over = true
	round.Phase = PhaseRoundOver
	round.Reason = "draw: live wall exhausted"
	round.RecordEvent(EventWallExhausted, player, -1, round.Reason)
}
