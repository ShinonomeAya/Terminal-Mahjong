package game

import (
	"encoding/json"
	"os"
	"testing"
)

type riichiDecompositionFixture struct {
	ID            string   `json:"id"`
	Source        string   `json:"source"`
	Hand          []string `json:"hand"`
	WinningTile   string   `json:"winning_tile"`
	ExpectedKind  string   `json:"expected_kind"`
	ExpectedWait  string   `json:"expected_wait"`
	ExpectedWaits []string `json:"expected_waits,omitempty"`
}

func TestRiichiDecomposeGoldenFixtures(t *testing.T) {
	for _, fixture := range loadRiichiDecompositionFixtures(t) {
		t.Run(fixture.ID, func(t *testing.T) {
			hand := parseFixtureTiles(t, fixture.Hand)
			winning := mustTile(t, fixture.WinningTile)

			got := RiichiDecompose(hand, nil, winning)
			if !hasRiichiShape(got, RiichiShapeKind(fixture.ExpectedKind)) {
				t.Fatalf("decompositions = %#v, want kind %s", got, fixture.ExpectedKind)
			}
			if fixture.ExpectedWait != "" && !hasRiichiWaitKind(got, RiichiWaitKind(fixture.ExpectedWait)) {
				t.Fatalf("decompositions = %#v, want wait %s", got, fixture.ExpectedWait)
			}
			if len(fixture.ExpectedWaits) > 0 {
				wantWaits := parseFixtureTiles(t, fixture.ExpectedWaits)
				if gotWaits := FormatTiles(RiichiWaits(hand, nil)); gotWaits != FormatTiles(wantWaits) {
					t.Fatalf("waits = %s, want %s", gotWaits, FormatTiles(wantWaits))
				}
			}
		})
	}
}

func TestRiichiWaitsIncludeEveryCompletingBaseTile(t *testing.T) {
	hand := mustTiles(t, "1m", "2m", "3m", "4m", "5m", "6m", "7p", "8p", "9p", "2s", "2s", "3s", "4s")
	waits := RiichiWaits(hand, nil)
	if got := FormatTiles(waits); got != "2s 5s" {
		t.Fatalf("waits = %s, want 2s 5s", got)
	}
	if !RiichiTenpai(hand, nil) {
		t.Fatal("hand should be tenpai")
	}
}

func TestRiichiDecomposeClassifiesWaits(t *testing.T) {
	tests := []struct {
		name    string
		hand    []Tile
		winning Tile
		wait    RiichiWaitKind
	}{
		{name: "ryanmen", hand: mustTiles(t, "1m", "2m", "3m", "4m", "5m", "6m", "7p", "8p", "9p", "2s", "2s", "3s", "4s"), winning: mustTile(t, "5s"), wait: RiichiWaitRyanmen},
		{name: "kanchan", hand: mustTiles(t, "1m", "2m", "3m", "4m", "5m", "6m", "7p", "8p", "9p", "2s", "2s", "3s", "5s"), winning: mustTile(t, "4s"), wait: RiichiWaitKanchan},
		{name: "penchan", hand: mustTiles(t, "1m", "2m", "3m", "4m", "5m", "6m", "7p", "8p", "9p", "2s", "2s", "8s", "9s"), winning: mustTile(t, "7s"), wait: RiichiWaitPenchan},
		{name: "tanki", hand: mustTiles(t, "1m", "2m", "3m", "4m", "5m", "6m", "7p", "8p", "9p", "2s", "3s", "4s", "E"), winning: mustTile(t, "E"), wait: RiichiWaitTanki},
		{name: "shanpon", hand: mustTiles(t, "1m", "2m", "3m", "4m", "5m", "6m", "7p", "8p", "9p", "2s", "2s", "E", "E"), winning: mustTile(t, "2s"), wait: RiichiWaitShanpon},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			decompositions := RiichiDecompose(test.hand, nil, test.winning)
			if !hasRiichiWaitKind(decompositions, test.wait) {
				t.Fatalf("decompositions = %#v, want wait %s", decompositions, test.wait)
			}
		})
	}
}

func TestRiichiDecomposeSupportsOnlyRiichiSpecialHands(t *testing.T) {
	sevenPairs := mustTiles(t, "1m", "1m", "2m", "2m", "3p", "3p", "4p", "4p", "5s", "5s", "E", "E", "Z")
	if values := RiichiDecompose(sevenPairs, nil, mustTile(t, "Z")); !hasRiichiShape(values, RiichiShapeSevenPairs) {
		t.Fatalf("seven pairs = %#v", values)
	}
	kokushi := mustTiles(t, "1m", "9m", "1p", "9p", "1s", "9s", "E", "S", "W", "N", "Z", "F", "B")
	if values := RiichiDecompose(kokushi, nil, mustTile(t, "E")); !hasRiichiShape(values, RiichiShapeThirteenOrphans) {
		t.Fatalf("kokushi = %#v", values)
	}
	invalidPairs := mustTiles(t, "1m", "1m", "1m", "1m", "2m", "2m", "3p", "3p", "4p", "4p", "5s", "5s", "E")
	if values := RiichiDecompose(invalidPairs, nil, mustTile(t, "E")); hasRiichiShape(values, RiichiShapeSevenPairs) {
		t.Fatalf("quad counted as two pairs: %#v", values)
	}
	knitted := mustTiles(t, "1m", "4m", "7m", "2p", "5p", "8p", "3s", "6s", "9s", "E", "S", "W", "N")
	if values := RiichiDecompose(knitted, nil, mustTile(t, "Z")); len(values) != 0 {
		t.Fatalf("MCR knitted hand leaked into riichi: %#v", values)
	}
}

func TestRiichiDecomposeNormalizesRedFiveAndOpenMelds(t *testing.T) {
	melds := []Meld{{Kind: MeldChow, Tiles: mustTiles(t, "4m", "0m", "6m")}}
	hand := mustTiles(t, "1p", "2p", "3p", "7p", "8p", "9p", "2s", "3s", "4s", "E")
	values := RiichiDecompose(hand, melds, mustTile(t, "E"))
	if len(values) == 0 || !containsRedRiichiTile(values) {
		t.Fatalf("red/open decomposition = %#v", values)
	}
}

func hasRiichiWaitKind(values []RiichiDecomposition, wait RiichiWaitKind) bool {
	for _, value := range values {
		if value.Wait == wait {
			return true
		}
	}
	return false
}

func hasRiichiShape(values []RiichiDecomposition, kind RiichiShapeKind) bool {
	for _, value := range values {
		if value.Kind == kind {
			return true
		}
	}
	return false
}

func containsRedRiichiTile(values []RiichiDecomposition) bool {
	for _, value := range values {
		for _, group := range value.Groups {
			for _, tile := range group.Tiles {
				if tile.IsRed() {
					return true
				}
			}
		}
	}
	return false
}

func loadRiichiDecompositionFixtures(t *testing.T) []riichiDecompositionFixture {
	t.Helper()
	data, err := os.ReadFile("../../testdata/rules/riichi/decompositions.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures []riichiDecompositionFixture
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatal(err)
	}
	return fixtures
}
