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
	out.WriteString(styleTitle("TERMINAL MAHJONG") + "\n")
	out.WriteString(styleMuted(fmt.Sprintf("Wall:%d  Events:%d  Turn:%s  Replay:ready", len(g.Wall), len(g.Events), g.Players[g.Current].Name)) + "\n\n")
	out.WriteString(styleSectionTitle("Opponents") + "\n")
	out.WriteString(renderOpponents(g, m.UnicodeTiles))
	out.WriteString("\n" + styleSectionTitle("Table") + "\n")
	out.WriteString(renderCenter(g, m.UnicodeTiles))
	out.WriteString("\n" + styleSectionTitle("You") + "\n")
	out.WriteString(renderStatus(m))
	out.WriteString(fmt.Sprintf("Melds: %s\n", g.Players[0].MeldSummary()))
	out.WriteString(fmt.Sprintf("Discards: %s\n", game.FormatTileLabels(g.Players[0].Discards, m.UnicodeTiles)))
	out.WriteString(renderHand(g.Players[0].Hand, m.SelectedIndex, m.UnicodeTiles))
	out.WriteString("\n" + styleSectionTitle("Controls") + "\n")
	out.WriteString(styleMuted(tableControls(m)) + "\n")
	return out.String()
}

func tableControls(m Model) string {
	if m.Width > 0 && m.Width < 80 {
		return "Arrows select | Enter discard | Click tile | Q quit"
	}
	return "Left/Right select | Enter/Space discard | click select | second click discard | Q quit"
}

func renderStatus(m Model) string {
	if m.StatusLine == "" {
		return ""
	}
	return styleStatus("Status: "+m.StatusLine) + "\n"
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

func renderCenter(g *game.Game, unicode bool) string {
	var out strings.Builder
	out.WriteString(renderEventLog(g.Events, unicode, 4))
	out.WriteString(fmt.Sprintf("Tips: %s\n", game.HandTips(g.Players[0].Hand)))
	return out.String()
}

func renderEventLog(events []game.GameEvent, unicode bool, limit int) string {
	var out strings.Builder
	out.WriteString(styleMuted("Recent Events") + "\n")
	recent := game.RecentEvents(events, limit)
	if len(recent) == 0 {
		out.WriteString("-\n")
		return out.String()
	}
	for _, event := range recent {
		out.WriteString(formatEventLine(event, unicode) + "\n")
	}
	return out.String()
}

func formatEventLine(event game.GameEvent, unicode bool) string {
	tileText := ""
	if event.Tile >= 0 && int(event.Tile) < game.TileTypeCount {
		tileText = " " + game.TileLabel(event.Tile, unicode)
	}
	note := ""
	if event.Note != "" {
		note = " - " + event.Note
	}
	return fmt.Sprintf("%02d. %s %s%s%s", event.Turn, eventPlayerName(event.Player), event.Kind, tileText, note)
}

func eventPlayerName(player int) string {
	switch player {
	case 0:
		return "You"
	case 1:
		return "AI-1"
	case 2:
		return "AI-2"
	case 3:
		return "AI-3"
	default:
		return fmt.Sprintf("Player-%d", player)
	}
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
		return padRightRunes(styleSelectedTile(fmt.Sprintf("▶ [%02d] %s ◀", index+1, label)), handCellW)
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
	out.WriteString(styleTitle("GAME OVER") + "\n")
	out.WriteString(fmt.Sprintf("Result: %s\n", m.Game.Reason))
	out.WriteString(fmt.Sprintf("Events: %d\n", len(m.Game.Events)))
	out.WriteString("Replay-ready event log: yes\n\n")
	for i, item := range gameOverItems {
		prefix := "  "
		if i == m.GameOverIndex {
			prefix = "> "
			out.WriteString(styleSelectedTile(prefix+item) + "\n")
			continue
		}
		out.WriteString(prefix + item + "\n")
	}
	out.WriteString("\n" + styleSectionTitle("Controls") + "\n")
	out.WriteString(styleMuted("Up/Down choose | Enter confirm | R restart | M menu | Q quit") + "\n")
	return out.String()
}
