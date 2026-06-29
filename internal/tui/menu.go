package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"mahjong/internal/game"
)

func updateMenu(m Model, key tea.KeyMsg) (tea.Model, tea.Cmd) {
	items := menuLabels(m)
	switch key.Type {
	case tea.KeyDown:
		m.MenuIndex = (m.MenuIndex + 1) % len(items)
	case tea.KeyUp:
		m.MenuIndex = (m.MenuIndex + len(items) - 1) % len(items)
	case tea.KeyEnter:
		switch m.MenuIndex {
		case 0:
			m.LocalMatch = newStartedMatchWithRules(m.SelectedMode, selectedRuleConfig(m))
			m = syncLocalRound(m)
			m.LastReplayPath = ""
			m.Online = false
			m.NetworkStatus = NetworkLocal
			m.Screen = ScreenTable
		case 1:
			m.NetworkStatus = NetworkWaiting
			m.StatusLine = "Connecting online room..."
			return m, createOnlineRoomCmd(m)
		case 2:
			m.NetworkStatus = NetworkWaiting
			m.StatusLine = "Loading online rooms..."
			return m, listOnlineRoomsCmd(m)
		case 3:
			m.Screen = ScreenJoinOnline
			m.StatusLine = ""
		case 4:
			m.NetworkStatus = NetworkReconnecting
			m.ReconnectAttempt = 1
			m.ReconnectMax = 5
			m.StatusLine = "Reconnecting online room..."
			return m, reconnectOnlineCmd(m)
		case 5:
			m = toggleSelectedMode(m)
		case 6:
			m = toggleSelectedRiichiRedFives(m)
		case 7:
			m.Screen = ScreenHelp
		case 8:
			m = toggleLanguage(m)
		case 9:
			return m, tea.Quit
		}
	}
	return m, nil
}

func toggleSelectedMode(m Model) Model {
	switch m.SelectedMode {
	case game.ModeRiichi:
		m.SelectedMode = game.ModeMCR
	case game.ModeMCR:
		m.SelectedMode = game.ModeCompatibility
	default:
		m.SelectedMode = game.ModeRiichi
	}
	return m
}

func toggleSelectedRiichiRedFives(m Model) Model {
	if m.SelectedRiichiRedFives == 0 {
		m.SelectedRiichiRedFives = 3
		return m
	}
	m.SelectedRiichiRedFives = 0
	return m
}

func updateOnlineRooms(m Model, key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.Type {
	case tea.KeyEsc:
		m.Screen = ScreenMenu
	case tea.KeyDown:
		if len(m.OnlineRooms) > 0 {
			m.RoomIndex = (m.RoomIndex + 1) % len(m.OnlineRooms)
		}
	case tea.KeyUp:
		if len(m.OnlineRooms) > 0 {
			m.RoomIndex = (m.RoomIndex + len(m.OnlineRooms) - 1) % len(m.OnlineRooms)
		}
	case tea.KeyEnter:
		if len(m.OnlineRooms) == 0 {
			m.StatusLine = "No waiting rooms"
			return m, nil
		}
		m.JoinRoomCode = m.OnlineRooms[m.RoomIndex].Code
		m.NetworkStatus = NetworkWaiting
		m.StatusLine = "Joining room " + m.JoinRoomCode + "..."
		return m, joinOnlineRoomCmd(m)
	}
	switch key.String() {
	case "r":
		m.StatusLine = "Refreshing rooms..."
		return m, listOnlineRoomsCmd(m)
	}
	return m, nil
}

func updateJoinOnline(m Model, key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.Type {
	case tea.KeyEsc:
		m.Screen = ScreenMenu
	case tea.KeyBackspace:
		if len(m.JoinRoomCode) > 0 {
			m.JoinRoomCode = m.JoinRoomCode[:len(m.JoinRoomCode)-1]
		}
	case tea.KeyEnter:
		if len(m.JoinRoomCode) == 0 {
			m.StatusLine = "Room code is required"
			return m, nil
		}
		m.NetworkStatus = NetworkWaiting
		m.StatusLine = "Joining room " + m.JoinRoomCode + "..."
		return m, joinOnlineRoomCmd(m)
	case tea.KeyRunes:
		for _, r := range key.Runes {
			if r >= '0' && r <= '9' && len(m.JoinRoomCode) < 6 {
				m.JoinRoomCode += string(r)
			}
		}
	}
	return m, nil
}

func updateHelp(m Model, key tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Type == tea.KeyEsc || key.String() == "q" {
		m.Screen = ScreenMenu
	}
	return m, nil
}

func renderMenu(m Model) string {
	var out strings.Builder
	items := menuLabels(m)
	if m.chinese() {
		out.WriteString(styleTitle("终端麻将") + "\n\n")
		out.WriteString(styleSectionTitle("开始菜单") + "\n")
	} else {
		out.WriteString(styleTitle("TERMINAL MAHJONG") + "\n\n")
		out.WriteString(styleSectionTitle("Menu") + "\n")
	}
	for i, item := range items {
		prefix := "  "
		if i == m.MenuIndex {
			prefix = "> "
			out.WriteString(styleSelectedTile(prefix+item) + "\n")
			continue
		}
		out.WriteString(fmt.Sprintf("%s%s\n", prefix, item))
	}
	if m.chinese() {
		out.WriteString("\n" + styleSectionTitle("操作") + "\n")
		out.WriteString(styleMuted("上下选择 | 回车确认 | Q 退出") + "\n")
	} else {
		out.WriteString("\n" + styleSectionTitle("Controls") + "\n")
		out.WriteString(styleMuted("Up/Down choose | Enter confirm | Q quit") + "\n")
	}
	return out.String()
}

func renderJoinOnline(m Model) string {
	var out strings.Builder
	if m.chinese() {
		out.WriteString(styleTitle("加入联网房间") + "\n\n")
		out.WriteString(styleSectionTitle("房间号") + "\n")
	} else {
		out.WriteString(styleTitle("JOIN ONLINE ROOM") + "\n\n")
		out.WriteString(styleSectionTitle("Room Code") + "\n")
	}
	code := m.JoinRoomCode
	if code == "" {
		code = "______"
	}
	out.WriteString(styleSelectedTile(code) + "\n\n")
	out.WriteString(renderStatus(m))
	if m.chinese() {
		out.WriteString("\n" + styleSectionTitle("操作") + "\n")
		out.WriteString(styleMuted("数字输入房间号 | 退格删除 | 回车加入 | Esc 返回菜单") + "\n")
	} else {
		out.WriteString("\n" + styleSectionTitle("Controls") + "\n")
		out.WriteString(styleMuted("Digits type room code | Backspace edit | Enter join | Esc menu") + "\n")
	}
	return out.String()
}

func renderOnlineRooms(m Model) string {
	var out strings.Builder
	if m.chinese() {
		out.WriteString(styleTitle("联网房间") + "\n\n")
		out.WriteString(styleSectionTitle("等待中的房间") + "\n")
	} else {
		out.WriteString(styleTitle("ONLINE ROOMS") + "\n\n")
		out.WriteString(styleSectionTitle("Waiting Rooms") + "\n")
	}
	if len(m.OnlineRooms) == 0 {
		if m.chinese() {
			out.WriteString(styleMuted("没有等待中的房间") + "\n")
		} else {
			out.WriteString(styleMuted("No waiting rooms") + "\n")
		}
	} else {
		for i, room := range m.OnlineRooms {
			line := fmt.Sprintf("%s  players:%d ready:%d wall:%d", room.Code, room.Occupied, room.Ready, room.Wall)
			if m.chinese() {
				line = fmt.Sprintf("%s  玩家:%d 准备:%d 牌墙:%d", room.Code, room.Occupied, room.Ready, room.Wall)
			}
			if i == m.RoomIndex {
				out.WriteString(styleSelectedTile("> "+line) + "\n")
			} else {
				out.WriteString("  " + line + "\n")
			}
		}
	}
	out.WriteString("\n" + renderStatus(m))
	if m.chinese() {
		out.WriteString("\n" + styleSectionTitle("操作") + "\n")
		out.WriteString(styleMuted("上下选择 | 回车加入 | R 刷新 | Esc 返回菜单") + "\n")
	} else {
		out.WriteString("\n" + styleSectionTitle("Controls") + "\n")
		out.WriteString(styleMuted("Up/Down choose | Enter join | R refresh | Esc menu") + "\n")
	}
	return out.String()
}

func renderHelp(m Model) string {
	if m.chinese() {
		return strings.Join([]string{
			styleTitle("终端麻将帮助"),
			"",
			styleSectionTitle("键盘"),
			"←/→ 选择手牌",
			"回车/空格 打出选中牌",
			"Q 退出当前对局",
			"",
			styleSectionTitle("鼠标"),
			"单击选择手牌",
			"再次单击打出选中牌",
			"",
			styleSectionTitle("操作"),
			styleMuted("Esc 返回菜单"),
			"",
		}, "\n")
	}
	return strings.Join([]string{
		styleTitle("TERMINAL MAHJONG HELP"),
		"",
		styleSectionTitle("Keyboard"),
		"Left/Right select tile",
		"Enter/Space discard selected tile",
		"Q quit current game",
		"",
		styleSectionTitle("Mouse"),
		"Click selects a hand tile",
		"Second click discards the selected tile",
		"",
		styleSectionTitle("Controls"),
		styleMuted("Esc returns to menu"),
		"",
	}, "\n")
}
