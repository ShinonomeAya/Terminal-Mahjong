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
