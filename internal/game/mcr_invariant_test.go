package game

import "testing"

func TestMCRGeneratedInvariantsAcrossOneThousandSeeds(t *testing.T) {
	rules := NewMCRRuleSet(DefaultRuleConfig(ModeMCR).MCR)
	score := MCRScoreBreakdown{NonFlowerPoints: 8, TotalPoints: 8, MeetsMinimum: true}
	for seed := int64(1); seed <= 1000; seed++ {
		round, err := NewGameWithRules(seed, rules)
		if err != nil {
			t.Fatalf("seed %d: %v", seed, err)
		}
		assertMCRRoundConservation(t, seed, round)
		if _, ok := round.EnsureCurrentTurnDraw(); !ok {
			t.Fatalf("seed %d: initial draw failed", seed)
		}
		actions := round.rules.LegalActions(round, playerID(round.Current))
		for _, action := range actions {
			copyRound, err := NewGameWithRules(seed, NewMCRRuleSet(DefaultRuleConfig(ModeMCR).MCR))
			if err != nil {
				t.Fatalf("seed %d action %s: %v", seed, action.Kind, err)
			}
			if _, ok := copyRound.EnsureCurrentTurnDraw(); !ok {
				t.Fatalf("seed %d action %s: copied draw failed", seed, action.Kind)
			}
			result := copyRound.ApplyCommand(GameCommand{
				PlayerID:  playerID(copyRound.Current),
				Kind:      action.Kind,
				TileIndex: action.TileIndex,
				Tile:      action.Tile,
			})
			if !result.OK {
				t.Fatalf("seed %d returned action %#v was rejected: %s", seed, action, result.Error)
			}
		}

		winner := int(seed % 4)
		discarder := int((seed + 1) % 4)
		settlement := SettleMCR(score, winner, discarder, WinDiscard)
		if sumMCRDeltas(settlement.Deltas) != 0 {
			t.Fatalf("seed %d settlement is not zero-sum: %v", seed, settlement.Deltas)
		}

		match, err := NewMatch(seed, NewMCRRuleSet(DefaultRuleConfig(ModeMCR).MCR))
		if err != nil {
			t.Fatalf("seed %d match: %v", seed, err)
		}
		for hand := 1; hand <= 16; hand++ {
			match.Round.Over = true
			match.Round.Phase = PhaseRoundOver
			match.Round.Winner = winner
			match.Round.WinType = WinSelfDraw
			match.Round.MCRScore = &score
			match.completeMCRRound()
		}
		if !match.Complete || match.RoundNumber != 16 || len(match.MCRSettlements) != 16 {
			t.Fatalf("seed %d match did not terminate at hand 16: complete=%t hand=%d settlements=%d", seed, match.Complete, match.RoundNumber, len(match.MCRSettlements))
		}
	}
}

func assertMCRRoundConservation(t *testing.T, seed int64, round *Game) {
	t.Helper()
	total := len(round.Wall)
	counts := make(map[Tile]int)
	for _, tile := range round.Wall {
		counts[tile]++
	}
	for playerIndex, player := range round.Players {
		for _, tile := range player.Hand {
			if tile.IsFlower() {
				t.Fatalf("seed %d player %d retained flower %s", seed, playerIndex, tile)
			}
			counts[tile]++
		}
		for _, tile := range player.Flowers {
			counts[tile]++
		}
		total += len(player.Hand) + len(player.Flowers)
	}
	if total != 144 {
		t.Fatalf("seed %d tile total = %d, want 144", seed, total)
	}
	for tile, count := range counts {
		want := 4
		if tile.IsFlower() {
			want = 1
		}
		if count != want {
			t.Fatalf("seed %d tile %s count = %d, want %d", seed, tile, count, want)
		}
	}
}

func sumMCRDeltas(deltas [4]int) int {
	total := 0
	for _, delta := range deltas {
		total += delta
	}
	return total
}
