package game

func ScoreRiichi(hand []Tile, melds []Meld, context RiichiYakuContext) RiichiScoreBreakdown {
	decompositions := RiichiDecompose(hand, melds, context.WinningTile)
	if len(decompositions) == 0 && len(context.Decomposition.Tiles) > 0 {
		decompositions = []RiichiDecomposition{context.Decomposition}
	}
	var best RiichiScoreBreakdown
	for _, decomposition := range decompositions {
		candidateContext := context
		candidateContext.Decomposition = decomposition
		yaku := DetectRiichiYaku(candidateContext)
		yakuHan, yakuman := riichiYakuTotals(yaku)
		bonusHan := riichiBonusHan(decomposition.Tiles, melds, candidateContext)
		if len(yaku) == 0 {
			if bonusHan > best.BonusHan {
				best.BonusHan = bonusHan
			}
			continue
		}
		fu := riichiFu(candidateContext)
		totalHan := yakuHan + bonusHan
		base, limit := riichiBasePoints(fu, totalHan, yakuman)
		score := RiichiScoreBreakdown{
			Yaku:          yaku,
			Fu:            fu,
			YakuHan:       yakuHan,
			BonusHan:      bonusHan,
			Yakuman:       yakuman,
			BasePoints:    base,
			LimitName:     limit,
			HasYaku:       true,
			WinningGroups: riichiGroupsAsMelds(decomposition.Groups),
		}
		if score.BasePoints > best.BasePoints {
			best = score
		}
	}
	return best
}

func riichiYakuTotals(yaku []RiichiYakuMatch) (int, int) {
	han, yakuman := 0, 0
	for _, match := range yaku {
		han += match.Han
		yakuman += match.Yakuman
	}
	return han, yakuman
}

func riichiBonusHan(tiles []Tile, melds []Meld, context RiichiYakuContext) int {
	all := append([]Tile(nil), tiles...)
	for _, meld := range melds {
		all = append(all, meld.Tiles...)
	}
	bonus := 0
	for _, tile := range all {
		if tile.IsRed() {
			bonus++
		}
		for _, indicator := range context.DoraIndicators {
			if tile.Base() == RiichiDoraTile(indicator).Base() {
				bonus++
			}
		}
		if context.Riichi == RiichiAccepted || context.Riichi == RiichiDeclared {
			for _, indicator := range context.UraIndicators {
				if tile.Base() == RiichiDoraTile(indicator).Base() {
					bonus++
				}
			}
		}
	}
	return bonus
}

func riichiFu(context RiichiYakuContext) int {
	if context.Decomposition.Kind == RiichiShapeSevenPairs {
		return 25
	}
	if riichiHasYaku(DetectRiichiYaku(context), "pinfu") && context.WinType == WinSelfDraw {
		return 20
	}
	fu := 20
	if context.Closed && context.WinType == WinDiscard {
		fu += 10
	}
	if context.WinType == WinSelfDraw {
		fu += 2
	}
	for _, group := range context.Decomposition.Groups {
		if group.Kind == MCRGroupPair && len(group.Tiles) > 0 {
			if riichiIsDragon(group.Tiles[0]) || group.Tiles[0].Base() == context.SeatWind.Base() || group.Tiles[0].Base() == context.PrevalentWind.Base() {
				fu += 2
			}
		}
		if riichiIsPungLike(group) && len(group.Tiles) > 0 {
			fu += riichiSetFu(group)
		}
	}
	if context.Decomposition.Wait == RiichiWaitKanchan || context.Decomposition.Wait == RiichiWaitPenchan || context.Decomposition.Wait == RiichiWaitTanki {
		fu += 2
	}
	if fu < 30 {
		fu = 30
	}
	return roundRiichiFu(fu)
}

func riichiSetFu(group MCRGroup) int {
	if len(group.Tiles) == 0 {
		return 0
	}
	base := 2
	if riichiIsTerminalOrHonor(group.Tiles[0]) {
		base *= 2
	}
	if !group.Open {
		base *= 2
	}
	if group.Kind == MCRGroupKong {
		base *= 4
	}
	return base
}

func riichiBasePoints(fu, han, yakuman int) (int, string) {
	if yakuman > 0 {
		return 8000 * yakuman, "yakuman"
	}
	if han >= 11 {
		return 6000, "sanbaiman"
	}
	if han >= 8 {
		return 4000, "baiman"
	}
	if han >= 6 {
		return 3000, "haneman"
	}
	if han >= 5 || (han == 4 && fu >= 40) || (han == 3 && fu >= 70) {
		return 2000, "mangan"
	}
	base := fu
	for i := 0; i < han+2; i++ {
		base *= 2
	}
	return base, ""
}

func roundRiichiFu(fu int) int {
	if fu%10 == 0 {
		return fu
	}
	return fu + (10 - fu%10)
}

func riichiHasYaku(values []RiichiYakuMatch, id string) bool {
	for _, value := range values {
		if value.ID == id {
			return true
		}
	}
	return false
}

func riichiGroupsAsMelds(groups []MCRGroup) []Meld {
	melds := make([]Meld, 0, len(groups))
	for _, group := range groups {
		kind := MeldPong
		switch group.Kind {
		case MCRGroupChow:
			kind = MeldChow
		case MCRGroupKong:
			kind = MeldKong
		case MCRGroupPair:
			continue
		}
		melds = append(melds, Meld{Kind: kind, Tiles: append([]Tile(nil), group.Tiles...)})
	}
	return melds
}
