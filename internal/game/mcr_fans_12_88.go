package game

func detectMCRHighFans(context MCRFanContext) []MCRFanOccurrence {
	var result []MCRFanOccurrence
	result = append(result, detectSpecialMCRShapes(context, "mcr_44", 12, MCRShapeLesserHonorsKnitted, MCRShapeGreaterHonorsKnitted)...)
	result = append(result, detectSpecialMCRShape(context, "mcr_45", 12, MCRShapeKnittedStraight)...)
	result = append(result, detectSuitedRankRange(context, "mcr_46", 12, 6, 9)...)
	result = append(result, detectSuitedRankRange(context, "mcr_47", 12, 1, 4)...)
	result = append(result, detectBigThreeWinds(context)...)
	result = append(result, detectPureStraight(context)...)
	result = append(result, detectThreeSuitedTerminalChows(context)...)
	result = append(result, detectPureShiftedChows(context)...)
	result = append(result, detectAllFives(context)...)
	result = append(result, detectTriplePung(context)...)
	result = append(result, detectThreeConcealedPungs(context)...)
	result = append(result, detectSpecialMCRShapes(context, "mcr_55", 24, MCRShapeSevenPairs, MCRShapeSevenShiftedPairs)...)
	result = append(result, detectSpecialMCRShape(context, "mcr_56", 24, MCRShapeGreaterHonorsKnitted)...)
	result = append(result, detectAllEvenPungs(context)...)
	result = append(result, detectFullFlush(context)...)
	result = append(result, detectPureTripleChow(context)...)
	result = append(result, detectPureShiftedPungs(context)...)
	result = append(result, detectSuitedRankRange(context, "mcr_61", 24, 7, 9)...)
	result = append(result, detectSuitedRankRange(context, "mcr_62", 24, 4, 6)...)
	result = append(result, detectSuitedRankRange(context, "mcr_63", 24, 1, 3)...)
	result = append(result, detectFourShiftedChows(context)...)
	result = append(result, detectKongCount(context, "mcr_65", 32, 3)...)
	result = append(result, detectAllTerminalsAndHonors(context)...)
	result = append(result, detectQuadrupleChow(context)...)
	result = append(result, detectFourPureShiftedPungs(context)...)
	result = append(result, detectAllTerminals(context)...)
	result = append(result, detectAllHonors(context)...)
	result = append(result, detectLittleFourWinds(context)...)
	result = append(result, detectLittleThreeDragons(context)...)
	result = append(result, detectFourConcealedPungs(context)...)
	result = append(result, detectPureTerminalChows(context)...)
	result = append(result, detectBigFourWinds(context)...)
	result = append(result, detectBigThreeDragons(context)...)
	result = append(result, detectAllGreen(context)...)
	result = append(result, detectSpecialMCRShape(context, "mcr_78", 88, MCRShapeNineGates)...)
	result = append(result, detectKongCount(context, "mcr_79", 88, 4)...)
	result = append(result, detectSpecialMCRShape(context, "mcr_80", 88, MCRShapeSevenShiftedPairs)...)
	result = append(result, detectSpecialMCRShape(context, "mcr_81", 88, MCRShapeThirteenOrphans)...)
	return result
}

func detectSpecialMCRShape(context MCRFanContext, id FanID, points int, kind MCRShapeKind) []MCRFanOccurrence {
	return detectSpecialMCRShapes(context, id, points, kind)
}

func detectSpecialMCRShapes(context MCRFanContext, id FanID, points int, kinds ...MCRShapeKind) []MCRFanOccurrence {
	for _, kind := range kinds {
		if context.Decomposition.Kind == kind {
			return []MCRFanOccurrence{occurrence(id, points, nil)}
		}
	}
	return nil
}

func detectSuitedRankRange(context MCRFanContext, id FanID, points, minimum, maximum int) []MCRFanOccurrence {
	tiles := mcrContextTiles(context)
	if len(tiles) == 0 {
		return nil
	}
	for _, tile := range tiles {
		if !tile.IsSuit() || tile.Rank() < minimum || tile.Rank() > maximum {
			return nil
		}
	}
	return []MCRFanOccurrence{occurrence(id, points, nil)}
}

func detectBigThreeWinds(context MCRFanContext) []MCRFanOccurrence {
	var groups []int
	for index, group := range context.Decomposition.Groups {
		if isPungGroup(group) && group.Tiles[0] >= 27 && group.Tiles[0] <= 30 {
			groups = append(groups, index)
		}
	}
	if len(groups) < 3 {
		return nil
	}
	return []MCRFanOccurrence{occurrence("mcr_48", 12, groups[:3])}
}

func detectPureStraight(context MCRFanContext) []MCRFanOccurrence {
	return matchingGroupTriples(context, "mcr_49", 16, func(first, second, third MCRGroup) bool {
		if first.Kind != MCRGroupChow || second.Kind != MCRGroupChow || third.Kind != MCRGroupChow {
			return false
		}
		if tileSuit(first.Tiles[0]) != tileSuit(second.Tiles[0]) || tileSuit(first.Tiles[0]) != tileSuit(third.Tiles[0]) {
			return false
		}
		var starts [10]bool
		starts[first.Tiles[0].Rank()] = true
		starts[second.Tiles[0].Rank()] = true
		starts[third.Tiles[0].Rank()] = true
		return starts[1] && starts[4] && starts[7]
	})
}

func detectThreeSuitedTerminalChows(context MCRFanContext) []MCRFanOccurrence {
	var terminalChows [3][2]int
	pairSuit := -1
	pairGroup := -1
	for index, group := range context.Decomposition.Groups {
		switch group.Kind {
		case MCRGroupChow:
			start := group.Tiles[0].Rank()
			if start != 1 && start != 7 {
				return nil
			}
			terminalChows[tileSuit(group.Tiles[0])][start/7] = index + 1
		case MCRGroupPair:
			if !group.Tiles[0].IsSuit() || group.Tiles[0].Rank() != 5 {
				return nil
			}
			pairSuit, pairGroup = tileSuit(group.Tiles[0]), index
		default:
			return nil
		}
	}
	if pairSuit < 0 {
		return nil
	}
	var groups []int
	for suit := 0; suit < 3; suit++ {
		if suit == pairSuit {
			if terminalChows[suit][0] != 0 || terminalChows[suit][1] != 0 {
				return nil
			}
			continue
		}
		if terminalChows[suit][0] == 0 || terminalChows[suit][1] == 0 {
			return nil
		}
		groups = append(groups, terminalChows[suit][0]-1, terminalChows[suit][1]-1)
	}
	groups = append(groups, pairGroup)
	return []MCRFanOccurrence{occurrence("mcr_50", 16, groups)}
}

func detectPureShiftedChows(context MCRFanContext) []MCRFanOccurrence {
	return matchingGroupTriples(context, "mcr_51", 16, func(first, second, third MCRGroup) bool {
		if first.Kind != MCRGroupChow || second.Kind != MCRGroupChow || third.Kind != MCRGroupChow ||
			tileSuit(first.Tiles[0]) != tileSuit(second.Tiles[0]) || tileSuit(first.Tiles[0]) != tileSuit(third.Tiles[0]) {
			return false
		}
		return ranksFormArithmeticSequence(first.Tiles[0].Rank(), second.Tiles[0].Rank(), third.Tiles[0].Rank(), 1, 2)
	})
}

func detectAllFives(context MCRFanContext) []MCRFanOccurrence {
	groups := context.Decomposition.Groups
	if len(groups) == 0 {
		return nil
	}
	for _, group := range groups {
		containsFive := false
		for _, tile := range group.Tiles {
			if tile.IsSuit() && tile.Rank() == 5 {
				containsFive = true
				break
			}
		}
		if !containsFive {
			return nil
		}
	}
	return []MCRFanOccurrence{occurrence("mcr_52", 16, nil)}
}

func detectTriplePung(context MCRFanContext) []MCRFanOccurrence {
	return matchingGroupTriples(context, "mcr_53", 16, func(first, second, third MCRGroup) bool {
		return isPungGroup(first) && isPungGroup(second) && isPungGroup(third) &&
			first.Tiles[0].IsSuit() && second.Tiles[0].IsSuit() && third.Tiles[0].IsSuit() &&
			threeDistinctSuits(first.Tiles[0], second.Tiles[0], third.Tiles[0]) &&
			first.Tiles[0].Rank() == second.Tiles[0].Rank() && first.Tiles[0].Rank() == third.Tiles[0].Rank()
	})
}

func detectThreeConcealedPungs(context MCRFanContext) []MCRFanOccurrence {
	groups := concealedPungGroups(context)
	if len(groups) < 3 {
		return nil
	}
	return []MCRFanOccurrence{occurrence("mcr_54", 16, groups[:3])}
}

func detectAllEvenPungs(context MCRFanContext) []MCRFanOccurrence {
	pungs, pairs := 0, 0
	for _, group := range context.Decomposition.Groups {
		if len(group.Tiles) == 0 || !group.Tiles[0].IsSuit() || group.Tiles[0].Rank()%2 != 0 {
			return nil
		}
		switch {
		case isPungGroup(group):
			pungs++
		case group.Kind == MCRGroupPair:
			pairs++
		default:
			return nil
		}
	}
	if pungs == 4 && pairs == 1 {
		return []MCRFanOccurrence{occurrence("mcr_57", 24, nil)}
	}
	return nil
}

func detectFullFlush(context MCRFanContext) []MCRFanOccurrence {
	tiles := mcrContextTiles(context)
	if len(tiles) == 0 || !tiles[0].IsSuit() {
		return nil
	}
	suit := tileSuit(tiles[0])
	for _, tile := range tiles[1:] {
		if !tile.IsSuit() || tileSuit(tile) != suit {
			return nil
		}
	}
	return []MCRFanOccurrence{occurrence("mcr_58", 24, nil)}
}

func detectPureTripleChow(context MCRFanContext) []MCRFanOccurrence {
	return matchingGroupTriples(context, "mcr_59", 24, func(first, second, third MCRGroup) bool {
		return first.Kind == MCRGroupChow && second.Kind == MCRGroupChow && third.Kind == MCRGroupChow &&
			first.Tiles[0] == second.Tiles[0] && first.Tiles[0] == third.Tiles[0]
	})
}

func detectPureShiftedPungs(context MCRFanContext) []MCRFanOccurrence {
	return matchingGroupTriples(context, "mcr_60", 24, func(first, second, third MCRGroup) bool {
		return isPungGroup(first) && isPungGroup(second) && isPungGroup(third) &&
			first.Tiles[0].IsSuit() && second.Tiles[0].IsSuit() && third.Tiles[0].IsSuit() &&
			tileSuit(first.Tiles[0]) == tileSuit(second.Tiles[0]) && tileSuit(first.Tiles[0]) == tileSuit(third.Tiles[0]) &&
			threeConsecutiveRanks(first.Tiles[0].Rank(), second.Tiles[0].Rank(), third.Tiles[0].Rank())
	})
}

func concealedPungGroups(context MCRFanContext) []int {
	var groups []int
	for index, group := range context.Decomposition.Groups {
		if !isPungGroup(group) || group.Open {
			continue
		}
		if context.WinType == WinDiscard && group.Kind == MCRGroupPung && groupContainsTile(group, context.WinningTile) {
			continue
		}
		groups = append(groups, index)
	}
	return groups
}

func ranksFormArithmeticSequence(first, second, third int, allowedSteps ...int) bool {
	minimum, maximum := first, first
	for _, rank := range []int{second, third} {
		if rank < minimum {
			minimum = rank
		}
		if rank > maximum {
			maximum = rank
		}
	}
	if first == second || first == third || second == third {
		return false
	}
	for _, step := range allowedSteps {
		if maximum-minimum != step*2 {
			continue
		}
		middle := minimum + step
		if first == middle || second == middle || third == middle {
			return true
		}
	}
	return false
}

func detectFourShiftedChows(context MCRFanContext) []MCRFanOccurrence {
	var groups []int
	var ranks []int
	suit := -1
	for index, group := range context.Decomposition.Groups {
		if group.Kind != MCRGroupChow {
			continue
		}
		groupSuit := tileSuit(group.Tiles[0])
		if suit == -1 {
			suit = groupSuit
		} else if suit != groupSuit {
			return nil
		}
		groups = append(groups, index)
		ranks = append(ranks, group.Tiles[0].Rank())
	}
	if len(groups) != 4 || !fourRanksFormArithmeticSequence(ranks, 1, 2) {
		return nil
	}
	return []MCRFanOccurrence{occurrence("mcr_64", 32, groups)}
}

func detectKongCount(context MCRFanContext, id FanID, points, minimum int) []MCRFanOccurrence {
	var groups []int
	for index, group := range context.Decomposition.Groups {
		if group.Kind == MCRGroupKong {
			groups = append(groups, index)
		}
	}
	if len(groups) < minimum {
		return nil
	}
	return []MCRFanOccurrence{occurrence(id, points, groups[:minimum])}
}

func detectAllTerminalsAndHonors(context MCRFanContext) []MCRFanOccurrence {
	tiles := mcrContextTiles(context)
	hasTerminal, hasHonor := false, false
	for _, tile := range tiles {
		if tile.IsSuit() && (tile.Rank() == 1 || tile.Rank() == 9) {
			hasTerminal = true
			continue
		}
		if tile >= 27 && tile < BaseTileTypeCount {
			hasHonor = true
			continue
		}
		return nil
	}
	if hasTerminal && hasHonor {
		return []MCRFanOccurrence{occurrence("mcr_66", 32, nil)}
	}
	return nil
}

func detectQuadrupleChow(context MCRFanContext) []MCRFanOccurrence {
	var groups []int
	var start Tile
	for index, group := range context.Decomposition.Groups {
		if group.Kind != MCRGroupChow {
			continue
		}
		if len(groups) == 0 {
			start = group.Tiles[0]
		} else if group.Tiles[0] != start {
			return nil
		}
		groups = append(groups, index)
	}
	if len(groups) != 4 {
		return nil
	}
	return []MCRFanOccurrence{occurrence("mcr_67", 48, groups)}
}

func detectFourPureShiftedPungs(context MCRFanContext) []MCRFanOccurrence {
	var groups []int
	var ranks []int
	suit := -1
	for index, group := range context.Decomposition.Groups {
		if !isPungGroup(group) {
			continue
		}
		if !group.Tiles[0].IsSuit() {
			return nil
		}
		groupSuit := tileSuit(group.Tiles[0])
		if suit == -1 {
			suit = groupSuit
		} else if suit != groupSuit {
			return nil
		}
		groups = append(groups, index)
		ranks = append(ranks, group.Tiles[0].Rank())
	}
	if len(groups) != 4 || !fourRanksFormArithmeticSequence(ranks, 1) {
		return nil
	}
	return []MCRFanOccurrence{occurrence("mcr_68", 48, groups)}
}

func detectAllTerminals(context MCRFanContext) []MCRFanOccurrence {
	tiles := mcrContextTiles(context)
	if len(tiles) == 0 {
		return nil
	}
	for _, tile := range tiles {
		if !tile.IsSuit() || (tile.Rank() != 1 && tile.Rank() != 9) {
			return nil
		}
	}
	return []MCRFanOccurrence{occurrence("mcr_69", 64, nil)}
}

func detectAllHonors(context MCRFanContext) []MCRFanOccurrence {
	tiles := mcrContextTiles(context)
	if len(tiles) == 0 {
		return nil
	}
	for _, tile := range tiles {
		if tile < 27 || tile >= BaseTileTypeCount {
			return nil
		}
	}
	return []MCRFanOccurrence{occurrence("mcr_70", 64, nil)}
}

func detectLittleFourWinds(context MCRFanContext) []MCRFanOccurrence {
	pungs, pair, groups := honorFamilyGroups(context, 27, 30)
	if pungs != 3 || pair != 1 {
		return nil
	}
	return []MCRFanOccurrence{occurrence("mcr_71", 64, groups)}
}

func detectLittleThreeDragons(context MCRFanContext) []MCRFanOccurrence {
	pungs, pair, groups := honorFamilyGroups(context, 31, 33)
	if pungs != 2 || pair != 1 {
		return nil
	}
	return []MCRFanOccurrence{occurrence("mcr_72", 64, groups)}
}

func detectFourConcealedPungs(context MCRFanContext) []MCRFanOccurrence {
	groups := concealedPungGroups(context)
	if len(groups) < 4 {
		return nil
	}
	return []MCRFanOccurrence{occurrence("mcr_73", 64, groups[:4])}
}

func detectPureTerminalChows(context MCRFanContext) []MCRFanOccurrence {
	var groups []int
	starts := [2]int{}
	suit := -1
	pairFound := false
	for index, group := range context.Decomposition.Groups {
		switch group.Kind {
		case MCRGroupChow:
			groupSuit, start := tileSuit(group.Tiles[0]), group.Tiles[0].Rank()
			if (start != 1 && start != 7) || (suit != -1 && suit != groupSuit) {
				return nil
			}
			suit = groupSuit
			starts[start/7]++
			groups = append(groups, index)
		case MCRGroupPair:
			if !group.Tiles[0].IsSuit() || group.Tiles[0].Rank() != 5 || (suit != -1 && suit != tileSuit(group.Tiles[0])) {
				return nil
			}
			suit = tileSuit(group.Tiles[0])
			pairFound = true
			groups = append(groups, index)
		default:
			return nil
		}
	}
	if len(groups) != 5 || starts[0] != 2 || starts[1] != 2 || !pairFound {
		return nil
	}
	return []MCRFanOccurrence{occurrence("mcr_74", 64, groups)}
}

func detectBigFourWinds(context MCRFanContext) []MCRFanOccurrence {
	pungs, _, groups := honorFamilyGroups(context, 27, 30)
	if pungs != 4 {
		return nil
	}
	return []MCRFanOccurrence{occurrence("mcr_75", 88, groups)}
}

func detectBigThreeDragons(context MCRFanContext) []MCRFanOccurrence {
	pungs, _, groups := honorFamilyGroups(context, 31, 33)
	if pungs != 3 {
		return nil
	}
	return []MCRFanOccurrence{occurrence("mcr_76", 88, groups)}
}

func detectAllGreen(context MCRFanContext) []MCRFanOccurrence {
	tiles := mcrContextTiles(context)
	if len(tiles) == 0 {
		return nil
	}
	for _, tile := range tiles {
		allowed := tile == 32 || (tileSuit(tile) == 2 && (tile.Rank() == 2 || tile.Rank() == 3 || tile.Rank() == 4 || tile.Rank() == 6 || tile.Rank() == 8))
		if !allowed {
			return nil
		}
	}
	return []MCRFanOccurrence{occurrence("mcr_77", 88, nil)}
}

func honorFamilyGroups(context MCRFanContext, minimum, maximum Tile) (int, int, []int) {
	pungs, pairs := 0, 0
	var groups []int
	for index, group := range context.Decomposition.Groups {
		if len(group.Tiles) == 0 || group.Tiles[0] < minimum || group.Tiles[0] > maximum {
			continue
		}
		if isPungGroup(group) {
			pungs++
			groups = append(groups, index)
		} else if group.Kind == MCRGroupPair {
			pairs++
			groups = append(groups, index)
		}
	}
	return pungs, pairs, groups
}

func fourRanksFormArithmeticSequence(ranks []int, allowedSteps ...int) bool {
	if len(ranks) != 4 {
		return false
	}
	minimum, maximum := ranks[0], ranks[0]
	seen := make(map[int]bool, 4)
	for _, rank := range ranks {
		if seen[rank] {
			return false
		}
		seen[rank] = true
		if rank < minimum {
			minimum = rank
		}
		if rank > maximum {
			maximum = rank
		}
	}
	for _, step := range allowedSteps {
		if maximum-minimum == step*3 && seen[minimum+step] && seen[minimum+2*step] {
			return true
		}
	}
	return false
}
