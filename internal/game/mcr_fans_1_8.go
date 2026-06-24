package game

func DetectMCRFans(context MCRFanContext) []MCRFanOccurrence {
	var result []MCRFanOccurrence
	result = append(result, detectPureDoubleChow(context)...)
	result = append(result, detectMixedDoubleChow(context)...)
	result = append(result, detectShortStraight(context)...)
	result = append(result, detectTwoTerminalChows(context)...)
	result = append(result, detectTerminalHonorPungs(context)...)
	result = append(result, detectMeldedKongs(context)...)
	result = append(result, detectOneVoidedSuit(context)...)
	result = append(result, detectNoHonors(context)...)
	result = append(result, detectSelfDrawn(context)...)
	result = append(result, detectFlowers(context)...)
	result = append(result, detectWaits(context)...)
	result = append(result, detectDragonPungs(context)...)
	result = append(result, detectWindPungs(context, "mcr_15", context.PrevalentWind)...)
	result = append(result, detectWindPungs(context, "mcr_16", context.SeatWind)...)
	result = append(result, detectConcealedHand(context)...)
	result = append(result, detectAllChows(context)...)
	result = append(result, detectTileHogs(context)...)
	result = append(result, detectDoublePungs(context)...)
	result = append(result, detectTwoConcealedPungs(context)...)
	result = append(result, detectConcealedKongs(context)...)
	result = append(result, detectAllSimples(context)...)
	result = append(result, detectOutsideHand(context)...)
	result = append(result, detectFullyConcealedHand(context)...)
	result = append(result, detectTwoMeldedKongs(context)...)
	result = append(result, detectLastOfKind(context)...)
	result = append(result, detectAllPungs(context)...)
	result = append(result, detectHalfFlush(context)...)
	result = append(result, detectMixedShiftedChows(context)...)
	result = append(result, detectAllTypes(context)...)
	result = append(result, detectMeldedHand(context)...)
	result = append(result, detectTwoDragonPungs(context)...)
	result = append(result, detectMixedStraight(context)...)
	result = append(result, detectReversibleTiles(context)...)
	result = append(result, detectMixedTripleChow(context)...)
	result = append(result, detectMixedShiftedPungs(context)...)
	result = append(result, detectTwoConcealedKongs(context)...)
	result = append(result, detectWinCircumstances(context)...)
	result = append(result, detectMCRHighFans(context)...)
	return result
}

func detectPureDoubleChow(context MCRFanContext) []MCRFanOccurrence {
	return matchingGroupPairs(context, "mcr_01", 1, func(left, right MCRGroup) bool {
		return left.Kind == MCRGroupChow && right.Kind == MCRGroupChow && left.Tiles[0] == right.Tiles[0]
	})
}

func detectMixedDoubleChow(context MCRFanContext) []MCRFanOccurrence {
	return matchingGroupPairs(context, "mcr_02", 1, func(left, right MCRGroup) bool {
		return left.Kind == MCRGroupChow && right.Kind == MCRGroupChow && left.Tiles[0].Rank() == right.Tiles[0].Rank() && tileSuit(left.Tiles[0]) != tileSuit(right.Tiles[0])
	})
}

func detectShortStraight(context MCRFanContext) []MCRFanOccurrence {
	return matchingGroupPairs(context, "mcr_03", 1, func(left, right MCRGroup) bool {
		if left.Kind != MCRGroupChow || right.Kind != MCRGroupChow || tileSuit(left.Tiles[0]) != tileSuit(right.Tiles[0]) {
			return false
		}
		difference := left.Tiles[0].Rank() - right.Tiles[0].Rank()
		return difference == 3 || difference == -3
	})
}

func detectTwoTerminalChows(context MCRFanContext) []MCRFanOccurrence {
	return matchingGroupPairs(context, "mcr_04", 1, func(left, right MCRGroup) bool {
		if left.Kind != MCRGroupChow || right.Kind != MCRGroupChow || tileSuit(left.Tiles[0]) != tileSuit(right.Tiles[0]) {
			return false
		}
		leftRank, rightRank := left.Tiles[0].Rank(), right.Tiles[0].Rank()
		return (leftRank == 1 && rightRank == 7) || (leftRank == 7 && rightRank == 1)
	})
}

func detectTerminalHonorPungs(context MCRFanContext) []MCRFanOccurrence {
	var result []MCRFanOccurrence
	for index, group := range context.Decomposition.Groups {
		if (group.Kind == MCRGroupPung || group.Kind == MCRGroupKong) && isTerminalOrHonor(group.Tiles[0]) {
			result = append(result, occurrence("mcr_05", 1, []int{index}))
		}
	}
	return result
}

func detectMeldedKongs(context MCRFanContext) []MCRFanOccurrence {
	var result []MCRFanOccurrence
	for index, group := range context.Decomposition.Groups {
		if group.Kind == MCRGroupKong && group.Open {
			result = append(result, occurrence("mcr_06", 1, []int{index}))
		}
	}
	return result
}

func detectOneVoidedSuit(context MCRFanContext) []MCRFanOccurrence {
	var suits [3]bool
	for _, tile := range mcrContextTiles(context) {
		if tile.IsSuit() {
			suits[tileSuit(tile)] = true
		}
	}
	used := 0
	for _, present := range suits {
		if present {
			used++
		}
	}
	if used < 3 {
		return []MCRFanOccurrence{occurrence("mcr_07", 1, nil)}
	}
	return nil
}

func detectNoHonors(context MCRFanContext) []MCRFanOccurrence {
	tiles := mcrContextTiles(context)
	for _, tile := range tiles {
		if tile >= 27 && tile < BaseTileTypeCount {
			return nil
		}
	}
	if len(tiles) > 0 {
		return []MCRFanOccurrence{occurrence("mcr_08", 1, nil)}
	}
	return nil
}

func detectSelfDrawn(context MCRFanContext) []MCRFanOccurrence {
	if context.WinType == WinSelfDraw {
		return []MCRFanOccurrence{occurrence("mcr_09", 1, nil)}
	}
	return nil
}

func detectFlowers(context MCRFanContext) []MCRFanOccurrence {
	if len(context.Flowers) == 0 {
		return nil
	}
	return []MCRFanOccurrence{{ID: "mcr_10", Points: 1, Count: len(context.Flowers)}}
}

func detectWaits(context MCRFanContext) []MCRFanOccurrence {
	for index, group := range context.Decomposition.Groups {
		if group.Kind == MCRGroupPair && group.Tiles[0] == context.WinningTile {
			return []MCRFanOccurrence{occurrence("mcr_13", 1, []int{index})}
		}
	}
	for index, group := range context.Decomposition.Groups {
		if group.Kind != MCRGroupChow || !groupContainsTile(group, context.WinningTile) {
			continue
		}
		start := group.Tiles[0].Rank()
		winning := context.WinningTile.Rank()
		if (start == 1 && winning == 3) || (start == 7 && winning == 7) {
			return []MCRFanOccurrence{occurrence("mcr_11", 1, []int{index})}
		}
		if winning == start+1 {
			return []MCRFanOccurrence{occurrence("mcr_12", 1, []int{index})}
		}
	}
	return nil
}

func detectDragonPungs(context MCRFanContext) []MCRFanOccurrence {
	var result []MCRFanOccurrence
	for index, group := range context.Decomposition.Groups {
		if isPungGroup(group) && group.Tiles[0] >= 31 && group.Tiles[0] <= 33 {
			result = append(result, occurrence("mcr_14", 2, []int{index}))
		}
	}
	return result
}

func detectWindPungs(context MCRFanContext, id FanID, wind Tile) []MCRFanOccurrence {
	if wind < 27 || wind > 30 {
		return nil
	}
	for index, group := range context.Decomposition.Groups {
		if isPungGroup(group) && group.Tiles[0] == wind {
			return []MCRFanOccurrence{occurrence(id, 2, []int{index})}
		}
	}
	return nil
}

func detectConcealedHand(context MCRFanContext) []MCRFanOccurrence {
	if context.WinType != WinDiscard || hasOpenMCRGroup(context.Decomposition.Groups) {
		return nil
	}
	return []MCRFanOccurrence{occurrence("mcr_17", 2, nil)}
}

func detectAllChows(context MCRFanContext) []MCRFanOccurrence {
	chows := 0
	for _, group := range context.Decomposition.Groups {
		switch group.Kind {
		case MCRGroupChow:
			chows++
		case MCRGroupPair:
			if group.Tiles[0] >= 27 {
				return nil
			}
		default:
			return nil
		}
	}
	if chows == 4 {
		return []MCRFanOccurrence{occurrence("mcr_18", 2, nil)}
	}
	return nil
}

func detectTileHogs(context MCRFanContext) []MCRFanOccurrence {
	counts := make(map[Tile]int)
	kongs := make(map[Tile]bool)
	for _, group := range context.Decomposition.Groups {
		for _, tile := range group.Tiles {
			counts[tile]++
		}
		if group.Kind == MCRGroupKong {
			kongs[group.Tiles[0]] = true
		}
	}
	var result []MCRFanOccurrence
	for tile, count := range counts {
		if count == 4 && !kongs[tile] {
			result = append(result, occurrence("mcr_19", 2, nil))
		}
	}
	return result
}

func detectDoublePungs(context MCRFanContext) []MCRFanOccurrence {
	return matchingGroupPairs(context, "mcr_20", 2, func(left, right MCRGroup) bool {
		return isPungGroup(left) && isPungGroup(right) && left.Tiles[0].IsSuit() && right.Tiles[0].IsSuit() && left.Tiles[0].Rank() == right.Tiles[0].Rank() && tileSuit(left.Tiles[0]) != tileSuit(right.Tiles[0])
	})
}

func detectTwoConcealedPungs(context MCRFanContext) []MCRFanOccurrence {
	var concealed []int
	for index, group := range context.Decomposition.Groups {
		if !isPungGroup(group) || group.Open {
			continue
		}
		if context.WinType == WinDiscard && group.Kind == MCRGroupPung && groupContainsTile(group, context.WinningTile) {
			continue
		}
		concealed = append(concealed, index)
	}
	if len(concealed) >= 2 {
		return []MCRFanOccurrence{occurrence("mcr_21", 2, concealed[:2])}
	}
	return nil
}

func detectConcealedKongs(context MCRFanContext) []MCRFanOccurrence {
	var result []MCRFanOccurrence
	for index, group := range context.Decomposition.Groups {
		if group.Kind == MCRGroupKong && !group.Open {
			result = append(result, occurrence("mcr_22", 2, []int{index}))
		}
	}
	return result
}

func detectAllSimples(context MCRFanContext) []MCRFanOccurrence {
	tiles := mcrContextTiles(context)
	if len(tiles) == 0 {
		return nil
	}
	for _, tile := range tiles {
		if !tile.IsSuit() || tile.Rank() == 1 || tile.Rank() == 9 {
			return nil
		}
	}
	return []MCRFanOccurrence{occurrence("mcr_23", 2, nil)}
}

func detectOutsideHand(context MCRFanContext) []MCRFanOccurrence {
	groups := context.Decomposition.Groups
	if len(groups) == 0 {
		return nil
	}
	for _, group := range groups {
		switch group.Kind {
		case MCRGroupChow:
			start := group.Tiles[0].Rank()
			if start != 1 && start != 7 {
				return nil
			}
		case MCRGroupPair, MCRGroupPung, MCRGroupKong:
			if !isTerminalOrHonor(group.Tiles[0]) {
				return nil
			}
		default:
			return nil
		}
	}
	return []MCRFanOccurrence{occurrence("mcr_24", 4, nil)}
}

func detectFullyConcealedHand(context MCRFanContext) []MCRFanOccurrence {
	if context.WinType != WinSelfDraw || hasOpenMCRGroup(context.Decomposition.Groups) {
		return nil
	}
	return []MCRFanOccurrence{occurrence("mcr_25", 4, nil)}
}

func detectTwoMeldedKongs(context MCRFanContext) []MCRFanOccurrence {
	var groups []int
	for index, group := range context.Decomposition.Groups {
		if group.Kind == MCRGroupKong && group.Open {
			groups = append(groups, index)
		}
	}
	if len(groups) < 2 {
		return nil
	}
	return []MCRFanOccurrence{occurrence("mcr_26", 4, groups[:2])}
}

func detectLastOfKind(context MCRFanContext) []MCRFanOccurrence {
	if !context.LastOfKind {
		return nil
	}
	return []MCRFanOccurrence{occurrence("mcr_27", 4, nil)}
}

func detectAllPungs(context MCRFanContext) []MCRFanOccurrence {
	pungs, pairs := 0, 0
	for _, group := range context.Decomposition.Groups {
		switch {
		case isPungGroup(group):
			pungs++
		case group.Kind == MCRGroupPair:
			pairs++
		default:
			return nil
		}
	}
	if pungs != 4 || pairs != 1 {
		return nil
	}
	return []MCRFanOccurrence{occurrence("mcr_28", 6, nil)}
}

func detectHalfFlush(context MCRFanContext) []MCRFanOccurrence {
	tiles := mcrContextTiles(context)
	suit := -1
	hasHonor := false
	for _, tile := range tiles {
		if tile.IsSuit() {
			if suit == -1 {
				suit = tileSuit(tile)
			} else if suit != tileSuit(tile) {
				return nil
			}
		} else if tile >= 27 && tile < BaseTileTypeCount {
			hasHonor = true
		}
	}
	if len(tiles) == 0 || suit == -1 || !hasHonor {
		return nil
	}
	return []MCRFanOccurrence{occurrence("mcr_29", 6, nil)}
}

func detectMixedShiftedChows(context MCRFanContext) []MCRFanOccurrence {
	return matchingGroupTriples(context, "mcr_30", 6, func(first, second, third MCRGroup) bool {
		if first.Kind != MCRGroupChow || second.Kind != MCRGroupChow || third.Kind != MCRGroupChow {
			return false
		}
		return threeDistinctSuits(first.Tiles[0], second.Tiles[0], third.Tiles[0]) &&
			threeConsecutiveRanks(first.Tiles[0].Rank(), second.Tiles[0].Rank(), third.Tiles[0].Rank())
	})
}

func detectAllTypes(context MCRFanContext) []MCRFanOccurrence {
	var suits [3]bool
	hasWind, hasDragon := false, false
	for _, tile := range mcrContextTiles(context) {
		switch {
		case tile.IsSuit():
			suits[tileSuit(tile)] = true
		case tile >= 27 && tile <= 30:
			hasWind = true
		case tile >= 31 && tile <= 33:
			hasDragon = true
		}
	}
	if suits[0] && suits[1] && suits[2] && hasWind && hasDragon {
		return []MCRFanOccurrence{occurrence("mcr_31", 6, nil)}
	}
	return nil
}

func detectMeldedHand(context MCRFanContext) []MCRFanOccurrence {
	if context.WinType != WinDiscard {
		return nil
	}
	melds, pairs := 0, 0
	for _, group := range context.Decomposition.Groups {
		if group.Kind == MCRGroupPair {
			pairs++
			continue
		}
		if !group.Open || (group.Kind != MCRGroupChow && !isPungGroup(group)) {
			return nil
		}
		melds++
	}
	if melds == 4 && pairs == 1 {
		return []MCRFanOccurrence{occurrence("mcr_32", 6, nil)}
	}
	return nil
}

func detectTwoDragonPungs(context MCRFanContext) []MCRFanOccurrence {
	var groups []int
	for index, group := range context.Decomposition.Groups {
		if isPungGroup(group) && group.Tiles[0] >= 31 && group.Tiles[0] <= 33 {
			groups = append(groups, index)
		}
	}
	if len(groups) < 2 {
		return nil
	}
	return []MCRFanOccurrence{occurrence("mcr_33", 6, groups[:2])}
}

func detectMixedStraight(context MCRFanContext) []MCRFanOccurrence {
	return matchingGroupTriples(context, "mcr_34", 8, func(first, second, third MCRGroup) bool {
		if first.Kind != MCRGroupChow || second.Kind != MCRGroupChow || third.Kind != MCRGroupChow ||
			!threeDistinctSuits(first.Tiles[0], second.Tiles[0], third.Tiles[0]) {
			return false
		}
		var starts [10]bool
		starts[first.Tiles[0].Rank()] = true
		starts[second.Tiles[0].Rank()] = true
		starts[third.Tiles[0].Rank()] = true
		return starts[1] && starts[4] && starts[7]
	})
}

func detectReversibleTiles(context MCRFanContext) []MCRFanOccurrence {
	tiles := mcrContextTiles(context)
	if len(tiles) == 0 {
		return nil
	}
	for _, tile := range tiles {
		suit, rank := tileSuit(tile), tile.Rank()
		allowed := (suit == 1 && (rank == 1 || rank == 2 || rank == 3 || rank == 4 || rank == 5 || rank == 8 || rank == 9)) ||
			(suit == 2 && (rank == 2 || rank == 4 || rank == 5 || rank == 6 || rank == 8 || rank == 9)) ||
			tile == 33
		if !allowed {
			return nil
		}
	}
	return []MCRFanOccurrence{occurrence("mcr_35", 8, nil)}
}

func detectMixedTripleChow(context MCRFanContext) []MCRFanOccurrence {
	return matchingGroupTriples(context, "mcr_36", 8, func(first, second, third MCRGroup) bool {
		return first.Kind == MCRGroupChow && second.Kind == MCRGroupChow && third.Kind == MCRGroupChow &&
			threeDistinctSuits(first.Tiles[0], second.Tiles[0], third.Tiles[0]) &&
			first.Tiles[0].Rank() == second.Tiles[0].Rank() && first.Tiles[0].Rank() == third.Tiles[0].Rank()
	})
}

func detectMixedShiftedPungs(context MCRFanContext) []MCRFanOccurrence {
	return matchingGroupTriples(context, "mcr_37", 8, func(first, second, third MCRGroup) bool {
		return isPungGroup(first) && isPungGroup(second) && isPungGroup(third) &&
			first.Tiles[0].IsSuit() && second.Tiles[0].IsSuit() && third.Tiles[0].IsSuit() &&
			threeDistinctSuits(first.Tiles[0], second.Tiles[0], third.Tiles[0]) &&
			threeConsecutiveRanks(first.Tiles[0].Rank(), second.Tiles[0].Rank(), third.Tiles[0].Rank())
	})
}

func detectTwoConcealedKongs(context MCRFanContext) []MCRFanOccurrence {
	var groups []int
	for index, group := range context.Decomposition.Groups {
		if group.Kind == MCRGroupKong && !group.Open {
			groups = append(groups, index)
		}
	}
	if len(groups) < 2 {
		return nil
	}
	return []MCRFanOccurrence{occurrence("mcr_38", 8, groups[:2])}
}

func detectWinCircumstances(context MCRFanContext) []MCRFanOccurrence {
	var result []MCRFanOccurrence
	if context.LastTileDraw {
		result = append(result, occurrence("mcr_39", 8, nil))
	}
	if context.LastTileClaim {
		result = append(result, occurrence("mcr_40", 8, nil))
	}
	if context.ReplacementDraw {
		result = append(result, occurrence("mcr_41", 8, nil))
	}
	if context.RobbingKong {
		result = append(result, occurrence("mcr_42", 8, nil))
	}
	return result
}

func matchingGroupPairs(context MCRFanContext, id FanID, points int, matches func(MCRGroup, MCRGroup) bool) []MCRFanOccurrence {
	var result []MCRFanOccurrence
	groups := context.Decomposition.Groups
	for left := 0; left < len(groups); left++ {
		for right := left + 1; right < len(groups); right++ {
			if matches(groups[left], groups[right]) {
				result = append(result, occurrence(id, points, []int{left, right}))
			}
		}
	}
	return result
}

func matchingGroupTriples(context MCRFanContext, id FanID, points int, matches func(MCRGroup, MCRGroup, MCRGroup) bool) []MCRFanOccurrence {
	var result []MCRFanOccurrence
	groups := context.Decomposition.Groups
	for first := 0; first < len(groups); first++ {
		for second := first + 1; second < len(groups); second++ {
			for third := second + 1; third < len(groups); third++ {
				if matches(groups[first], groups[second], groups[third]) {
					result = append(result, occurrence(id, points, []int{first, second, third}))
				}
			}
		}
	}
	return result
}

func threeDistinctSuits(first, second, third Tile) bool {
	firstSuit, secondSuit, thirdSuit := tileSuit(first), tileSuit(second), tileSuit(third)
	return firstSuit >= 0 && secondSuit >= 0 && thirdSuit >= 0 &&
		firstSuit != secondSuit && firstSuit != thirdSuit && secondSuit != thirdSuit
}

func threeConsecutiveRanks(first, second, third int) bool {
	minRank, maxRank := first, first
	for _, rank := range []int{second, third} {
		if rank < minRank {
			minRank = rank
		}
		if rank > maxRank {
			maxRank = rank
		}
	}
	return maxRank-minRank == 2 && first != second && first != third && second != third
}

func occurrence(id FanID, points int, groups []int) MCRFanOccurrence {
	return MCRFanOccurrence{ID: id, Points: points, Count: 1, Groups: append([]int(nil), groups...)}
}

func mcrContextTiles(context MCRFanContext) []Tile {
	if len(context.Decomposition.Groups) == 0 {
		return append([]Tile(nil), context.Decomposition.Tiles...)
	}
	var tiles []Tile
	for _, group := range context.Decomposition.Groups {
		tiles = append(tiles, group.Tiles...)
	}
	return tiles
}

func tileSuit(tile Tile) int {
	if !tile.IsSuit() {
		return -1
	}
	return int(tile) / 9
}

func isTerminalOrHonor(tile Tile) bool {
	return tile >= 27 || (tile.IsSuit() && (tile.Rank() == 1 || tile.Rank() == 9))
}

func groupContainsTile(group MCRGroup, tile Tile) bool {
	for _, groupTile := range group.Tiles {
		if groupTile == tile {
			return true
		}
	}
	return false
}

func isPungGroup(group MCRGroup) bool {
	return group.Kind == MCRGroupPung || group.Kind == MCRGroupKong
}

func hasOpenMCRGroup(groups []MCRGroup) bool {
	for _, group := range groups {
		if group.Kind != MCRGroupPair && group.Open {
			return true
		}
	}
	return false
}
