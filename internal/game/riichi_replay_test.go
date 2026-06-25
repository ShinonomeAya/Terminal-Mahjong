package game

import "testing"

func TestRiichiReplayHidesUraUntilRoundEndAndIncludesSettlement(t *testing.T) {
	round := newRiichiActionTestGame(t)
	round.RiichiScore = &RiichiScoreBreakdown{BasePoints: 2000, HasYaku: true}

	before := round.ReplayLog()
	if len(before.UraIndicators) != 0 {
		t.Fatalf("ura indicators visible before end: %#v", before.UraIndicators)
	}
	if before.RiichiScore != nil {
		t.Fatalf("riichi score visible before end: %#v", before.RiichiScore)
	}

	round.Over = true
	round.Phase = PhaseRoundOver
	after := round.ReplayLog()
	if len(after.UraIndicators) != len(round.Riichi.UraIndicators) {
		t.Fatalf("ura indicators after end = %#v, want %#v", after.UraIndicators, round.Riichi.UraIndicators)
	}
	if after.RiichiScore == nil || after.RiichiScore.BasePoints != 2000 {
		t.Fatalf("riichi score after end = %#v", after.RiichiScore)
	}
}

func TestRiichiMatchReplayIncludesSettlements(t *testing.T) {
	match, err := NewMatch(11, NewRiichiRuleSet(DefaultRuleConfig(ModeRiichi).Riichi))
	if err != nil {
		t.Fatal(err)
	}
	match.RiichiSettlements = []RiichiSettlement{{
		Winners:     []int{0},
		Discarder:   -1,
		Deltas:      [4]int{4000, -2000, -1000, -1000},
		Scores:      []RiichiScoreBreakdown{{BasePoints: 2000, HasYaku: true}},
		HonbaAfter:  1,
		SticksAfter: 0,
	}}
	match.LastRiichiSettlement = &match.RiichiSettlements[0]
	match.Round.Over = true
	match.Round.Phase = PhaseRoundOver

	log := match.ReplayLog()

	if len(log.RiichiSettlements) != 1 || log.RiichiSettlements[0].HonbaAfter != 1 {
		t.Fatalf("riichi settlements missing from replay: %#v", log.RiichiSettlements)
	}
	if log.RiichiScore == nil || log.RiichiScore.BasePoints != 2000 {
		t.Fatalf("last riichi score missing from replay: %#v", log.RiichiScore)
	}
}
