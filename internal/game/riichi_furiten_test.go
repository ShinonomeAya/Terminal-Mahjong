package game

import (
	"encoding/json"
	"os"
	"testing"
)

type riichiFuritenFixture struct {
	ID       string   `json:"id"`
	Expected []string `json:"expected"`
}

func TestRiichiFuritenFixtureManifest(t *testing.T) {
	file, err := os.Open("../../testdata/rules/riichi/furiten.json")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	var fixtures []riichiFuritenFixture
	if err := decoder.Decode(&fixtures); err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{
		"closed-tenpai-riichi": true, "riichi-accepts-after-safe-discard": true,
		"own-discard-furiten": true, "temporary-furiten-after-pass": true,
		"riichi-pass-furiten": true, "ippatsu-cancelled-by-call": true,
	}
	for _, fixture := range fixtures {
		if !want[fixture.ID] || len(fixture.Expected) == 0 {
			t.Fatalf("invalid riichi furiten fixture: %#v", fixture)
		}
		delete(want, fixture.ID)
	}
	if len(want) != 0 {
		t.Fatalf("missing riichi furiten fixtures: %#v", want)
	}
}

func TestRiichiDeclarationDiscardsAndAcceptsAfterNoClaim(t *testing.T) {
	round := newRiichiActionTestGame(t)
	round.Current = 0
	round.Wall = mustTiles(t, "7s", "8s", "9s", "1p")
	round.Players[0].Hand = mustTiles(t, "1m", "2m", "3m", "4m", "5m", "6m", "7p", "8p", "9p", "2s", "2s", "3s", "4s", "E")

	actions := round.rules.LegalActions(round, "0")
	if !hasLegalAction(actions, CommandRiichi, "") {
		t.Fatalf("actions = %#v, want riichi", actions)
	}
	result := round.ApplyCommand(GameCommand{PlayerID: "0", Kind: CommandRiichi, TileIndex: 13})

	if !result.OK || round.Riichi.Declarations[0] != RiichiAccepted || !round.Riichi.Ippatsu[0] || round.Riichi.RiichiSticks != 1 {
		t.Fatalf("result=%#v riichi=%#v", result, round.Riichi)
	}
	if got := FormatTiles(round.Players[0].Discards); got != "E" {
		t.Fatalf("discards = %s, want E", got)
	}
}

func TestRiichiOwnDiscardFuritenSuppressesRon(t *testing.T) {
	round := newRiichiActionTestGame(t)
	discard := mustTile(t, "5s")
	round.Players[1].Hand = mustTiles(t, "1m", "2m", "3m", "4m", "5m", "6m", "7p", "8p", "9p", "2s", "2s", "3s", "4s")
	round.Players[1].Discards = []Tile{discard}

	pending := round.buildPendingClaim(0, discard)

	if hasClaimOption(pending, ClaimWin, 1) {
		t.Fatalf("furiten player received ron claim: %#v", pending)
	}
}

func TestRiichiPassingRonCreatesTemporaryFuriten(t *testing.T) {
	round := newRiichiActionTestGame(t)
	discard := mustTile(t, "5s")
	round.Players[1].Hand = mustTiles(t, "1m", "2m", "3m", "4m", "5m", "6m", "7p", "8p", "9p", "2s", "2s", "3s", "4s")
	round.PendingClaim = &PendingClaim{Discarder: 0, Tile: discard, Options: []ClaimOption{{Kind: ClaimWin, Player: 1}}}
	round.Phase = PhaseAwaitingClaim
	round.Current = 1

	result := round.ApplyCommand(GameCommand{PlayerID: "1", Kind: CommandPass})

	if !result.OK || !round.Riichi.TemporaryFuriten[1] {
		t.Fatalf("result=%#v riichi=%#v", result, round.Riichi)
	}
	if next := round.buildPendingClaim(2, discard); hasClaimOption(next, ClaimWin, 1) {
		t.Fatalf("temporary furiten player received ron claim: %#v", next)
	}
}

func TestRiichiPassingRonAfterRiichiCreatesRiichiFuriten(t *testing.T) {
	round := newRiichiActionTestGame(t)
	discard := mustTile(t, "5s")
	round.Riichi.Declarations[1] = RiichiAccepted
	round.Players[1].Hand = mustTiles(t, "1m", "2m", "3m", "4m", "5m", "6m", "7p", "8p", "9p", "2s", "2s", "3s", "4s")
	round.PendingClaim = &PendingClaim{Discarder: 0, Tile: discard, Options: []ClaimOption{{Kind: ClaimWin, Player: 1}}}
	round.Phase = PhaseAwaitingClaim
	round.Current = 1

	result := round.ApplyCommand(GameCommand{PlayerID: "1", Kind: CommandPass})

	if !result.OK || !round.Riichi.RiichiFuriten[1] || round.Riichi.TemporaryFuriten[1] {
		t.Fatalf("result=%#v riichi=%#v", result, round.Riichi)
	}
}

func TestRiichiIppatsuCancelledByAcceptedCall(t *testing.T) {
	round := newRiichiActionTestGame(t)
	discard := mustTile(t, "5p")
	round.Riichi.Ippatsu[0] = true
	round.Players[0].Discards = []Tile{discard}
	round.Players[2].Hand = mustTiles(t, "5p", "5p", "2m")
	round.PendingClaim = &PendingClaim{Discarder: 0, Tile: discard, Options: []ClaimOption{{Kind: ClaimPong, Player: 2, Consumed: []Tile{discard, discard}}}}
	round.Phase = PhaseAwaitingClaim
	round.Current = 2

	result := round.ApplyCommand(GameCommand{PlayerID: "2", Kind: CommandPong})

	if !result.OK || round.Riichi.Ippatsu[0] {
		t.Fatalf("result=%#v riichi=%#v", result, round.Riichi)
	}
}

func hasClaimOption(pending *PendingClaim, kind ClaimKind, player int) bool {
	if pending == nil {
		return false
	}
	for _, option := range pending.Options {
		if option.Kind == kind && option.Player == player {
			return true
		}
	}
	return false
}
