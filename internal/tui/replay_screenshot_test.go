package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"mahjong/internal/game"
)

func TestGeneratePhase14ReplaySnapshots(t *testing.T) {
	outputDir := os.Getenv("MAHJONG_PHASE14_CAPTURE_DIR")
	if outputDir == "" {
		t.Skip("set MAHJONG_PHASE14_CAPTURE_DIR to generate visual artifacts")
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name     string
		mode     game.RuleMode
		language Language
		width    int
		details  bool
	}{
		{name: "riichi-replay-wide-zh", mode: game.ModeRiichi, language: LanguageChinese, width: 140},
		{name: "riichi-replay-wide-en", mode: game.ModeRiichi, language: LanguageEnglish, width: 140},
		{name: "mcr-replay-wide-zh", mode: game.ModeMCR, language: LanguageChinese, width: 140},
		{name: "mcr-replay-wide-en", mode: game.ModeMCR, language: LanguageEnglish, width: 140},
		{name: "riichi-replay-details-zh", mode: game.ModeRiichi, language: LanguageChinese, width: 140, details: true},
		{name: "riichi-replay-fallback-80", mode: game.ModeRiichi, language: LanguageChinese, width: 80},
	}
	for _, capture := range cases {
		var file game.ReplayFile
		if capture.mode == game.ModeMCR {
			file = completedMCRReplay(t)
		} else {
			file = completedRiichiReplay(t)
		}
		model := replayViewerModel(t, file)
		model.Language = capture.language
		model.Width = capture.width
		model.Height = 42
		model.ReplayShowDetails = capture.details
		rendered := renderReplayViewer(model)
		lines := strings.Split(strings.TrimRight(rendered, "\n"), "\n")
		if len(lines) > model.Height {
			t.Fatalf("%s rendered %d lines, want at most %d", capture.name, len(lines), model.Height)
		}
		for lineNumber, line := range lines {
			if width := visibleWidth(line); width > capture.width {
				t.Fatalf("%s line %d is %d cells wide, want at most %d", capture.name, lineNumber+1, width, capture.width)
			}
		}
		page := phase13SnapshotHTML(capture.name, ansi.Strip(rendered))
		path := filepath.Join(outputDir, capture.name+".html")
		if err := os.WriteFile(path, []byte(page), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
