package replay

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"mahjong/internal/game"
)

func TestSaveAndLoadReplayAtomically(t *testing.T) {
	dir := t.TempDir()
	file := replayStoreFixture(t, "atomic", time.Unix(20, 0).UTC())

	path, err := Save(dir, file)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Ext(path) != ".json" {
		t.Fatalf("path = %q", path)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || strings.HasPrefix(entries[0].Name(), ".replay-") {
		t.Fatalf("directory entries = %#v", entries)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	wantJSON, _ := json.Marshal(file)
	gotJSON, _ := json.Marshal(loaded)
	if !bytes.Equal(wantJSON, gotJSON) {
		t.Fatalf("loaded replay differs\nwant=%s\ngot=%s", wantJSON, gotJSON)
	}
}

func TestSaveRejectsInvalidReplayWithoutCreatingFile(t *testing.T) {
	dir := t.TempDir()
	file := replayStoreFixture(t, "invalid", time.Unix(20, 0).UTC())
	file.ApplicationVersion = "tampered"

	if _, err := Save(dir, file); err == nil {
		t.Fatal("Save accepted checksum-invalid replay")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("invalid save left files: %#v", entries)
	}
}

func TestListSkipsCorruptReplayAndKeepsValidFiles(t *testing.T) {
	dir := t.TempDir()
	older := replayStoreFixture(t, "older", time.Unix(10, 0).UTC())
	newer := replayStoreFixture(t, "newer", time.Unix(20, 0).UTC())
	if _, err := Save(dir, older); err != nil {
		t.Fatal(err)
	}
	newerPath, err := Save(dir, newer)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "broken.json"), []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}

	entries, issues, err := List(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 || entries[0].ReplayID != "newer" || entries[0].Path != newerPath {
		t.Fatalf("entries = %#v", entries)
	}
	if len(issues) != 1 || filepath.Base(issues[0].Path) != "broken.json" {
		t.Fatalf("issues = %#v", issues)
	}
}

func TestLoadRejectsTrailingJSON(t *testing.T) {
	dir := t.TempDir()
	file := replayStoreFixture(t, "trailing", time.Unix(20, 0).UTC())
	data, err := json.Marshal(file)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "trailing.json")
	if err := os.WriteFile(path, append(data, []byte("\n{}")...), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Load(path); err == nil {
		t.Fatal("Load accepted trailing JSON")
	}
}

func TestListMissingDirectoryIsEmpty(t *testing.T) {
	entries, issues, err := List(filepath.Join(t.TempDir(), "missing"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 || len(issues) != 0 {
		t.Fatalf("entries=%#v issues=%#v", entries, issues)
	}
}

func TestApplicationVersionIsNeverEmpty(t *testing.T) {
	if ApplicationVersion() == "" {
		t.Fatal("application version is empty")
	}
}

func replayStoreFixture(t *testing.T, id string, createdAt time.Time) game.ReplayFile {
	t.Helper()
	round := game.NewGame(140014).Snapshot()
	round.Over = true
	match := game.MatchSnapshot{
		Mode:       game.ModeCompatibility,
		RuleConfig: game.RuleConfig{},
		Complete:   true,
		Round:      round,
	}
	file, err := game.SealReplay(game.ReplayFile{
		ApplicationVersion: "test",
		ReplayID:           id,
		CreatedAt:          createdAt,
		Mode:               game.ModeCompatibility,
		RuleConfig:         game.RuleConfig{},
		ShuffleProof:       round.ShuffleProof,
		Participants: []game.ReplayParticipant{
			{Seat: 0, ID: "0", Name: "You"},
			{Seat: 1, ID: "1", Name: "AI-1"},
			{Seat: 2, ID: "2", Name: "AI-2"},
			{Seat: 3, ID: "3", Name: "AI-3"},
		},
		Initial:        match,
		Frames:         []game.ReplayFrame{{Index: 0, Match: match}},
		FinalStandings: match.Points,
		Complete:       true,
	})
	if err != nil {
		t.Fatal(err)
	}
	return file
}
