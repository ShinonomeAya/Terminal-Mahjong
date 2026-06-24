package game

import (
	"encoding/json"
	"os"
	"testing"
)

type mcrWallFlowFixture struct {
	ID              string   `json:"id"`
	Source          string   `json:"source"`
	Wall            []string `json:"wall"`
	Player          int      `json:"player"`
	ExpectedTile    string   `json:"expected_tile"`
	ExpectedFlowers []string `json:"expected_flowers"`
	Exhausted       bool     `json:"exhausted"`
}

func TestMCRDealReplacesEveryFlower(t *testing.T) {
	rules := NewMCRRuleSet(DefaultRuleConfig(ModeMCR).MCR)
	round, err := NewGameWithRules(73, rules)
	if err != nil {
		t.Fatal(err)
	}
	for seat := range round.Players {
		if len(round.Players[seat].Hand) != 13 {
			t.Fatalf("seat %d hand = %d, want 13", seat, len(round.Players[seat].Hand))
		}
		if containsFlower(round.Players[seat].Hand) {
			t.Fatalf("seat %d kept a flower: %v", seat, round.Players[seat].Hand)
		}
	}
	assertMCRTileConservation(t, round)
}

func TestMCRFlowerDrawFixtures(t *testing.T) {
	for _, fixture := range loadMCRWallFlowFixtures(t) {
		t.Run(fixture.ID, func(t *testing.T) {
			rules := NewMCRRuleSet(DefaultRuleConfig(ModeMCR).MCR)
			round := &Game{Players: NewPlayers(), Wall: parseFixtureTiles(t, fixture.Wall), Mode: ModeMCR, rules: rules, Phase: PhaseAwaitingDiscard}

			tile, ok := rules.Draw(round, fixture.Player, DrawNormal)

			if ok == fixture.Exhausted {
				t.Fatalf("draw ok = %v, exhausted = %v", ok, fixture.Exhausted)
			}
			if !fixture.Exhausted {
				want, parsed := ParseTile(fixture.ExpectedTile)
				if !parsed || tile != want {
					t.Fatalf("tile = %s, want %s", tile, fixture.ExpectedTile)
				}
			}
			if got := tileStrings(round.Players[fixture.Player].Flowers); !equalStrings(got, fixture.ExpectedFlowers) {
				t.Fatalf("flowers = %v, want %v", got, fixture.ExpectedFlowers)
			}
			if fixture.Exhausted && (!round.Over || round.Reason == "") {
				t.Fatalf("exhausted round = over:%v reason:%q", round.Over, round.Reason)
			}
		})
	}
}

func TestMCRReplacementDrawUsesBackOfWall(t *testing.T) {
	rules := NewMCRRuleSet(DefaultRuleConfig(ModeMCR).MCR)
	round := &Game{Players: NewPlayers(), Wall: mustTiles(t, "1m", "9p"), Mode: ModeMCR, rules: rules, Phase: PhaseAwaitingDiscard}

	tile, ok := rules.Draw(round, 0, DrawReplacement)

	if !ok || tile != mustTile(t, "9p") || len(round.Wall) != 1 || round.Wall[0] != mustTile(t, "1m") {
		t.Fatalf("replacement tile=%s ok=%v wall=%v", tile, ok, round.Wall)
	}
}

func TestMCRPrivateSnapshotRedactsReplacementDrawButShowsFlowers(t *testing.T) {
	rules := NewMCRRuleSet(DefaultRuleConfig(ModeMCR).MCR)
	round := &Game{Players: NewPlayers(), Wall: mustTiles(t, "P1", "2m"), Mode: ModeMCR, rules: rules, Phase: PhaseAwaitingDiscard, Current: 1}
	if _, ok := rules.Draw(round, 1, DrawNormal); !ok {
		t.Fatal("replacement draw failed")
	}

	snapshot := round.SnapshotFor("0")

	if len(snapshot.Players[1].Flowers) != 1 || snapshot.Players[1].Flowers[0] != FlowerPlum {
		t.Fatalf("public flowers = %v", snapshot.Players[1].Flowers)
	}
	for _, event := range snapshot.Events {
		if event.Kind == EventReplacementDraw && event.Tile != -1 {
			t.Fatalf("replacement draw leaked: %#v", event)
		}
	}
}

func loadMCRWallFlowFixtures(t *testing.T) []mcrWallFlowFixture {
	t.Helper()
	data, err := os.ReadFile("../../testdata/rules/mcr/wall_flow.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures []mcrWallFlowFixture
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatal(err)
	}
	return fixtures
}

func parseFixtureTiles(t *testing.T, values []string) []Tile {
	t.Helper()
	tiles := make([]Tile, len(values))
	for i, value := range values {
		tile, ok := ParseTile(value)
		if !ok {
			t.Fatalf("invalid fixture tile %q", value)
		}
		tiles[i] = tile
	}
	return tiles
}

func containsFlower(tiles []Tile) bool {
	for _, tile := range tiles {
		if tile.IsFlower() {
			return true
		}
	}
	return false
}

func assertMCRTileConservation(t *testing.T, round *Game) {
	t.Helper()
	total := len(round.Wall)
	for _, player := range round.Players {
		total += len(player.Hand) + len(player.Flowers) + len(player.Discards)
		for _, meld := range player.Melds {
			total += len(meld.Tiles)
		}
	}
	if total != 144 {
		t.Fatalf("tile total = %d, want 144", total)
	}
}

func tileStrings(tiles []Tile) []string {
	values := make([]string, len(tiles))
	for i, tile := range tiles {
		values[i] = tile.String()
	}
	return values
}

func equalStrings(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
