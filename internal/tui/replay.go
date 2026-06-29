package tui

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"mahjong/internal/game"
	"mahjong/internal/replay"
)

type replaySavedMsg struct {
	Path     string
	ReplayID string
}

type replaySaveErrorMsg struct {
	ReplayID string
	Err      error
}

type replayListMsg struct {
	Entries []replay.Entry
	Issues  []replay.FileIssue
}

type replayLoadedMsg struct {
	File game.ReplayFile
}

type replayListErrorMsg struct {
	Err error
}

type replayLoadErrorMsg struct {
	Err error
}

type replayTickMsg struct{}

func saveCompletedReplayCmd(match *game.Match, dir string) tea.Cmd {
	return func() tea.Msg {
		file, err := match.CompletedReplay(
			replay.ApplicationVersion(),
			time.Now().UTC(),
			replayParticipants(match),
		)
		if err != nil {
			return replaySaveErrorMsg{Err: err}
		}
		path, err := replay.Save(dir, file)
		if err != nil {
			return replaySaveErrorMsg{Err: err}
		}
		return replaySavedMsg{Path: path}
	}
}

func saveReplayFileCmd(file game.ReplayFile, dir string) tea.Cmd {
	return func() tea.Msg {
		path, err := replay.Save(dir, file)
		if err != nil {
			return replaySaveErrorMsg{ReplayID: file.ReplayID, Err: err}
		}
		return replaySavedMsg{Path: path, ReplayID: file.ReplayID}
	}
}

func listReplaysCmd(dir string) tea.Cmd {
	return func() tea.Msg {
		entries, issues, err := replay.List(dir)
		if err != nil {
			return replayListErrorMsg{Err: err}
		}
		return replayListMsg{Entries: entries, Issues: issues}
	}
}

func loadReplayCmd(path string) tea.Cmd {
	return func() tea.Msg {
		file, err := replay.Load(path)
		if err != nil {
			return replayLoadErrorMsg{Err: err}
		}
		return replayLoadedMsg{File: file}
	}
}

func replayTickCmd() tea.Cmd {
	return tea.Tick(750*time.Millisecond, func(time.Time) tea.Msg {
		return replayTickMsg{}
	})
}

func applyReplayTick(m Model) (Model, tea.Cmd) {
	if m.ReplayFile == nil || len(m.ReplayFile.Frames) == 0 {
		m.ReplayPlaying = false
		m.ReplayFrame = 0
		return m, nil
	}
	last := len(m.ReplayFile.Frames) - 1
	if m.ReplayFrame < 0 {
		m.ReplayFrame = 0
	}
	if m.ReplayFrame >= last {
		m.ReplayFrame = last
		m.ReplayPlaying = false
		return m, nil
	}
	m.ReplayFrame++
	if m.ReplayFrame >= last {
		m.ReplayPlaying = false
		return m, nil
	}
	return m, replayTickCmd()
}

func updateReplayBrowser(m Model, key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.Type {
	case tea.KeyEsc:
		m.Screen = ScreenMenu
		m.StatusLine = ""
	case tea.KeyDown:
		if m.ReplayIndex < len(m.ReplayEntries)-1 {
			m.ReplayIndex++
		}
	case tea.KeyUp:
		if m.ReplayIndex > 0 {
			m.ReplayIndex--
		}
	case tea.KeyEnter:
		if len(m.ReplayEntries) > 0 {
			return m, loadReplayCmd(m.ReplayEntries[m.ReplayIndex].Path)
		}
	}
	if key.String() == "r" {
		m.StatusLine = replayLoadingStatus(m)
		return m, listReplaysCmd(m.ReplayDir)
	}
	return m, nil
}

func renderReplayBrowser(m Model) string {
	var out strings.Builder
	if m.chinese() {
		out.WriteString(styleTitle("回放") + "\n\n")
		out.WriteString(styleSectionTitle("已保存对局") + "\n")
	} else {
		out.WriteString(styleTitle("REPLAYS") + "\n\n")
		out.WriteString(styleSectionTitle("Saved Matches") + "\n")
	}
	if len(m.ReplayEntries) == 0 {
		if m.chinese() {
			out.WriteString(styleMuted("暂无可用回放") + "\n")
		} else {
			out.WriteString(styleMuted("No valid replays") + "\n")
		}
	} else {
		for index, entry := range m.ReplayEntries {
			label := replayEntryLabel(m, entry)
			prefix := "  "
			if index == m.ReplayIndex {
				prefix = "> "
				out.WriteString(styleSelectedTile(prefix+strings.ReplaceAll(label, "\n", "\n  ")) + "\n")
			} else {
				out.WriteString(prefix + strings.ReplaceAll(label, "\n", "\n  ") + "\n")
			}
		}
	}
	if m.StatusLine != "" {
		out.WriteString("\n" + styleStatus(m.StatusLine) + "\n")
	}
	if m.chinese() {
		out.WriteString("\n" + styleSectionTitle("操作") + "\n")
		out.WriteString(styleMuted("上下选择 | 回车打开 | R 刷新 | Esc 返回菜单") + "\n")
	} else {
		out.WriteString("\n" + styleSectionTitle("Controls") + "\n")
		out.WriteString(styleMuted("Up/Down choose | Enter open | R refresh | Esc menu") + "\n")
	}
	return out.String()
}

func replayEntryLabel(m Model, entry replay.Entry) string {
	width := m.Width
	if width <= 0 {
		width = 100
	}
	contentWidth := width - 2
	if contentWidth < 20 {
		contentWidth = 20
	}
	framesLabel := fmt.Sprintf("frames:%d", entry.Frames)
	idLabel := "ID: " + entry.ReplayID
	fileLabel := "File: "
	if m.chinese() {
		framesLabel = fmt.Sprintf("帧:%d", entry.Frames)
		fileLabel = "文件: "
	}
	summary := fmt.Sprintf("%s | %s | %s", entry.CreatedAt.Local().Format("2006-01-02 15:04:05"), entry.Mode, framesLabel)
	nameWidth := contentWidth - visibleWidth(fileLabel)
	return strings.Join([]string{
		summary,
		idLabel,
		fileLabel + truncateVisible(filepath.Base(entry.Path), nameWidth),
	}, "\n")
}

func renderReplayViewer(m Model) string {
	frame, ok := currentReplayFrame(m)
	if !ok {
		if m.chinese() {
			return styleTitle("回放") + "\n\n" + styleStatus("回放数据不可用") + "\n"
		}
		return styleTitle("REPLAY") + "\n\n" + styleStatus("Replay data is unavailable") + "\n"
	}
	state := "paused"
	title := "REPLAY"
	modeLabel := "Mode"
	frameLabel := "Frame"
	idLabel := "Replay ID"
	createdLabel := "Created"
	roundLabel := "Round snapshot"
	controls := "Left/Right frame | Home/End boundary | Space play/pause | Tab details | Esc browser"
	if m.ReplayPlaying {
		state = "playing"
	}
	if m.chinese() {
		title = "回放"
		modeLabel = "模式"
		frameLabel = "帧"
		idLabel = "回放 ID"
		createdLabel = "创建时间"
		roundLabel = "对局快照"
		state = "已暂停"
		if m.ReplayPlaying {
			state = "播放中"
		}
		controls = "左右切帧 | Home/End 首尾 | 空格播放/暂停 | Tab 详情 | Esc 返回"
	}

	var out strings.Builder
	out.WriteString(styleTitle(title) + "\n")
	out.WriteString(fmt.Sprintf(
		"%s: %s | %s: %d/%d | %s\n",
		modeLabel,
		ruleModeName(m, m.ReplayFile.Mode),
		frameLabel,
		m.ReplayFrame+1,
		len(m.ReplayFile.Frames),
		state,
	))
	out.WriteString(fmt.Sprintf("%s: %s\n", idLabel, m.ReplayFile.ReplayID))
	out.WriteString(fmt.Sprintf("%s: %s\n\n", createdLabel, m.ReplayFile.CreatedAt.Local().Format("2006-01-02 15:04:05")))
	out.WriteString(styleSectionTitle(roundLabel) + "\n")
	out.WriteString(fmt.Sprintf(
		"wall:%d current:%d events:%d\n",
		frame.Match.Round.WallCount,
		frame.Match.Round.Current+1,
		len(frame.Match.Round.Events),
	))
	for seat, player := range frame.Match.Round.Players {
		out.WriteString(fmt.Sprintf(
			"%d. %s  points:%d  hand:%d  discards:%d\n",
			seat+1,
			playerName(m, player.Name),
			frame.Match.Points[seat],
			len(player.Hand),
			len(player.Discards),
		))
	}
	if m.ReplayFrame == len(m.ReplayFile.Frames)-1 {
		if m.chinese() {
			out.WriteString("\n" + styleSectionTitle("最终积分") + "\n")
		} else {
			out.WriteString("\n" + styleSectionTitle("Final Standings") + "\n")
		}
		out.WriteString(fmt.Sprintf(
			"%d | %d | %d | %d\n",
			m.ReplayFile.FinalStandings[0],
			m.ReplayFile.FinalStandings[1],
			m.ReplayFile.FinalStandings[2],
			m.ReplayFile.FinalStandings[3],
		))
	}
	out.WriteString("\n" + styleSectionTitle(replayControlsTitle(m)) + "\n")
	out.WriteString(styleMuted(controls) + "\n")
	return out.String()
}

func currentReplayFrame(m Model) (game.ReplayFrame, bool) {
	if m.ReplayFile == nil || m.ReplayFrame < 0 || m.ReplayFrame >= len(m.ReplayFile.Frames) {
		return game.ReplayFrame{}, false
	}
	return m.ReplayFile.Frames[m.ReplayFrame], true
}

func replayControlsTitle(m Model) string {
	if m.chinese() {
		return "操作"
	}
	return "Controls"
}

func truncateVisible(text string, width int) string {
	if width <= 0 {
		return ""
	}
	if visibleWidth(text) <= width {
		return text
	}
	ellipsis := "..."
	target := width - visibleWidth(ellipsis)
	if target <= 0 {
		return ellipsis[:width]
	}
	var out strings.Builder
	for _, r := range text {
		next := out.String() + string(r)
		if visibleWidth(next) > target {
			break
		}
		out.WriteRune(r)
	}
	return out.String() + ellipsis
}

func replayLoadingStatus(m Model) string {
	if m.chinese() {
		return "正在加载回放..."
	}
	return "Loading replays..."
}

func replayListStatus(m Model) string {
	if len(m.ReplayIssues) == 0 {
		if m.chinese() {
			return fmt.Sprintf("已加载 %d 个回放", len(m.ReplayEntries))
		}
		return fmt.Sprintf("Loaded %d replays", len(m.ReplayEntries))
	}
	if m.chinese() {
		return fmt.Sprintf("已加载 %d 个回放，跳过 %d 个损坏文件", len(m.ReplayEntries), len(m.ReplayIssues))
	}
	return fmt.Sprintf("Loaded %d replays; skipped %d corrupt files", len(m.ReplayEntries), len(m.ReplayIssues))
}

func replayErrorStatus(m Model, err error) string {
	if m.chinese() {
		return "回放读取失败: " + err.Error()
	}
	return "Replay read failed: " + err.Error()
}

func replayParticipants(match *game.Match) []game.ReplayParticipant {
	participants := make([]game.ReplayParticipant, 0, len(match.Round.Players))
	for seat, player := range match.Round.Players {
		participants = append(participants, game.ReplayParticipant{
			Seat: seat,
			ID:   strconv.Itoa(seat),
			Name: player.Name,
		})
	}
	return participants
}
