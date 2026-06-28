package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"mahjong/internal/game"
)

const wideTableMinWidth = 110

const (
	wideTableBodyWidth  = 88
	wideSeatWidth       = 20
	wideCenterWidth     = 38
	tacticalRailWidth   = 20
	mediumTableMinWidth = 80
)

func renderWideTable(m Model, state tableViewState) string {
	if len(state.Snapshot.Players) == 0 {
		return waitingSnapshotText(m) + "\n"
	}
	table := lipgloss.JoinVertical(
		lipgloss.Left,
		renderTableHeader(m, state),
		centerLine(wideTableBodyWidth+4, renderWideSeat(m, state, (state.ViewerSeat+2)%4, wideSeatLabel(m, "opposite"))),
		renderWideMiddle(m, state),
		centerLine(wideTableBodyWidth+4, renderWideSeat(m, state, state.ViewerSeat, wideSeatLabel(m, "self"))),
		renderWideHandAndActions(m, state),
	)
	rail := renderTacticalRail(m, tacticalViewFor(m, state))
	return lipgloss.JoinHorizontal(lipgloss.Top, table, "  ", rail) + "\n"
}

func renderTacticalFallback(m Model, state tableViewState, board string) string {
	if m.Width < mediumTableMinWidth && (m.Width <= 0 || !m.ShowTactical) {
		return board
	}
	withRail := strings.TrimRight(board, "\n") + "\n" + renderTacticalRail(m, tacticalViewFor(m, state)) + "\n"
	if !m.ShowTactical && m.Height > 0 && lipgloss.Height(withRail) > m.Height {
		return board
	}
	return withRail
}

func renderTableHeader(m Model, state tableViewState) string {
	mode := ruleModeName(m, state.Mode)
	round := fmt.Sprintf("%d", state.Snapshot.HandNumber)
	if state.Snapshot.HandNumber == 0 {
		round = "1"
	}
	title := centerLine(wideTableBodyWidth+4, styleTitle(appTitle(m)))
	if m.chinese() {
		return title + "\n" + styleSectionTitle(mode) + styleMuted(fmt.Sprintf("  第%s局  庄家:%d  牌墙:%d  %s", round, state.Snapshot.Dealer+1, state.Snapshot.WallCount, renderNetworkStatus(m)))
	}
	return title + "\n" + styleSectionTitle(mode) + styleMuted(fmt.Sprintf("  Hand %s  Dealer:%d  Wall:%d  %s", round, state.Snapshot.Dealer+1, state.Snapshot.WallCount, renderNetworkStatus(m)))
}

func renderWideMiddle(m Model, state tableViewState) string {
	left := renderWideSeat(m, state, (state.ViewerSeat+1)%4, wideSeatLabel(m, "left"))
	center := stylePanelWidth(tableTitle(m), renderWideCenter(m, state), wideCenterWidth)
	right := renderWideSeat(m, state, (state.ViewerSeat+3)%4, wideSeatLabel(m, "right"))
	return lipgloss.JoinHorizontal(lipgloss.Center, left, "  ", center, "  ", right)
}

func renderWideSeat(m Model, state tableViewState, seat int, label string) string {
	if seat < 0 || seat >= len(state.Snapshot.Players) {
		return ""
	}
	player := state.Snapshot.Players[seat]
	active := state.Snapshot.Current == seat
	name := playerName(m, player.Name)
	if name == "" {
		name = "-"
	}
	marker := " "
	if active {
		marker = ">"
	}
	body := fmt.Sprintf("%s %s  %s", marker, label, name)
	points := state.Match.Points[seat]
	if seat == state.ViewerSeat {
		if m.chinese() {
			body += fmt.Sprintf("  %d张", playerHandCount(player))
		} else {
			body += fmt.Sprintf("  %d tiles", playerHandCount(player))
		}
	} else if m.chinese() {
		body += fmt.Sprintf("\n手牌:%d", playerHandCount(player))
	} else {
		body += fmt.Sprintf("\nhand:%d", playerHandCount(player))
	}
	if points != 0 {
		body += fmt.Sprintf("  %d", points)
	}
	if m.chinese() {
		body += fmt.Sprintf("\n副露:%d", len(player.Melds))
		if len(player.Flowers) > 0 {
			body += fmt.Sprintf("  花:%d", len(player.Flowers))
		}
		if state.Snapshot.Riichi != nil && state.Snapshot.Riichi.Declarations[seat] == game.RiichiAccepted {
			body += "  立直"
		}
	} else {
		body += fmt.Sprintf("\nmelds:%d", len(player.Melds))
		if len(player.Flowers) > 0 {
			body += fmt.Sprintf("  flowers:%d", len(player.Flowers))
		}
		if state.Snapshot.Riichi != nil && state.Snapshot.Riichi.Declarations[seat] == game.RiichiAccepted {
			body += "  RIICHI"
		}
	}
	return stylePanelWidth("", body, wideSeatWidth)
}

func renderWideCenter(m Model, state tableViewState) string {
	var lines []string
	for seat, player := range state.Snapshot.Players {
		label := fmt.Sprintf("%d", seat+1)
		if seat == state.ViewerSeat {
			label = wideSeatLabel(m, "self")
		}
		river := renderDiscardRiver(m, player.Discards)
		lines = append(lines, label+" "+river)
	}
	if m.chinese() {
		return "牌河\n" + strings.Join(lines, "\n")
	}
	return "Discard rivers\n" + strings.Join(lines, "\n")
}

func renderDiscardRiver(m Model, tiles []game.Tile) string {
	if len(tiles) == 0 {
		return "-"
	}
	const tilesPerRow = 6
	rows := make([]string, 0, (len(tiles)+tilesPerRow-1)/tilesPerRow)
	for start := 0; start < len(tiles); start += tilesPerRow {
		end := start + tilesPerRow
		if end > len(tiles) {
			end = len(tiles)
		}
		cells := make([]string, 0, end-start)
		for index := start; index < end; index++ {
			label := game.TileLabel(tiles[index], m.UnicodeTiles)
			if index == len(tiles)-1 {
				label = "◆" + label
			}
			cells = append(cells, label)
		}
		rows = append(rows, strings.Join(cells, " "))
	}
	return strings.Join(rows, "\n")
}

func renderWideHandAndActions(m Model, state tableViewState) string {
	player := state.Snapshot.Players[state.ViewerSeat]
	hand := stylePanelWidth(handTitle(m), renderHand(m, player.Hand, m.SelectedIndex, m.UnicodeTiles), wideTableBodyWidth)
	actions := styleSectionTitle(controlsTitle(m)) + "\n" + styleMuted(tableControls(m))
	return hand + "\n" + actions
}

func wideSeatLabel(m Model, position string) string {
	if !m.chinese() {
		switch position {
		case "opposite":
			return "Opposite"
		case "left":
			return "Next"
		case "right":
			return "Previous"
		default:
			return "You"
		}
	}
	switch position {
	case "opposite":
		return "对家"
	case "left":
		return "下家"
	case "right":
		return "上家"
	default:
		return "你"
	}
}
