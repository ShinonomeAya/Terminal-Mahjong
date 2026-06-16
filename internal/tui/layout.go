package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"mahjong/internal/game"
)

type TileHitBox struct {
	Index int
	X1    int
	X2    int
	Y     int
	Y1    int
	Y2    int
}

const (
	handStartX = 6
	handRowY   = 29
	handRowGap = 1
	handCellW  = 4
	handCols   = 14
)

func handHitBoxes(count int, startX int, y int) []TileHitBox {
	boxes := make([]TileHitBox, count)
	for i := 0; i < count; i++ {
		col := i % handCols
		row := i / handCols
		x := startX + col*handCellW
		faceY := y + row*handRowGap
		boxes[i] = TileHitBox{Index: i, X1: x, X2: x + handCellW - 1, Y: faceY, Y1: faceY, Y2: faceY}
	}
	return boxes
}

func tileIndexAt(boxes []TileHitBox, x int, y int) (int, bool) {
	for _, box := range boxes {
		y1, y2 := box.Y1, box.Y2
		if y1 == 0 && y2 == 0 {
			y1, y2 = box.Y, box.Y
		}
		if y >= y1 && y <= y2 && x >= box.X1 && x <= box.X2 {
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
	topSeat := renderSeatPanel(
		m,
		seatLabel(m, "Opposite"),
		g.Players[2].Name,
		len(g.Players[2].Hand),
		g.Players[2].MeldSummary(),
		game.FormatTileLabels(recentTiles(g.Players[2].Discards, 4), m.UnicodeTiles),
		g.Current == 2,
	)
	leftSeat := renderSeatPanel(
		m,
		seatLabel(m, "Left"),
		g.Players[1].Name,
		len(g.Players[1].Hand),
		g.Players[1].MeldSummary(),
		game.FormatTileLabels(recentTiles(g.Players[1].Discards, 4), m.UnicodeTiles),
		g.Current == 1,
	)
	rightSeat := renderSeatPanel(
		m,
		seatLabel(m, "Right"),
		g.Players[3].Name,
		len(g.Players[3].Hand),
		g.Players[3].MeldSummary(),
		game.FormatTileLabels(recentTiles(g.Players[3].Discards, 4), m.UnicodeTiles),
		g.Current == 3,
	)
	center := stylePanelWidth(tableTitle(m), renderCenter(m, g, m.UnicodeTiles), 38)
	middle := renderTableMiddle(m, leftSeat, center, rightSeat)
	hand := stylePanelWidth(
		handTitle(m),
		renderStatus(m)+
			renderLastAction(m)+
			meldsLine(m, g.Players[0].MeldSummary())+
			discardsLine(m, game.FormatTileLabels(g.Players[0].Discards, m.UnicodeTiles))+
			renderHand(m, g.Players[0].Hand, m.SelectedIndex, m.UnicodeTiles),
		handPanelWidth(m),
	)
	return renderBoardFrame(
		m,
		styleMuted(tableMeta(m, len(g.Wall), len(g.Events), playerName(m, g.Players[g.Current].Name), "")),
		topSeat,
		middle,
		hand,
		styleSectionTitle(controlsTitle(m))+"\n"+styleMuted(tableControls(m)),
	)
}

func renderOnlineTable(m Model) string {
	snapshot := m.OnlineSnapshot
	if len(snapshot.Players) == 0 {
		return styleTitle(appTitle(m)) + "\n" + styleMuted(renderNetworkStatus(m)) + "\n" + waitingSnapshotText(m) + "\n"
	}
	player := onlinePlayer(m, m.OnlineSeat)
	opposite := onlinePlayer(m, (m.OnlineSeat+2)%4)
	left := onlinePlayer(m, (m.OnlineSeat+1)%4)
	right := onlinePlayer(m, (m.OnlineSeat+3)%4)
	topSeat := renderSeatPanel(m, seatLabel(m, "Opposite"), opposite.Name, len(opposite.Hand), meldSummary(opposite.Melds), game.FormatTileLabels(recentTiles(opposite.Discards, 4), m.UnicodeTiles), snapshot.Current == (m.OnlineSeat+2)%4)
	leftSeat := renderSeatPanel(m, seatLabel(m, "Left"), left.Name, len(left.Hand), meldSummary(left.Melds), game.FormatTileLabels(recentTiles(left.Discards, 4), m.UnicodeTiles), snapshot.Current == (m.OnlineSeat+1)%4)
	rightSeat := renderSeatPanel(m, seatLabel(m, "Right"), right.Name, len(right.Hand), meldSummary(right.Melds), game.FormatTileLabels(recentTiles(right.Discards, 4), m.UnicodeTiles), snapshot.Current == (m.OnlineSeat+3)%4)
	center := stylePanelWidth(tableTitle(m), renderOnlineCenter(m), 38)
	middle := renderTableMiddle(m, leftSeat, center, rightSeat)
	hand := stylePanelWidth(
		handTitle(m),
		renderStatus(m)+
			renderLastAction(m)+
			meldsLine(m, meldSummary(player.Melds))+
			discardsLine(m, game.FormatTileLabels(player.Discards, m.UnicodeTiles))+
			renderHand(m, player.Hand, m.SelectedIndex, m.UnicodeTiles),
		handPanelWidth(m),
	)
	return renderBoardFrame(
		m,
		styleMuted(onlineMeta(m, snapshot.WallCount, len(snapshot.Events))),
		topSeat,
		middle,
		hand,
		styleSectionTitle(controlsTitle(m))+"\n"+styleMuted(tableControls(m)),
	)
}

func tableWidth(m Model) int {
	width := m.Width
	if width <= 0 {
		width = 96
	}
	if width < 80 {
		return width
	}
	if width > 120 {
		return 120
	}
	return width
}

func centerLine(width int, text string) string {
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, text)
}

func handPanelWidth(m Model) int {
	if tableWidth(m) < 80 {
		return 74
	}
	return 92
}

func renderTableMiddle(m Model, leftSeat string, center string, rightSeat string) string {
	if tableWidth(m) < 80 {
		return lipgloss.JoinVertical(lipgloss.Center, leftSeat, center, rightSeat)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, leftSeat, "  ", center, "  ", rightSeat)
}

func appTitle(m Model) string {
	if m.chinese() {
		return "终端麻将"
	}
	return "TERMINAL MAHJONG"
}

func opponentsTitle(m Model) string {
	if m.chinese() {
		return "对手"
	}
	return "Opponents"
}

func tableTitle(m Model) string {
	if m.chinese() {
		return "牌桌"
	}
	return "Table CENTER"
}

func handTitle(m Model) string {
	if m.chinese() {
		return "手牌"
	}
	return "Hand Tray"
}

func controlsTitle(m Model) string {
	if m.chinese() {
		return "操作"
	}
	return "Controls"
}

func tableMeta(m Model, wall int, events int, turn string, prefix string) string {
	if m.chinese() {
		body := fmt.Sprintf("牌墙:%d  事件:%d  轮到:%s  回放:就绪", wall, events, turn)
		if prefix != "" {
			return prefix + "\n" + body
		}
		return body
	}
	body := fmt.Sprintf("Wall:%d  Events:%d  Turn:%s  Replay:ready", wall, events, turn)
	if prefix != "" {
		return prefix + "\n" + body
	}
	return body
}

func onlineMeta(m Model, wall int, events int) string {
	if m.chinese() {
		return fmt.Sprintf("房间:%s  座位:%d  牌墙:%d  事件:%d\n%s", m.OnlineRoomCode, m.OnlineSeat+1, wall, events, renderOnlineRoomState(m))
	}
	return fmt.Sprintf("Room:%s  Seat:%d  Wall:%d  Events:%d\n%s", m.OnlineRoomCode, m.OnlineSeat+1, wall, events, renderOnlineRoomState(m))
}

func meldsLine(m Model, value string) string {
	if m.chinese() {
		return fmt.Sprintf("副露：%s\n", value)
	}
	return fmt.Sprintf("Melds: %s\n", value)
}

func discardsLine(m Model, value string) string {
	if m.chinese() {
		return fmt.Sprintf("弃牌：%s\n", value)
	}
	return fmt.Sprintf("Discards: %s\n", value)
}

func waitingSnapshotText(m Model) string {
	if m.chinese() {
		return "等待联网牌桌同步"
	}
	return "Waiting for online snapshot"
}

func renderBoardFrame(m Model, meta string, topSeat string, middle string, hand string, prompt string) string {
	width := tableWidth(m)
	sections := []string{
		centerLine(width, styleTitle(appTitle(m))),
		centerLine(width, meta),
		centerLine(width, styleMuted(renderNetworkStatus(m))),
		centerLine(width, styleSectionTitle(opponentsTitle(m))),
		centerLine(width, topSeat),
		centerLine(width, middle),
		centerLine(width, hand),
		centerLine(width, prompt),
	}
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, strings.Join(sections, "\n")) + "\n"
}

func renderSeatPanel(m Model, label string, name string, handCount int, melds string, discards string, active bool) string {
	if discards == "" {
		discards = "-"
	}
	title := fmt.Sprintf("%s %s", label, playerName(m, name))
	if active {
		title = "▶ " + title
	}
	body := fmt.Sprintf("手牌:%02d  副露:%s\n最近:%s", handCount, melds, discards)
	if !m.chinese() {
		body = fmt.Sprintf("hand:%02d  melds:%s\nlast:%s", handCount, melds, discards)
	}
	return stylePanelWidth(title, body, 24)
}

func renderOnlineRoomState(m Model) string {
	total := len(m.OnlineOccupiedSeats)
	if total == 0 {
		total = 1
	}
	return roomStateText(m, len(m.OnlineReadySeats), total, m.OnlineStarted)
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
		turnName = playerName(m, snapshot.Players[snapshot.Current].Name)
	}
	out.WriteString(styleMuted(roundStatusTitle(m)) + "\n")
	out.WriteString(roundStatusLine(m, snapshot.WallCount, turnName, len(snapshot.Events)))
	out.WriteString(renderEventLog(m, snapshot.Events, m.UnicodeTiles, 4))
	out.WriteString(fmt.Sprintf("%s %s\n", tipsLabel(m), handTipsText(m, game.HandTips(onlineHand(m)))))
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
	if m.chinese() {
		if m.Online {
			if m.Width > 0 && m.Width < 80 {
				return strings.Join([]string{"方向键选牌", "R 准备", commandLabel(m, "[H] Win", canOnlineWin(m)), commandLabel(m, "[K] Kong", canOnlineKong(m)), "回车打出", "Q 菜单"}, " | ")
			}
			return strings.Join([]string{"←/→ 选牌", "R 准备", commandLabel(m, "[H] Win", canOnlineWin(m)), commandLabel(m, "[K] Kong", canOnlineKong(m)), "回车/空格打出", "单击选牌", "再次单击打出", "Q 菜单"}, " | ")
		}
		if m.Width > 0 && m.Width < 80 {
			return strings.Join([]string{"方向键选牌", "回车打出", "单击选牌", commandLabel(m, "[H] Win", canHumanWin(m)), commandLabel(m, "[K] Kong", canHumanKong(m)), "Q 退出"}, " | ")
		}
		return strings.Join([]string{"←/→ 选牌", "回车/空格打出", "单击选牌", "再次单击打出", commandLabel(m, "[H] Win", canHumanWin(m)), commandLabel(m, "[K] Kong", canHumanKong(m)), "Q 退出"}, " | ")
	}
	if m.Online {
		if m.Width > 0 && m.Width < 80 {
			return strings.Join([]string{"Arrows select", "R ready", commandLabel(m, "[H] Win", canOnlineWin(m)), commandLabel(m, "[K] Kong", canOnlineKong(m)), "Enter discard", "Q menu"}, " | ")
		}
		return strings.Join([]string{"Left/Right select", "R ready", commandLabel(m, "[H] Win", canOnlineWin(m)), commandLabel(m, "[K] Kong", canOnlineKong(m)), "Enter/Space discard", "click select", "second click discard", "Q menu"}, " | ")
	}
	if m.Width > 0 && m.Width < 80 {
		return strings.Join([]string{"Arrows select", "Enter discard", "Click tile", commandLabel(m, "[H] Win", canHumanWin(m)), commandLabel(m, "[K] Kong", canHumanKong(m)), "Q quit"}, " | ")
	}
	return strings.Join([]string{"Left/Right select", "Enter/Space discard", "click select", "second click discard", commandLabel(m, "[H] Win", canHumanWin(m)), commandLabel(m, "[K] Kong", canHumanKong(m)), "Q quit"}, " | ")
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
		winAction,
		kongAction,
		"[←/→] Select",
		"[Enter/Space] Discard",
		"[Click] Tile",
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
	return styleStatus(statusText(m)) + "\n"
}

func renderLastAction(m Model) string {
	return styleStatus(lastActionText(m)) + "\n"
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

func renderCenter(m Model, g *game.Game, unicode bool) string {
	var out strings.Builder
	out.WriteString(renderRoundStatus(m, g))
	out.WriteString("\n")
	out.WriteString(renderEventLog(m, g.Events, unicode, 4))
	out.WriteString(fmt.Sprintf("%s %s\n", tipsLabel(m), handTipsText(m, game.HandTips(g.Players[0].Hand))))
	return out.String()
}

func renderRoundStatus(m Model, g *game.Game) string {
	return fmt.Sprintf("%s\n%s",
		styleMuted(roundStatusTitle(m)),
		roundStatusLine(m,
			len(g.Wall),
			playerName(m, g.Players[g.Current].Name),
			len(g.Events),
		),
	)
}

func roundStatusTitle(m Model) string {
	if m.chinese() {
		return "局况"
	}
	return "Round Status"
}

func roundStatusLine(m Model, wall int, turn string, events int) string {
	if m.chinese() {
		return fmt.Sprintf("牌墙: %02d | 轮到: %s | 事件: %d\n",
			wall,
			turn,
			events,
		)
	}
	return fmt.Sprintf("Wall: %02d | Turn: %s | Events: %d\n",
		wall,
		turn,
		events,
	)
}

func tipsLabel(m Model) string {
	if m.chinese() {
		return "提示："
	}
	return "Tips:"
}

func renderEventLog(m Model, events []game.GameEvent, unicode bool, limit int) string {
	var out strings.Builder
	if m.chinese() {
		out.WriteString(styleMuted("最近事件") + "\n")
	} else {
		out.WriteString(styleMuted("Recent Events") + "\n")
	}
	recent := game.RecentEvents(events, limit)
	for i := 0; i < limit; i++ {
		if i < len(recent) {
			out.WriteString(formatEventLine(m, recent[i], unicode) + "\n")
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

func formatEventLine(m Model, event game.GameEvent, unicode bool) string {
	tileText := ""
	if event.Tile >= 0 && int(event.Tile) < game.TileTypeCount {
		tileText = " " + game.TileLabel(event.Tile, unicode)
	}
	note := ""
	if event.Note != "" {
		note = " - " + event.Note
	}
	return fmt.Sprintf("%02d. %s %s%s%s", event.Turn, eventPlayerName(m, event.Player), eventKindText(m, string(event.Kind)), tileText, note)
}

func eventPlayerName(m Model, player int) string {
	switch player {
	case 0:
		return playerName(m, "You")
	case 1:
		return playerName(m, "AI-1")
	case 2:
		return playerName(m, "AI-2")
	case 3:
		return playerName(m, "AI-3")
	default:
		return fmt.Sprintf("Player-%d", player)
	}
}

func renderHand(m Model, hand []game.Tile, selected int, unicode bool) string {
	var out strings.Builder
	if m.chinese() {
		out.WriteString(styleMuted("手牌托盘") + "\n")
		out.WriteString("焦点：" + selectedHandText(hand, selected, unicode) + "\n")
	} else {
		out.WriteString(styleMuted("Hand Tray") + "\n")
		out.WriteString("Focus: " + selectedHandText(hand, selected, unicode) + "\n")
	}
	for rowStart := 0; rowStart < len(hand); rowStart += handCols {
		rowEnd := min(rowStart+handCols, len(hand))
		if m.chinese() {
			out.WriteString("手牌：")
		} else {
			out.WriteString("Hand: ")
		}
		out.WriteString(renderHandTileRow(hand[rowStart:rowEnd], selected-rowStart, unicode))
		out.WriteString("\n")
	}
	out.WriteString("\n")
	return out.String()
}

func renderHandTileRow(hand []game.Tile, selected int, unicode bool) string {
	var out strings.Builder
	for i, tile := range hand {
		out.WriteString(renderTileCell(i, tile, i == selected, unicode))
	}
	return out.String()
}

func selectedHandText(hand []game.Tile, selected int, unicode bool) string {
	if selected < 0 || selected >= len(hand) {
		return "-"
	}
	return selectedTileText(selected, hand[selected], unicode)
}

func renderTileCell(index int, tile game.Tile, selected bool, unicode bool) string {
	label := game.TileLabel(tile, unicode)
	content := centeredCardContent(label, handCellW)
	if selected {
		return styleSelectedTile(content)
	}
	return styledCardContent(tile, label, content)
}

func selectedTileText(index int, tile game.Tile, unicode bool) string {
	return fmt.Sprintf("[%02d] %s (%s)", index+1, game.TileLabel(tile, unicode), tile.String())
}

func centeredCardContent(label string, width int) string {
	labelWidth := visibleWidth(label)
	if labelWidth >= width {
		return label
	}
	left := (width - labelWidth) / 2
	right := width - labelWidth - left
	return strings.Repeat(" ", left) + label + strings.Repeat(" ", right)
}

func styledCardContent(tile game.Tile, label string, content string) string {
	return strings.Replace(content, label, styleMahjongTile(tile, label, false), 1)
}

var gameOverItems = []string{"Restart", "Main Menu", "Quit"}

func gameOverItemLabel(m Model, index int) string {
	if !m.chinese() {
		return gameOverItems[index]
	}
	switch index {
	case 0:
		return "重新开始"
	case 1:
		return "返回菜单"
	case 2:
		return "退出"
	default:
		return gameOverItems[index]
	}
}

func renderGameOver(m Model) string {
	if m.Online {
		return renderOnlineGameOver(m)
	}
	if m.Game == nil {
		return "GAME OVER\n"
	}
	var out strings.Builder
	if m.chinese() {
		out.WriteString(styleTitle("对局结束") + "\n")
		out.WriteString(fmt.Sprintf("结果：%s\n", m.Game.Reason))
		out.WriteString(fmt.Sprintf("事件：%d\n", len(m.Game.Events)))
		out.WriteString("可回放事件日志：是\n\n")
	} else {
		out.WriteString(styleTitle("GAME OVER") + "\n")
		out.WriteString(fmt.Sprintf("Result: %s\n", m.Game.Reason))
		out.WriteString(fmt.Sprintf("Events: %d\n", len(m.Game.Events)))
		out.WriteString("Replay-ready event log: yes\n\n")
	}
	for i, item := range gameOverItems {
		prefix := "  "
		label := item
		if m.chinese() {
			label = gameOverItemLabel(m, i)
		}
		if i == m.GameOverIndex {
			prefix = "> "
			out.WriteString(styleSelectedTile(prefix+label) + "\n")
			continue
		}
		out.WriteString(prefix + label + "\n")
	}
	if m.chinese() {
		out.WriteString("\n" + styleSectionTitle("操作") + "\n")
		out.WriteString(styleMuted("上下选择 | 回车确认 | R 重开 | M 菜单 | Q 退出") + "\n")
	} else {
		out.WriteString("\n" + styleSectionTitle("Controls") + "\n")
		out.WriteString(styleMuted("Up/Down choose | Enter confirm | R restart | M menu | Q quit") + "\n")
	}
	return out.String()
}

func renderOnlineGameOver(m Model) string {
	snapshot := m.OnlineSnapshot
	var out strings.Builder
	if m.chinese() {
		out.WriteString(styleTitle("对局结束") + "\n")
		out.WriteString(fmt.Sprintf("房间：%s\n", m.OnlineRoomCode))
		out.WriteString(fmt.Sprintf("结果：%s\n", snapshot.Reason))
		if snapshot.Winner >= 0 {
			out.WriteString(fmt.Sprintf("赢家：座位 %d\n", snapshot.Winner+1))
		} else {
			out.WriteString("赢家：-\n")
		}
		out.WriteString(fmt.Sprintf("事件：%d\n", len(snapshot.Events)))
		out.WriteString("可回放事件日志：是\n\n")
	} else {
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
	}
	for i, item := range gameOverItems {
		prefix := "  "
		label := item
		if m.chinese() {
			label = gameOverItemLabel(m, i)
		}
		if i == m.GameOverIndex {
			prefix = "> "
			out.WriteString(styleSelectedTile(prefix+label) + "\n")
			continue
		}
		out.WriteString(prefix + label + "\n")
	}
	if m.chinese() {
		out.WriteString("\n" + styleSectionTitle("操作") + "\n")
		out.WriteString(styleMuted("上下选择 | 回车确认 | M 菜单 | Q 退出") + "\n")
	} else {
		out.WriteString("\n" + styleSectionTitle("Controls") + "\n")
		out.WriteString(styleMuted("Up/Down choose | Enter confirm | M menu | Q quit") + "\n")
	}
	return out.String()
}
