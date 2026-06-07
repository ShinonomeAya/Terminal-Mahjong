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
