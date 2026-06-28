package game

import (
	"reflect"
	"strings"
	"testing"
)

func TestShantenStandardWinningHand(t *testing.T) {
	hand := mustTiles(t,
		"1m", "2m", "3m",
		"4m", "5m", "6m",
		"2p", "3p", "4p",
		"7s", "7s", "7s",
		"E", "E",
	)
	if got := ShantenStandard(hand); got != -1 {
		t.Fatalf("shanten = %d, want -1", got)
	}
}

func TestShantenStandardTenpaiHand(t *testing.T) {
	hand := mustTiles(t,
		"1m", "2m", "3m",
		"4m", "5m", "6m",
		"2p", "3p", "4p",
		"7s", "7s", "7s",
		"E",
	)
	if got := ShantenStandard(hand); got != 0 {
		t.Fatalf("shanten = %d, want 0", got)
	}
}

func TestWinningTilesFindsSinglePairWait(t *testing.T) {
	hand := mustTiles(t,
		"1m", "2m", "3m",
		"4m", "5m", "6m",
		"2p", "3p", "4p",
		"7s", "7s", "7s",
		"E",
	)
	waits := WinningTiles(hand)
	if FormatTiles(waits) != "E" {
		t.Fatalf("waits = %s, want E", FormatTiles(waits))
	}
}

func TestHandTipsShowsTenpaiWaits(t *testing.T) {
	tips := HandTips(mustTiles(t,
		"1m", "2m", "3m",
		"4m", "5m", "6m",
		"2p", "3p", "4p",
		"7s", "7s", "7s",
		"E",
	))
	if !strings.Contains(tips, "tenpai") || !strings.Contains(tips, "E") {
		t.Fatalf("tips = %q", tips)
	}
}

func TestWinningTilesIncludesSevenPairsWait(t *testing.T) {
	hand := mustTiles(t, "1m", "1m", "2m", "2m", "3m", "3m", "4p", "4p", "5p", "5p", "E", "E", "B")
	waits := WinningTiles(hand)
	if FormatTiles(waits) != "B" {
		t.Fatalf("waits = %s, want B", FormatTiles(waits))
	}
}

func TestHandTipsShowsSevenPairsWait(t *testing.T) {
	tips := HandTips(mustTiles(t, "1m", "1m", "2m", "2m", "3m", "3m", "4p", "4p", "5p", "5p", "E", "E", "B"))
	if !strings.Contains(tips, "tenpai") || !strings.Contains(tips, "B") {
		t.Fatalf("tips = %q", tips)
	}
}

func TestBestDiscardIndexKeepsCompleteMelds(t *testing.T) {
	hand := mustTiles(t,
		"1m", "2m", "3m",
		"4m", "5m", "6m",
		"2p", "3p", "4p",
		"7s", "7s",
		"E", "E", "9m",
	)
	index := BestDiscardIndex(hand)
	if index < 0 || hand[index] != mustTile(t, "9m") {
		t.Fatalf("discard = %d:%s, want 9m", index, hand[index])
	}
}

func TestEffectiveTilesReduceShanten(t *testing.T) {
	hand := mustTiles(t, "1m", "2m", "3m", "4m", "5m", "6m", "7p", "8p", "9p", "2s", "2s", "3s", "4s")

	got := EffectiveTiles(hand)

	if FormatTiles(got) != "2s 5s" {
		t.Fatalf("effective tiles = %s, want 2s 5s", FormatTiles(got))
	}
}

func TestImprovementTilesDoNotMutateHand(t *testing.T) {
	hand := mustTiles(t, "1m", "2m", "3m", "4m", "5m", "6m", "7p", "8p", "9p", "2s", "2s", "3s", "4s", "E")
	before := append([]Tile(nil), hand...)

	got := ImprovementTiles(hand)

	if !reflect.DeepEqual(hand, before) {
		t.Fatalf("hand mutated: before=%v after=%v", before, hand)
	}
	if len(got) == 0 {
		t.Fatal("expected discard improvements")
	}
	for _, improvement := range got {
		if len(improvement.Effective) == 0 {
			t.Fatalf("discard %s has no effective tiles", improvement.Discard)
		}
	}
}
