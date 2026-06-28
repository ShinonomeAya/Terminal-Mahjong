package tui

import (
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"mahjong/internal/game"
)

func TestGeneratePhase13Snapshots(t *testing.T) {
	outputDir := os.Getenv("MAHJONG_PHASE13_CAPTURE_DIR")
	if outputDir == "" {
		t.Skip("set MAHJONG_PHASE13_CAPTURE_DIR to generate visual artifacts")
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name  string
		mode  game.RuleMode
		lang  Language
		width int
	}{
		{name: "riichi-wide-zh", mode: game.ModeRiichi, lang: LanguageChinese, width: 140},
		{name: "riichi-wide-en", mode: game.ModeRiichi, lang: LanguageEnglish, width: 140},
		{name: "mcr-wide-zh", mode: game.ModeMCR, lang: LanguageChinese, width: 140},
		{name: "mcr-wide-en", mode: game.ModeMCR, lang: LanguageEnglish, width: 140},
		{name: "riichi-fallback-80", mode: game.ModeRiichi, lang: LanguageChinese, width: 80},
	}
	for _, capture := range cases {
		model := phase13CaptureModel(t, capture.mode, capture.lang, capture.width)
		rendered := renderTable(model)
		lines := strings.Split(strings.TrimRight(rendered, "\n"), "\n")
		if len(lines) > model.Height {
			t.Fatalf("%s rendered %d lines, want at most %d", capture.name, len(lines), model.Height)
		}
		for lineNumber, line := range lines {
			if width := visibleWidth(line); width > capture.width {
				t.Fatalf("%s line %d is %d cells wide, want at most %d", capture.name, lineNumber+1, width, capture.width)
			}
		}
		view := ansi.Strip(rendered)
		page := phase13SnapshotHTML(capture.name, view)
		path := filepath.Join(outputDir, capture.name+".html")
		if err := os.WriteFile(path, []byte(page), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func phase13CaptureModel(t *testing.T, mode game.RuleMode, language Language, width int) Model {
	t.Helper()
	var rules game.RuleSet
	switch mode {
	case game.ModeMCR:
		rules = game.NewMCRRuleSet(game.DefaultRuleConfig(game.ModeMCR).MCR)
	case game.ModeRiichi:
		rules = game.NewRiichiRuleSet(game.DefaultRuleConfig(game.ModeRiichi).Riichi)
	default:
		t.Fatalf("unsupported capture mode %q", mode)
	}
	match, err := game.NewMatch(1313, rules)
	if err != nil {
		t.Fatal(err)
	}
	match.EnsureCurrentTurnDraw()
	model := NewModel()
	model.Screen = ScreenTable
	model.Game = match.Round
	model.Language = language
	model.Width = width
	model.Height = 42
	return model
}

func phase13SnapshotHTML(title string, view string) string {
	return fmt.Sprintf(`<!doctype html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<title>%s</title>
<style>
html,body{margin:0;background:#070b0f;color:#d7dde5}
body{padding:20px}
pre{margin:0;font:18px/1.2 "Cascadia Mono","Segoe UI Symbol","Noto Sans Mono CJK SC",monospace;letter-spacing:0;white-space:pre}
</style>
</head>
<body><pre>%s</pre></body>
</html>`, html.EscapeString(title), html.EscapeString(view))
}
