package game

import "fmt"

const maxStandardShanten = 6

func ShantenStandard(tiles []Tile) int {
	if CanWin(tiles) {
		return -1
	}
	if len(tiles)%3 == 1 && len(WinningTiles(tiles)) > 0 {
		return 0
	}
	counts := TileCounts(tiles)
	best := maxStandardShanten
	searchShanten(counts, 0, 0, 0, 0, &best)
	if best < 0 {
		return 0
	}
	if best > maxStandardShanten {
		return maxStandardShanten
	}
	return best
}

func searchShanten(counts [TileTypeCount]int, index int, melds int, pairs int, taatsu int, best *int) {
	for index < TileTypeCount && counts[index] == 0 {
		index++
	}
	if index >= TileTypeCount {
		value := standardShantenValue(melds, pairs, taatsu)
		if value < *best {
			*best = value
		}
		return
	}
	if melds > 4 {
		melds = 4
	}
	if pairs > 1 {
		pairs = 1
	}
	if taatsu > 4-melds {
		taatsu = 4 - melds
	}
	currentLowerBound := standardShantenValue(melds, pairs, taatsu)
	if currentLowerBound >= *best && counts[index] <= 0 {
		return
	}
	if counts[index] >= 3 {
		counts[index] -= 3
		searchShanten(counts, index, melds+1, pairs, taatsu, best)
		counts[index] += 3
	}
	tile := Tile(index)
	if tile.IsSuit() && tile.Rank() <= 7 && counts[index+1] > 0 && counts[index+2] > 0 {
		counts[index]--
		counts[index+1]--
		counts[index+2]--
		searchShanten(counts, index, melds+1, pairs, taatsu, best)
		counts[index]++
		counts[index+1]++
		counts[index+2]++
	}
	if counts[index] >= 2 {
		counts[index] -= 2
		nextPairs := pairs
		nextTaatsu := taatsu
		if pairs == 0 {
			nextPairs = 1
		} else {
			nextTaatsu++
		}
		searchShanten(counts, index, melds, nextPairs, nextTaatsu, best)
		counts[index] += 2
	}
	if tile.IsSuit() && tile.Rank() <= 8 && counts[index+1] > 0 {
		counts[index]--
		counts[index+1]--
		searchShanten(counts, index, melds, pairs, taatsu+1, best)
		counts[index]++
		counts[index+1]++
	}
	if tile.IsSuit() && tile.Rank() <= 7 && counts[index+2] > 0 {
		counts[index]--
		counts[index+2]--
		searchShanten(counts, index, melds, pairs, taatsu+1, best)
		counts[index]++
		counts[index+2]++
	}
	counts[index]--
	searchShanten(counts, index, melds, pairs, taatsu, best)
	counts[index]++
}

func standardShantenValue(melds int, pairs int, taatsu int) int {
	if melds > 4 {
		melds = 4
	}
	if taatsu > 4-melds {
		taatsu = 4 - melds
	}
	needPair := 1
	if pairs > 0 {
		needPair = 0
	}
	return 8 - melds*2 - taatsu - pairs - needPair
}

func WinningTiles(tiles []Tile) []Tile {
	counts := TileCounts(tiles)
	waits := make([]Tile, 0)
	for tile := 0; tile < TileTypeCount; tile++ {
		if counts[tile] >= 4 {
			continue
		}
		candidate := append([]Tile(nil), tiles...)
		candidate = append(candidate, Tile(tile))
		SortTiles(candidate)
		if CanWin(candidate) {
			waits = append(waits, Tile(tile))
		}
	}
	return waits
}

func HandTips(tiles []Tile) string {
	if CanWin(tiles) {
		return "winning hand"
	}
	shanten := ShantenStandard(tiles)
	if shanten == 0 {
		waits := WinningTiles(tiles)
		if len(waits) > 0 {
			return fmt.Sprintf("tenpai: waits %s", FormatTiles(waits))
		}
		return "tenpai"
	}
	return fmt.Sprintf("shanten: %d", shanten)
}

type TileImprovement struct {
	Discard   Tile   `json:"discard"`
	Effective []Tile `json:"effective"`
}

func EffectiveTiles(hand []Tile) []Tile {
	current := ShantenStandard(hand)
	counts := TileCounts(hand)
	effective := make([]Tile, 0)
	for tile := 0; tile < TileTypeCount; tile++ {
		if counts[tile] >= 4 {
			continue
		}
		candidate := append([]Tile(nil), hand...)
		candidate = append(candidate, Tile(tile))
		if ShantenStandard(candidate) < current {
			effective = append(effective, Tile(tile))
		}
	}
	return effective
}

func ImprovementTiles(hand []Tile) []TileImprovement {
	if len(hand) == 0 {
		return nil
	}
	improvements := make([]TileImprovement, 0, len(hand))
	seen := make(map[Tile]bool)
	for index, discard := range hand {
		if seen[discard] {
			continue
		}
		seen[discard] = true
		candidate := append([]Tile(nil), hand[:index]...)
		candidate = append(candidate, hand[index+1:]...)
		effective := EffectiveTiles(candidate)
		if len(effective) == 0 {
			continue
		}
		improvements = append(improvements, TileImprovement{
			Discard:   discard,
			Effective: append([]Tile(nil), effective...),
		})
	}
	return improvements
}

func BestDiscardIndex(hand []Tile) int {
	if len(hand) == 0 {
		return -1
	}
	counts := TileCounts(hand)
	bestIndex := -1
	bestShanten := 99
	bestUsefulness := 999
	for i, tile := range hand {
		candidate := append([]Tile(nil), hand[:i]...)
		candidate = append(candidate, hand[i+1:]...)
		shanten := ShantenStandard(candidate)
		usefulness := tileUsefulness(tile, counts)
		if shanten < bestShanten || (shanten == bestShanten && usefulness < bestUsefulness) {
			bestIndex = i
			bestShanten = shanten
			bestUsefulness = usefulness
		}
	}
	if bestIndex < 0 {
		return chooseAIDiscardByUsefulness(hand)
	}
	return bestIndex
}
