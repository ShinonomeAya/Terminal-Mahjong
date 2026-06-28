package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"mahjong/internal/game"
)

const wideTableMinWidth = 110

const (
	wideTableBodyWidth = 88
	wideSeatWidth      = 20
	wideCenterWidth    = 38
	tacticalRailWidth  = 20
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
	rail := renderTacticalPlaceholder(m)
	return lipgloss.JoinHorizontal(lipgloss.Top, table, "  ", rail) + "\n"
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
	return stylePanelWidth("", body, wideSeatWidth)
}

func renderWideCenter(m Model, state tableViewState) string {
	var lines []string
	for _, player := range state.Snapshot.Players {
		line := game.FormatTileLabels(recentTiles(player.Discards, 6), m.UnicodeTiles)
		if line == "" {
			line = "-"
		}
		lines = append(lines, line)
	}
	if m.chinese() {
		return "牌河\n" + strings.Join(lines, "\n")
	}
	return "Discard rivers\n" + strings.Join(lines, "\n")
}

func renderWideHandAndActions(m Model, state tableViewState) string {
	player := state.Snapshot.Players[state.ViewerSeat]
	hand := stylePanelWidth(handTitle(m), renderHand(m, player.Hand, m.SelectedIndex, m.UnicodeTiles), wideTableBodyWidth)
	actions := styleSectionTitle(controlsTitle(m)) + "\n" + styleMuted(tableControls(m))
	return hand + "\n" + actions
}

func renderTacticalPlaceholder(m Model) string {
	if m.chinese() {
		return stylePanelWidth("战术分析", "向听：-\n有效牌：-\n改良牌：-\n\n最近事件\n-", tacticalRailWidth)
	}
	return stylePanelWidth("Tactical Analysis", "Shanten: -\nEffective: -\nImprovements: -\n\nRecent Events\n-", tacticalRailWidth)
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
