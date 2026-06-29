package tui

import (
	"fmt"
	"strings"

	"mahjong/internal/game"
)

type Language string

const (
	LanguageChinese Language = "zh"
	LanguageEnglish Language = "en"
)

func (m Model) chinese() bool {
	return m.Language != LanguageEnglish
}

func languageMenuLabel(m Model) string {
	if m.chinese() {
		return "语言：中文"
	}
	return "Language: English"
}

func toggleLanguage(m Model) Model {
	if m.chinese() {
		m.Language = LanguageEnglish
		return m
	}
	m.Language = LanguageChinese
	return m
}

func menuLabels(m Model) []string {
	if !m.chinese() {
		return []string{"Solo Game", "Create Online Room", "Browse Online Rooms", "Join Online Room", "Reconnect Online", ruleModeMenuLabel(m), redFivesMenuLabel(m), "How to Play", "Replays", languageMenuLabel(m), "Quit"}
	}
	return []string{"单机对局", "创建联网房间", "浏览联网房间", "加入联网房间", "断线重连", ruleModeMenuLabel(m), redFivesMenuLabel(m), "玩法说明", "回放", languageMenuLabel(m), "退出"}
}

func ruleModeMenuLabel(m Model) string {
	if !m.chinese() {
		return "Rules: " + ruleModeName(m, m.SelectedMode)
	}
	return "规则：" + ruleModeName(m, m.SelectedMode)
}

func redFivesMenuLabel(m Model) string {
	label := "Red fives: "
	if m.chinese() {
		label = "红五："
	}
	value := "3"
	if m.SelectedRiichiRedFives == 0 {
		value = "off"
		if m.chinese() {
			value = "关闭"
		}
	} else if m.chinese() {
		value = "三张"
	}
	if m.SelectedMode != game.ModeRiichi {
		if m.chinese() {
			return label + value + "（仅日麻）"
		}
		return label + value + " (Riichi only)"
	}
	return label + value
}

func ruleModeName(m Model, mode game.RuleMode) string {
	if !m.chinese() {
		switch mode {
		case game.ModeRiichi:
			return "Riichi"
		case game.ModeMCR:
			return "Chinese Official"
		default:
			return "Classic"
		}
	}
	switch mode {
	case game.ModeRiichi:
		return "日麻"
	case game.ModeMCR:
		return "国标"
	default:
		return "经典"
	}
}

func playerName(m Model, name string) string {
	if !m.chinese() || name == "You" {
		if m.chinese() && name == "You" {
			return "你"
		}
		return name
	}
	if strings.HasPrefix(name, "AI-") {
		return strings.Replace(name, "AI-", "电脑", 1)
	}
	return name
}

func seatLabel(m Model, seat string) string {
	if !m.chinese() {
		return seat + ":"
	}
	switch seat {
	case "Opposite":
		return "对家:"
	case "Left":
		return "左家:"
	case "Right":
		return "右家:"
	default:
		return seat + ":"
	}
}

func statusText(m Model) string {
	if !m.chinese() {
		if m.StatusLine == "" {
			return "Status: Ready"
		}
		return "Status: " + m.StatusLine
	}
	if m.StatusLine == "" {
		return "状态：就绪"
	}
	return "状态：" + localizeStatusLine(m, m.StatusLine)
}

func lastActionText(m Model) string {
	if !m.chinese() {
		if m.StatusLine == "" {
			return "Last Action: Waiting for input"
		}
		return "Last Action: " + m.StatusLine
	}
	if m.StatusLine == "" {
		return "上步：等待操作"
	}
	return "上步：" + localizeStatusLine(m, m.StatusLine)
}

func localizeStatusLine(m Model, text string) string {
	if !m.chinese() {
		return text
	}
	replacer := strings.NewReplacer(
		"Mouse selected", "鼠标选中",
		"Selected", "选中",
		"Discarded", "已打出",
		"Discarding", "正在打出",
		"Winning", "胡牌",
		"Kong", "杠",
		"Riichi", "立直",
		"Ready sent", "已准备",
		"Reconnected", "已重连",
		"Waiting for players to ready", "等待玩家准备",
		"Online room is not ready", "房间尚未开始",
		"Waiting for your turn", "等待轮到你",
		"Win is not available", "当前不能胡",
		"Kong is not available", "当前不能杠",
		"Riichi is not available", "当前不能立直",
		"Rooms found", "找到房间",
		"Connecting online room...", "正在创建联网房间...",
		"Loading online rooms...", "正在加载联网房间...",
		"Reconnecting online room...", "正在重连联网房间...",
		"Refreshing rooms...", "正在刷新房间...",
		"Room code is required", "请输入房间号",
		"No waiting rooms", "没有等待中的房间",
		"Passed claim", "已过",
		"Passing claim", "正在过",
		"Claimed win", "已胡",
		"Claimed pong", "已碰",
		"Claimed chow", "已吃",
		"Claiming win", "正在胡",
		"Claiming pong", "正在碰",
		"Claiming chow", "正在吃",
	)
	if strings.HasPrefix(text, "Joining room ") {
		return strings.Replace(text, "Joining room ", "正在加入房间 ", 1)
	}
	if strings.HasPrefix(text, "Room:") && strings.Contains(text, " Seat:") {
		return strings.NewReplacer("Room:", "房间：", " Seat:", " 座位：").Replace(text)
	}
	return replacer.Replace(text)
}

func actionVerb(m Model, verb string) string {
	if !m.chinese() {
		return verb
	}
	switch verb {
	case "Selected":
		return "选中"
	case "Mouse selected":
		return "鼠标选中"
	case "Discarded":
		return "已打出"
	case "Discarding":
		return "正在打出"
	default:
		return verb
	}
}

func commandLabel(m Model, label string, ready bool) string {
	if !m.chinese() {
		return actionState(label, ready)
	}
	switch label {
	case "[H] Win":
		label = "[H] 胡"
	case "[K] Kong":
		label = "[K] 杠"
	case "[L] Riichi":
		label = "[L] 立直"
	case "[H]Win":
		label = "[H]胡"
	case "[K]Kong":
		label = "[K]杠"
	}
	if ready {
		return styleSelectedTile(label + ":可用")
	}
	return label + ":不可用"
}

func eventKindText(m Model, kind string) string {
	if !m.chinese() {
		return kind
	}
	switch kind {
	case "draw":
		return "摸牌"
	case "discard":
		return "打出"
	case "win":
		return "胡牌"
	case "kong":
		return "杠"
	case "pong":
		return "碰"
	case "chow":
		return "吃"
	default:
		return kind
	}
}

func handTipsText(m Model, tips string) string {
	if !m.chinese() {
		return tips
	}
	return strings.ReplaceAll(tips, "shanten:", "向听：")
}

func roomStateText(m Model, ready int, total int, started bool) string {
	if !m.chinese() {
		status := "Waiting for players"
		if started {
			status = "Started"
		}
		return fmt.Sprintf("Ready: %d/%d  State: %s  Press R ready", ready, total, status)
	}
	status := "等待玩家"
	if started {
		status = "已开始"
	}
	return fmt.Sprintf("准备：%d/%d  状态：%s  按 R 准备", ready, total, status)
}
