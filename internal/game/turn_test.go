package game

import "testing"

func TestStartHumanTurnDrawsTile(t *testing.T) {
	game := NewGame(1)
	startWall := len(game.Wall)
	startHand := len(game.Players[0].Hand)

	tile, ok := game.StartHumanTurn()

	if !ok {
		t.Fatal("StartHumanTurn returned false")
	}
	if tile < 0 || int(tile) >= TileTypeCount {
		t.Fatalf("drawn tile = %v", tile)
	}
	if len(game.Wall) != startWall-1 {
		t.Fatalf("wall = %d, want %d", len(game.Wall), startWall-1)
	}
	if len(game.Players[0].Hand) != startHand+1 {
		t.Fatalf("hand = %d, want %d", len(game.Players[0].Hand), startHand+1)
	}
}

func TestHumanDiscardSelectedRemovesTileAndRecordsEvent(t *testing.T) {
	game := NewGame(1)
	game.Players[0].Hand = mustTiles(t, "1m", "2m", "3m")

	discard, err := game.HumanDiscardSelected(1)

	if err != nil {
		t.Fatal(err)
	}
	if discard.String() != "2m" {
		t.Fatalf("discard = %s, want 2m", discard)
	}
	if FormatTiles(game.Players[0].Discards) != "2m" {
		t.Fatalf("discards = %s, want 2m", FormatTiles(game.Players[0].Discards))
	}
	if len(game.Events) == 0 || game.Events[len(game.Events)-1].Kind != EventDiscard {
		t.Fatalf("last event = %#v, want discard", game.Events)
	}
	if game.Current != 1 {
		t.Fatalf("current = %d, want next player 1", game.Current)
	}
}

func TestHumanDiscardSelectedRejectsInvalidIndex(t *testing.T) {
	game := NewGame(1)
	game.Players[0].Hand = mustTiles(t, "1m")

	if _, err := game.HumanDiscardSelected(3); err == nil {
		t.Fatal("expected invalid index error")
	}
}

func TestAdvanceAIUntilHumanTurnReturnsToHumanWithDraw(t *testing.T) {
	game := NewGame(1)
	game.StartHumanTurn()
	if _, err := game.HumanDiscardSelected(0); err != nil {
		t.Fatal(err)
	}
	startEvents := len(game.Events)

	game.AdvanceAIUntilHumanTurn()

	if game.Over {
		t.Fatalf("game ended early: %s", game.Reason)
	}
	if game.Current != 0 {
		t.Fatalf("current = %d, want human player 0", game.Current)
	}
	if len(game.Players[0].Hand) != 14 {
		t.Fatalf("human hand = %d, want 14 after next draw", len(game.Players[0].Hand))
	}
	if len(game.Events) <= startEvents {
		t.Fatalf("events did not advance: %d <= %d", len(game.Events), startEvents)
	}
}
