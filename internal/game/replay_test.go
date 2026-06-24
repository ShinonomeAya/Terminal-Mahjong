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

func TestReplayLogCapturesSchemaModeAndRuleConfig(t *testing.T) {
	config := DefaultRuleConfig(ModeRiichi)
	game, err := NewGameWithRules(7, NewCompatibilityRuleSet(ModeRiichi, config))
	if err != nil {
		t.Fatal(err)
	}

	log := game.ReplayLog()

	if ReplaySchemaVersion != 3 {
		t.Fatalf("ReplaySchemaVersion = %d", ReplaySchemaVersion)
	}
	if log.SchemaVersion != ReplaySchemaVersion {
		t.Fatalf("log schema version = %d", log.SchemaVersion)
	}
	if log.Mode != ModeRiichi || log.RuleConfig != config {
		t.Fatalf("replay rules = %q/%#v, want %q/%#v", log.Mode, log.RuleConfig, ModeRiichi, config)
	}
	log.RuleConfig.Riichi.StartingPoints = 1000
	if game.RuleConfig.Riichi.StartingPoints != 25000 {
		t.Fatal("replay config should not mutate game config")
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
	for _, want := range []string{`"schema_version": 3`, `"mode": "compatibility"`, `"rule_config"`, `"seed": 7`, `"shuffle_proof"`, `"wall_hash"`, `"events"`, `"discard"`} {
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
	for _, want := range []string{"Replay", "Mode: compatibility", "Seed: 7", "Wall Hash:", "Winner: You", "Events: 2"} {
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

func TestMCRMatchReplayContainsScoreAndSettlementHistory(t *testing.T) {
	match, err := NewMatch(71, NewMCRRuleSet(DefaultRuleConfig(ModeMCR).MCR))
	if err != nil {
		t.Fatal(err)
	}
	score := MCRScoreBreakdown{
		Fans:            []FanMatch{{ID: "mcr_34", NameZH: "花龙", NameEN: "Mixed Straight", Points: 8, Count: 1}},
		NonFlowerPoints: 8,
		TotalPoints:     8,
		MeetsMinimum:    true,
	}
	for hand := 0; hand < 2; hand++ {
		match.Round.Over = true
		match.Round.Phase = PhaseRoundOver
		match.Round.Winner = hand
		match.Round.WinType = WinSelfDraw
		match.Round.MCRScore = &score
		match.completeMCRRound()
	}

	log := match.ReplayLog()

	if log.SchemaVersion != ReplaySchemaVersion || ReplaySchemaVersion != 3 || log.MCRScore == nil || len(log.MCRSettlements) != 2 {
		t.Fatalf("MCR replay = %#v", log)
	}
	if log.MCRScore.Fans[0].ID != "mcr_34" || log.MCRSettlements[1].Winner != 1 {
		t.Fatalf("MCR replay details = %#v", log)
	}
}
