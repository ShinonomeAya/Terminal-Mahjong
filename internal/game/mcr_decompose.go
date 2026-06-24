package game

import (
	"fmt"
	"sort"
	"strings"
)

type MCRShapeKind string

const (
	MCRShapeStandard             MCRShapeKind = "standard"
	MCRShapeSevenPairs           MCRShapeKind = "seven_pairs"
	MCRShapeSevenShiftedPairs    MCRShapeKind = "seven_shifted_pairs"
	MCRShapeThirteenOrphans      MCRShapeKind = "thirteen_orphans"
	MCRShapeGreaterHonorsKnitted MCRShapeKind = "greater_honors_knitted"
	MCRShapeLesserHonorsKnitted  MCRShapeKind = "lesser_honors_knitted"
	MCRShapeKnittedStraight      MCRShapeKind = "knitted_straight"
	MCRShapeNineGates            MCRShapeKind = "nine_gates"
)

type MCRGroupKind string

const (
	MCRGroupPair    MCRGroupKind = "pair"
	MCRGroupChow    MCRGroupKind = "chow"
	MCRGroupPung    MCRGroupKind = "pung"
	MCRGroupKong    MCRGroupKind = "kong"
	MCRGroupKnitted MCRGroupKind = "knitted"
)

type MCRGroup struct {
	Kind  MCRGroupKind `json:"kind"`
	Tiles []Tile       `json:"tiles"`
	Open  bool         `json:"open"`
}

type MCRDecomposition struct {
	Kind   MCRShapeKind `json:"kind"`
	Groups []MCRGroup   `json:"groups"`
	Tiles  []Tile       `json:"tiles"`
}

func MCRDecompose(hand []Tile, declared []Meld, winningTile Tile) []MCRDecomposition {
	tiles := append([]Tile(nil), hand...)
	tiles = append(tiles, winningTile)
	SortTiles(tiles)
	if containsFlower(tiles) {
		return nil
	}

	result := standardMCRDecompositions(tiles, declared)
	if len(declared) == 0 {
		result = append(result, specialMCRDecompositions(tiles)...)
	}
	result = append(result, knittedStraightDecompositions(tiles, declared)...)
	return uniqueMCRDecompositions(result)
}

func standardMCRDecompositions(tiles []Tile, declared []Meld) []MCRDecomposition {
	neededMelds := 4 - len(declared)
	if neededMelds < 0 || len(tiles) != neededMelds*3+2 {
		return nil
	}
	counts := TileCounts(tiles)
	declaredGroups := mcrDeclaredGroups(declared)
	var result []MCRDecomposition
	for pair := 0; pair < TileTypeCount; pair++ {
		if counts[pair] < 2 {
			continue
		}
		remaining := counts
		remaining[pair] -= 2
		for _, groups := range mcrMeldGroupings(remaining, neededMelds) {
			allGroups := append(copyMCRGroups(declaredGroups), MCRGroup{Kind: MCRGroupPair, Tiles: []Tile{Tile(pair), Tile(pair)}})
			allGroups = append(allGroups, groups...)
			result = append(result, MCRDecomposition{Kind: MCRShapeStandard, Groups: allGroups, Tiles: append([]Tile(nil), tiles...)})
		}
	}
	return result
}

func mcrMeldGroupings(counts [TileTypeCount]int, remaining int) [][]MCRGroup {
	first := -1
	for tile, count := range counts {
		if count > 0 {
			first = tile
			break
		}
	}
	if first == -1 {
		if remaining == 0 {
			return [][]MCRGroup{{}}
		}
		return nil
	}
	if remaining == 0 {
		return nil
	}

	var result [][]MCRGroup
	if counts[first] >= 3 {
		next := counts
		next[first] -= 3
		group := MCRGroup{Kind: MCRGroupPung, Tiles: []Tile{Tile(first), Tile(first), Tile(first)}}
		for _, rest := range mcrMeldGroupings(next, remaining-1) {
			result = append(result, append([]MCRGroup{group}, rest...))
		}
	}
	tile := Tile(first)
	if tile.IsSuit() && tile.Rank() <= 7 && counts[first+1] > 0 && counts[first+2] > 0 {
		next := counts
		next[first]--
		next[first+1]--
		next[first+2]--
		group := MCRGroup{Kind: MCRGroupChow, Tiles: []Tile{tile, Tile(first + 1), Tile(first + 2)}}
		for _, rest := range mcrMeldGroupings(next, remaining-1) {
			result = append(result, append([]MCRGroup{group}, rest...))
		}
	}
	return result
}

func specialMCRDecompositions(tiles []Tile) []MCRDecomposition {
	var result []MCRDecomposition
	if groups, ok := mcrSevenPairGroups(tiles); ok {
		result = append(result, MCRDecomposition{Kind: MCRShapeSevenPairs, Groups: groups, Tiles: append([]Tile(nil), tiles...)})
	}
	if mcrIsSevenShiftedPairs(tiles) {
		result = append(result, MCRDecomposition{Kind: MCRShapeSevenShiftedPairs, Groups: pairGroupsForTiles(tiles), Tiles: append([]Tile(nil), tiles...)})
	}
	if mcrIsThirteenOrphans(tiles) {
		result = append(result, MCRDecomposition{Kind: MCRShapeThirteenOrphans, Tiles: append([]Tile(nil), tiles...)})
	}
	for _, pattern := range mcrKnittedPatterns() {
		if kind, ok := mcrHonorsKnittedKind(tiles, pattern); ok {
			result = append(result, MCRDecomposition{Kind: kind, Groups: []MCRGroup{{Kind: MCRGroupKnitted, Tiles: append([]Tile(nil), tiles...)}}, Tiles: append([]Tile(nil), tiles...)})
		}
	}
	if mcrIsNineGates(tiles) {
		result = append(result, MCRDecomposition{Kind: MCRShapeNineGates, Tiles: append([]Tile(nil), tiles...)})
	}
	return result
}

func knittedStraightDecompositions(tiles []Tile, declared []Meld) []MCRDecomposition {
	if len(declared) > 1 {
		return nil
	}
	counts := TileCounts(tiles)
	var result []MCRDecomposition
	for _, pattern := range mcrKnittedPatterns() {
		remaining := counts
		valid := true
		for _, tile := range pattern {
			if remaining[tile] == 0 {
				valid = false
				break
			}
			remaining[tile]--
		}
		if !valid {
			continue
		}
		neededMelds := 1 - len(declared)
		for pair := 0; pair < TileTypeCount; pair++ {
			if remaining[pair] < 2 {
				continue
			}
			afterPair := remaining
			afterPair[pair] -= 2
			for _, groups := range mcrMeldGroupings(afterPair, neededMelds) {
				all := mcrDeclaredGroups(declared)
				all = append(all, MCRGroup{Kind: MCRGroupKnitted, Tiles: append([]Tile(nil), pattern...)})
				all = append(all, MCRGroup{Kind: MCRGroupPair, Tiles: []Tile{Tile(pair), Tile(pair)}})
				all = append(all, groups...)
				result = append(result, MCRDecomposition{Kind: MCRShapeKnittedStraight, Groups: all, Tiles: append([]Tile(nil), tiles...)})
			}
		}
	}
	return result
}

func mcrSevenPairGroups(tiles []Tile) ([]MCRGroup, bool) {
	if len(tiles) != 14 {
		return nil, false
	}
	counts := TileCounts(tiles)
	var groups []MCRGroup
	for tile, count := range counts {
		if count%2 != 0 {
			return nil, false
		}
		for pairs := 0; pairs < count/2; pairs++ {
			groups = append(groups, MCRGroup{Kind: MCRGroupPair, Tiles: []Tile{Tile(tile), Tile(tile)}})
		}
	}
	return groups, len(groups) == 7
}

func pairGroupsForTiles(tiles []Tile) []MCRGroup {
	groups, _ := mcrSevenPairGroups(tiles)
	return groups
}

func mcrIsSevenShiftedPairs(tiles []Tile) bool {
	if len(tiles) != 14 || !tiles[0].IsSuit() {
		return false
	}
	suit := int(tiles[0]) / 9
	start := tiles[0].Rank()
	if start > 3 {
		return false
	}
	counts := TileCounts(tiles)
	for rank := 1; rank <= 9; rank++ {
		want := 0
		if rank >= start && rank < start+7 {
			want = 2
		}
		if counts[suit*9+rank-1] != want {
			return false
		}
	}
	for tile := 0; tile < TileTypeCount; tile++ {
		if tile/9 != suit && counts[tile] != 0 {
			return false
		}
	}
	return true
}

func mcrIsThirteenOrphans(tiles []Tile) bool {
	if len(tiles) != 14 {
		return false
	}
	required := []Tile{0, 8, 9, 17, 18, 26, 27, 28, 29, 30, 31, 32, 33}
	counts := TileCounts(tiles)
	pairs := 0
	for _, tile := range required {
		if counts[tile] == 0 {
			return false
		}
		if counts[tile] == 2 {
			pairs++
		} else if counts[tile] != 1 {
			return false
		}
	}
	return pairs == 1
}

func mcrHonorsKnittedKind(tiles []Tile, pattern []Tile) (MCRShapeKind, bool) {
	if len(tiles) != 14 {
		return "", false
	}
	allowed := make(map[Tile]bool)
	for _, tile := range pattern {
		allowed[tile] = true
	}
	for tile := Tile(27); tile <= Tile(33); tile++ {
		allowed[tile] = true
	}
	honors := 0
	seen := make(map[Tile]bool)
	for _, tile := range tiles {
		if seen[tile] || !allowed[tile] {
			return "", false
		}
		seen[tile] = true
		if tile >= 27 {
			honors++
		}
	}
	if honors == 7 {
		return MCRShapeGreaterHonorsKnitted, true
	}
	if honors >= 5 {
		return MCRShapeLesserHonorsKnitted, true
	}
	return "", false
}

func mcrIsNineGates(tiles []Tile) bool {
	if len(tiles) != 14 || !tiles[0].IsSuit() {
		return false
	}
	suit := int(tiles[0]) / 9
	counts := TileCounts(tiles)
	for tile, count := range counts {
		if count > 0 && tile/9 != suit {
			return false
		}
	}
	base := suit * 9
	if counts[base] < 3 || counts[base+8] < 3 {
		return false
	}
	for rank := 1; rank <= 7; rank++ {
		if counts[base+rank] < 1 {
			return false
		}
	}
	return true
}

func mcrKnittedPatterns() [][]Tile {
	rankGroups := [][]int{{1, 4, 7}, {2, 5, 8}, {3, 6, 9}}
	permutations := [][3]int{{0, 1, 2}, {0, 2, 1}, {1, 0, 2}, {1, 2, 0}, {2, 0, 1}, {2, 1, 0}}
	patterns := make([][]Tile, 0, len(permutations))
	for _, permutation := range permutations {
		var pattern []Tile
		for suit, groupIndex := range permutation {
			for _, rank := range rankGroups[groupIndex] {
				pattern = append(pattern, Tile(suit*9+rank-1))
			}
		}
		SortTiles(pattern)
		patterns = append(patterns, pattern)
	}
	return patterns
}

func mcrDeclaredGroups(melds []Meld) []MCRGroup {
	groups := make([]MCRGroup, len(melds))
	for i, meld := range melds {
		kind := MCRGroupPung
		switch meld.Kind {
		case MeldChow:
			kind = MCRGroupChow
		case MeldKong:
			kind = MCRGroupKong
		}
		tiles := append([]Tile(nil), meld.Tiles...)
		SortTiles(tiles)
		groups[i] = MCRGroup{Kind: kind, Tiles: tiles, Open: true}
	}
	return groups
}

func uniqueMCRDecompositions(values []MCRDecomposition) []MCRDecomposition {
	seen := make(map[string]bool)
	result := make([]MCRDecomposition, 0, len(values))
	for _, value := range values {
		key := mcrCanonicalDecompositionKey(value)
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool {
		return mcrCanonicalDecompositionKey(result[i]) < mcrCanonicalDecompositionKey(result[j])
	})
	return result
}

func mcrCanonicalDecompositionKey(value MCRDecomposition) string {
	groups := make([]string, len(value.Groups))
	for i, group := range value.Groups {
		groups[i] = fmt.Sprintf("%s:%s:%t", group.Kind, FormatTiles(group.Tiles), group.Open)
	}
	sort.Strings(groups)
	return fmt.Sprintf("%s|%s", value.Kind, strings.Join(groups, ";"))
}

func copyMCRGroups(groups []MCRGroup) []MCRGroup {
	result := make([]MCRGroup, len(groups))
	for i, group := range groups {
		result[i] = group
		result[i].Tiles = append([]Tile(nil), group.Tiles...)
	}
	return result
}
