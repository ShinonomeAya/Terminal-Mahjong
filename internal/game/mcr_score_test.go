package game

import (
	"encoding/json"
	"os"
	"testing"
)

type mcrScoringFixture struct {
	ID           string             `json:"id"`
	Occurrences  []MCRFanOccurrence `json:"occurrences"`
	Flowers      int                `json:"flowers"`
	ExpectedFans []FanID            `json:"expected_fans"`
	NonFlower    int                `json:"non_flower"`
	Total        int                `json:"total"`
	Eligible     bool               `json:"eligible"`
}

func TestMCRScoringGoldenFixtures(t *testing.T) {
	file, err := os.Open("../../testdata/rules/mcr/scoring.json")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	var fixtures []mcrScoringFixture
	if err := decoder.Decode(&fixtures); err != nil {
		t.Fatal(err)
	}
	for _, fixture := range fixtures {
		t.Run(fixture.ID, func(t *testing.T) {
			got := scoreMCRFanOccurrences(fixture.Occurrences, fixture.Flowers, 8)
			assertMCRScoreFanIDs(t, got, fixture.ExpectedFans)
			if got.NonFlowerPoints != fixture.NonFlower || got.TotalPoints != fixture.Total || got.MeetsMinimum != fixture.Eligible {
				t.Fatalf("score = %#v", got)
			}
		})
	}
}

func TestMCRScoringAppliesExclusionsAndCountingPrinciples(t *testing.T) {
	tests := []struct {
		name        string
		occurrences []MCRFanOccurrence
		wantFans    []FanID
		wantPoints  int
	}{
		{
			name: "all simples excludes no honors",
			occurrences: []MCRFanOccurrence{
				occurrence("mcr_08", 1, nil),
				occurrence("mcr_23", 2, nil),
			},
			wantFans: []FanID{"mcr_23"}, wantPoints: 2,
		},
		{
			name: "seven shifted pairs excludes lower flush and pairs",
			occurrences: []MCRFanOccurrence{
				occurrence("mcr_08", 1, nil),
				occurrence("mcr_55", 24, nil),
				occurrence("mcr_58", 24, nil),
				occurrence("mcr_80", 88, nil),
			},
			wantFans: []FanID{"mcr_80"}, wantPoints: 88,
		},
		{
			name: "repeatable groups use each group once",
			occurrences: []MCRFanOccurrence{
				occurrence("mcr_01", 1, []int{0, 1}),
				occurrence("mcr_01", 1, []int{0, 2}),
				occurrence("mcr_01", 1, []int{2, 3}),
			},
			wantFans: []FanID{"mcr_01"}, wantPoints: 2,
		},
		{
			name: "non-repeatable fan is counted once",
			occurrences: []MCRFanOccurrence{
				occurrence("mcr_28", 6, nil),
				occurrence("mcr_28", 6, nil),
			},
			wantFans: []FanID{"mcr_28"}, wantPoints: 6,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := scoreMCRFanOccurrences(test.occurrences, 0, 8)
			assertMCRScoreFanIDs(t, got, test.wantFans)
			if got.NonFlowerPoints != test.wantPoints {
				t.Fatalf("non-flower points = %d, want %d", got.NonFlowerPoints, test.wantPoints)
			}
		})
	}
}

func TestMCRScoringMinimumIgnoresFlowersAndAddsChickenHand(t *testing.T) {
	tests := []struct {
		name        string
		occurrences []MCRFanOccurrence
		flowers     int
		wantFans    []FanID
		wantNon     int
		wantTotal   int
		wantMinimum bool
	}{
		{name: "no fan becomes chicken hand", wantFans: []FanID{"mcr_43"}, wantNon: 8, wantTotal: 8, wantMinimum: true},
		{name: "seven points plus flowers remains ineligible", occurrences: []MCRFanOccurrence{
			occurrence("mcr_28", 6, nil), occurrence("mcr_09", 1, nil), occurrence("mcr_10", 1, nil), occurrence("mcr_10", 1, nil),
		}, flowers: 2, wantFans: []FanID{"mcr_28", "mcr_09", "mcr_10"}, wantNon: 7, wantTotal: 9, wantMinimum: false},
		{name: "exactly eight points is eligible", occurrences: []MCRFanOccurrence{occurrence("mcr_34", 8, nil)}, wantFans: []FanID{"mcr_34"}, wantNon: 8, wantTotal: 8, wantMinimum: true},
		{name: "nine points is eligible", occurrences: []MCRFanOccurrence{occurrence("mcr_34", 8, nil), occurrence("mcr_09", 1, nil)}, wantFans: []FanID{"mcr_34", "mcr_09"}, wantNon: 9, wantTotal: 9, wantMinimum: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := scoreMCRFanOccurrences(test.occurrences, test.flowers, 8)
			assertMCRScoreFanIDs(t, got, test.wantFans)
			if got.NonFlowerPoints != test.wantNon || got.TotalPoints != test.wantTotal || got.MeetsMinimum != test.wantMinimum {
				t.Fatalf("score = %#v", got)
			}
		})
	}
}

func TestScoreMCRChoosesAReadableWinningBreakdown(t *testing.T) {
	tiles := mustFanTiles(t, "2m", "3m", "4m", "3m", "4m", "5m", "4p", "5p", "6p", "6s", "7s", "8s", "5p", "5p")
	winningTile := tiles[len(tiles)-1]
	hand := append([]Tile(nil), tiles[:len(tiles)-1]...)
	got := ScoreMCR(hand, nil, MCRScoreContext{WinningTile: winningTile, WinType: WinSelfDraw, Flowers: 0})
	if !got.MeetsMinimum || got.NonFlowerPoints < 8 || len(got.WinningGrouping) == 0 {
		t.Fatalf("score = %#v", got)
	}
	for _, fan := range got.Fans {
		if fan.NameZH == "" || fan.NameEN == "" {
			t.Fatalf("fan is not readable: %#v", fan)
		}
	}
}

func assertMCRScoreFanIDs(t *testing.T, score MCRScoreBreakdown, want []FanID) {
	t.Helper()
	if len(score.Fans) != len(want) {
		t.Fatalf("fans = %#v, want IDs %v", score.Fans, want)
	}
	for index, id := range want {
		if score.Fans[index].ID != id {
			t.Fatalf("fan %d = %s, want %s; fans=%#v", index, score.Fans[index].ID, id, score.Fans)
		}
	}
}
