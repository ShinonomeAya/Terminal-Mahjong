package game

func CanWin(tiles []Tile) bool {
	if len(tiles)%3 != 2 {
		return false
	}
	counts := TileCounts(tiles)
	for tile := 0; tile < TileTypeCount; tile++ {
		if counts[tile] >= 2 {
			counts[tile] -= 2
			if canFormMelds(counts) {
				return true
			}
			counts[tile] += 2
		}
	}
	return false
}

func canFormMelds(counts [TileTypeCount]int) bool {
	first := -1
	for i, count := range counts {
		if count > 0 {
			first = i
			break
		}
	}
	if first == -1 {
		return true
	}
	if counts[first] >= 3 {
		counts[first] -= 3
		if canFormMelds(counts) {
			return true
		}
		counts[first] += 3
	}
	tile := Tile(first)
	if tile.IsSuit() && tile.Rank() <= 7 && counts[first+1] > 0 && counts[first+2] > 0 {
		counts[first]--
		counts[first+1]--
		counts[first+2]--
		if canFormMelds(counts) {
			return true
		}
	}
	return false
}
