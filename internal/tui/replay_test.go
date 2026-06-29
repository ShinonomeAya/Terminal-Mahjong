package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"mahjong/internal/game"
	"mahjong/internal/replay"
)

func TestReplayMenuOpensLocalizedBrowser(t *testing.T) {
	model := NewModel()
	model.ReplayDir = t.TempDir()
	model.MenuIndex = 8

	next, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := next.(Model)
	if updated.Screen != ScreenReplayBrowser || cmd == nil {
		t.Fatalf("screen=%v cmd=%v", updated.Screen, cmd)
	}
	if _, ok := cmd().(replayListMsg); !ok {
		t.Fatalf("command result = %#v", cmd())
	}
	if !strings.Contains(updated.View(), "回放") {
		t.Fatalf("Chinese browser title missing:\n%s", updated.View())
	}

	updated.Language = LanguageEnglish
	if !strings.Contains(updated.View(), "REPLAYS") {
		t.Fatalf("English browser title missing:\n%s", updated.View())
	}
}

func TestReplayBrowserListsNewestValidFiles(t *testing.T) {
	dir := t.TempDir()
	saveTUIReplayFixture(t, dir, "older", game.ModeMCR, time.Unix(10, 0).UTC())
	newest := saveTUIReplayFixture(t, dir, "newest", game.ModeRiichi, time.Unix(20, 0).UTC())
	if err := os.WriteFile(filepath.Join(dir, "broken.json"), []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}

	message := listReplaysCmd(dir)()
	listed, ok := message.(replayListMsg)
	if !ok {
		t.Fatalf("message = %#v", message)
	}
	model := NewModel()
	model.Screen = ScreenReplayBrowser
	next, _ := model.Update(listed)
	updated := next.(Model)
	if len(updated.ReplayEntries) != 2 || updated.ReplayEntries[0].Path != newest {
		t.Fatalf("entries=%#v", updated.ReplayEntries)
	}
	if len(updated.ReplayIssues) != 1 || !strings.Contains(updated.View(), "1") {
		t.Fatalf("issues=%#v\n%s", updated.ReplayIssues, updated.View())
	}
}

func TestReplayBrowserNavigatesRefreshesLoadsAndReturns(t *testing.T) {
	dir := t.TempDir()
	first := saveTUIReplayFixture(t, dir, "first", game.ModeMCR, time.Unix(10, 0).UTC())
	second := saveTUIReplayFixture(t, dir, "second", game.ModeRiichi, time.Unix(20, 0).UTC())
	model := NewModel()
	model.Screen = ScreenReplayBrowser
	model.ReplayDir = dir
	model.ReplayEntries = []replay.Entry{{Path: second}, {Path: first}}

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	updated := next.(Model)
	if updated.ReplayIndex != 1 {
		t.Fatalf("index=%d", updated.ReplayIndex)
	}
	next, cmd := updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("Enter did not schedule replay load")
	}
	loaded, ok := cmd().(replayLoadedMsg)
	if !ok || loaded.File.ReplayID != "first" {
		t.Fatalf("loaded=%#v", loaded)
	}
	next, _ = next.(Model).Update(loaded)
	if next.(Model).Screen != ScreenReplayViewer {
		t.Fatalf("screen=%v", next.(Model).Screen)
	}

	updated.Screen = ScreenReplayBrowser
	next, cmd = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	if cmd == nil {
		t.Fatal("R did not schedule refresh")
	}
	next, _ = next.(Model).Update(tea.KeyMsg{Type: tea.KeyEsc})
	if next.(Model).Screen != ScreenMenu {
		t.Fatalf("screen=%v", next.(Model).Screen)
	}
}

func TestReplayBrowserClampsSelectionAndKeepsLinesWithinWidth(t *testing.T) {
	model := NewModel()
	model.Screen = ScreenReplayBrowser
	model.Width = 80
	model.ReplayIndex = 4
	model.ReplayEntries = []replay.Entry{{
		Path:      filepath.Join("replays", strings.Repeat("long-name-", 20)+".json"),
		ReplayID:  "0123456789abcdef",
		Mode:      game.ModeRiichi,
		CreatedAt: time.Unix(20, 0).UTC(),
		Frames:    3,
	}}

	next, _ := model.Update(replayListMsg{Entries: model.ReplayEntries})
	updated := next.(Model)
	if updated.ReplayIndex != 0 {
		t.Fatalf("index=%d", updated.ReplayIndex)
	}
	for _, line := range strings.Split(updated.View(), "\n") {
		if visibleWidth(line) > model.Width {
			t.Fatalf("line width=%d:\n%s", visibleWidth(line), line)
		}
	}
	for _, want := range []string{"0123456789abcdef", string(game.ModeRiichi)} {
		if !strings.Contains(updated.View(), want) {
			t.Fatalf("view missing %q:\n%s", want, updated.View())
		}
	}
}

func saveTUIReplayFixture(t *testing.T, dir string, id string, mode game.RuleMode, createdAt time.Time) string {
	t.Helper()
	round := game.NewGame(140017).Snapshot()
	round.Over = true
	match := game.MatchSnapshot{
		Mode:       mode,
		RuleConfig: game.DefaultRuleConfig(mode),
		Complete:   true,
		Round:      round,
	}
	file, err := game.SealReplay(game.ReplayFile{
		ApplicationVersion: "test",
		ReplayID:           id,
		CreatedAt:          createdAt,
		Mode:               mode,
		RuleConfig:         game.DefaultRuleConfig(mode),
		ShuffleProof:       round.ShuffleProof,
		Participants: []game.ReplayParticipant{
			{Seat: 0, ID: "0", Name: "You"},
			{Seat: 1, ID: "1", Name: "AI-1"},
			{Seat: 2, ID: "2", Name: "AI-2"},
			{Seat: 3, ID: "3", Name: "AI-3"},
		},
		Initial:  match,
		Frames:   []game.ReplayFrame{{Index: 0, Match: match}},
		Complete: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	path, err := replay.Save(dir, file)
	if err != nil {
		t.Fatal(err)
	}
	return path
}
