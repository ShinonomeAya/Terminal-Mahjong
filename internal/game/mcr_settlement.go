package game

func SettleMCR(score MCRScoreBreakdown, winner, discarder int, winType WinType) MCRSettlement {
	settlement := MCRSettlement{Winner: winner, Discarder: discarder, Score: score}
	if winner < 0 || winner >= len(settlement.Deltas) || !score.MeetsMinimum {
		return settlement
	}
	switch winType {
	case WinSelfDraw:
		payment := 8 + score.TotalPoints
		for player := range settlement.Deltas {
			if player == winner {
				continue
			}
			settlement.Deltas[player] = -payment
			settlement.Deltas[winner] += payment
		}
	case WinDiscard:
		if discarder < 0 || discarder >= len(settlement.Deltas) || discarder == winner {
			return MCRSettlement{Winner: winner, Discarder: discarder, Score: score}
		}
		for player := range settlement.Deltas {
			if player == winner {
				continue
			}
			payment := 8
			if player == discarder {
				payment += score.TotalPoints
			}
			settlement.Deltas[player] = -payment
			settlement.Deltas[winner] += payment
		}
	}
	return settlement
}
