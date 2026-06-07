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

const (
	handStartX = 6
	handRowY   = 8
	handCellW  = 10
	handCols   = 7
)

func handHitBoxes(count int, startX int, y int) []TileHitBox {
	boxes := make([]TileHitBox, count)
	for i := 0; i < count; i++ {
		col := i % handCols
		row := i / handCols
		x := startX + col*handCellW
		boxes[i] = TileHitBox{Index: i, X1: x, X2: x + handCellW - 1, Y: y + row}
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
	return handHitBoxes(len(m.Game.Players[0].Hand), handStartX, handRowY)
}

func renderTable(m Model) string {
	if m.Game == nil {
		return "No game\n"
	}
	g := m.Game
	var out strings.Builder
	out.WriteString("TERMINAL MAHJONG\n")
	out.WriteString(fmt.Sprintf("Wall:%d  Events:%d  Turn:%s  Replay:ready\n\n", len(g.Wall), len(g.Events), g.Players[g.Current].Name))
	out.WriteString("Opponents\n")
	out.WriteString(renderOpponents(g, m.UnicodeTiles))
	out.WriteString("\nTable\n")
	out.WriteString(renderCenter(g))
	out.WriteString("\nYou\n")
	out.WriteString(fmt.Sprintf("Melds: %s\n", g.Players[0].MeldSummary()))
	out.WriteString(fmt.Sprintf("Discards: %s\n", game.FormatTileLabels(g.Players[0].Discards, m.UnicodeTiles)))
	out.WriteString(renderHand(g.Players[0].Hand, m.SelectedIndex, m.UnicodeTiles))
	out.WriteString("\nControls\n")
	out.WriteString("Left/Right select | Enter/Space discard | click select | second click discard | Q quit\n")
	return out.String()
}

func renderOpponents(g *game.Game, unicode bool) string {
	var out strings.Builder
	out.WriteString(renderOpponent(g.Players[2], unicode))
	out.WriteString(renderOpponent(g.Players[1], unicode))
	out.WriteString(renderOpponent(g.Players[3], unicode))
	return out.String()
}

func renderOpponent(player game.Player, unicode bool) string {
	return fmt.Sprintf("%s  hand:%02d  melds:%s  discards:%s\n", player.Name, len(player.Hand), player.MeldSummary(), game.FormatTileLabels(player.Discards, unicode))
}

func renderCenter(g *game.Game) string {
	recent := game.RecentEvents(g.Events, 3)
	center := "No events yet"
	if len(recent) > 0 {
		center = recent[len(recent)-1].String()
	}
	return fmt.Sprintf("Last: %s\nTips: %s\n", center, game.HandTips(g.Players[0].Hand))
}

func renderHand(hand []game.Tile, selected int, unicode bool) string {
	var tiles strings.Builder
	tiles.WriteString("Hand: ")
	selectedText := "-"
	for i, tile := range hand {
		if i > 0 && i%handCols == 0 {
			tiles.WriteString("\n      ")
		}
		if i == selected {
			selectedText = selectedTileText(i, tile, unicode)
		}
		tiles.WriteString(renderTileCell(i, tile, i == selected, unicode))
	}
	tiles.WriteString("\n")
	tiles.WriteString("Selected: " + selectedText + "\n")
	return tiles.String()
}

func renderTileCell(index int, tile game.Tile, selected bool, unicode bool) string {
	label := game.TileLabel(tile, unicode)
	if selected {
		return padRightRunes(fmt.Sprintf("▶ [%02d] %s ◀", index+1, label), handCellW)
	}
	return padRightRunes(fmt.Sprintf("  [%02d] %s", index+1, label), handCellW)
}

func selectedTileText(index int, tile game.Tile, unicode bool) string {
	return fmt.Sprintf("[%02d] %s (%s)", index+1, game.TileLabel(tile, unicode), tile.String())
}

func runeWidth(text string) int {
	return len([]rune(text))
}

func padRightRunes(text string, width int) string {
	if runeWidth(text) >= width {
		return text
	}
	return text + strings.Repeat(" ", width-runeWidth(text))
}

var gameOverItems = []string{"Restart", "Main Menu", "Quit"}

func renderGameOver(m Model) string {
	if m.Game == nil {
		return "GAME OVER\n"
	}
	var out strings.Builder
	out.WriteString("GAME OVER\n")
	out.WriteString(fmt.Sprintf("Result: %s\n", m.Game.Reason))
	out.WriteString(fmt.Sprintf("Events: %d\n", len(m.Game.Events)))
	out.WriteString("Replay-ready event log: yes\n\n")
	for i, item := range gameOverItems {
		prefix := "  "
		if i == m.GameOverIndex {
			prefix = "> "
		}
		out.WriteString(prefix + item + "\n")
	}
	out.WriteString("\nUp/Down choose | Enter confirm | R restart | M menu | Q quit\n")
	return out.String()
}
