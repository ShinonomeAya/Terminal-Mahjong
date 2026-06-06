package game

import (
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
