package game

import "testing"

func TestRiichiGeneratedInvariantsAcrossOneThousandSeeds(t *testing.T) {
	for _, redFives := range []int{0, 3} {
		config := DefaultRuleConfig(ModeRiichi)
		config.Riichi.RedFives = redFives
		rules := NewRiichiRuleSet(config.Riichi)
		score := RiichiScoreBreakdown{BasePoints: 2000, HasYaku: true}
		for seed := int64(1); seed <= 1000; seed++ {
			round, err := NewGameWithRules(seed, rules)
			if err != nil {
				t.Fatalf("red %d seed %d: %v", redFives, seed, err)
			}
			assertRiichiRoundConservation(t, redFives, seed, round)
			assertRiichiIndicatorBounds(t, redFives, seed, round)
			actions := round.rules.LegalActions(round, playerID(round.Current))
			for _, action := range actions {
				copyRound, err := NewGameWithRules(seed, NewRiichiRuleSet(config.Riichi))
				if err != nil {
					t.Fatalf("red %d seed %d action %s: %v", redFives, seed, action.Kind, err)
				}
				result := copyRound.ApplyCommand(GameCommand{
					PlayerID:  playerID(copyRound.Current),
					Kind:      action.Kind,
					TileIndex: action.TileIndex,
					Tile:      action.Tile,
				})
				if !result.OK {
					t.Fatalf("red %d seed %d returned action %#v was rejected: %s", redFives, seed, action, result.Error)
				}
			}

			for _, settlement := range []RiichiSettlement{
				SettleRiichi(RiichiSettlementInput{Winners: []int{1}, Discarder: 0, Dealer: 0, WinType: WinDiscard, Scores: []RiichiScoreBreakdown{score}}),
				SettleRiichi(RiichiSettlementInput{Winners: []int{0}, Discarder: -1, Dealer: 0, WinType: WinSelfDraw, Scores: []RiichiScoreBreakdown{score}, Honba: int(seed % 3), RiichiSticks: int(seed % 2)}),
			} {
				if sumRiichiDeltas(settlement.Deltas) != 0 {
					t.Fatalf("red %d seed %d settlement is not zero-sum: %#v", redFives, seed, settlement)
				}
			}
			if deltas := SettleRiichiExhaustiveDraw([4]bool{seed%2 == 0, true, seed%3 == 0, false}); sumRiichiDeltas(deltas) != 0 {
				t.Fatalf("red %d seed %d exhaustive draw is not zero-sum: %v", redFives, seed, deltas)
			}
		}
	}
}

func TestRiichiGeneratedMatchTerminatesDeterministically(t *testing.T) {
	score := RiichiScoreBreakdown{BasePoints: 2000, HasYaku: true}
	for _, redFives := range []int{0, 3} {
		config := DefaultRuleConfig(ModeRiichi)
		config.Riichi.RedFives = redFives
		for seed := int64(1); seed <= 1000; seed++ {
			match, err := NewMatch(seed, NewRiichiRuleSet(config.Riichi))
			if err != nil {
				t.Fatalf("red %d seed %d match: %v", redFives, seed, err)
			}
			for !match.Complete {
				winner := (match.Dealer + 1) % 4
				match.Round.Over = true
				match.Round.Phase = PhaseRoundOver
				match.Round.Winner = winner
				match.Round.Discarder = -1
				match.Round.WinType = WinSelfDraw
				match.Round.RiichiScore = &score
				match.completeRiichiRound()
			}
			if match.RoundNumber != 9 || len(match.RiichiSettlements) != 8 {
				t.Fatalf("red %d seed %d match did not terminate after east-south: hand=%d settlements=%d", redFives, seed, match.RoundNumber, len(match.RiichiSettlements))
			}
		}
	}
}

func TestRiichiFuritenRonNeverExposedAsLegalClaim(t *testing.T) {
	round := newRiichiActionTestGame(t)
	rules := round.rules.(*RiichiRuleSet)
	discard := mustTile(t, "3m")
	round.Players[1].Hand = mustTiles(t, "1m", "2m", "4m", "5m", "6m", "2p", "3p", "4p", "7s", "7s", "7s", "E", "E")
	round.Riichi.TemporaryFuriten[1] = true

	pending := rules.buildPendingClaim(round, 0, discard)

	if pending == nil {
		return
	}
	for _, option := range pending.Options {
		if option.Player == 1 && option.Kind == ClaimWin {
			t.Fatalf("furiten ron exposed: %#v", pending.Options)
		}
	}
}

func assertRiichiRoundConservation(t *testing.T, redFives int, seed int64, round *Game) {
	t.Helper()
	total := len(round.Wall)
	counts := make(map[Tile]int)
	for _, tile := range round.Wall {
		counts[tile.Base()]++
	}
	if round.Riichi == nil || len(round.Riichi.DeadWall) != 14 {
		t.Fatalf("red %d seed %d riichi state = %#v", redFives, seed, round.Riichi)
	}
	total += len(round.Riichi.DeadWall)
	for _, tile := range round.Riichi.DeadWall {
		counts[tile.Base()]++
	}
	for playerIndex, player := range round.Players {
		for _, tile := range player.Hand {
			counts[tile.Base()]++
		}
		for _, tile := range player.Discards {
			counts[tile.Base()]++
		}
		for _, meld := range player.Melds {
			for _, tile := range meld.Tiles {
				counts[tile.Base()]++
			}
		}
		total += len(player.Hand) + len(player.Discards)
		for _, meld := range player.Melds {
			total += len(meld.Tiles)
		}
		if len(player.Flowers) != 0 {
			t.Fatalf("red %d seed %d player %d has riichi flowers: %#v", redFives, seed, playerIndex, player.Flowers)
		}
	}
	if total != 136 {
		t.Fatalf("red %d seed %d tile total = %d, want 136", redFives, seed, total)
	}
	for tile, count := range counts {
		if count < 0 || count > 4 {
			t.Fatalf("red %d seed %d tile %s count = %d", redFives, seed, tile, count)
		}
	}
}

func assertRiichiIndicatorBounds(t *testing.T, redFives int, seed int64, round *Game) {
	t.Helper()
	if round.Riichi.KanCount < 0 || round.Riichi.KanCount > 4 || round.Riichi.RinshanDraws < 0 || round.Riichi.RinshanDraws > 4 {
		t.Fatalf("red %d seed %d kan/rinshan state = %#v", redFives, seed, round.Riichi)
	}
	if len(round.Riichi.DoraIndicators) == 0 || len(round.Riichi.DoraIndicators) > 5 || len(round.Riichi.UraIndicators) == 0 || len(round.Riichi.UraIndicators) > 5 {
		t.Fatalf("red %d seed %d indicators = %#v", redFives, seed, round.Riichi)
	}
}

func sumRiichiDeltas(deltas [4]int) int {
	total := 0
	for _, delta := range deltas {
		total += delta
	}
	return total
}
