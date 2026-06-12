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
	if m.Online {
		return handHitBoxes(len(onlineHand(m)), handStartX, handRowY)
	}
	if m.Game == nil {
		return nil
	}
	return handHitBoxes(len(m.Game.Players[0].Hand), handStartX, handRowY)
}

func renderTable(m Model) string {
	if m.Online {
		return renderOnlineTable(m)
	}
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

func renderOnlineTable(m Model) string {
	snapshot := m.OnlineSnapshot
	if len(snapshot.Players) == 0 {
		return styleTitle("TERMINAL MAHJONG") + "\n" + styleMuted(renderNetworkStatus(m)) + "\nWaiting for online snapshot\n"
	}
	var out strings.Builder
	out.WriteString(styleTitle("TERMINAL MAHJONG") + "\n")
	out.WriteString(styleMuted(fmt.Sprintf("Room:%s  Seat:%d  Wall:%d  Events:%d", m.OnlineRoomCode, m.OnlineSeat+1, snapshot.WallCount, len(snapshot.Events))) + "\n")
	out.WriteString(styleMuted(renderOnlineRoomState(m)) + "\n\n")
	out.WriteString(styleMuted(renderNetworkStatus(m)) + "\n\n")
	out.WriteString(styleSectionTitle("Opponents") + "\n")
	out.WriteString(renderOnlineOpponents(m))
	out.WriteString("\n" + styleSectionTitle("Table") + "\n")
	out.WriteString(renderOnlineCenter(m))
	out.WriteString("\n" + styleSectionTitle("You") + "\n")
	out.WriteString(renderStatus(m))
	out.WriteString(renderLastAction(m))
	player := onlinePlayer(m, m.OnlineSeat)
	out.WriteString(fmt.Sprintf("Melds: %s\n", meldSummary(player.Melds)))
	out.WriteString(fmt.Sprintf("Discards: %s\n", game.FormatTileLabels(player.Discards, m.UnicodeTiles)))
	out.WriteString(renderHand(player.Hand, m.SelectedIndex, m.UnicodeTiles))
	out.WriteString(renderOnlineActionBar(m))
	out.WriteString("\n" + styleSectionTitle("Controls") + "\n")
	out.WriteString(styleMuted(tableControls(m)) + "\n")
	return out.String()
}

func renderOnlineRoomState(m Model) string {
	status := "Waiting for players"
	if m.OnlineStarted {
		status = "Started"
	}
	total := len(m.OnlineOccupiedSeats)
	if total == 0 {
		total = 1
	}
	return fmt.Sprintf("Ready: %d/%d  State: %s  Press R ready", len(m.OnlineReadySeats), total, status)
}

func renderOnlineOpponents(m Model) string {
	opposite := onlinePlayer(m, (m.OnlineSeat+2)%4)
	left := onlinePlayer(m, (m.OnlineSeat+1)%4)
	right := onlinePlayer(m, (m.OnlineSeat+3)%4)
	var out strings.Builder
	out.WriteString("                 " + renderOnlineSeat("Opposite", opposite, m.UnicodeTiles) + "\n")
	out.WriteString(padRightVisible(renderOnlineSeat("Left", left, m.UnicodeTiles), 44))
	out.WriteString(renderOnlineSeat("Right", right, m.UnicodeTiles) + "\n")
	return out.String()
}

func renderOnlineSeat(seat string, player game.PlayerView, unicode bool) string {
	discards := game.FormatTileLabels(recentTiles(player.Discards, 4), unicode)
	if discards == "" {
		discards = "-"
	}
	return fmt.Sprintf("%s: %s hand:%02d melds:%s last:%s", seat, player.Name, len(player.Hand), meldSummary(player.Melds), discards)
}

func renderOnlineCenter(m Model) string {
	var out strings.Builder
	snapshot := m.OnlineSnapshot
	turnName := "-"
	if snapshot.Current >= 0 && snapshot.Current < len(snapshot.Players) {
		turnName = snapshot.Players[snapshot.Current].Name
	}
	out.WriteString(styleMuted("Round Status") + "\n")
	out.WriteString(fmt.Sprintf("Wall: %02d | Turn: %s | Events: %d\n", snapshot.WallCount, turnName, len(snapshot.Events)))
	out.WriteString(renderEventLog(snapshot.Events, m.UnicodeTiles, 4))
	out.WriteString(fmt.Sprintf("Tips: %s\n", game.HandTips(onlineHand(m))))
	return out.String()
}

func onlineHand(m Model) []game.Tile {
	return onlinePlayer(m, m.OnlineSeat).Hand
}

func onlinePlayer(m Model, seat int) game.PlayerView {
	if seat >= 0 && seat < len(m.OnlineSnapshot.Players) {
		return m.OnlineSnapshot.Players[seat]
	}
	return game.PlayerView{Name: "-"}
}

func meldSummary(melds []game.Meld) string {
	if len(melds) == 0 {
		return "-"
	}
	parts := make([]string, len(melds))
	for i, meld := range melds {
		parts[i] = string(meld.Kind)
	}
	return strings.Join(parts, ",")
}

func tableControls(m Model) string {
	if m.Online {
		if m.Width > 0 && m.Width < 80 {
			return "Arrows select | R ready | H win | K kong | Enter discard | Q menu"
		}
		return "Left/Right select | R ready | H win | K kong | Enter/Space discard | click select | second click discard | Q menu"
	}
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

func renderOnlineActionBar(m Model) string {
	winReady := canOnlineWin(m)
	kongReady := canOnlineKong(m)
	winAction := actionState("[H] Win", winReady)
	kongAction := actionState("[K] Kong", kongReady)
	if m.Width > 0 && m.Width < 80 {
		return actionBarLine(
			"Actions:",
			"[R]Ready",
			"[Enter] Discard",
			"[Click] Tile",
			compactActionState("[H]Win", winReady),
			compactActionState("[K]Kong", kongReady),
			"[Q] Menu",
		) + "\n"
	}
	return actionBarLine(
		"Actions:",
		"[R] Ready",
		"[←/→] Select",
		"[Enter/Space] Discard",
		"[Click] Tile",
		winAction,
		kongAction,
		"[Q] Menu",
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

func canOnlineWin(m Model) bool {
	if !m.OnlineStarted || m.OnlineSnapshot.Current != m.OnlineSeat {
		return false
	}
	return game.CanWin(onlineHand(m))
}

func canOnlineKong(m Model) bool {
	if !m.OnlineStarted || m.OnlineSnapshot.Current != m.OnlineSeat {
		return false
	}
	_, ok := onlineKongTile(m)
	return ok
}

func onlineKongTile(m Model) (game.Tile, bool) {
	hand := onlineHand(m)
	if len(hand) == 0 {
		return 0, false
	}
	if m.SelectedIndex >= 0 && m.SelectedIndex < len(hand) {
		selected := hand[m.SelectedIndex]
		if game.TileCounts(hand)[selected] >= 4 {
			return selected, true
		}
	}
	counts := game.TileCounts(hand)
	for tile, count := range counts {
		if count >= 4 {
			return game.Tile(tile), true
		}
	}
	return 0, false
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
	if m.Online {
		return renderOnlineGameOver(m)
	}
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

func renderOnlineGameOver(m Model) string {
	snapshot := m.OnlineSnapshot
	var out strings.Builder
	out.WriteString(styleTitle("GAME OVER") + "\n")
	out.WriteString(fmt.Sprintf("Room: %s\n", m.OnlineRoomCode))
	out.WriteString(fmt.Sprintf("Result: %s\n", snapshot.Reason))
	if snapshot.Winner >= 0 {
		out.WriteString(fmt.Sprintf("Winner: Seat %d\n", snapshot.Winner+1))
	} else {
		out.WriteString("Winner: -\n")
	}
	out.WriteString(fmt.Sprintf("Events: %d\n", len(snapshot.Events)))
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
	out.WriteString(styleMuted("Up/Down choose | Enter confirm | M menu | Q quit") + "\n")
	return out.String()
}
