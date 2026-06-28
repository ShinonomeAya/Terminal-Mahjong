package game

import (
	"encoding/json"
	"testing"
)

type dualModeAcceptanceFixture struct {
	name  string
	rules func() RuleSet
}

func TestDualModeFixedSeedReplayAndLegalActionClosure(t *testing.T) {
	for _, fixture := range dualModeAcceptanceFixtures() {
		t.Run(fixture.name, func(t *testing.T) {
			first := mustDualModeMatch(t, 120012, fixture.rules())
			second := mustDualModeMatch(t, 120012, fixture.rules())
			first.EnsureCurrentTurnDraw()
			second.EnsureCurrentTurnDraw()

			assertDualModeCanonicalJSONEqual(t, first.Snapshot(), second.Snapshot())
			assertDualModeCanonicalJSONEqual(t, first.ReplayLog(), second.ReplayLog())

			for _, action := range first.Round.Snapshot().LegalActions {
				fresh := mustDualModeMatch(t, 120012, fixture.rules())
				fresh.EnsureCurrentTurnDraw()
				result := fresh.ApplyCommand(GameCommand{
					PlayerID:  playerID(fresh.Round.Current),
					Kind:      action.Kind,
					TileIndex: action.TileIndex,
					Tile:      action.Tile,
				})
				if !result.OK {
					t.Fatalf("returned action %#v was rejected: %s", action, result.Error)
				}
			}
		})
	}
}

func TestDualModeRepresentativeSettlementsAreZeroSum(t *testing.T) {
	mcrScore := MCRScoreBreakdown{NonFlowerPoints: 8, TotalPoints: 8, MeetsMinimum: true}
	for _, settlement := range []MCRSettlement{
		SettleMCR(mcrScore, 1, 0, WinDiscard),
		SettleMCR(mcrScore, 0, -1, WinSelfDraw),
	} {
		if sumDualModeDeltas(settlement.Deltas) != 0 {
			t.Fatalf("MCR settlement is not zero-sum: %#v", settlement)
		}
	}

	riichiScore := RiichiScoreBreakdown{BasePoints: 2000, HasYaku: true}
	for _, settlement := range []RiichiSettlement{
		SettleRiichi(RiichiSettlementInput{Winners: []int{1}, Discarder: 0, Dealer: 0, WinType: WinDiscard, Scores: []RiichiScoreBreakdown{riichiScore}, Honba: 2}),
		SettleRiichi(RiichiSettlementInput{Winners: []int{0}, Discarder: -1, Dealer: 0, WinType: WinSelfDraw, Scores: []RiichiScoreBreakdown{riichiScore}, Honba: 1}),
	} {
		if sumDualModeDeltas(settlement.Deltas) != 0 {
			t.Fatalf("Riichi settlement is not zero-sum: %#v", settlement)
		}
	}
	if deltas := SettleRiichiExhaustiveDraw([4]bool{true, false, true, false}); sumDualModeDeltas(deltas) != 0 {
		t.Fatalf("Riichi exhaustive draw is not zero-sum: %v", deltas)
	}
}

func dualModeAcceptanceFixtures() []dualModeAcceptanceFixture {
	return []dualModeAcceptanceFixture{
		{
			name: "mcr",
			rules: func() RuleSet {
				return NewMCRRuleSet(DefaultRuleConfig(ModeMCR).MCR)
			},
		},
		{
			name: "riichi",
			rules: func() RuleSet {
				return NewRiichiRuleSet(DefaultRuleConfig(ModeRiichi).Riichi)
			},
		},
	}
}

func mustDualModeMatch(t *testing.T, seed int64, rules RuleSet) *Match {
	t.Helper()
	match, err := NewMatch(seed, rules)
	if err != nil {
		t.Fatal(err)
	}
	return match
}

func assertDualModeCanonicalJSONEqual(t *testing.T, first any, second any) {
	t.Helper()
	firstJSON, err := json.Marshal(first)
	if err != nil {
		t.Fatal(err)
	}
	secondJSON, err := json.Marshal(second)
	if err != nil {
		t.Fatal(err)
	}
	if string(firstJSON) != string(secondJSON) {
		t.Fatalf("canonical JSON differs\nfirst=%s\nsecond=%s", firstJSON, secondJSON)
	}
}

func sumDualModeDeltas(deltas [4]int) int {
	total := 0
	for _, delta := range deltas {
		total += delta
	}
	return total
}
