package game

import "testing"

func TestRiichiSnapshotForRedactsPrivateStateAndKeepsPublicTableState(t *testing.T) {
	round := newRiichiActionTestGame(t)
	round.Dealer = 2
	round.HandNumber = 3
	round.Current = 0
	round.Players[0].Hand = mustTiles(t, "1m", "2m", "3m")
	round.Players[1].Hand = mustTiles(t, "4m", "5m", "6m")
	round.Riichi.Honba = 2
	round.Riichi.RiichiSticks = 1
	round.Riichi.KanCount = 1
	round.Riichi.RinshanDraws = 1
	round.Riichi.Declarations[0] = RiichiAccepted
	round.Riichi.Ippatsu[0] = true
	round.Riichi.TemporaryFuriten[0] = true
	round.Riichi.RiichiFuriten[0] = true

	snapshot := round.SnapshotFor("0")

	if snapshot.Riichi == nil {
		t.Fatal("snapshot missing riichi state")
	}
	if snapshot.Riichi.DeadWallCount != 14 || len(snapshot.Riichi.DoraIndicators) != 1 {
		t.Fatalf("public riichi state = %#v", snapshot.Riichi)
	}
	if len(snapshot.Riichi.UraIndicators) != 0 {
		t.Fatalf("ura indicators leaked before round end: %#v", snapshot.Riichi.UraIndicators)
	}
	if snapshot.Riichi.Honba != 2 || snapshot.Riichi.RiichiSticks != 1 || snapshot.Riichi.KanCount != 1 || snapshot.Riichi.RinshanDraws != 1 {
		t.Fatalf("round counters = %#v", snapshot.Riichi)
	}
	if snapshot.Riichi.Declarations[0] != RiichiAccepted || !snapshot.Riichi.Ippatsu[0] {
		t.Fatalf("public declarations = %#v", snapshot.Riichi)
	}
	if !snapshot.Riichi.OwnTemporaryFuriten || !snapshot.Riichi.OwnRiichiFuriten {
		t.Fatalf("viewer furiten not preserved: %#v", snapshot.Riichi)
	}
	if len(snapshot.Players[0].Hand) == 0 || len(snapshot.Players[1].Hand) != 0 {
		t.Fatalf("private hands not redacted for viewer: %#v", snapshot.Players)
	}

	other := round.SnapshotFor("1")
	if other.Riichi.OwnTemporaryFuriten || other.Riichi.OwnRiichiFuriten {
		t.Fatalf("other viewer received player 0 furiten: %#v", other.Riichi)
	}
	if len(other.Players[0].Hand) != 0 || len(other.Players[1].Hand) == 0 {
		t.Fatalf("private hands not redacted for other viewer: %#v", other.Players)
	}
}

func TestRiichiSnapshotForRevealsUraOnlyAfterRoundEnd(t *testing.T) {
	round := newRiichiActionTestGame(t)

	before := round.SnapshotFor("0")
	if len(before.Riichi.UraIndicators) != 0 {
		t.Fatalf("ura indicators visible before end: %#v", before.Riichi.UraIndicators)
	}

	round.Over = true
	round.Phase = PhaseRoundOver
	after := round.SnapshotFor("0")
	if len(after.Riichi.UraIndicators) != len(round.Riichi.UraIndicators) {
		t.Fatalf("ura indicators after end = %#v, want %#v", after.Riichi.UraIndicators, round.Riichi.UraIndicators)
	}
}
