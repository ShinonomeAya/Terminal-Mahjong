package game

import (
	"encoding/json"
	"os"
	"testing"
)

type riichiScoringFixture struct {
	ID     string `json:"id"`
	Source string `json:"source"`
}

func TestRiichiScoringFixtureManifest(t *testing.T) {
	file, err := os.Open("../../testdata/rules/riichi/scoring.json")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	var fixtures []riichiScoringFixture
	if err := decoder.Decode(&fixtures); err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{
		"pinfu-tsumo": true, "seven-pairs": true, "red-and-normal-dora": true,
		"dora-only-rejected": true, "haneman-limit": true,
	}
	for _, fixture := range fixtures {
		if !want[fixture.ID] || fixture.Source == "" {
			t.Fatalf("invalid riichi scoring fixture: %#v", fixture)
		}
		delete(want, fixture.ID)
	}
	if len(want) != 0 {
		t.Fatalf("missing riichi scoring fixtures: %#v", want)
	}
}

func TestRiichiScoringGoldenFixtures(t *testing.T) {
	tests := []struct {
		name     string
		hand     []Tile
		melds    []Meld
		context  RiichiYakuContext
		want     RiichiScoreBreakdown
		wantYaku []string
	}{
		{
			name:     "pinfu tsumo",
			hand:     mustTiles(t, "2m", "3m", "1p", "2p", "3p", "4p", "5p", "6p", "7s", "8s", "9s", "5s", "5s"),
			context:  RiichiYakuContext{WinningTile: mustTile(t, "4m"), WinType: WinSelfDraw, Closed: true, SeatWind: mustTile(t, "E"), PrevalentWind: mustTile(t, "S")},
			want:     RiichiScoreBreakdown{Fu: 20, YakuHan: 2, BonusHan: 0, Yakuman: 0, BasePoints: 320, HasYaku: true},
			wantYaku: []string{"menzen_tsumo", "pinfu"},
		},
		{
			name:     "seven pairs",
			hand:     mustTiles(t, "1m", "1m", "2m", "2m", "3p", "3p", "4p", "4p", "5s", "5s", "E", "E", "Z"),
			context:  RiichiYakuContext{WinningTile: mustTile(t, "Z"), WinType: WinDiscard, Closed: true, SeatWind: mustTile(t, "E"), PrevalentWind: mustTile(t, "S")},
			want:     RiichiScoreBreakdown{Fu: 25, YakuHan: 2, BonusHan: 0, Yakuman: 0, BasePoints: 400, HasYaku: true},
			wantYaku: []string{"chiitoitsu"},
		},
		{
			name:     "red and normal dora",
			hand:     mustTiles(t, "3m", "4m", "0m", "3p", "4p", "5p", "4s", "5s", "6s", "6p", "7p", "8p", "5s"),
			context:  RiichiYakuContext{WinningTile: mustTile(t, "5s"), WinType: WinDiscard, Closed: false, SeatWind: mustTile(t, "E"), PrevalentWind: mustTile(t, "S"), DoraIndicators: []Tile{mustTile(t, "4m")}},
			want:     RiichiScoreBreakdown{Fu: 30, YakuHan: 1, BonusHan: 2, Yakuman: 0, BasePoints: 960, HasYaku: true},
			wantYaku: []string{"tanyao"},
		},
		{
			name:    "dora only rejected",
			hand:    mustTiles(t, "1m", "2m", "3m", "4p", "5p", "6p", "7s", "8s", "9s", "E", "E", "2m", "2m"),
			context: RiichiYakuContext{WinningTile: mustTile(t, "E"), WinType: WinDiscard, Closed: true, SeatWind: mustTile(t, "S"), PrevalentWind: mustTile(t, "W"), DoraIndicators: []Tile{mustTile(t, "N")}},
			want:    RiichiScoreBreakdown{Fu: 0, YakuHan: 0, BonusHan: 3, Yakuman: 0, BasePoints: 0, HasYaku: false},
		},
		{
			name:     "haneman limit",
			hand:     mustTiles(t, "1m", "1m", "1m", "2m", "3m", "4m", "4m", "5m", "6m", "7m", "8m", "9m", "9m"),
			context:  RiichiYakuContext{WinningTile: mustTile(t, "6m"), WinType: WinDiscard, Closed: true, SeatWind: mustTile(t, "E"), PrevalentWind: mustTile(t, "S")},
			want:     RiichiScoreBreakdown{Fu: 40, YakuHan: 6, BonusHan: 0, Yakuman: 0, BasePoints: 3000, LimitName: "haneman", HasYaku: true},
			wantYaku: []string{"chinitsu"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := ScoreRiichi(test.hand, test.melds, test.context)
			if got.Fu != test.want.Fu || got.YakuHan != test.want.YakuHan || got.BonusHan != test.want.BonusHan || got.Yakuman != test.want.Yakuman || got.BasePoints != test.want.BasePoints || got.LimitName != test.want.LimitName || got.HasYaku != test.want.HasYaku {
				t.Fatalf("score = %#v, want %#v", got, test.want)
			}
			for _, id := range test.wantYaku {
				if !hasRiichiYakuID(got.Yaku, id) {
					t.Fatalf("score yaku = %#v, want %s", got.Yaku, id)
				}
			}
		})
	}
}
