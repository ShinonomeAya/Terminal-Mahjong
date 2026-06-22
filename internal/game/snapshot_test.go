package game

import "testing"

func TestSnapshotIncludesPublicRoundState(t *testing.T) {
	g := NewGame(9)
	g.RecordEvent(EventDraw, 0, mustTile(t, "1m"), "")

	snapshot := g.Snapshot()

	if snapshot.Seed != 9 || snapshot.ShuffleProof.WallHash == "" {
		t.Fatalf("snapshot proof = %#v seed=%d", snapshot.ShuffleProof, snapshot.Seed)
	}
	if snapshot.WallCount != len(g.Wall) || snapshot.Current != g.Current {
		t.Fatalf("snapshot wall/current = %d/%d, want %d/%d", snapshot.WallCount, snapshot.Current, len(g.Wall), g.Current)
	}
	if len(snapshot.Players) != len(g.Players) || snapshot.Players[0].Name != "You" {
		t.Fatalf("players = %#v", snapshot.Players)
	}
	if len(snapshot.Players[0].Hand) != len(g.Players[0].Hand) {
		t.Fatalf("human hand length = %d, want %d", len(snapshot.Players[0].Hand), len(g.Players[0].Hand))
	}
	if len(snapshot.Events) != 1 {
		t.Fatalf("events = %#v", snapshot.Events)
	}
}

func TestSnapshotCopiesSlices(t *testing.T) {
	g := NewGame(9)

	snapshot := g.Snapshot()
	snapshot.Players[0].Hand[0] = mustTile(t, "E")

	if g.Players[0].Hand[0] == mustTile(t, "E") {
		t.Fatal("snapshot hand should not alias game hand")
	}
}

func TestApplyGameCommandDiscardsSelectedTile(t *testing.T) {
	g := NewGame(9)
	g.Current = 0
	startEvents := len(g.Events)
	tile := g.Players[0].Hand[0]

	result := g.ApplyCommand(GameCommand{PlayerID: "0", Kind: CommandDiscard, TileIndex: 0})

	if !result.OK || result.Tile != tile.String() {
		t.Fatalf("result = %#v want discard %s", result, tile)
	}
	if len(g.Players[0].Discards) != 1 || len(g.Events) != startEvents+1 {
		t.Fatalf("discard/events = %d/%d", len(g.Players[0].Discards), len(g.Events))
	}
}

func TestApplyGameCommandRejectsWrongPlayer(t *testing.T) {
	g := NewGame(9)

	result := g.ApplyCommand(GameCommand{PlayerID: "1", Kind: CommandDiscard, TileIndex: 0})

	if result.OK || result.Error == "" {
		t.Fatalf("result = %#v, want rejection", result)
	}
}

func TestApplyGameCommandSupportsCurrentSeatDiscard(t *testing.T) {
	g := NewGame(9)
	g.Current = 1
	tile := g.Players[1].Hand[0]

	result := g.ApplyCommand(GameCommand{PlayerID: "1", Kind: CommandDiscard, TileIndex: 0})

	if !result.OK || result.Tile != tile.String() {
		t.Fatalf("result = %#v want discard %s", result, tile)
	}
	if len(g.Players[1].Discards) != 1 || g.Current != 2 {
		t.Fatalf("player 1 discards/current = %d/%d", len(g.Players[1].Discards), g.Current)
	}
}

func TestEnsureCurrentTurnDrawDrawsOnceWhenHandNeedsTile(t *testing.T) {
	g := NewGame(9)
	startWall := len(g.Wall)
	startEvents := len(g.Events)

	tile, ok := g.EnsureCurrentTurnDraw()
	if !ok {
		t.Fatal("expected current player to draw")
	}
	if tile < 0 || len(g.Players[g.Current].Hand) != 14 {
		t.Fatalf("tile=%v hand=%d", tile, len(g.Players[g.Current].Hand))
	}
	if len(g.Wall) != startWall-1 || len(g.Events) != startEvents+1 {
		t.Fatalf("wall/events = %d/%d", len(g.Wall), len(g.Events))
	}

	_, ok = g.EnsureCurrentTurnDraw()
	if ok {
		t.Fatal("expected second ensure not to draw again")
	}
	if len(g.Players[g.Current].Hand) != 14 || len(g.Wall) != startWall-1 {
		t.Fatalf("second ensure hand/wall = %d/%d", len(g.Players[g.Current].Hand), len(g.Wall))
	}
}

func TestApplyGameCommandWinsWhenHandIsComplete(t *testing.T) {
	g := NewGame(9)
	g.Players[0].Hand = mustTiles(t,
		"1m", "2m", "3m",
		"4m", "5m", "6m",
		"2p", "3p", "4p",
		"7s", "7s", "7s",
		"E", "E",
	)

	result := g.ApplyCommand(GameCommand{PlayerID: "0", Kind: CommandWin})

	if !result.OK || !g.Over || g.Winner != 0 {
		t.Fatalf("result = %#v winner=%d over=%v", result, g.Winner, g.Over)
	}
}

func TestBuildPendingClaimOrdersWinBeforePongBeforeChow(t *testing.T) {
	g := NewGame(9)
	discard := mustTile(t, "3m")
	g.Players[1].Hand = mustTiles(t,
		"1m", "2m",
		"4m", "5m", "6m",
		"2p", "3p", "4p",
		"7s", "7s", "7s",
		"E", "E",
	)
	g.Players[2].Hand = mustTiles(t,
		"3m", "3m", "1p", "2p", "4p", "5p", "7p",
		"1s", "2s", "4s", "5s", "7s", "N",
	)
	g.Players[3].Hand = mustTiles(t,
		"1p", "2p", "4p", "5p", "7p", "8p", "1s",
		"2s", "4s", "5s", "7s", "8s", "N",
	)

	pending := g.buildPendingClaim(0, discard)

	if pending == nil || len(pending.Options) < 3 {
		t.Fatalf("pending = %#v, want win, pong, and chow options", pending)
	}
	if pending.Options[0].Kind != ClaimWin || pending.Options[0].Player != 1 {
		t.Fatalf("first option = %#v, want player 1 win", pending.Options[0])
	}
	if pending.Options[1].Kind != ClaimPong || pending.Options[1].Player != 2 {
		t.Fatalf("second option = %#v, want player 2 pong", pending.Options[1])
	}
	for _, option := range pending.Options[2:] {
		if option.Kind != ClaimChow || option.Player != 1 {
			t.Fatalf("lower-priority option = %#v, want player 1 chow", option)
		}
	}
}

func TestSnapshotCopiesPendingClaim(t *testing.T) {
	g := NewGame(9)
	g.Phase = PhaseAwaitingClaim
	g.PendingClaim = &PendingClaim{
		Discarder: 0,
		Tile:      mustTile(t, "3m"),
		Options: []ClaimOption{{
			Kind:     ClaimChow,
			Player:   1,
			Consumed: mustTiles(t, "1m", "2m"),
		}},
	}

	snapshot := g.Snapshot()

	if snapshot.Phase != PhaseAwaitingClaim || snapshot.PendingClaim == nil {
		t.Fatalf("snapshot phase/pending = %q/%#v", snapshot.Phase, snapshot.PendingClaim)
	}
	snapshot.PendingClaim.Options[0].Consumed[0] = mustTile(t, "9p")
	if g.PendingClaim.Options[0].Consumed[0] == mustTile(t, "9p") {
		t.Fatal("snapshot pending claim should not alias game state")
	}
}

func TestDiscardCommandEntersPendingClaimState(t *testing.T) {
	g := NewGame(9)
	g.Current = 0
	g.Players[0].Hand = mustTiles(t, "3m", "1p", "2p", "4p", "5p", "7p", "8p", "1s", "2s", "4s", "5s", "7s", "N", "N")
	g.Players[1].Hand = mustTiles(t, "3m", "3m", "1p", "2p", "4p", "5p", "7p", "8p", "1s", "2s", "4s", "5s", "N")

	result := g.ApplyCommand(GameCommand{PlayerID: "0", Kind: CommandDiscard, TileIndex: 0})

	if !result.OK {
		t.Fatalf("discard failed: %s", result.Error)
	}
	if g.Phase != PhaseAwaitingClaim || g.PendingClaim == nil {
		t.Fatalf("phase/pending = %q/%#v", g.Phase, g.PendingClaim)
	}
	if g.Current != 1 || g.PendingClaim.Options[g.PendingClaim.Active].Kind != ClaimPong {
		t.Fatalf("current/pending = %d/%#v", g.Current, g.PendingClaim)
	}
}

func TestPassCompletesClaimWindowAndAdvancesTurn(t *testing.T) {
	g := gameWithPendingPong(t)

	result := g.ApplyCommand(GameCommand{PlayerID: "1", Kind: CommandPass})

	if !result.OK {
		t.Fatalf("pass failed: %s", result.Error)
	}
	if g.Phase != PhaseAwaitingDiscard || g.PendingClaim != nil || g.Current != 1 {
		t.Fatalf("phase/pending/current = %q/%#v/%d", g.Phase, g.PendingClaim, g.Current)
	}
	startWall := len(g.Wall)
	if _, ok := g.EnsureCurrentTurnDraw(); !ok || len(g.Wall) != startWall-1 {
		t.Fatal("next player should draw after all claims pass")
	}
}

func TestPongCommandClaimsDiscardWithoutDrawing(t *testing.T) {
	g := gameWithPendingPong(t)
	startWall := len(g.Wall)

	result := g.ApplyCommand(GameCommand{PlayerID: "1", Kind: CommandPong})

	if !result.OK {
		t.Fatalf("pong failed: %s", result.Error)
	}
	if g.Phase != PhaseAwaitingDiscard || g.Current != 1 || g.PendingClaim != nil {
		t.Fatalf("phase/current/pending = %q/%d/%#v", g.Phase, g.Current, g.PendingClaim)
	}
	if len(g.Players[1].Melds) != 1 || g.Players[1].Melds[0].Kind != MeldPong {
		t.Fatalf("melds = %#v", g.Players[1].Melds)
	}
	if len(g.Players[0].Discards) != 0 || len(g.Wall) != startWall {
		t.Fatalf("discard/wall = %v/%d, want claimed discard and no draw", g.Players[0].Discards, len(g.Wall))
	}
}

func TestClaimWinCommandFinishesOnDiscard(t *testing.T) {
	g := NewGame(9)
	discard := mustTile(t, "3m")
	g.Players[0].Discards = []Tile{discard}
	g.Phase = PhaseAwaitingClaim
	g.Current = 1
	g.PendingClaim = &PendingClaim{Discarder: 0, Tile: discard, Options: []ClaimOption{{Kind: ClaimWin, Player: 1}}}

	result := g.ApplyCommand(GameCommand{PlayerID: "1", Kind: CommandClaimWin})

	if !result.OK || !g.Over || g.Winner != 1 || g.WinType != WinDiscard || g.Phase != PhaseRoundOver {
		t.Fatalf("result=%#v winner=%d type=%q phase=%q", result, g.Winner, g.WinType, g.Phase)
	}
}

func TestChowCommandSelectsActiveCombination(t *testing.T) {
	g := NewGame(9)
	discard := mustTile(t, "3m")
	g.Players[0].Discards = []Tile{discard}
	g.Players[1].Hand = mustTiles(t, "1m", "2m", "2m", "4m", "1p", "2p", "4p", "5p", "7p", "1s", "2s", "4s", "N")
	g.Phase = PhaseAwaitingClaim
	g.Current = 1
	g.PendingClaim = &PendingClaim{
		Discarder: 0,
		Tile:      discard,
		Options: []ClaimOption{
			{Kind: ClaimChow, Player: 1, Consumed: mustTiles(t, "1m", "2m")},
			{Kind: ClaimChow, Player: 1, Consumed: mustTiles(t, "2m", "4m")},
		},
	}

	result := g.ApplyCommand(GameCommand{PlayerID: "1", Kind: CommandChow, TileIndex: 1})

	if !result.OK {
		t.Fatalf("chow failed: %s", result.Error)
	}
	if got := FormatTiles(g.Players[1].Melds[0].Tiles); got != "2m 3m 4m" {
		t.Fatalf("chow meld = %s", got)
	}
	if g.Players[1].Count(mustTile(t, "1m")) != 1 || g.Players[1].Count(mustTile(t, "2m")) != 1 {
		t.Fatalf("wrong chow tiles removed: %s", FormatTiles(g.Players[1].Hand))
	}
}

func TestClaimStateRejectsDiscardAndWrongPlayer(t *testing.T) {
	g := gameWithPendingPong(t)

	if result := g.ApplyCommand(GameCommand{PlayerID: "1", Kind: CommandDiscard, TileIndex: 0}); result.OK {
		t.Fatal("discard should be rejected during a claim response")
	}
	if result := g.ApplyCommand(GameCommand{PlayerID: "2", Kind: CommandPass}); result.OK {
		t.Fatal("non-active player pass should be rejected")
	}
}

func gameWithPendingPong(t *testing.T) *Game {
	t.Helper()
	g := NewGame(9)
	g.Current = 0
	g.Players[0].Hand = mustTiles(t, "3m", "1p", "2p", "4p", "5p", "7p", "8p", "1s", "2s", "4s", "5s", "7s", "N", "N")
	g.Players[1].Hand = mustTiles(t, "3m", "3m", "1p", "2p", "4p", "5p", "7p", "8p", "1s", "2s", "4s", "5s", "N")
	result := g.ApplyCommand(GameCommand{PlayerID: "0", Kind: CommandDiscard, TileIndex: 0})
	if !result.OK || g.PendingClaim == nil {
		t.Fatalf("setup discard failed: %#v pending=%#v", result, g.PendingClaim)
	}
	return g
}
