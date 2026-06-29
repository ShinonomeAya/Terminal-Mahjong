package game

import (
	"errors"
	"testing"
	"time"
)

func TestReplayFileSealAndValidate(t *testing.T) {
	for _, mode := range []RuleMode{ModeMCR, ModeRiichi} {
		t.Run(string(mode), func(t *testing.T) {
			file := completedReplayFileFixture(t, mode)
			sealed, err := SealReplay(file)
			if err != nil {
				t.Fatal(err)
			}
			if sealed.SchemaVersion != ReplayFileSchemaVersion || sealed.Checksum == "" {
				t.Fatalf("sealed replay = %#v", sealed)
			}
			if err := ValidateReplay(sealed); err != nil {
				t.Fatal(err)
			}

			sealed.ApplicationVersion = "tampered"
			if !errors.Is(ValidateReplay(sealed), ErrReplayChecksum) {
				t.Fatal("mutated replay should fail checksum validation")
			}
		})
	}
}

func TestSealReplayDoesNotMutateInput(t *testing.T) {
	file := completedReplayFileFixture(t, ModeRiichi)

	sealed, err := SealReplay(file)
	if err != nil {
		t.Fatal(err)
	}

	if file.SchemaVersion != 0 || file.Checksum != "" {
		t.Fatalf("input replay was mutated: %#v", file)
	}
	if sealed.SchemaVersion != ReplayFileSchemaVersion || sealed.Checksum == "" {
		t.Fatalf("sealed replay = %#v", sealed)
	}
}

func TestReplayFileRejectsUnsupportedAndIncompleteFiles(t *testing.T) {
	file := sealedReplayFileFixture(t, ModeMCR)

	file.SchemaVersion = ReplayFileSchemaVersion + 1
	if !errors.Is(ValidateReplay(file), ErrUnsupportedReplayVersion) {
		t.Fatal("unsupported version was accepted")
	}

	file.SchemaVersion = ReplayFileSchemaVersion
	file.Complete = false
	if !errors.Is(ValidateReplay(file), ErrIncompleteReplay) {
		t.Fatal("incomplete replay was accepted")
	}
}

func TestReplayFileRejectsInvalidStructure(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*ReplayFile)
	}{
		{
			name: "missing frames",
			mutate: func(file *ReplayFile) {
				file.Frames = nil
			},
		},
		{
			name: "frame index gap",
			mutate: func(file *ReplayFile) {
				file.Frames[0].Index = 1
			},
		},
		{
			name: "duplicate participant seat",
			mutate: func(file *ReplayFile) {
				file.Participants[1].Seat = file.Participants[0].Seat
			},
		},
		{
			name: "participant seat out of range",
			mutate: func(file *ReplayFile) {
				file.Participants[0].Seat = 4
			},
		},
		{
			name: "invalid rule mode",
			mutate: func(file *ReplayFile) {
				file.Mode = RuleMode("unknown")
			},
		},
		{
			name: "final standings mismatch",
			mutate: func(file *ReplayFile) {
				file.FinalStandings[0]++
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			file := sealedReplayFileFixture(t, ModeRiichi)
			test.mutate(&file)
			if !errors.Is(ValidateReplay(file), ErrInvalidReplay) {
				t.Fatalf("invalid replay was accepted: %#v", file)
			}
		})
	}
}

func sealedReplayFileFixture(t *testing.T, mode RuleMode) ReplayFile {
	t.Helper()
	sealed, err := SealReplay(completedReplayFileFixture(t, mode))
	if err != nil {
		t.Fatal(err)
	}
	return sealed
}

func completedReplayFileFixture(t *testing.T, mode RuleMode) ReplayFile {
	t.Helper()
	config := DefaultRuleConfig(mode)
	var rules RuleSet
	switch mode {
	case ModeMCR:
		rules = NewMCRRuleSet(config.MCR)
	case ModeRiichi:
		rules = NewRiichiRuleSet(config.Riichi)
	default:
		t.Fatalf("unsupported fixture mode %q", mode)
	}
	match, err := NewMatch(140014, rules)
	if err != nil {
		t.Fatal(err)
	}
	match.EnsureCurrentTurnDraw()
	snapshot := match.Snapshot()
	snapshot.Complete = true
	return ReplayFile{
		ApplicationVersion: "test",
		ReplayID:           "replay-140014",
		CreatedAt:          time.Unix(140014, 0).UTC(),
		Mode:               mode,
		RuleConfig:         config,
		ShuffleProof:       snapshot.Round.ShuffleProof,
		Participants: []ReplayParticipant{
			{Seat: 0, ID: "0", Name: "You"},
			{Seat: 1, ID: "1", Name: "AI-1"},
			{Seat: 2, ID: "2", Name: "AI-2"},
			{Seat: 3, ID: "3", Name: "AI-3"},
		},
		Initial:        snapshot,
		Frames:         []ReplayFrame{{Index: 0, Match: snapshot}},
		FinalStandings: snapshot.Points,
		Complete:       true,
	}
}
