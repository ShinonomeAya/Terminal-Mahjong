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
