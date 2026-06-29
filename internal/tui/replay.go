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
	if _, ok := currentReplayFrame(m); !ok {
		if m.chinese() {
			return styleTitle("回放") + "\n\n" + styleStatus("回放数据不可用") + "\n"
		}
		return styleTitle("REPLAY") + "\n\n" + styleStatus("Replay data is unavailable") + "\n"
	}
	state := tableStateFor(m)
	if m.Width >= wideTableMinWidth {
		return renderWideTable(m, state)
	}
	return renderReplayCompactTable(m, state)
}

func renderReplayCompactTable(m Model, state tableViewState) string {
	if len(state.Snapshot.Players) != 4 {
		return replayViewerTitle(m) + "\n"
	}
	topSeat := renderReplaySeatPanel(m, state, 2, "Opposite")
	leftSeat := renderReplaySeatPanel(m, state, 1, "Left")
	rightSeat := renderReplaySeatPanel(m, state, 3, "Right")
	center := stylePanelWidth(tableTitle(m), renderWideCenter(m, state), 38)
	middle := renderTableMiddle(m, leftSeat, center, rightSeat)
	player := state.Snapshot.Players[state.ViewerSeat]
	hand := stylePanelWidth(handTitle(m), renderHand(m, player.Hand, -1, m.UnicodeTiles), handPanelWidth(m))
	board := renderBoardFrame(
		m,
		styleMuted(renderReplayMeta(m, state)),
		topSeat,
		middle,
		hand,
		renderReplayControls(m),
	)
	frame, _ := currentReplayFrame(m)
	return strings.TrimRight(board, "\n") + "\n" + renderReplayDetailRail(m, *m.ReplayFile, frame) + "\n"
}

func renderReplaySeatPanel(m Model, state tableViewState, seat int, position string) string {
	player := state.Snapshot.Players[seat]
	return renderSeatPanel(
		m,
		seatLabel(m, position),
		player.Name,
		len(player.Hand),
		meldSummary(player.Melds),
		game.FormatTileLabels(recentTiles(player.Discards, 4), m.UnicodeTiles),
		state.Snapshot.Current == seat,
	)
}

func renderReplayMeta(m Model, state tableViewState) string {
	frameState := "paused"
	modeLabel := "Mode"
	frameLabel := "Frame"
	idLabel := "Replay"
	if m.ReplayPlaying {
		frameState = "playing"
	}
	if m.chinese() {
		modeLabel = "模式"
		frameLabel = "帧"
		idLabel = "回放"
		frameState = "已暂停"
		if m.ReplayPlaying {
			frameState = "播放中"
		}
	}
	return fmt.Sprintf(
		"%s:%s  %s:%d/%d  %s  %s:%s",
		modeLabel,
		ruleModeName(m, state.Mode),
		frameLabel,
		m.ReplayFrame+1,
		len(m.ReplayFile.Frames),
		frameState,
		idLabel,
		m.ReplayFile.ReplayID,
	)
}

func replayViewerTitle(m Model) string {
	if m.chinese() {
		return "终端麻将 · 回放"
	}
	return "TERMINAL MAHJONG · REPLAY"
}

func renderReplayControls(m Model) string {
	controls := "←/→ frame | Home/End | Space play/pause | Tab details | Esc library"
	if m.chinese() {
		controls = "←/→ 切帧 | Home/End 首尾 | Space 播放/暂停 | Tab 详情 | Esc 回放库"
	}
	return styleSectionTitle(replayControlsTitle(m)) + "\n" + styleMuted(controls)
}

func renderReplayDetailRail(m Model, file game.ReplayFile, frame game.ReplayFrame) string {
	var lines []string
	if m.chinese() {
		lines = append(lines, "帧详情")
	} else {
		lines = append(lines, "Frame Details")
	}
	command := "-"
	if frame.Command != nil {
		command = string(frame.Command.Kind)
	}
	if m.chinese() {
		lines = append(lines, "命令: "+command)
	} else {
		lines = append(lines, "Command: "+command)
	}
	lines = append(lines, renderReplayIndicators(m, frame.Match.Round)...)
	lines = append(lines, renderReplayNewEvents(m, file, frame)...)
	if m.ReplayShowDetails {
		lines = append(lines, "", renderReplayHands(m, frame.Match.Round.Players))
		lines = append(lines, "", renderReplaySettlement(m, frame))
	} else if frame.Match.Complete {
		lines = append(lines, "", renderReplaySettlement(m, frame))
	} else if m.chinese() {
		lines = append(lines, "", "Tab 查看完整手牌与结算")
	} else {
		lines = append(lines, "", "Tab shows all hands and settlement")
	}
	return stylePanelWidth("", strings.Join(lines, "\n"), tacticalRailWidth)
}

func renderReplayIndicators(m Model, snapshot game.GameSnapshot) []string {
	var lines []string
	if snapshot.Riichi != nil {
		dora := game.FormatTileLabels(snapshot.Riichi.DoraIndicators, m.UnicodeTiles)
		ura := game.FormatTileLabels(snapshot.Riichi.UraIndicators, m.UnicodeTiles)
		if m.chinese() {
			return []string{"宝牌: " + dora, "里宝牌: " + ura}
		}
		return []string{"Dora: " + dora, "Ura: " + ura}
	}
	for _, player := range snapshot.Players {
		if len(player.Flowers) == 0 {
			continue
		}
		label := "Flowers: "
		if m.chinese() {
			label = "花牌: "
		}
		lines = append(lines, playerName(m, player.Name)+" "+label+game.FormatTileLabels(player.Flowers, m.UnicodeTiles))
	}
	return lines
}

func renderReplayNewEvents(m Model, file game.ReplayFile, frame game.ReplayFrame) []string {
	start := 0
	if frame.Index > 0 && frame.Index < len(file.Frames) {
		start = len(file.Frames[frame.Index-1].Match.Round.Events)
	}
	events := frame.Match.Round.Events
	if start > len(events) {
		start = len(events)
	}
	events = events[start:]
	if len(events) > 3 {
		events = events[len(events)-3:]
	}
	title := "New events"
	if m.chinese() {
		title = "新增事件"
	}
	lines := []string{title + ":"}
	if len(events) == 0 {
		return append(lines, "-")
	}
	for _, event := range events {
		lines = append(lines, event.String())
	}
	return lines
}

func renderReplayHands(m Model, players []game.PlayerView) string {
	title := "All hands"
	if m.chinese() {
		title = "全部手牌"
	}
	lines := []string{title}
	for _, player := range players {
		lines = append(lines, playerName(m, player.Name)+":")
		if len(player.Hand) == 0 {
			lines = append(lines, "-")
			continue
		}
		for start := 0; start < len(player.Hand); start += 6 {
			end := min(start+6, len(player.Hand))
			lines = append(lines, game.FormatTileLabels(player.Hand[start:end], m.UnicodeTiles))
		}
	}
	return strings.Join(lines, "\n")
}

func renderReplaySettlement(m Model, frame game.ReplayFrame) string {
	title := "Settlement"
	finalTitle := "Final standings"
	if m.chinese() {
		title = "本局结算"
		finalTitle = "最终积分"
	}
	lines := []string{title}
	switch {
	case frame.Match.LastRiichiSettlement != nil:
		lines = append(lines, formatReplayDeltas(frame.Match.LastRiichiSettlement.Deltas))
	case frame.Match.LastMCRSettlement != nil:
		lines = append(lines, formatReplayDeltas(frame.Match.LastMCRSettlement.Deltas))
	default:
		lines = append(lines, "-")
	}
	if frame.Match.Complete {
		lines = append(lines, finalTitle, formatReplayPoints(frame.Match.Points))
	}
	return strings.Join(lines, "\n")
}

func formatReplayDeltas(values [4]int) string {
	return fmt.Sprintf("%+d %+d %+d %+d", values[0], values[1], values[2], values[3])
}

func formatReplayPoints(values [4]int) string {
	return fmt.Sprintf("%d %d %d %d", values[0], values[1], values[2], values[3])
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
