package game

import (
	"encoding/json"
	"os"
	"testing"
)

type mcrSettlementFixture struct {
	ID        string            `json:"id"`
	Score     MCRScoreBreakdown `json:"score"`
	Winner    int               `json:"winner"`
	Discarder int               `json:"discarder"`
	WinType   WinType           `json:"win_type"`
	Deltas    [4]int            `json:"deltas"`
}

func TestMCRSettlementGoldenFixtures(t *testing.T) {
	file, err := os.Open("../../testdata/rules/mcr/settlement.json")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	var fixtures []mcrSettlementFixture
	if err := decoder.Decode(&fixtures); err != nil {
		t.Fatal(err)
	}
	for _, fixture := range fixtures {
		got := SettleMCR(fixture.Score, fixture.Winner, fixture.Discarder, fixture.WinType)
		if got.Deltas != fixture.Deltas {
			t.Fatalf("%s deltas = %v, want %v", fixture.ID, got.Deltas, fixture.Deltas)
		}
	}
}

func TestMCRSettlementIsZeroSum(t *testing.T) {
	tests := []struct {
		name      string
		score     MCRScoreBreakdown
		winner    int
		discarder int
		winType   WinType
		want      [4]int
	}{
		{
			name:   "self draw",
			score:  MCRScoreBreakdown{NonFlowerPoints: 8, FlowerPoints: 2, TotalPoints: 10, MeetsMinimum: true},
			winner: 1, discarder: -1, winType: WinSelfDraw,
			want: [4]int{-18, 54, -18, -18},
		},
		{
			name:   "discard win",
			score:  MCRScoreBreakdown{NonFlowerPoints: 8, FlowerPoints: 2, TotalPoints: 10, MeetsMinimum: true},
			winner: 2, discarder: 0, winType: WinDiscard,
			want: [4]int{-18, -8, 34, -8},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := SettleMCR(test.score, test.winner, test.discarder, test.winType)
			if got.Deltas != test.want || got.Winner != test.winner || got.Discarder != test.discarder {
				t.Fatalf("settlement = %#v, want deltas %v", got, test.want)
			}
			total := 0
			for _, delta := range got.Deltas {
				total += delta
			}
			if total != 0 {
				t.Fatalf("settlement is not zero-sum: %v", got.Deltas)
			}
		})
	}
}

func TestMCRMatchAdvancesDealerAndCompletesAfterSixteenHands(t *testing.T) {
	match, err := NewMatch(41, NewMCRRuleSet(DefaultRuleConfig(ModeMCR).MCR))
	if err != nil {
		t.Fatal(err)
	}
	score := MCRScoreBreakdown{NonFlowerPoints: 8, TotalPoints: 8, MeetsMinimum: true}
	for hand := 1; hand <= 16; hand++ {
		match.Round.Over = true
		match.Round.Phase = PhaseRoundOver
		match.Round.Winner = 0
		match.Round.WinType = WinSelfDraw
		match.Round.MCRScore = &score

		match.completeMCRRound()

		if hand < 16 {
			if match.Complete || match.RoundNumber != hand+1 || match.Dealer != hand%4 || match.Round.Over || match.Round.Current != match.Dealer {
				t.Fatalf("after hand %d match=%#v", hand, match)
			}
		} else if !match.Complete || match.RoundNumber != 16 {
			t.Fatalf("final match=%#v", match)
		}
	}
	if match.Points != [4]int{768, -256, -256, -256} {
		t.Fatalf("points = %v", match.Points)
	}
}

func TestMCRWallExhaustionAdvancesWithoutSettlement(t *testing.T) {
	match, err := NewMatch(53, NewMCRRuleSet(DefaultRuleConfig(ModeMCR).MCR))
	if err != nil {
		t.Fatal(err)
	}
	match.Round.Over = true
	match.Round.Phase = PhaseRoundOver
	match.Round.Winner = -1
	match.Round.Reason = "draw: wall exhausted"

	match.completeMCRRound()

	if match.Points != [4]int{} || match.RoundNumber != 2 || match.Dealer != 1 || match.LastMCRSettlement != nil {
		t.Fatalf("match after draw = %#v", match)
	}
}

func TestMCRRoundMetadataDeterminesSeatAndPrevalentWinds(t *testing.T) {
	round := newMCRActionTestGame()
	round.Dealer = 2
	round.HandNumber = 5
	wantSeats := []Tile{mustFanTiles(t, "W")[0], mustFanTiles(t, "N")[0], mustFanTiles(t, "E")[0], mustFanTiles(t, "S")[0]}
	for player, want := range wantSeats {
		if got := mcrSeatWind(round, player); got != want {
			t.Fatalf("player %d seat wind = %s, want %s", player, got, want)
		}
	}
	if got, want := mcrPrevalentWind(round), mustFanTiles(t, "S")[0]; got != want {
		t.Fatalf("prevalent wind = %s, want %s", got, want)
	}
}

func TestMCRMatchCommandSettlesAndStartsNextHand(t *testing.T) {
	round := newMCRActionTestGame()
	round.HandNumber = 1
	round.Players[0].Hand = mustFanTiles(t, "1m", "9m", "1p", "9p", "1s", "9s", "E", "E", "S", "W", "N", "Z", "F", "B")
	round.RecordEvent(EventDraw, 0, mustFanTiles(t, "E")[0], "")
	match := &Match{Mode: ModeMCR, RuleConfig: round.RuleConfig, RoundNumber: 1, Round: round, rules: round.rules}

	result := match.ApplyCommand(GameCommand{PlayerID: "0", Kind: CommandWin})

	if !result.OK || match.Complete || match.RoundNumber != 2 || match.Dealer != 1 || match.LastMCRSettlement == nil {
		t.Fatalf("result=%#v match=%#v", result, match)
	}
	total := 0
	for _, points := range match.Points {
		total += points
	}
	if total != 0 || match.Points == [4]int{} {
		t.Fatalf("points = %v", match.Points)
	}
}
