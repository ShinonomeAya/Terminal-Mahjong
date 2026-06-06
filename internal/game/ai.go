package game

func ChooseAIDiscard(hand []Tile) int {
	if len(hand) == 0 {
		return -1
	}
	counts := TileCounts(hand)
	bestIndex := 0
	bestScore := 999
	for i, tile := range hand {
		score := tileUsefulness(tile, counts)
		if score < bestScore {
			bestScore = score
			bestIndex = i
		}
	}
	return bestIndex
}

func tileUsefulness(tile Tile, counts [TileTypeCount]int) int {
	score := counts[tile] * 5
	if !tile.IsSuit() {
		return score
	}
	index := int(tile)
	if tile.Rank() > 1 {
		score += counts[index-1] * 2
	}
	if tile.Rank() > 2 {
		score += counts[index-2]
	}
	if tile.Rank() < 9 {
		score += counts[index+1] * 2
	}
	if tile.Rank() < 8 {
		score += counts[index+2]
	}
	return score
}
