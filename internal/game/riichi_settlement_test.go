package game

import (
	"encoding/json"
	"os"
	"testing"
)

type riichiSettlementFixture struct {
	ID     string `json:"id"`
	Source string `json:"source"`
}

func TestRiichiSettlementFixtureManifest(t *testing.T) {
	file, err := os.Open("../../testdata/rules/riichi/settlement.json")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	var fixtures []riichiSettlementFixture
	if err := decoder.Decode(&fixtures); err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{
		"nondealer-ron": true, "dealer-tsumo": true, "nondealer-tsumo": true,
		"exhaustive-one-tenpai": true, "dealer-repeat": true,
	}
	for _, fixture := range fixtures {
		if !want[fixture.ID] || fixture.Source == "" {
			t.Fatalf("invalid riichi settlement fixture: %#v", fixture)
		}
		delete(want, fixture.ID)
	}
	if len(want) != 0 {
		t.Fatalf("missing riichi settlement fixtures: %#v", want)
	}
}

func TestRiichiSettlementGoldenFixtures(t *testing.T) {
	tests := []struct {
		name  string
		input RiichiSettlementInput
		want  [4]int
	}{
		{
			name:  "nondealer ron",
			input: RiichiSettlementInput{Winners: []int{1}, Discarder: 0, Dealer: 0, WinType: WinDiscard, Scores: []RiichiScoreBreakdown{{BasePoints: 2000, HasYaku: true}}},
			want:  [4]int{-8000, 8000, 0, 0},
		},
		{
			name:  "dealer tsumo",
			input: RiichiSettlementInput{Winners: []int{0}, Discarder: -1, Dealer: 0, WinType: WinSelfDraw, Scores: []RiichiScoreBreakdown{{BasePoints: 2000, HasYaku: true}}},
			want:  [4]int{12000, -4000, -4000, -4000},
		},
		{
			name:  "nondealer tsumo",
			input: RiichiSettlementInput{Winners: []int{1}, Discarder: -1, Dealer: 0, WinType: WinSelfDraw, Scores: []RiichiScoreBreakdown{{BasePoints: 2000, HasYaku: true}}},
			want:  [4]int{-4000, 8000, -2000, -2000},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := SettleRiichi(test.input)
			if got.Deltas != test.want || sumInts(got.Deltas) != 0 {
				t.Fatalf("settlement = %#v, want deltas %v", got, test.want)
			}
		})
	}
}

func TestRiichiExhaustiveDrawPayments(t *testing.T) {
	got := SettleRiichiExhaustiveDraw([4]bool{true, false, false, false})
	want := [4]int{3000, -1000, -1000, -1000}
	if got != want || sumInts(got) != 0 {
		t.Fatalf("draw deltas = %v, want %v", got, want)
	}
}

func TestRiichiMatchDealerRepeatsOnDealerWin(t *testing.T) {
	match, err := NewMatch(70, NewRiichiRuleSet(DefaultRuleConfig(ModeRiichi).Riichi))
	if err != nil {
		t.Fatal(err)
	}
	match.Round.Over = true
	match.Round.Phase = PhaseRoundOver
	match.Round.Winner = 0
	match.Round.WinType = WinSelfDraw
	match.Round.RiichiScore = &RiichiScoreBreakdown{BasePoints: 2000, HasYaku: true}

	match.completeRiichiRound()

	if match.Complete || match.RoundNumber != 1 || match.Dealer != 0 || match.Round.Dealer != 0 || match.LastRiichiSettlement == nil || match.LastRiichiSettlement.HonbaAfter != 1 {
		t.Fatalf("match after dealer win = %#v", match)
	}
}

func sumInts(values [4]int) int {
	total := 0
	for _, value := range values {
		total += value
	}
	return total
}
