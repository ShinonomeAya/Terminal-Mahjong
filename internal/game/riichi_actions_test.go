package game

import (
	"encoding/json"
	"os"
	"testing"
)

type riichiLegalActionFixture struct {
	ID       string   `json:"id"`
	Expected []string `json:"expected"`
}

func TestRiichiLegalActionFixtureManifest(t *testing.T) {
	file, err := os.Open("../../testdata/rules/riichi/actions.json")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	var fixtures []riichiLegalActionFixture
	if err := decoder.Decode(&fixtures); err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{
		"ron-priority": true, "daiminkan-before-chi": true, "next-seat-chi": true,
		"concealed-kong": true, "added-kong-chankan": true,
	}
	for _, fixture := range fixtures {
		if !want[fixture.ID] || len(fixture.Expected) == 0 {
			t.Fatalf("invalid riichi action fixture: %#v", fixture)
		}
		delete(want, fixture.ID)
	}
	if len(want) != 0 {
		t.Fatalf("missing riichi action fixtures: %#v", want)
	}
}

func TestRiichiClaimPriorityIsRonThenKanOrPonThenChi(t *testing.T) {
	round := newRiichiActionTestGame(t)
	discard := mustTile(t, "5p")
	round.Players[1].Hand = mustTiles(t, "3p", "4p")
	round.Players[2].Hand = mustTiles(t, "5p", "5p", "5p")
	round.Players[3].Hand = mustTiles(t, "1m", "2m", "3m", "2m", "3m", "4m", "4s", "5s", "6s", "7p", "8p", "9p", "5p")

	pending := round.buildPendingClaim(0, discard)

	if pending == nil {
		t.Fatal("expected riichi claim options")
	}
	want := []ClaimOption{
		{Kind: ClaimWin, Player: 3},
		{Kind: ClaimKong, Player: 2, Consumed: []Tile{discard, discard, discard}},
		{Kind: ClaimPong, Player: 2, Consumed: []Tile{discard, discard}},
		{Kind: ClaimChow, Player: 1, Consumed: mustTiles(t, "3p", "4p")},
	}
	assertClaimOptions(t, pending.Options, want)
}

func TestRiichiChowIsOnlyAvailableToNextSeat(t *testing.T) {
	round := newRiichiActionTestGame(t)
	discard := mustTile(t, "5p")
	round.Players[1].Hand = mustTiles(t, "3p", "4p")
	round.Players[2].Hand = mustTiles(t, "3p", "4p")

	pending := round.buildPendingClaim(0, discard)

	if pending == nil {
		t.Fatal("expected next-seat chow")
	}
	for _, option := range pending.Options {
		if option.Kind == ClaimChow && option.Player != 1 {
			t.Fatalf("non-next seat received chow: %#v", pending.Options)
		}
	}
}

func TestRiichiConcealedKongDrawsRinshanAndRevealsKanDora(t *testing.T) {
	round := newRiichiActionTestGame(t)
	tile := mustTile(t, "1m")
	replacement := round.Riichi.DeadWall[0]
	round.Current = 0
	round.Players[0].Hand = mustTiles(t, "1m", "1m", "1m", "1m", "2m")

	result := round.ApplyCommand(GameCommand{PlayerID: "0", Kind: CommandKong, Tile: "1m"})

	if !result.OK || round.PendingClaim != nil || !hasMeld(round.Players[0].Melds, MeldKong, tile, 4) {
		t.Fatalf("result=%#v player=%#v pending=%#v", result, round.Players[0], round.PendingClaim)
	}
	if round.Players[0].Count(replacement) != 1 || len(round.Riichi.DoraIndicators) != 2 || round.Riichi.RinshanDraws != 1 {
		t.Fatalf("riichi kan state: hand=%s riichi=%#v", FormatTiles(round.Players[0].Hand), round.Riichi)
	}
}

func TestRiichiAddedKongOpensChankanWindowBeforeMutation(t *testing.T) {
	round := riichiAddedKongTestGame(t)
	tile := mustTile(t, "5m")

	result := round.ApplyCommand(GameCommand{PlayerID: "0", Kind: CommandKong, Tile: "5m"})

	if !result.OK || round.PendingClaim == nil || !round.PendingClaim.RobbingKong || round.Phase != PhaseAwaitingClaim || round.Current != 1 {
		t.Fatalf("result=%#v pending=%#v", result, round.PendingClaim)
	}
	if round.Players[0].Count(tile) != 1 || !hasMeld(round.Players[0].Melds, MeldPong, tile, 3) {
		t.Fatalf("added kong mutated before chankan window: %#v", round.Players[0])
	}
}

func TestRiichiPassedChankanWindowCompletesAddedKong(t *testing.T) {
	round := riichiAddedKongTestGame(t)
	tile := mustTile(t, "5m")
	replacement := round.Riichi.DeadWall[0]
	if result := round.ApplyCommand(GameCommand{PlayerID: "0", Kind: CommandKong, Tile: "5m"}); !result.OK {
		t.Fatal(result.Error)
	}

	result := round.ApplyCommand(GameCommand{PlayerID: "1", Kind: CommandPass})

	if !result.OK || round.PendingClaim != nil || round.Current != 0 || round.Phase != PhaseAwaitingDiscard {
		t.Fatalf("result=%#v pending=%#v", result, round.PendingClaim)
	}
	if round.Players[0].Count(tile) != 0 || !hasMeld(round.Players[0].Melds, MeldKong, tile, 4) || round.Players[0].Count(replacement) != 1 {
		t.Fatalf("completed added kong state: %#v", round.Players[0])
	}
}

func assertClaimOptions(t *testing.T, got []ClaimOption, want []ClaimOption) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("options = %#v, want %#v", got, want)
	}
	for index := range want {
		if got[index].Kind != want[index].Kind || got[index].Player != want[index].Player || FormatTiles(got[index].Consumed) != FormatTiles(want[index].Consumed) {
			t.Fatalf("option %d = %#v, want %#v", index, got[index], want[index])
		}
	}
}

func hasMeld(melds []Meld, kind MeldKind, tile Tile, count int) bool {
	for _, meld := range melds {
		if meld.Kind == kind && len(meld.Tiles) == count && meld.Tiles[0].Base() == tile.Base() {
			return true
		}
	}
	return false
}

func riichiAddedKongTestGame(t *testing.T) *Game {
	t.Helper()
	round := newRiichiActionTestGame(t)
	tile := mustTile(t, "5m")
	round.Current = 0
	round.Players[0].Melds = []Meld{{Kind: MeldPong, Tiles: []Tile{tile, tile, tile}}}
	round.Players[0].Hand = mustTiles(t, "5m", "2p")
	round.Players[1].Hand = mustTiles(t, "1p", "2p", "3p", "2p", "3p", "4p", "4s", "5s", "6s", "7m", "8m", "9m", "5m")
	round.Players[2].Hand = mustTiles(t, "1m")
	round.Players[3].Hand = mustTiles(t, "2m")
	return round
}

func newRiichiActionTestGame(t *testing.T) *Game {
	t.Helper()
	rules := NewRiichiRuleSet(DefaultRuleConfig(ModeRiichi).Riichi)
	dead := mustTiles(t, "1p", "2p", "3p", "4p", "5p", "6p", "7p", "8p", "9p", "1s", "2s", "3s", "4s", "5s")
	return &Game{
		Players:    NewPlayers(),
		Wall:       mustTiles(t, "7s", "8s", "9s"),
		Mode:       ModeRiichi,
		RuleConfig: rules.Config(),
		Winner:     -1,
		Phase:      PhaseAwaitingDiscard,
		Riichi: &RiichiRoundState{
			DeadWall:       dead,
			DoraIndicators: []Tile{dead[4]},
			UraIndicators:  []Tile{dead[5]},
		},
		rules: rules,
	}
}
