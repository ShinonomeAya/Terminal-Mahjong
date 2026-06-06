package game

import (
	"bufio"
	"fmt"
	"io"
)

func ChowOptions(player Player, discard Tile) [][]Tile {
	if !discard.IsSuit() {
		return nil
	}
	options := make([][]Tile, 0, 3)
	discardIndex := int(discard)
	for start := discardIndex - 2; start <= discardIndex; start++ {
		if start < 0 || start+2 >= 27 {
			continue
		}
		first := Tile(start)
		second := Tile(start + 1)
		third := Tile(start + 2)
		if first.Rank() > third.Rank() {
			continue
		}
		if first != discard && player.Count(first) == 0 {
			continue
		}
		if second != discard && player.Count(second) == 0 {
			continue
		}
		if third != discard && player.Count(third) == 0 {
			continue
		}
		option := []Tile{first, second, third}
		SortTiles(option)
		options = append(options, option)
	}
	return options
}

func (g *Game) resolveDiscardClaims(reader *bufio.Reader, out io.Writer, discarder int, discard Tile) bool {
	for offset := 1; offset < len(g.Players); offset++ {
		claimer := (discarder + offset) % len(g.Players)
		claimHand := append([]Tile(nil), g.Players[claimer].Hand...)
		claimHand = append(claimHand, discard)
		SortTiles(claimHand)
		if CanWin(claimHand) {
			if g.Players[claimer].Human {
				if askYes(reader, out, fmt.Sprintf("You can win on %s. Win? [y/N] ", discard)) {
					g.finish(claimer, fmt.Sprintf("discard-win on %s from %s", discard, g.Players[discarder].Name), WinDiscard)
					return true
				}
			} else {
				g.finish(claimer, fmt.Sprintf("discard-win on %s from %s", discard, g.Players[discarder].Name), WinDiscard)
				fmt.Fprintf(out, "%s wins on %s.\n", g.Players[claimer].Name, discard)
				return true
			}
		}
	}
	for offset := 1; offset < len(g.Players); offset++ {
		claimer := (discarder + offset) % len(g.Players)
		if g.Players[claimer].Count(discard) < 2 {
			continue
		}
		if g.Players[claimer].Human {
			if !askYes(reader, out, fmt.Sprintf("Pong %s? [y/N] ", discard)) {
				continue
			}
		} else if !shouldAIPong(g.Players[claimer], discard) {
			continue
		}
		g.claimPong(claimer, discard)
		fmt.Fprintf(out, "%s pongs %s.\n", g.Players[claimer].Name, discard)
		claimedDiscard, ok := g.takeDiscardTurn(reader, out, claimer)
		if !ok {
			return true
		}
		g.Current = claimer
		if g.resolveDiscardClaims(reader, out, claimer, claimedDiscard) {
			return true
		}
		g.Current = (claimer + 1) % len(g.Players)
		return true
	}
	nextPlayer := (discarder + 1) % len(g.Players)
	chowOptions := ChowOptions(g.Players[nextPlayer], discard)
	if len(chowOptions) == 0 {
		return false
	}
	chowIndex := -1
	if g.Players[nextPlayer].Human {
		for i, option := range chowOptions {
			handTiles := chowHandTiles(option, discard)
			if askYes(reader, out, fmt.Sprintf("Chow %s with %s? [y/N] ", discard, FormatTiles(handTiles))) {
				chowIndex = i
				break
			}
		}
	} else if index, ok := shouldAIChow(g.Players[nextPlayer], discard, chowOptions); ok {
		chowIndex = index
	}
	if chowIndex < 0 {
		return false
	}
	g.claimChow(nextPlayer, discard, chowOptions[chowIndex])
	fmt.Fprintf(out, "%s chows %s.\n", g.Players[nextPlayer].Name, discard)
	claimedDiscard, ok := g.takeDiscardTurn(reader, out, nextPlayer)
	if !ok {
		return true
	}
	g.Current = nextPlayer
	if g.resolveDiscardClaims(reader, out, nextPlayer, claimedDiscard) {
		return true
	}
	g.Current = (nextPlayer + 1) % len(g.Players)
	return true
}

func shouldAIChow(player Player, discard Tile, options [][]Tile) (int, bool) {
	if len(options) == 0 {
		return 0, false
	}
	counts := TileCounts(player.Hand)
	bestIndex := -1
	bestScore := 999
	for i, option := range options {
		score := 0
		for _, tile := range chowHandTiles(option, discard) {
			score += tileUsefulness(tile, counts)
		}
		if score < bestScore {
			bestScore = score
			bestIndex = i
		}
	}
	return bestIndex, bestIndex >= 0 && bestScore <= 12
}

func chowHandTiles(option []Tile, discard Tile) []Tile {
	handTiles := make([]Tile, 0, 2)
	removedDiscard := false
	for _, tile := range option {
		if tile == discard && !removedDiscard {
			removedDiscard = true
			continue
		}
		handTiles = append(handTiles, tile)
	}
	return handTiles
}

func (g *Game) claimChow(playerIndex int, discard Tile, option []Tile) {
	player := &g.Players[playerIndex]
	for _, tile := range chowHandTiles(option, discard) {
		player.RemoveTile(tile)
	}
	player.AddMeld(MeldChow, option)
	g.RecordEvent(EventChow, playerIndex, discard, FormatTiles(option))
}

func shouldAIPong(player Player, tile Tile) bool {
	if !tile.IsSuit() {
		return true
	}
	counts := TileCounts(player.Hand)
	return tileUsefulness(tile, counts) >= 6
}

func (g *Game) claimPong(playerIndex int, tile Tile) {
	player := &g.Players[playerIndex]
	player.RemoveTile(tile)
	player.RemoveTile(tile)
	player.AddMeld(MeldPong, []Tile{tile, tile, tile})
	g.RecordEvent(EventPong, playerIndex, tile, "")
}
