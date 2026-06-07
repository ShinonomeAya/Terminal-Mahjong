package tui

import (
	"fmt"
	"strings"

	"mahjong/internal/game"
)

type TileHitBox struct {
	Index int
	X1    int
	X2    int
	Y     int
}

func handHitBoxes(count int, startX int, y int) []TileHitBox {
	boxes := make([]TileHitBox, count)
	x := startX
	for i := 0; i < count; i++ {
		boxes[i] = TileHitBox{Index: i, X1: x, X2: x + 5, Y: y}
		x += 6
	}
	return boxes
}

func tileIndexAt(boxes []TileHitBox, x int, y int) (int, bool) {
	for _, box := range boxes {
		if y == box.Y && x >= box.X1 && x <= box.X2 {
			return box.Index, true
		}
	}
	return 0, false
}

func currentHandHitBoxes(m Model) []TileHitBox {
	if len(m.HandHitBoxes) > 0 {
		return m.HandHitBoxes
	}
	if m.Game == nil {
		return nil
	}
	return handHitBoxes(len(m.Game.Players[0].Hand), 2, 10)
}

func renderTable(m Model) string {
	if m.Game == nil {
		return "No game\n"
	}
	g := m.Game
	var out strings.Builder
	out.WriteString("╔════════════════════════ TERMINAL MAHJONG ════════════════════════╗\n")
	out.WriteString(fmt.Sprintf("║ Wall %-3d  Events %-3d  Turn %-10s  Replay ready                 ║\n", len(g.Wall), len(g.Events), g.Players[g.Current].Name))
	out.WriteString("╠════════════════════════════ AI-2 ═════════════════════════════════╣\n")
	out.WriteString(renderOpponent(g.Players[2], m.UnicodeTiles))
	out.WriteString("╠══════════════ AI-1 ═══════════╦════════ CENTER ════════╦════════ AI-3 ════════╣\n")
	out.WriteString(renderSidePlayers(g, m.UnicodeTiles))
	out.WriteString("╠══════════════════════════════ YOU ════════════════════════════════╣\n")
	out.WriteString(fmt.Sprintf("║ Melds: %-24s Discards: %-24s ║\n", g.Players[0].MeldSummary(), game.FormatTileLabels(g.Players[0].Discards, m.UnicodeTiles)))
	out.WriteString(renderHand(g.Players[0].Hand, m.SelectedIndex, m.UnicodeTiles))
	out.WriteString("╚══ ←/→ select  Enter/Space discard  mouse click tile  H win  K kong  Q quit ═╝\n")
	return out.String()
}

func renderOpponent(player game.Player, unicode bool) string {
	return fmt.Sprintf("║ %-8s hand:%2d  melds:%-12s discards:%-24s ║\n", player.Name, len(player.Hand), player.MeldSummary(), game.FormatTileLabels(player.Discards, unicode))
}

func renderSidePlayers(g *game.Game, unicode bool) string {
	left := g.Players[1]
	right := g.Players[3]
	recent := game.RecentEvents(g.Events, 3)
	center := "No events yet"
	if len(recent) > 0 {
		center = recent[len(recent)-1].String()
	}
	return fmt.Sprintf("║ %-25s ║ %-20s ║ %-20s ║\n", left.Name+" discards "+game.FormatTileLabels(left.Discards, unicode), center, right.Name+" discards "+game.FormatTileLabels(right.Discards, unicode))
}

func renderHand(hand []game.Tile, selected int, unicode bool) string {
	var tiles strings.Builder
	var markers strings.Builder
	tiles.WriteString("║ ")
	markers.WriteString("║ ")
	for i, tile := range hand {
		label := game.TileLabel(tile, unicode)
		cell := fmt.Sprintf("[%2d]%s ", i+1, label)
		tiles.WriteString(cell)
		if i == selected {
			markers.WriteString(strings.Repeat(" ", len(cell)/2))
			markers.WriteString("▲ selected ")
		} else {
			markers.WriteString(strings.Repeat(" ", len(cell)))
		}
	}
	tiles.WriteString("\n")
	markers.WriteString("\n")
	return tiles.String() + markers.String()
}

func renderGameOver(m Model) string {
	if m.Game == nil {
		return "GAME OVER\n"
	}
	return fmt.Sprintf("GAME OVER\nResult: %s\nEvents: %d\nReplay-ready event log: yes\n", m.Game.Reason, len(m.Game.Events))
}
