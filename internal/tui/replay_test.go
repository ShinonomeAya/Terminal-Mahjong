package tui

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
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

func TestReplayViewerNavigationIsBounded(t *testing.T) {
	model := replayViewerModel(t, threeFrameReplay(t))

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyRight})
	model = next.(Model)
	if model.ReplayFrame != 1 {
		t.Fatalf("right frame=%d", model.ReplayFrame)
	}
	next, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnd})
	model = next.(Model)
	if model.ReplayFrame != 2 {
		t.Fatalf("end frame=%d", model.ReplayFrame)
	}
	next, _ = model.Update(tea.KeyMsg{Type: tea.KeyRight})
	model = next.(Model)
	if model.ReplayFrame != 2 {
		t.Fatalf("right exceeded final frame: %d", model.ReplayFrame)
	}
	next, _ = model.Update(tea.KeyMsg{Type: tea.KeyHome})
	model = next.(Model)
	if model.ReplayFrame != 0 {
		t.Fatalf("home frame=%d", model.ReplayFrame)
	}
	next, _ = model.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if next.(Model).ReplayFrame != 0 {
		t.Fatalf("left went below zero: %d", next.(Model).ReplayFrame)
	}
}

func TestReplayPlaybackAdvancesAndStopsAtFinalFrame(t *testing.T) {
	model := replayViewerModel(t, threeFrameReplay(t))

	next, cmd := model.Update(tea.KeyMsg{Type: tea.KeySpace})
	model = next.(Model)
	if !model.ReplayPlaying || cmd == nil {
		t.Fatalf("playing=%t cmd=%v", model.ReplayPlaying, cmd)
	}
	next, cmd = model.Update(replayTickMsg{})
	model = next.(Model)
	if model.ReplayFrame != 1 || !model.ReplayPlaying || cmd == nil {
		t.Fatalf("first tick frame=%d playing=%t cmd=%v", model.ReplayFrame, model.ReplayPlaying, cmd)
	}
	next, cmd = model.Update(replayTickMsg{})
	model = next.(Model)
	if model.ReplayFrame != 2 || model.ReplayPlaying || cmd != nil {
		t.Fatalf("final tick frame=%d playing=%t cmd=%v", model.ReplayFrame, model.ReplayPlaying, cmd)
	}
}

func TestReplayViewerTogglesDetailsAndReturnsToBrowser(t *testing.T) {
	model := replayViewerModel(t, threeFrameReplay(t))

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = next.(Model)
	if !model.ReplayShowDetails {
		t.Fatal("Tab did not enable replay details")
	}
	next, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = next.(Model)
	if model.Screen != ScreenReplayBrowser || model.ReplayPlaying {
		t.Fatalf("screen=%v playing=%t", model.Screen, model.ReplayPlaying)
	}
}

func TestReplayViewerIgnoresLiveGameCommands(t *testing.T) {
	model := replayViewerModel(t, threeFrameReplay(t))
	before, err := json.Marshal(model.ReplayFile)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"h", "k", "l", "p", "c", "q"} {
		next, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
		if cmd != nil {
			t.Fatalf("key %q scheduled live command", key)
		}
		model = next.(Model)
	}
	after, err := json.Marshal(model.ReplayFile)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("replay viewer mutated recorded data")
	}
}

func TestReplayViewerRendersFrameAndFinalStandings(t *testing.T) {
	model := replayViewerModel(t, threeFrameReplay(t))
	model.ReplayFrame = 2
	view := model.View()
	for _, want := range []string{"3/3", "viewer-replay", "25000"} {
		if !strings.Contains(view, want) {
			t.Fatalf("view missing %q:\n%s", want, view)
		}
	}
}

func TestReplayTableUsesSharedFourSeatLayout(t *testing.T) {
	model := replayViewerModel(t, completedRiichiReplay(t))
	model.Width = 140
	view := renderReplayViewer(model)

	for _, want := range []string{"对家", "下家", "上家", "你", "牌河", "手牌", "25000"} {
		if !strings.Contains(view, want) {
			t.Fatalf("replay table missing %q:\n%s", want, view)
		}
	}
	if strings.Contains(view, "战术分析") {
		t.Fatalf("replay table exposed tactical analysis:\n%s", view)
	}
	state := tableStateFor(model)
	if !state.Replay || !state.ReadOnly || len(state.Snapshot.Players[1].Hand) == 0 {
		t.Fatalf("replay table state = %#v", state)
	}
}

func TestReplayDetailsRevealAllHandsAndRiichiSettlement(t *testing.T) {
	model := replayViewerModel(t, completedRiichiReplay(t))
	model.Width = 140
	model.ReplayShowDetails = true
	view := renderReplayViewer(model)
	frame, _ := currentReplayFrame(model)

	for _, player := range frame.Match.Round.Players {
		if !strings.Contains(view, playerName(model, player.Name)) {
			t.Fatalf("detail view missing player %s:\n%s", player.Name, view)
		}
		for _, tile := range player.Hand {
			if !strings.Contains(view, game.TileLabel(tile, model.UnicodeTiles)) {
				t.Fatalf("detail view missing tile %s for %s:\n%s", tile, player.Name, view)
			}
		}
	}
	for _, want := range []string{"宝牌", "里宝牌", "+3000", "-1000", "最终积分"} {
		if !strings.Contains(view, want) {
			t.Fatalf("detail view missing %q:\n%s", want, view)
		}
	}
	for _, forbidden := range []string{"[H]", "[K]", "战术分析"} {
		if strings.Contains(view, forbidden) {
			t.Fatalf("replay exposed live control %q:\n%s", forbidden, view)
		}
	}
}

func TestReplayDetailsShowMCRFlowersAndSettlement(t *testing.T) {
	model := replayViewerModel(t, completedMCRReplay(t))
	model.Width = 140
	model.ReplayShowDetails = true
	view := renderReplayViewer(model)

	for _, want := range []string{"花牌", game.TileLabel(mustUITiles(t, "P1")[0], true), "+24", "-8"} {
		if !strings.Contains(view, want) {
			t.Fatalf("MCR detail missing %q:\n%s", want, view)
		}
	}
}

func TestReplayWidthFitsSupportedViewports(t *testing.T) {
	for _, language := range []Language{LanguageChinese, LanguageEnglish} {
		for _, width := range []int{140, 96, 80} {
			t.Run(string(language)+"-"+strconv.Itoa(width), func(t *testing.T) {
				model := replayViewerModel(t, completedRiichiReplay(t))
				model.Language = language
				model.Width = width
				model.ReplayShowDetails = true
				view := renderReplayViewer(model)
				for _, line := range strings.Split(strings.TrimRight(view, "\n"), "\n") {
					if visibleWidth(line) > width {
						t.Fatalf("line width=%d want<=%d:\n%s", visibleWidth(line), width, view)
					}
				}
			})
		}
	}
}

func replayViewerModel(t *testing.T, file game.ReplayFile) Model {
	t.Helper()
	model := NewModel()
	model.Screen = ScreenReplayViewer
	model.ReplayFile = &file
	model.Width = 100
	return model
}

func completedRiichiReplay(t *testing.T) game.ReplayFile {
	t.Helper()
	config := game.DefaultRuleConfig(game.ModeRiichi)
	match, err := game.NewMatch(140019, game.NewRiichiRuleSet(config.Riichi))
	if err != nil {
		t.Fatal(err)
	}
	match.Round.Players[0].Discards = append(match.Round.Players[0].Discards, mustUITiles(t, "1m", "2m")...)
	match.Round.Over = true
	match.Round.Reason = "self-draw"
	match.Complete = true
	settlement := game.RiichiSettlement{
		Winners:    []int{0},
		Discarder:  1,
		Deltas:     [4]int{3000, -1000, -1000, -1000},
		HonbaAfter: 1,
	}
	match.LastRiichiSettlement = &settlement
	match.RiichiSettlements = []game.RiichiSettlement{settlement}
	snapshot := match.Snapshot()
	return sealTUIReplay(t, "riichi-details", time.Unix(40, 0).UTC(), snapshot)
}

func completedMCRReplay(t *testing.T) game.ReplayFile {
	t.Helper()
	config := game.DefaultRuleConfig(game.ModeMCR)
	match, err := game.NewMatch(140020, game.NewMCRRuleSet(config.MCR))
	if err != nil {
		t.Fatal(err)
	}
	match.Round.Players[0].Flowers = mustUITiles(t, "P1")
	match.Round.Over = true
	match.Round.Reason = "self-draw"
	match.Complete = true
	settlement := game.MCRSettlement{
		Winner:    0,
		Discarder: 1,
		Deltas:    [4]int{24, -8, -8, -8},
	}
	match.LastMCRSettlement = &settlement
	match.MCRSettlements = []game.MCRSettlement{settlement}
	snapshot := match.Snapshot()
	return sealTUIReplay(t, "mcr-details", time.Unix(50, 0).UTC(), snapshot)
}

func sealTUIReplay(t *testing.T, id string, createdAt time.Time, snapshot game.MatchSnapshot) game.ReplayFile {
	t.Helper()
	file, err := game.SealReplay(game.ReplayFile{
		ApplicationVersion: "test",
		ReplayID:           id,
		CreatedAt:          createdAt,
		Mode:               snapshot.Mode,
		RuleConfig:         snapshot.RuleConfig,
		ShuffleProof:       snapshot.Round.ShuffleProof,
		Participants: []game.ReplayParticipant{
			{Seat: 0, ID: "0", Name: snapshot.Round.Players[0].Name},
			{Seat: 1, ID: "1", Name: snapshot.Round.Players[1].Name},
			{Seat: 2, ID: "2", Name: snapshot.Round.Players[2].Name},
			{Seat: 3, ID: "3", Name: snapshot.Round.Players[3].Name},
		},
		Initial:           snapshot,
		Frames:            []game.ReplayFrame{{Index: 0, Match: snapshot}},
		MCRSettlements:    append([]game.MCRSettlement(nil), snapshot.MCRSettlements...),
		RiichiSettlements: append([]game.RiichiSettlement(nil), snapshot.RiichiSettlements...),
		FinalStandings:    snapshot.Points,
		Complete:          true,
	})
	if err != nil {
		t.Fatal(err)
	}
	return file
}

func threeFrameReplay(t *testing.T) game.ReplayFile {
	t.Helper()
	round := game.NewGame(140018).Snapshot()
	initial := game.MatchSnapshot{
		Mode:       game.ModeCompatibility,
		RuleConfig: game.RuleConfig{},
		Round:      round,
		Points:     [4]int{25000, 25000, 25000, 25000},
	}
	middle := initial
	middle.Round.Events = append([]game.GameEvent(nil), round.Events...)
	middle.Round.WallCount--
	final := middle
	final.Complete = true
	final.Round.Over = true
	final.Round.Reason = "test complete"
	file, err := game.SealReplay(game.ReplayFile{
		ApplicationVersion: "test",
		ReplayID:           "viewer-replay",
		CreatedAt:          time.Unix(30, 0).UTC(),
		Mode:               game.ModeCompatibility,
		RuleConfig:         game.RuleConfig{},
		ShuffleProof:       round.ShuffleProof,
		Participants: []game.ReplayParticipant{
			{Seat: 0, ID: "0", Name: "You"},
			{Seat: 1, ID: "1", Name: "AI-1"},
			{Seat: 2, ID: "2", Name: "AI-2"},
			{Seat: 3, ID: "3", Name: "AI-3"},
		},
		Initial: initial,
		Frames: []game.ReplayFrame{
			{Index: 0, Match: initial},
			{Index: 1, Match: middle},
			{Index: 2, Match: final},
		},
		FinalStandings: final.Points,
		Complete:       true,
	})
	if err != nil {
		t.Fatal(err)
	}
	return file
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
