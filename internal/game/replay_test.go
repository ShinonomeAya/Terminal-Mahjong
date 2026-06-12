package game

import (
	"strings"
	"testing"
)

func TestReplayLogCapturesGameMetadata(t *testing.T) {
	game := NewGame(7)
	game.RecordEvent(EventDraw, 0, mustTile(t, "1m"), "")
	game.finish(0, "self-draw", WinSelfDraw)
	log := game.ReplayLog()
	if log.Seed != 7 || log.Winner != "You" || log.Result == "" || len(log.Events) != 2 {
		t.Fatalf("log = %#v", log)
	}
	if log.ShuffleProof.Seed != 7 || log.ShuffleProof.WallHash == "" {
		t.Fatalf("shuffle proof = %#v", log.ShuffleProof)
	}
}

func TestReplayLogToJSON(t *testing.T) {
	game := NewGame(7)
	game.RecordEvent(EventDiscard, 0, mustTile(t, "1m"), "")
	data, err := game.ReplayLog().ToJSON()
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, want := range []string{`"seed": 7`, `"shuffle_proof"`, `"wall_hash"`, `"events"`, `"discard"`} {
		if !strings.Contains(text, want) {
			t.Fatalf("json missing %s:\n%s", want, text)
		}
	}
}

func TestReplaySummaryIncludesResultAndEvents(t *testing.T) {
	game := NewGame(7)
	game.RecordEvent(EventDiscard, 0, mustTile(t, "1m"), "")
	game.finish(0, "self-draw", WinSelfDraw)
	summary := ReplaySummary(game.ReplayLog())
	for _, want := range []string{"Replay", "Seed: 7", "Wall Hash:", "Winner: You", "Events: 2"} {
		if !strings.Contains(summary, want) {
			t.Fatalf("summary missing %q:\n%s", want, summary)
		}
	}
}

func TestPrintResultMentionsReplayReadyLog(t *testing.T) {
	game := NewGame(7)
	game.finish(0, "self-draw", WinSelfDraw)
	var out strings.Builder
	game.printResult(&out)
	if !strings.Contains(out.String(), "Replay-ready event log") {
		t.Fatalf("result output:\n%s", out.String())
	}
}
