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
	handRowY   = 27
	handRowGap = 2
	handCellW  = 10
	handCols   = 7
)

func handHitBoxes(count int, startX int, y int) []TileHitBox {
	boxes := make([]TileHitBox, count)
	for i := 0; i < count; i++ {
		col := i % handCols
		row := i / handCols
		x := startX + col*handCellW
		boxes[i] = TileHitBox{Index: i, X1: x, X2: x + handCellW - 1, Y: y + row*handRowGap}
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
	out.WriteString(styleMuted(renderNetworkStatus(m)) + "\n\n")
	out.WriteString(styleSectionTitle("Opponents") + "\n")
	out.WriteString(renderOpponents(g, m.UnicodeTiles))
	out.WriteString("\n" + styleSectionTitle("Table") + "\n")
	out.WriteString(renderCenter(g, m.UnicodeTiles))
	out.WriteString("\n" + styleSectionTitle("You") + "\n")
	out.WriteString(renderStatus(m))
	out.WriteString(renderLastAction(m))
	out.WriteString(fmt.Sprintf("Melds: %s\n", g.Players[0].MeldSummary()))
	out.WriteString(fmt.Sprintf("Discards: %s\n", game.FormatTileLabels(g.Players[0].Discards, m.UnicodeTiles)))
	out.WriteString(renderHand(g.Players[0].Hand, m.SelectedIndex, m.UnicodeTiles))
	out.WriteString(renderActionBar(m))
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

func renderActionBar(m Model) string {
	winReady := canHumanWin(m)
	kongReady := canHumanKong(m)
	winAction := actionState("[H] Win", winReady)
	kongAction := actionState("[K] Kong", kongReady)
	if m.Width > 0 && m.Width < 80 {
		return actionBarLine(
			"Actions:",
			"[Enter] Discard",
			"[Click] Tile",
			compactActionState("[H]Win", winReady),
			compactActionState("[K]Kong", kongReady),
			"[Q] Quit",
		) + "\n"
	}
	return actionBarLine(
		"Actions:",
		"[←/→] Select",
		"[Enter/Space] Discard",
		"[Click] Tile",
		winAction,
		kongAction,
		"[Q] Quit",
	) + "\n"
}

func actionBarLine(parts ...string) string {
	return styleMuted(strings.Join(parts, " "))
}

func actionState(label string, ready bool) string {
	if ready {
		return styleSelectedTile(label + ":READY")
	}
	return label + ":off"
}

func compactActionState(label string, ready bool) string {
	if ready {
		return styleSelectedTile(label + ":READY")
	}
	return label + ":off"
}

func canHumanWin(m Model) bool {
	if m.Game == nil || len(m.Game.Players) == 0 {
		return false
	}
	return game.CanWin(m.Game.Players[0].Hand)
}

func canHumanKong(m Model) bool {
	if m.Game == nil || len(m.Game.Players) == 0 {
		return false
	}
	counts := game.TileCounts(m.Game.Players[0].Hand)
	for _, count := range counts {
		if count >= 4 {
			return true
		}
	}
	return false
}

func renderStatus(m Model) string {
	if m.StatusLine == "" {
		return styleStatus("Status: Ready") + "\n"
	}
	return styleStatus("Status: "+m.StatusLine) + "\n"
}

func renderLastAction(m Model) string {
	if m.StatusLine == "" {
		return styleStatus("Last Action: Waiting for input") + "\n"
	}
	return styleStatus("Last Action: "+m.StatusLine) + "\n"
}

func renderOpponents(g *game.Game, unicode bool) string {
	var out strings.Builder
	out.WriteString("                 " + renderOpponentSeat("Opposite", g.Players[2], unicode) + "\n")
	out.WriteString(padRightVisible(renderOpponentSeat("Left", g.Players[1], unicode), 44))
	out.WriteString(renderOpponentSeat("Right", g.Players[3], unicode) + "\n")
	return out.String()
}

func renderOpponentSeat(seat string, player game.Player, unicode bool) string {
	discards := game.FormatTileLabels(recentTiles(player.Discards, 4), unicode)
	if discards == "" {
		discards = "-"
	}
	return fmt.Sprintf("%s: %s hand:%02d melds:%s last:%s", seat, player.Name, len(player.Hand), player.MeldSummary(), discards)
}

func recentTiles(tiles []game.Tile, limit int) []game.Tile {
	if limit <= 0 || len(tiles) == 0 {
		return nil
	}
	if len(tiles) <= limit {
		return tiles
	}
	return tiles[len(tiles)-limit:]
}

func padRightVisible(text string, width int) string {
	if visibleWidth(text) >= width {
		return text + "  "
	}
	return text + strings.Repeat(" ", width-visibleWidth(text))
}

func renderCenter(g *game.Game, unicode bool) string {
	var out strings.Builder
	out.WriteString(renderRoundStatus(g))
	out.WriteString("\n")
	out.WriteString(renderEventLog(g.Events, unicode, 4))
	out.WriteString(fmt.Sprintf("Tips: %s\n", game.HandTips(g.Players[0].Hand)))
	return out.String()
}

func renderRoundStatus(g *game.Game) string {
	return fmt.Sprintf("%s\nWall: %02d | Turn: %s | Events: %d\n",
		styleMuted("Round Status"),
		len(g.Wall),
		g.Players[g.Current].Name,
		len(g.Events),
	)
}

func renderEventLog(events []game.GameEvent, unicode bool, limit int) string {
	var out strings.Builder
	out.WriteString(styleMuted("Recent Events") + "\n")
	recent := game.RecentEvents(events, limit)
	for i := 0; i < limit; i++ {
		if i < len(recent) {
			out.WriteString(formatEventLine(recent[i], unicode) + "\n")
			continue
		}
		if i == 0 {
			out.WriteString("-\n")
			continue
		}
		out.WriteString(" \n")
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
	var out strings.Builder
	out.WriteString(styleMuted("Hand Tray") + "\n")
	out.WriteString("Focus: " + selectedHandText(hand, selected, unicode) + "\n")
	for rowStart := 0; rowStart < len(hand); rowStart += handCols {
		rowEnd := min(rowStart+handCols, len(hand))
		if rowStart == 0 {
			out.WriteString("Hand: ")
		} else {
			out.WriteString("      ")
		}
		for i := rowStart; i < rowEnd; i++ {
			tile := hand[i]
			out.WriteString(renderTileCell(i, tile, i == selected, unicode))
		}
		out.WriteString("\n")
		out.WriteString(renderHandFocusMarker(rowStart, rowEnd, selected))
		out.WriteString("\n")
	}
	return out.String()
}

func selectedHandText(hand []game.Tile, selected int, unicode bool) string {
	if selected < 0 || selected >= len(hand) {
		return "-"
	}
	return selectedTileText(selected, hand[selected], unicode)
}

func renderHandFocusMarker(rowStart int, rowEnd int, selected int) string {
	var out strings.Builder
	out.WriteString("      ")
	for i := rowStart; i < rowEnd; i++ {
		if i == selected {
			out.WriteString(padRightRunes("    ▲", handCellW))
			continue
		}
		out.WriteString(strings.Repeat(" ", handCellW))
	}
	return out.String()
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
