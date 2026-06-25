package game

import (
	"fmt"
	"sort"
	"strings"
)

func RiichiWaits(hand []Tile, declared []Meld) []Tile {
	owned := riichiOwnedBaseCounts(hand, declared)
	waits := make([]Tile, 0, TileTypeCount)
	for tile := 0; tile < TileTypeCount; tile++ {
		candidate := Tile(tile)
		if owned[candidate] >= 4 {
			continue
		}
		if len(RiichiDecompose(hand, declared, candidate)) > 0 {
			waits = append(waits, candidate)
		}
	}
	SortTiles(waits)
	return waits
}

func RiichiTenpai(hand []Tile, declared []Meld) bool {
	return len(RiichiWaits(hand, declared)) > 0
}

func RiichiDecompose(hand []Tile, declared []Meld, winning Tile) []RiichiDecomposition {
	if winning.IsFlower() {
		return nil
	}
	originalTiles := append([]Tile(nil), hand...)
	originalTiles = append(originalTiles, winning)
	SortTiles(originalTiles)
	originalAll := append([]Tile(nil), originalTiles...)
	for _, meld := range declared {
		originalAll = append(originalAll, meld.Tiles...)
	}
	normalizedHand := normalizeRiichiTiles(hand)
	normalizedDeclared := normalizeRiichiMelds(declared)
	normalizedWinning := winning.Base()
	candidates := MCRDecompose(normalizedHand, normalizedDeclared, normalizedWinning)
	result := make([]RiichiDecomposition, 0, len(candidates))
	for _, candidate := range candidates {
		kind, ok := riichiShapeKind(candidate.Kind)
		if !ok {
			continue
		}
		if kind == RiichiShapeSevenPairs && !riichiStrictSevenPairs(candidate.Tiles) {
			continue
		}
		groups := restoreRiichiRedTiles(candidate.Groups, originalAll)
		waits := riichiWaitKinds(candidate, normalizedWinning)
		for _, wait := range waits {
			result = append(result, RiichiDecomposition{Kind: kind, Groups: groups, Tiles: append([]Tile(nil), originalTiles...), Wait: wait})
		}
	}
	return uniqueRiichiDecompositions(result)
}

func normalizeRiichiTiles(tiles []Tile) []Tile {
	result := make([]Tile, len(tiles))
	for i, tile := range tiles {
		result[i] = tile.Base()
	}
	SortTiles(result)
	return result
}

func normalizeRiichiMelds(melds []Meld) []Meld {
	result := make([]Meld, len(melds))
	for i, meld := range melds {
		result[i] = Meld{Kind: meld.Kind, Tiles: normalizeRiichiTiles(meld.Tiles)}
	}
	return result
}

func riichiShapeKind(kind MCRShapeKind) (RiichiShapeKind, bool) {
	switch kind {
	case MCRShapeStandard:
		return RiichiShapeStandard, true
	case MCRShapeSevenPairs:
		return RiichiShapeSevenPairs, true
	case MCRShapeThirteenOrphans:
		return RiichiShapeThirteenOrphans, true
	default:
		return "", false
	}
}

func riichiStrictSevenPairs(tiles []Tile) bool {
	if len(tiles) != 14 {
		return false
	}
	counts := TileCounts(tiles)
	pairs := 0
	for _, count := range counts {
		if count == 0 {
			continue
		}
		if count != 2 {
			return false
		}
		pairs++
	}
	return pairs == 7
}

func riichiWaitKinds(value MCRDecomposition, winning Tile) []RiichiWaitKind {
	switch value.Kind {
	case MCRShapeSevenPairs:
		return []RiichiWaitKind{RiichiWaitTanki}
	case MCRShapeThirteenOrphans:
		return []RiichiWaitKind{RiichiWaitKokushi}
	}
	seen := make(map[RiichiWaitKind]bool)
	var result []RiichiWaitKind
	for _, group := range value.Groups {
		wait, ok := riichiWaitKindForGroup(group, winning)
		if !ok || seen[wait] {
			continue
		}
		seen[wait] = true
		result = append(result, wait)
	}
	if len(result) == 0 {
		result = append(result, RiichiWaitUnknown)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func riichiWaitKindForGroup(group MCRGroup, winning Tile) (RiichiWaitKind, bool) {
	if !riichiGroupContains(group, winning) {
		return "", false
	}
	switch group.Kind {
	case MCRGroupPair:
		return RiichiWaitTanki, true
	case MCRGroupPung, MCRGroupKong:
		return RiichiWaitShanpon, true
	case MCRGroupChow:
		return riichiChowWaitKind(group.Tiles, winning)
	default:
		return "", false
	}
}

func riichiGroupContains(group MCRGroup, winning Tile) bool {
	for _, tile := range group.Tiles {
		if tile.Base() == winning.Base() {
			return true
		}
	}
	return false
}

func riichiChowWaitKind(tiles []Tile, winning Tile) (RiichiWaitKind, bool) {
	if len(tiles) != 3 || !winning.IsSuit() {
		return "", false
	}
	normalized := normalizeRiichiTiles(tiles)
	start := normalized[0].Rank()
	winningRank := winning.Rank()
	if winningRank == start+1 {
		return RiichiWaitKanchan, true
	}
	if start == 1 && winningRank == 3 {
		return RiichiWaitPenchan, true
	}
	if start == 7 && winningRank == 7 {
		return RiichiWaitPenchan, true
	}
	if winningRank == start || winningRank == start+2 {
		return RiichiWaitRyanmen, true
	}
	return "", false
}

func restoreRiichiRedTiles(groups []MCRGroup, originals []Tile) []MCRGroup {
	result := copyMCRGroups(groups)
	for _, red := range originals {
		if !red.IsRed() {
			continue
		}
		restored := false
		base := red.Base()
		for groupIndex := range result {
			if restored {
				break
			}
			for tileIndex, tile := range result[groupIndex].Tiles {
				if tile.Base() == base && !tile.IsRed() {
					result[groupIndex].Tiles[tileIndex] = red
					restored = true
					break
				}
			}
		}
	}
	return result
}

func riichiOwnedBaseCounts(hand []Tile, declared []Meld) [TileTypeCount]int {
	counts := TileCounts(hand)
	for _, meld := range declared {
		for _, tile := range meld.Tiles {
			base := tile.Base()
			if base >= 0 && int(base) < TileTypeCount {
				counts[base]++
			}
		}
	}
	return counts
}

func uniqueRiichiDecompositions(values []RiichiDecomposition) []RiichiDecomposition {
	seen := make(map[string]bool)
	result := make([]RiichiDecomposition, 0, len(values))
	for _, value := range values {
		key := riichiCanonicalDecompositionKey(value)
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool {
		return riichiCanonicalDecompositionKey(result[i]) < riichiCanonicalDecompositionKey(result[j])
	})
	return result
}

func riichiCanonicalDecompositionKey(value RiichiDecomposition) string {
	groups := make([]string, len(value.Groups))
	for i, group := range value.Groups {
		groups[i] = fmt.Sprintf("%s:%s:%t", group.Kind, FormatTiles(group.Tiles), group.Open)
	}
	sort.Strings(groups)
	return fmt.Sprintf("%s|%s|%s", value.Kind, value.Wait, strings.Join(groups, ";"))
}
