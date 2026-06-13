package tui

import (
	"strings"
	"testing"
)

func TestVisibleWidthIgnoresAnsiAndCountsWideGlyphs(t *testing.T) {
	styled := styleSelectedTile("🀈")

	if got := visibleWidth(styled); got != visibleWidth("🀈") {
		t.Fatalf("visibleWidth(styled tile) = %d, want plain tile width", got)
	}
}

func TestStyleSectionTitleKeepsText(t *testing.T) {
	styled := styleSectionTitle("Opponents")

	if !strings.Contains(styled, "Opponents") {
		t.Fatalf("styled section = %q, want text", styled)
	}
}

func TestStyleSelectedTileKeepsMarkers(t *testing.T) {
	styled := styleSelectedTile("▶ [02] 🀈 ◀")

	if !strings.Contains(styled, "▶ [02]") || !strings.Contains(styled, "◀") {
		t.Fatalf("styled selected tile = %q, want markers", styled)
	}
}

func TestStyleMutedKeepsText(t *testing.T) {
	styled := styleMuted("Controls")

	if !strings.Contains(styled, "Controls") {
		t.Fatalf("styled muted text = %q, want text", styled)
	}
}

func TestStylePanelKeepsContentVisible(t *testing.T) {
	panel := stylePanel("Seat", "AI-1\nhand:13")

	for _, want := range []string{"Seat", "AI-1", "hand:13"} {
		if !strings.Contains(panel, want) {
			t.Fatalf("panel missing %q:\n%s", want, panel)
		}
	}
}

func TestStylePanelRespectsVisibleWidth(t *testing.T) {
	panel := stylePanel("Center", "Wall: 67")

	if got := visibleWidth(panel); got > 40 {
		t.Fatalf("panel width = %d, want <= 40:\n%s", got, panel)
	}
}

func TestStyleTileFaceKeepsUnicodeTile(t *testing.T) {
	tile := styleTileFace("🀇", true)

	if !strings.Contains(tile, "🀇") {
		t.Fatalf("tile face missing unicode tile: %q", tile)
	}
	if got := visibleWidth(tile); got > 6 {
		t.Fatalf("tile face width = %d, want <= 6: %q", got, tile)
	}
}
