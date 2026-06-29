package game

import (
	"errors"
	"testing"
	"time"
)

func TestNewMatchCarriesModeConfigAndInitialPoints(t *testing.T) {
	tests := []struct {
		name   string
		mode   RuleMode
		config RuleConfig
		points [4]int
	}{
		{name: "compatibility", mode: ModeCompatibility, config: RuleConfig{}},
		{name: "MCR", mode: ModeMCR, config: DefaultRuleConfig(ModeMCR)},
		{name: "riichi", mode: ModeRiichi, config: DefaultRuleConfig(ModeRiichi), points: [4]int{25000, 25000, 25000, 25000}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			match, err := NewMatch(9, NewCompatibilityRuleSet(test.mode, test.config))
			if err != nil {
				t.Fatalf("NewMatch() error = %v", err)
			}
			if match.Mode != test.mode || match.RuleConfig != test.config || match.Points != test.points {
				t.Fatalf("match = %#v, want mode=%q config=%#v points=%v", match, test.mode, test.config, test.points)
			}
			if match.Round == nil || match.Round.Mode != test.mode {
				t.Fatalf("round = %#v, want mode %q", match.Round, test.mode)
			}
		})
	}
}

func TestMatchDelegatesDrawAndCommand(t *testing.T) {
	match, err := NewMatch(9, NewCompatibilityRuleSet(ModeCompatibility, RuleConfig{}))
	if err != nil {
		t.Fatal(err)
	}
	startWall := len(match.Round.Wall)

	if _, ok := match.EnsureCurrentTurnDraw(); !ok {
		t.Fatal("expected delegated draw")
	}
	result := match.ApplyCommand(GameCommand{PlayerID: "0", Kind: CommandDiscard, TileIndex: 0})

	if !result.OK || len(match.Round.Wall) != startWall-1 || len(match.Round.Players[0].Discards) != 1 {
		t.Fatalf("result=%#v wall=%d discards=%v", result, len(match.Round.Wall), match.Round.Players[0].Discards)
	}
}

func TestMatchMarksCompleteAfterRoundEnds(t *testing.T) {
	match, err := NewMatch(9, NewCompatibilityRuleSet(ModeCompatibility, RuleConfig{}))
	if err != nil {
		t.Fatal(err)
	}
	match.Round.Players[0].Hand = mustTiles(t,
		"1m", "2m", "3m",
		"4m", "5m", "6m",
		"2p", "3p", "4p",
		"7s", "7s", "7s",
		"E", "E",
	)

	result := match.ApplyCommand(GameCommand{PlayerID: "0", Kind: CommandWin})

	if !result.OK || !match.Complete {
		t.Fatalf("result=%#v complete=%v", result, match.Complete)
	}
}

func TestMatchSnapshotIsCopied(t *testing.T) {
	config := DefaultRuleConfig(ModeRiichi)
	match, err := NewMatch(9, NewCompatibilityRuleSet(ModeRiichi, config))
	if err != nil {
		t.Fatal(err)
	}

	snapshot := match.Snapshot()
	snapshot.RuleConfig.Riichi.StartingPoints = 1000
	snapshot.Round.Players[0].Hand[0] = mustTile(t, "E")

	if match.RuleConfig.Riichi.StartingPoints != 25000 {
		t.Fatal("snapshot config should not mutate match")
	}
	if match.Round.Players[0].Hand[0] == mustTile(t, "E") {
		t.Fatal("snapshot round should not mutate match")
	}
}

func TestMCRMatchSnapshotCopiesSettlementHistory(t *testing.T) {
	match, err := NewMatch(73, NewMCRRuleSet(DefaultRuleConfig(ModeMCR).MCR))
	if err != nil {
		t.Fatal(err)
	}
	score := MCRScoreBreakdown{Fans: []FanMatch{{ID: "mcr_43", NameZH: "无番和", NameEN: "Chicken Hand", Points: 8, Count: 1}}, NonFlowerPoints: 8, TotalPoints: 8, MeetsMinimum: true}
	match.Round.Over = true
	match.Round.Phase = PhaseRoundOver
	match.Round.Winner = 0
	match.Round.WinType = WinSelfDraw
	match.Round.MCRScore = &score
	match.completeMCRRound()

	snapshot := match.Snapshot()
	if len(snapshot.MCRSettlements) != 1 || snapshot.LastMCRSettlement == nil {
		t.Fatalf("snapshot = %#v", snapshot)
	}
	snapshot.MCRSettlements[0].Score.Fans[0].NameEN = "mutated"
	snapshot.LastMCRSettlement.Score.Fans[0].NameEN = "mutated"
	if match.MCRSettlements[0].Score.Fans[0].NameEN != "Chicken Hand" || match.LastMCRSettlement.Score.Fans[0].NameEN != "Chicken Hand" {
		t.Fatal("snapshot mutated match settlement history")
	}
}

func TestMatchReplayJournalCapturesAcceptedCommandsAndFrames(t *testing.T) {
	match, err := NewMatch(140014, NewCompatibilityRuleSet(ModeCompatibility, RuleConfig{}))
	if err != nil {
		t.Fatal(err)
	}
	if match.ReplayFrameCount() != 1 || match.ReplayCommandCount() != 0 {
		t.Fatalf("initial journal frames=%d commands=%d", match.ReplayFrameCount(), match.ReplayCommandCount())
	}

	if _, ok := match.EnsureCurrentTurnDraw(); !ok {
		t.Fatal("expected initial draw")
	}
	if match.ReplayFrameCount() != 2 || match.replay.frames[1].Command != nil {
		t.Fatalf("draw journal = %#v", match.replay)
	}

	result := match.ApplyCommand(GameCommand{PlayerID: "0", Kind: CommandDiscard, TileIndex: 0})
	if !result.OK {
		t.Fatal(result.Error)
	}
	if match.ReplayCommandCount() != 1 || match.ReplayFrameCount() != 3 {
		t.Fatalf("journal frames=%d commands=%d", match.ReplayFrameCount(), match.ReplayCommandCount())
	}
	frame := match.replay.frames[2]
	if frame.Command == nil || frame.Command.Kind != CommandDiscard || frame.Match.Round.Players[0].Discards == nil {
		t.Fatalf("command frame = %#v", frame)
	}
}

func TestMatchRejectedCommandDoesNotRecordReplayFrame(t *testing.T) {
	match, err := NewMatch(140014, NewCompatibilityRuleSet(ModeCompatibility, RuleConfig{}))
	if err != nil {
		t.Fatal(err)
	}
	beforeFrames := match.ReplayFrameCount()
	beforeCommands := match.ReplayCommandCount()

	result := match.ApplyCommand(GameCommand{PlayerID: "1", Kind: CommandDiscard, TileIndex: 0})

	if result.OK {
		t.Fatal("out-of-turn command was accepted")
	}
	if match.ReplayFrameCount() != beforeFrames || match.ReplayCommandCount() != beforeCommands {
		t.Fatalf("rejected command changed journal frames=%d commands=%d", match.ReplayFrameCount(), match.ReplayCommandCount())
	}
}

func TestMatchReplayCapturesCompletedRoundBeforeNextRound(t *testing.T) {
	match, err := NewMatch(140014, NewMCRRuleSet(DefaultRuleConfig(ModeMCR).MCR))
	if err != nil {
		t.Fatal(err)
	}
	match.Round.Players[0].Hand = mustTiles(t,
		"1m", "9m", "1p", "9p", "1s", "9s",
		"E", "E", "S", "W", "N", "Z", "F", "B",
	)
	match.Round.RecordEvent(EventDraw, 0, mustTile(t, "E"), "")

	result := match.ApplyCommand(GameCommand{PlayerID: "0", Kind: CommandWin})
	if !result.OK {
		t.Fatal(result.Error)
	}

	frames := match.replay.frames
	if len(frames) < 3 {
		t.Fatalf("frames = %#v", frames)
	}
	completedRound := frames[len(frames)-2].Match
	nextRound := frames[len(frames)-1].Match
	if !completedRound.Round.Over || completedRound.RoundNumber != 1 {
		t.Fatalf("completed round frame = %#v", completedRound)
	}
	if nextRound.Complete || nextRound.Round.Over || nextRound.RoundNumber != 2 {
		t.Fatalf("next round frame = %#v", nextRound)
	}
}

func TestMatchQuitIsAbandonedAndCannotExportReplay(t *testing.T) {
	match, err := NewMatch(140014, NewCompatibilityRuleSet(ModeCompatibility, RuleConfig{}))
	if err != nil {
		t.Fatal(err)
	}

	result := match.ApplyCommand(GameCommand{PlayerID: "0", Kind: CommandQuit})
	if !result.OK {
		t.Fatal(result.Error)
	}
	if !match.Abandoned || match.Complete || !match.Snapshot().Abandoned {
		t.Fatalf("quit match = %#v snapshot=%#v", match, match.Snapshot())
	}
	if _, err := match.CompletedReplay("test", time.Unix(140014, 0), replayParticipantsFixture()); !errors.Is(err, ErrIncompleteReplay) {
		t.Fatalf("CompletedReplay() error = %v, want incomplete", err)
	}
}

func TestMatchCompletedReplayIsSealedAndCopied(t *testing.T) {
	match, err := NewMatch(140014, NewCompatibilityRuleSet(ModeCompatibility, RuleConfig{}))
	if err != nil {
		t.Fatal(err)
	}
	match.Round.Players[0].Hand = mustTiles(t,
		"1m", "2m", "3m",
		"4m", "5m", "6m",
		"2p", "3p", "4p",
		"7s", "7s", "7s",
		"E", "E",
	)
	result := match.ApplyCommand(GameCommand{PlayerID: "0", Kind: CommandWin})
	if !result.OK || !match.Complete {
		t.Fatalf("result=%#v complete=%t", result, match.Complete)
	}

	file, err := match.CompletedReplay("test", time.Unix(140014, 0).UTC(), replayParticipantsFixture())
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateReplay(file); err != nil {
		t.Fatal(err)
	}
	if file.ReplayID == "" || len(file.Commands) != 1 || len(file.Frames) < 3 {
		t.Fatalf("completed replay = %#v", file)
	}

	file.Frames[0].Match.Round.Players[0].Hand[0] = mustTile(t, "B")
	if match.replay.frames[0].Match.Round.Players[0].Hand[0] == mustTile(t, "B") {
		t.Fatal("exported replay mutated match journal")
	}
}

func TestAdvanceMatchAIUntilHumanTurnRecordsCommands(t *testing.T) {
	match, err := NewMatch(140014, NewCompatibilityRuleSet(ModeCompatibility, RuleConfig{}))
	if err != nil {
		t.Fatal(err)
	}
	match.EnsureCurrentTurnDraw()
	result := match.ApplyCommand(GameCommand{PlayerID: "0", Kind: CommandDiscard, TileIndex: 0})
	if !result.OK {
		t.Fatal(result.Error)
	}

	match.AdvanceAIUntilHumanTurn()

	if !match.Complete && match.Round.Current != 0 {
		t.Fatalf("current=%d complete=%t", match.Round.Current, match.Complete)
	}
	if match.ReplayCommandCount() < 2 {
		t.Fatalf("AI commands were not journaled: %#v", match.replay.commands)
	}
	for _, frame := range match.replay.frames {
		if frame.Command != nil && frame.Command.PlayerID == "" {
			t.Fatalf("command frame missing player: %#v", frame)
		}
	}
}

func replayParticipantsFixture() []ReplayParticipant {
	return []ReplayParticipant{
		{Seat: 0, ID: "0", Name: "You"},
		{Seat: 1, ID: "1", Name: "AI-1"},
		{Seat: 2, ID: "2", Name: "AI-2"},
		{Seat: 3, ID: "3", Name: "AI-3"},
	}
}
