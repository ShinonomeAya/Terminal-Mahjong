package game

import (
	"encoding/json"
	"os"
	"testing"
)

type riichiWallFixture struct {
	ID             string   `json:"id"`
	Source         string   `json:"source"`
	DeadWall       []string `json:"dead_wall"`
	Rinshan        []string `json:"rinshan"`
	DoraIndicators []string `json:"dora_indicators"`
	UraIndicators  []string `json:"ura_indicators"`
}

func TestRiichiDeadWallGoldenLayout(t *testing.T) {
	file, err := os.Open("../../testdata/rules/riichi/wall.json")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	var fixtures []riichiWallFixture
	if err := decoder.Decode(&fixtures); err != nil {
		t.Fatal(err)
	}
	for _, fixture := range fixtures {
		dead := parseFixtureTiles(t, fixture.DeadWall)
		if fixture.Source == "" || len(dead) != 14 {
			t.Fatalf("invalid fixture: %#v", fixture)
		}
		if got := FormatTiles(dead[:4]); got != FormatTiles(parseFixtureTiles(t, fixture.Rinshan)) {
			t.Fatalf("%s rinshan=%s", fixture.ID, got)
		}
		var dora, ura []Tile
		for index := 4; index < 14; index += 2 {
			dora = append(dora, dead[index])
			ura = append(ura, dead[index+1])
		}
		if FormatTiles(dora) != FormatTiles(parseFixtureTiles(t, fixture.DoraIndicators)) || FormatTiles(ura) != FormatTiles(parseFixtureTiles(t, fixture.UraIndicators)) {
			t.Fatalf("%s dora=%v ura=%v", fixture.ID, dora, ura)
		}
	}
}

func TestBuildRiichiWallSupportsConfiguredRedFives(t *testing.T) {
	for _, redFives := range []int{0, 3} {
		wall := BuildRiichiWall(redFives)
		if len(wall) != 136 {
			t.Fatalf("red=%d wall=%d", redFives, len(wall))
		}
		counts := TileCounts(wall)
		for tile, count := range counts {
			if count != 4 {
				t.Fatalf("red=%d tile=%d count=%d", redFives, tile, count)
			}
		}
		reds := 0
		for _, tile := range wall {
			if tile.IsRed() {
				reds++
			}
		}
		if reds != redFives {
			t.Fatalf("red tiles = %d, want %d", reds, redFives)
		}
	}
}

func TestRiichiDealCreatesDeadWallAndDealerHand(t *testing.T) {
	rules := NewRiichiRuleSet(DefaultRuleConfig(ModeRiichi).Riichi)
	round, err := NewGameWithRules(83, rules)
	if err != nil {
		t.Fatal(err)
	}
	if round.Riichi == nil || len(round.Riichi.DeadWall) != 14 || len(round.Riichi.DoraIndicators) != 1 || len(round.Riichi.UraIndicators) != 1 {
		t.Fatalf("riichi state = %#v", round.Riichi)
	}
	if len(round.Wall) != 69 || len(round.Players[0].Hand) != 14 {
		t.Fatalf("live wall=%d dealer hand=%d", len(round.Wall), len(round.Players[0].Hand))
	}
	for player := 1; player < 4; player++ {
		if len(round.Players[player].Hand) != 13 {
			t.Fatalf("player %d hand=%d", player, len(round.Players[player].Hand))
		}
	}
	assertRiichiTileConservation(t, round)
}

func TestRiichiRinshanDrawReplenishesDeadWallFromLiveWall(t *testing.T) {
	rules := NewRiichiRuleSet(DefaultRuleConfig(ModeRiichi).Riichi)
	dead := mustTiles(t, "9p", "8p", "7p", "6p", "1m", "2m", "3m", "4m", "5m", "6m", "7m", "8m", "9m", "E")
	round := &Game{
		Players: NewPlayers(),
		Wall:    mustTiles(t, "1s", "2s"),
		Mode:    ModeRiichi,
		Riichi: &RiichiRoundState{
			DeadWall:       append([]Tile(nil), dead...),
			DoraIndicators: []Tile{dead[4]},
			UraIndicators:  []Tile{dead[5]},
		},
		rules: rules,
	}

	tile, ok := rules.Draw(round, 0, DrawReplacement)

	if !ok || tile != mustTile(t, "9p") || round.Riichi.RinshanDraws != 1 || len(round.Riichi.DeadWall) != 14 || round.Riichi.DeadWall[0] != mustTile(t, "2s") || len(round.Wall) != 1 {
		t.Fatalf("tile=%s ok=%t wall=%v state=%#v", tile, ok, round.Wall, round.Riichi)
	}
}

func TestRiichiKanRevealsNextDoraAndStopsAfterFour(t *testing.T) {
	rules := NewRiichiRuleSet(DefaultRuleConfig(ModeRiichi).Riichi)
	dead := mustTiles(t, "1m", "2m", "3m", "4m", "5m", "6m", "7m", "8m", "9m", "1p", "2p", "3p", "4p", "5p")
	round := &Game{Players: NewPlayers(), Mode: ModeRiichi, Riichi: &RiichiRoundState{DeadWall: dead, DoraIndicators: []Tile{dead[4]}, UraIndicators: []Tile{dead[5]}}, rules: rules}

	for kan := 1; kan <= 4; kan++ {
		if !revealRiichiKanDora(round) {
			t.Fatalf("kan %d reveal failed", kan)
		}
		if len(round.Riichi.DoraIndicators) != kan+1 || round.Riichi.DoraIndicators[kan] != dead[4+kan*2] {
			t.Fatalf("kan %d indicators=%v", kan, round.Riichi.DoraIndicators)
		}
	}
	if revealRiichiKanDora(round) {
		t.Fatal("fifth kan dora reveal should fail")
	}
}

func TestRiichiDoraIndicatorWraps(t *testing.T) {
	tests := map[string]string{
		"9m": "1m",
		"N":  "E",
		"Z":  "B",
		"B":  "F",
		"F":  "Z",
		"0p": "6p",
	}
	for indicatorText, wantText := range tests {
		indicator := mustTile(t, indicatorText)
		want := mustTile(t, wantText)
		if got := RiichiDoraTile(indicator); got != want {
			t.Fatalf("indicator %s dora=%s, want %s", indicator, got, want)
		}
	}
}

func assertRiichiTileConservation(t *testing.T, round *Game) {
	t.Helper()
	total := len(round.Wall) + len(round.Riichi.DeadWall)
	for _, player := range round.Players {
		total += len(player.Hand) + len(player.Discards)
		for _, meld := range player.Melds {
			total += len(meld.Tiles)
		}
	}
	if total != 136 {
		t.Fatalf("tile total = %d, want 136", total)
	}
}
