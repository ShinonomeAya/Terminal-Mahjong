package game

func SettleRiichi(input RiichiSettlementInput) RiichiSettlement {
	settlement := RiichiSettlement{
		Winners:     append([]int(nil), input.Winners...),
		Discarder:   input.Discarder,
		Scores:      append([]RiichiScoreBreakdown(nil), input.Scores...),
		HonbaAfter:  input.Honba,
		SticksAfter: input.RiichiSticks,
	}
	for index, winner := range input.Winners {
		if winner < 0 || winner >= len(settlement.Deltas) || index >= len(input.Scores) || !input.Scores[index].HasYaku {
			continue
		}
		score := input.Scores[index]
		switch input.WinType {
		case WinDiscard:
			if input.Discarder < 0 || input.Discarder >= len(settlement.Deltas) || input.Discarder == winner {
				continue
			}
			payment := riichiRonPayment(score.BasePoints, winner == input.Dealer) + input.Honba*300
			settlement.Deltas[input.Discarder] -= payment
			settlement.Deltas[winner] += payment
		case WinSelfDraw:
			for player := range settlement.Deltas {
				if player == winner {
					continue
				}
				payment := riichiTsumoPayment(score.BasePoints, winner == input.Dealer, player == input.Dealer) + input.Honba*100
				settlement.Deltas[player] -= payment
				settlement.Deltas[winner] += payment
			}
		}
	}
	dealerWon := false
	for _, winner := range input.Winners {
		if winner == input.Dealer {
			dealerWon = true
			break
		}
	}
	if dealerWon {
		settlement.HonbaAfter = input.Honba + 1
	} else {
		settlement.HonbaAfter = 0
	}
	if len(input.Winners) > 0 {
		settlement.SticksAfter = 0
	}
	return settlement
}

func SettleRiichiExhaustiveDraw(tenpai [4]bool) [4]int {
	count := 0
	for _, ready := range tenpai {
		if ready {
			count++
		}
	}
	var deltas [4]int
	if count == 0 || count == 4 {
		return deltas
	}
	tenpaiGain := 3000 / count
	notenLoss := 3000 / (4 - count)
	for player, ready := range tenpai {
		if ready {
			deltas[player] = tenpaiGain
		} else {
			deltas[player] = -notenLoss
		}
	}
	return deltas
}

func riichiRonPayment(basePoints int, dealer bool) int {
	multiplier := 4
	if dealer {
		multiplier = 6
	}
	return roundRiichiPayment(basePoints * multiplier)
}

func riichiTsumoPayment(basePoints int, winnerDealer bool, payerDealer bool) int {
	if winnerDealer || payerDealer {
		return roundRiichiPayment(basePoints * 2)
	}
	return roundRiichiPayment(basePoints)
}

func roundRiichiPayment(points int) int {
	if points%100 == 0 {
		return points
	}
	return points + (100 - points%100)
}
