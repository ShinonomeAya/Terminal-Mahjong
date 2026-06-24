package game

import "sort"

const mcrStandardMinimum = 8

func ScoreMCR(hand []Tile, melds []Meld, context MCRScoreContext) MCRScoreBreakdown {
	decompositions := MCRDecompose(hand, melds, context.WinningTile)
	if len(decompositions) == 0 {
		return MCRScoreBreakdown{}
	}

	best := MCRScoreBreakdown{}
	for _, decomposition := range decompositions {
		fanContext := MCRFanContext{
			Decomposition:   decomposition,
			WinningTile:     context.WinningTile,
			WinType:         context.WinType,
			SeatWind:        context.SeatWind,
			PrevalentWind:   context.PrevalentWind,
			Flowers:         mcrFlowerPlaceholders(context.Flowers),
			LastTileDraw:    context.LastTileDraw,
			LastTileClaim:   context.LastTileClaim,
			LastOfKind:      context.LastOfKind,
			ReplacementDraw: context.ReplacementDraw,
			RobbingKong:     context.RobbingKong,
		}
		score := scoreMCRFanOccurrences(DetectMCRFans(fanContext), context.Flowers, mcrStandardMinimum)
		score.WinningGrouping = mcrWinningGrouping(decomposition)
		if score.NonFlowerPoints > best.NonFlowerPoints ||
			(score.NonFlowerPoints == best.NonFlowerPoints && score.TotalPoints > best.TotalPoints) {
			best = score
		}
	}
	return best
}

func scoreMCRFanOccurrences(occurrences []MCRFanOccurrence, flowers, minimum int) MCRScoreBreakdown {
	byID := make(map[FanID][]MCRFanOccurrence)
	for _, value := range occurrences {
		if value.ID == "mcr_10" {
			continue
		}
		if _, ok := mcrFanMetadataByID[value.ID]; ok {
			byID[value.ID] = append(byID[value.ID], value)
		}
	}

	ids := make([]FanID, 0, len(byID))
	for id := range byID {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		left, right := mcrFanMetadataByID[ids[i]], mcrFanMetadataByID[ids[j]]
		if left.points != right.points {
			return left.points > right.points
		}
		return ids[i] < ids[j]
	})

	excluded := make(map[FanID]bool)
	var fans []FanMatch
	nonFlowerPoints := 0
	for _, id := range ids {
		if excluded[id] {
			continue
		}
		metadata := mcrFanMetadataByID[id]
		count := mcrOccurrenceCount(metadata, byID[id])
		if count == 0 {
			continue
		}
		fans = append(fans, fanMatch(id, count))
		nonFlowerPoints += metadata.points * count
		for _, blocked := range metadata.excludes {
			excluded[blocked] = true
		}
	}

	if len(fans) == 0 {
		fans = append(fans, fanMatch("mcr_43", 1))
		nonFlowerPoints = mcrFanMetadataByID["mcr_43"].points
	}
	if flowers > 0 {
		fans = append(fans, fanMatch("mcr_10", flowers))
	}
	sort.SliceStable(fans, func(i, j int) bool {
		if fans[i].Points != fans[j].Points {
			return fans[i].Points > fans[j].Points
		}
		return fans[i].ID < fans[j].ID
	})

	return MCRScoreBreakdown{
		Fans:            fans,
		NonFlowerPoints: nonFlowerPoints,
		FlowerPoints:    flowers,
		TotalPoints:     nonFlowerPoints + flowers,
		MeetsMinimum:    nonFlowerPoints >= minimum,
	}
}

func mcrOccurrenceCount(metadata mcrFanMetadata, occurrences []MCRFanOccurrence) int {
	if len(occurrences) == 0 {
		return 0
	}
	if !metadata.repeatable {
		return 1
	}
	ungrouped := 0
	var grouped []MCRFanOccurrence
	for _, value := range occurrences {
		if len(value.Groups) == 0 {
			count := value.Count
			if count < 1 {
				count = 1
			}
			ungrouped += count
		} else {
			grouped = append(grouped, value)
		}
	}
	return ungrouped + maximumDisjointMCROccurrences(grouped)
}

func maximumDisjointMCROccurrences(occurrences []MCRFanOccurrence) int {
	best := 0
	var visit func(int, map[int]bool, int)
	visit = func(index int, used map[int]bool, count int) {
		if index == len(occurrences) {
			if count > best {
				best = count
			}
			return
		}
		visit(index+1, used, count)
		for _, group := range occurrences[index].Groups {
			if used[group] {
				return
			}
		}
		next := make(map[int]bool, len(used)+len(occurrences[index].Groups))
		for group := range used {
			next[group] = true
		}
		for _, group := range occurrences[index].Groups {
			next[group] = true
		}
		visit(index+1, next, count+1)
	}
	visit(0, map[int]bool{}, 0)
	return best
}

func fanMatch(id FanID, count int) FanMatch {
	metadata := mcrFanMetadataByID[id]
	return FanMatch{ID: id, NameZH: metadata.zh, NameEN: metadata.en, Points: metadata.points, Count: count}
}

func mcrFlowerPlaceholders(count int) []Tile {
	if count <= 0 {
		return nil
	}
	flowers := make([]Tile, count)
	for index := range flowers {
		flowers[index] = FlowerPlum
	}
	return flowers
}

func mcrWinningGrouping(decomposition MCRDecomposition) []Meld {
	if len(decomposition.Groups) == 0 {
		if len(decomposition.Tiles) == 0 {
			return nil
		}
		return []Meld{{Kind: MeldKind(decomposition.Kind), Tiles: append([]Tile(nil), decomposition.Tiles...)}}
	}
	groups := make([]Meld, 0, len(decomposition.Groups))
	for _, group := range decomposition.Groups {
		kind := MeldKind(group.Kind)
		if group.Kind == MCRGroupPung {
			kind = MeldPong
		}
		groups = append(groups, Meld{Kind: kind, Tiles: append([]Tile(nil), group.Tiles...)})
	}
	return groups
}
