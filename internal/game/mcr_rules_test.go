package game

import (
	"encoding/json"
	"os"
	"testing"
)

type mcrLegalActionFixture struct {
	ID       string   `json:"id"`
	Expected []string `json:"expected"`
}

func TestMCRLegalActionFixtureManifest(t *testing.T) {
	file, err := os.Open("../../testdata/rules/mcr/legal_actions.json")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	var fixtures []mcrLegalActionFixture
	if err := decoder.Decode(&fixtures); err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{
		"minimum-win": true, "win-priority": true, "exposed-kong": true,
		"pong-before-chow": true, "next-seat-chow": true, "concealed-kong": true,
		"added-kong": true, "robbing-kong": true, "all-pass": true,
	}
	for _, fixture := range fixtures {
		if !want[fixture.ID] || len(fixture.Expected) == 0 {
			t.Fatalf("invalid legal-action fixture: %#v", fixture)
		}
		delete(want, fixture.ID)
	}
	if len(want) != 0 {
		t.Fatalf("missing legal-action fixtures: %#v", want)
	}
}

func TestMCRLegalActionsExposeOnlyEligibleWin(t *testing.T) {
	round := newMCRActionTestGame()
	round.Current = 0
	round.Players[0].Hand = mustFanTiles(t, "2m", "3m", "4m", "3m", "4m", "5m", "4p", "5p", "6p", "6s", "7s", "8s", "5p", "5p")
	round.RecordEvent(EventDraw, 0, mustFanTiles(t, "5p")[0], "")

	actions := round.rules.LegalActions(round, "0")

	if !hasLegalAction(actions, CommandWin, "") {
		t.Fatalf("eligible self draw actions = %#v, want win", actions)
	}
	if got := round.rules.LegalActions(round, "1"); len(got) != 0 {
		t.Fatalf("non-current player actions = %#v, want none", got)
	}
}

func TestMCRDiscardWinBelowEightPointsIsNotAClaim(t *testing.T) {
	round := newMCRActionTestGame()
	round.Players[1].Melds = []Meld{
		{Kind: MeldChow, Tiles: mustFanTiles(t, "1m", "2m", "3m")},
		{Kind: MeldChow, Tiles: mustFanTiles(t, "2p", "3p", "4p")},
		{Kind: MeldPong, Tiles: mustFanTiles(t, "W", "W", "W")},
	}
	round.Players[1].Hand = mustFanTiles(t, "6s", "7s", "Z", "Z")
	discard := mustFanTiles(t, "8s")[0]

	pending := round.buildPendingClaim(0, discard)

	if pending != nil {
		for _, option := range pending.Options {
			if option.Kind == ClaimWin && option.Player == 1 {
				t.Fatalf("seven-point hand received win claim: %#v", pending)
			}
		}
	}
}

func TestMCRClaimPriorityIsWinThenKongOrPongThenChow(t *testing.T) {
	round := newMCRActionTestGame()
	discard := mustFanTiles(t, "5p")[0]
	round.Players[1].Hand = mustFanTiles(t, "3p", "4p")
	round.Players[2].Hand = mustFanTiles(t, "5p", "5p", "5p")
	round.Players[3].Hand = mustFanTiles(t, "1m", "2m", "3m", "2m", "3m", "4m", "4s", "5s", "6s", "7p", "8p", "9p", "5p")

	pending := round.buildPendingClaim(0, discard)

	if pending == nil {
		t.Fatal("expected claim options")
	}
	want := []ClaimOption{
		{Kind: ClaimWin, Player: 3},
		{Kind: ClaimKong, Player: 2, Consumed: []Tile{discard, discard, discard}},
		{Kind: ClaimPong, Player: 2, Consumed: []Tile{discard, discard}},
		{Kind: ClaimChow, Player: 1, Consumed: mustFanTiles(t, "3p", "4p")},
	}
	if len(pending.Options) != len(want) {
		t.Fatalf("options = %#v, want %#v", pending.Options, want)
	}
	for index := range want {
		got := pending.Options[index]
		if got.Kind != want[index].Kind || got.Player != want[index].Player || FormatTiles(got.Consumed) != FormatTiles(want[index].Consumed) {
			t.Fatalf("option %d = %#v, want %#v", index, got, want[index])
		}
	}
}

func TestMCRExposedKongClaimConsumesDiscardAndDrawsReplacement(t *testing.T) {
	round := newMCRActionTestGame()
	tile := mustFanTiles(t, "5p")[0]
	replacement := mustFanTiles(t, "9s")[0]
	round.Wall = []Tile{replacement}
	round.Players[0].Discards = []Tile{tile}
	round.Players[2].Hand = mustFanTiles(t, "5p", "5p", "5p", "2m")
	round.PendingClaim = &PendingClaim{Discarder: 0, Tile: tile, Options: []ClaimOption{{Kind: ClaimKong, Player: 2, Consumed: []Tile{tile, tile, tile}}}}
	round.Phase = PhaseAwaitingClaim
	round.Current = 2

	result := round.ApplyCommand(GameCommand{PlayerID: "2", Kind: CommandKong})

	if !result.OK || round.PendingClaim != nil || round.Current != 2 || round.Phase != PhaseAwaitingDiscard {
		t.Fatalf("result=%#v round=%#v", result, round)
	}
	if len(round.Players[0].Discards) != 0 || !hasMCRMeld(round.Players[2].Melds, MeldKong, tile, 4) || round.Players[2].Count(replacement) != 1 {
		t.Fatalf("exposed kong state = %#v", round.Players)
	}
}

func TestMCRAddedKongOpensRobbingWindowBeforeMutation(t *testing.T) {
	round := mcrAddedKongTestGame(t)
	tile := mustFanTiles(t, "5m")[0]

	result := round.ApplyCommand(GameCommand{PlayerID: "0", Kind: CommandKong, Tile: "5m"})

	if !result.OK || round.PendingClaim == nil || !round.PendingClaim.RobbingKong || round.Phase != PhaseAwaitingClaim || round.Current != 1 {
		t.Fatalf("result=%#v pending=%#v", result, round.PendingClaim)
	}
	if round.Players[0].Count(tile) != 1 || !hasMCRMeld(round.Players[0].Melds, MeldPong, tile, 3) {
		t.Fatalf("added kong mutated before robbing window: %#v", round.Players[0])
	}
	snapshot := round.SnapshotFor("1")
	if snapshot.PendingClaim == nil || !snapshot.PendingClaim.RobbingKong {
		t.Fatalf("robbing window missing from snapshot: %#v", snapshot.PendingClaim)
	}
}

func TestMCRRobbingKongWinLeavesOriginalPong(t *testing.T) {
	round := mcrAddedKongTestGame(t)
	tile := mustFanTiles(t, "5m")[0]
	if result := round.ApplyCommand(GameCommand{PlayerID: "0", Kind: CommandKong, Tile: "5m"}); !result.OK {
		t.Fatal(result.Error)
	}

	result := round.ApplyCommand(GameCommand{PlayerID: "1", Kind: CommandClaimWin})

	if !result.OK || !round.Over || round.Winner != 1 || round.Reason != "robbing-kong" {
		t.Fatalf("result=%#v winner=%d reason=%q", result, round.Winner, round.Reason)
	}
	if round.MCRScore == nil || !mcrScoreHasFan(*round.MCRScore, "mcr_42") || round.Discarder != 0 {
		t.Fatalf("robbing score=%#v discarder=%d", round.MCRScore, round.Discarder)
	}
	if round.Players[0].Count(tile) != 1 || !hasMCRMeld(round.Players[0].Melds, MeldPong, tile, 3) {
		t.Fatalf("robbed added kong mutated source player: %#v", round.Players[0])
	}
}

func TestMCRPassedRobbingWindowCompletesAddedKong(t *testing.T) {
	round := mcrAddedKongTestGame(t)
	tile := mustFanTiles(t, "5m")[0]
	replacement := round.Wall[len(round.Wall)-1]
	if result := round.ApplyCommand(GameCommand{PlayerID: "0", Kind: CommandKong, Tile: "5m"}); !result.OK {
		t.Fatal(result.Error)
	}

	result := round.ApplyCommand(GameCommand{PlayerID: "1", Kind: CommandPass})

	if !result.OK || round.PendingClaim != nil || round.Current != 0 || round.Phase != PhaseAwaitingDiscard {
		t.Fatalf("result=%#v pending=%#v", result, round.PendingClaim)
	}
	if round.Players[0].Count(tile) != 0 || !hasMCRMeld(round.Players[0].Melds, MeldKong, tile, 4) || round.Players[0].Count(replacement) != 1 {
		t.Fatalf("completed added kong state: %#v", round.Players[0])
	}
}

func TestMCRConcealedKongDoesNotOpenRobbingWindow(t *testing.T) {
	round := newMCRActionTestGame()
	tile := mustFanTiles(t, "1m")[0]
	replacement := mustFanTiles(t, "9p")[0]
	round.Wall = []Tile{replacement}
	round.Players[0].Hand = mustFanTiles(t, "1m", "1m", "1m", "1m", "2m")

	result := round.ApplyCommand(GameCommand{PlayerID: "0", Kind: CommandKong, Tile: "1m"})

	if !result.OK || round.PendingClaim != nil || !hasMCRMeld(round.Players[0].Melds, MeldKong, tile, 4) || round.Players[0].Count(replacement) != 1 {
		t.Fatalf("result=%#v player=%#v pending=%#v", result, round.Players[0], round.PendingClaim)
	}
}

func TestMCRSpecialShapeWinCommandUsesMCRRuleSet(t *testing.T) {
	round := newMCRActionTestGame()
	round.Players[0].Hand = mustFanTiles(t, "1m", "9m", "1p", "9p", "1s", "9s", "E", "E", "S", "W", "N", "Z", "F", "B")
	round.RecordEvent(EventDraw, 0, mustFanTiles(t, "E")[0], "")
	if !hasLegalAction(round.rules.LegalActions(round, "0"), CommandWin, "") {
		t.Fatal("thirteen orphans should expose MCR win")
	}

	result := round.ApplyCommand(GameCommand{PlayerID: "0", Kind: CommandWin})

	if !result.OK || !round.Over || round.Winner != 0 {
		t.Fatalf("result=%#v round winner=%d over=%t", result, round.Winner, round.Over)
	}
	if round.MCRScore == nil || !mcrScoreHasFan(*round.MCRScore, "mcr_81") {
		t.Fatalf("special win score = %#v", round.MCRScore)
	}
}

func mcrAddedKongTestGame(t *testing.T) *Game {
	t.Helper()
	round := newMCRActionTestGame()
	tile := mustFanTiles(t, "5m")[0]
	round.Players[0].Melds = []Meld{{Kind: MeldPong, Tiles: []Tile{tile, tile, tile}}}
	round.Players[0].Hand = mustFanTiles(t, "5m", "2p")
	round.Players[1].Hand = mustFanTiles(t, "1p", "2p", "3p", "2p", "3p", "4p", "4s", "5s", "6s", "7m", "8m", "9m", "5m")
	round.Players[2].Hand = mustFanTiles(t, "1m")
	round.Players[3].Hand = mustFanTiles(t, "2m")
	round.Wall = mustFanTiles(t, "3m", "9p")
	return round
}

func hasMCRMeld(melds []Meld, kind MeldKind, tile Tile, count int) bool {
	for _, meld := range melds {
		if meld.Kind == kind && len(meld.Tiles) == count && meld.Tiles[0] == tile {
			return true
		}
	}
	return false
}

func mcrScoreHasFan(score MCRScoreBreakdown, id FanID) bool {
	for _, fan := range score.Fans {
		if fan.ID == id {
			return true
		}
	}
	return false
}

func newMCRActionTestGame() *Game {
	rules := NewMCRRuleSet(DefaultRuleConfig(ModeMCR).MCR)
	return &Game{
		Players:    NewPlayers(),
		Wall:       BuildMCRWall(),
		Mode:       ModeMCR,
		RuleConfig: rules.Config(),
		Winner:     -1,
		Phase:      PhaseAwaitingDiscard,
		rules:      rules,
	}
}
