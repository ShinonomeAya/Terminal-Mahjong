package tui

import (
	"strings"
	"testing"

	"mahjong/internal/game"
)

func TestTileCellRendersUnselectedUnicodeTile(t *testing.T) {
	tile := mustUITiles(t, "2m")[0]

	cell := renderTileCell(1, tile, false, true)

	if strings.Contains(cell, "[02]") || !strings.Contains(cell, "🀈") || containsTileFrame(cell) {
		t.Fatalf("cell = %q, want bare unicode glyph without visible index or frame", cell)
	}
}

func TestTileCellRendersSelectedUnicodeTile(t *testing.T) {
	tile := mustUITiles(t, "2m")[0]

	cell := renderTileCell(1, tile, true, true)

	if strings.Contains(cell, "[02]") || !strings.Contains(cell, "🀈") || containsTileFrame(cell) {
		t.Fatalf("selected cell = %q, want highlighted bare tile without visible index or frame", cell)
	}
}

func TestTileCellRendersFallbackNotation(t *testing.T) {
	tile := mustUITiles(t, "2m")[0]

	cell := renderTileCell(1, tile, false, false)

	if !strings.Contains(cell, "2m") {
		t.Fatalf("fallback cell = %q, want 2m", cell)
	}
}

func TestTileCellStaysWithinWidthBudget(t *testing.T) {
	tile := mustUITiles(t, "2m")[0]

	cell := renderTileCell(1, tile, true, true)

	if got := visibleWidth(cell); got != handCellW {
		t.Fatalf("cell width = %d, want %d: %q", got, handCellW, cell)
	}
}

func TestSelectedHandTileAlignsInFixedWidthRow(t *testing.T) {
	hand := mustUITiles(t, "1m", "2m", "3m")

	view := renderHand(NewModel(), hand, 1, true)
	lines := strings.Split(strings.TrimRight(view, "\n"), "\n")
	handLine := ""
	for _, line := range lines {
		if strings.Contains(line, "手牌：") {
			handLine = line
			break
		}
	}
	if containsTileFrame(view) {
		t.Fatalf("bare hand should not render tile frames:\n%s", view)
	}
	if got := visibleWidth(strings.TrimPrefix(handLine, "手牌：")); got != len(hand)*handCellW {
		t.Fatalf("hand row width = %d, want %d:\n%s", got, len(hand)*handCellW, view)
	}
}

func TestRenderTableAddsChineseLabels(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable

	view := renderTable(model)

	for _, text := range []string{"终端麻将", "对手", "牌桌", "手牌"} {
		if !strings.Contains(view, text) {
			t.Fatalf("view missing Chinese label %q:\n%s", text, view)
		}
	}
}

func TestRenderHandShowsTilesInOneRowWithoutIndices(t *testing.T) {
	hand := mustUITiles(t, "1m", "2m", "3m", "4m", "5m", "6m", "7m", "8m", "9m", "1p", "2p", "3p", "4p", "5p")

	view := renderHand(NewModel(), hand, 1, true)

	if strings.Contains(view, "[01]") || strings.Contains(view, "[14]") {
		t.Fatalf("hand should not show numeric tile prefixes:\n%s", view)
	}
	if strings.Count(view, "🀇") != 1 || !strings.Contains(view, "🀝") {
		t.Fatalf("hand missing unicode mahjong tiles:\n%s", view)
	}
	if strings.Count(view, "手牌：") != 1 {
		t.Fatalf("hand should render as one row:\n%s", view)
	}
}

func TestRenderTableIncludesUnicodeTiles(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Game.Players[0].Hand = mustUITiles(t, "1m", "1p", "1s")
	model.Screen = ScreenTable
	model.UnicodeTiles = true

	view := renderTable(model)

	if !strings.Contains(view, "🀇") || !strings.Contains(view, "🀙") || !strings.Contains(view, "🀐") {
		t.Fatalf("view does not appear to include unicode tiles:\n%s", view)
	}
}

func TestRenderTableIncludesFallbackLabels(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Game.Players[0].Hand = mustUITiles(t, "1m", "2m", "E")
	model.Screen = ScreenTable
	model.UnicodeTiles = false

	view := renderTable(model)

	if !strings.Contains(view, "1m") || !strings.Contains(view, "2m") || !strings.Contains(view, "E") {
		t.Fatalf("view missing fallback labels:\n%s", view)
	}
}

func TestRenderTableIncludesRecentEventsPanel(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Game.Events = nil
	model.Game.RecordEvent(game.EventDraw, 0, mustUITiles(t, "1m")[0], "")
	model.Game.RecordEvent(game.EventDiscard, 0, mustUITiles(t, "1m")[0], "")
	model.Game.RecordEvent(game.EventDraw, 1, mustUITiles(t, "2m")[0], "")
	model.Game.RecordEvent(game.EventDiscard, 1, mustUITiles(t, "2m")[0], "")
	model.Screen = ScreenTable
	model.UnicodeTiles = false

	view := renderTable(model)

	for _, text := range []string{"最近事件", "01. 你 摸牌 1m", "04. 电脑1 打出 2m"} {
		if !strings.Contains(view, text) {
			t.Fatalf("view missing recent event text %q:\n%s", text, view)
		}
	}
	if strings.Contains(view, "Last:") {
		t.Fatalf("view still uses one-line Last summary:\n%s", view)
	}
}

func TestRenderTableUsesUnicodeInRecentEvents(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Game.Events = nil
	model.Game.RecordEvent(game.EventDiscard, 0, mustUITiles(t, "1m")[0], "")
	model.Screen = ScreenTable
	model.UnicodeTiles = true

	view := renderTable(model)

	if !strings.Contains(view, "01. 你 打出 🀇") {
		t.Fatalf("view missing unicode event tile:\n%s", view)
	}
}

func TestRenderTableIncludesRoundStatusPanel(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Game.Events = nil
	model.Game.RecordEvent(game.EventDraw, 0, mustUITiles(t, "1m")[0], "")
	model.Screen = ScreenTable

	view := renderTable(model)

	for _, text := range []string{"局况", "牌墙:", "轮到: 你", "事件: 1"} {
		if !strings.Contains(view, text) {
			t.Fatalf("view missing round status text %q:\n%s", text, view)
		}
	}
}

func TestRenderTableShowsNetworkStatus(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable
	model.NetworkStatus = NetworkReconnecting
	model.ReconnectAttempt = 2
	model.ReconnectMax = 5

	view := renderTable(model)

	if !strings.Contains(view, "网络：重连中 2/5") {
		t.Fatalf("view missing reconnecting status:\n%s", view)
	}
}

func TestRenderTableDefaultsToLocalNetworkStatus(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable

	view := renderTable(model)

	if !strings.Contains(view, "网络：本地") {
		t.Fatalf("view missing local network status:\n%s", view)
	}
}

func TestRenderTableLocalizesPendingClaimPrompt(t *testing.T) {
	model := localClaimModel(t, game.ClaimPong,
		game.ClaimOption{Kind: game.ClaimPong, Player: 0, Consumed: mustUITiles(t, "3m", "3m")},
	)
	model.UnicodeTiles = false

	chinese := renderTable(model)
	for _, want := range []string{"响应 3m", "[P] 碰", "空格/Esc 过"} {
		if !strings.Contains(chinese, want) {
			t.Fatalf("Chinese claim view missing %q:\n%s", want, chinese)
		}
	}

	model.Language = LanguageEnglish
	english := renderTable(model)
	for _, want := range []string{"Respond to 3m", "[P] Pong", "Space/Esc Pass"} {
		if !strings.Contains(english, want) {
			t.Fatalf("English claim view missing %q:\n%s", want, english)
		}
	}
	if strings.Contains(english, "响应") {
		t.Fatalf("English claim view contains Chinese prompt:\n%s", english)
	}
}

func TestRenderTableLimitsRecentEventsPanel(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Game.Events = nil
	for _, tile := range mustUITiles(t, "1m", "2m", "3m", "4m", "5m") {
		model.Game.RecordEvent(game.EventDiscard, 0, tile, "")
	}
	model.Screen = ScreenTable
	model.UnicodeTiles = false

	view := renderTable(model)

	if strings.Contains(view, "01. 你 打出 1m") {
		t.Fatalf("view should trim oldest event from compact event panel:\n%s", view)
	}
	if !strings.Contains(view, "02. 你 打出 2m") || !strings.Contains(view, "05. 你 打出 5m") {
		t.Fatalf("view missing recent tail events:\n%s", view)
	}
}

func TestRenderTableMarksSelectedTile(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable
	model.SelectedIndex = 2

	view := renderTable(model)

	if !strings.Contains(view, "焦点：") {
		t.Fatalf("view missing focus marker:\n%s", view)
	}
}

func TestRenderTableShowsSelectedTileInHandRow(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Game.Players[0].Hand = mustUITiles(t, "1m", "2m", "3m")
	model.Screen = ScreenTable
	model.SelectedIndex = 1

	view := renderTable(model)

	if strings.Contains(view, "▶ [02]") || !strings.Contains(view, "焦点：[02]") {
		t.Fatalf("view does not clearly show selected tile:\n%s", view)
	}
}

func TestRenderTableShowsHandTrayFocus(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Game.Players[0].Hand = mustUITiles(t, "1m", "2m", "3m")
	model.Screen = ScreenTable
	model.SelectedIndex = 1

	view := renderTable(model)

	for _, text := range []string{"手牌", "焦点：[02]", "🀈"} {
		if !strings.Contains(view, text) {
			t.Fatalf("view missing hand tray focus text %q:\n%s", text, view)
		}
	}
}

func TestRenderTableShowsWaitingActionFeedback(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable

	view := renderTable(model)

	if !strings.Contains(view, "上步：等待操作") {
		t.Fatalf("view missing waiting action feedback:\n%s", view)
	}
}

func TestRenderTableMergesActionsIntoBottomControls(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable

	view := renderTable(model)

	for _, text := range []string{"操作", "回车/空格打出", "单击选牌", "[H] 胡:不可用", "[K] 杠:不可用", "Q 退出"} {
		if !strings.Contains(view, text) {
			t.Fatalf("view missing controls text %q:\n%s", text, view)
		}
	}
	if strings.Contains(view, "Actions:") {
		t.Fatalf("hand panel should not contain duplicated Actions line:\n%s", view)
	}
}

func TestRenderTableHighlightsReadyWinAction(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Game.Players[0].Hand = mustUITiles(t,
		"1m", "2m", "3m",
		"4m", "5m", "6m",
		"2p", "3p", "4p",
		"7s", "7s", "7s",
		"E", "E",
	)
	model.Screen = ScreenTable

	view := renderTable(model)

	if !strings.Contains(view, "[H] 胡:可用") {
		t.Fatalf("view missing ready win action:\n%s", view)
	}
}

func TestRenderTableHighlightsReadyKongAction(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Game.Players[0].Hand = mustUITiles(t, "1m", "1m", "1m", "1m", "2m", "3m")
	model.Screen = ScreenTable

	view := renderTable(model)

	if !strings.Contains(view, "[K] 杠:可用") {
		t.Fatalf("view missing ready kong action:\n%s", view)
	}
}

func TestReadyActionStateKeepsStableVisibleWidth(t *testing.T) {
	action := actionState("[H] Win", true)

	if !strings.Contains(action, "[H] Win:READY") {
		t.Fatalf("ready action missing label: %q", action)
	}
	if got := visibleWidth(action); got != visibleWidth("[H] Win:READY") {
		t.Fatalf("ready action width = %d, want plain width", got)
	}
}

func TestRenderTableUsesCompactActionBarForNarrowWidth(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable
	model.Width = 64

	view := renderTable(model)

	if !strings.Contains(view, "回车打出") || strings.Contains(view, "回车/空格打出") {
		t.Fatalf("view should use compact action bar:\n%s", view)
	}
	for _, line := range strings.Split(strings.TrimRight(view, "\n"), "\n") {
		if strings.Contains(line, "操作") && visibleWidth(line) > 80 {
			t.Fatalf("compact action bar too wide (%d cells):\n%s", visibleWidth(line), line)
		}
	}
}

func TestRenderTableKeepsReadableLineWidth(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable

	view := renderTable(model)

	for _, line := range strings.Split(strings.TrimRight(view, "\n"), "\n") {
		if visibleWidth(line) > 96 {
			t.Fatalf("line too wide (%d cells):\n%s", visibleWidth(line), line)
		}
	}
}

func TestRenderTableUsesCompactControlsForNarrowWidth(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable
	model.Width = 64

	view := renderTable(model)

	if !strings.Contains(view, "方向键选牌") || !strings.Contains(view, "回车打出") || !strings.Contains(view, "Q 退出") {
		t.Fatalf("view missing compact controls:\n%s", view)
	}
	for _, line := range strings.Split(strings.TrimRight(view, "\n"), "\n") {
		if visibleWidth(line) > 80 {
			t.Fatalf("compact line too wide (%d cells):\n%s", visibleWidth(line), line)
		}
	}
}

func TestRenderTableUsesClientSections(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable

	view := renderTable(model)

	for _, text := range []string{"终端麻将", "对手", "牌桌", "你", "操作"} {
		if !strings.Contains(view, text) {
			t.Fatalf("view missing section %q:\n%s", text, view)
		}
	}
}

func TestRenderTableShowsAllOpponentsByName(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable

	view := renderTable(model)

	for _, text := range []string{"电脑1", "电脑2", "电脑3"} {
		if !strings.Contains(view, text) {
			t.Fatalf("view missing opponent %q:\n%s", text, view)
		}
	}
}

func TestRenderTableArrangesOpponentsAsSeats(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable

	view := renderTable(model)

	for _, text := range []string{"对家: 电脑2", "左家: 电脑1", "右家: 电脑3"} {
		if !strings.Contains(view, text) {
			t.Fatalf("view missing seat label %q:\n%s", text, view)
		}
	}
	if !lineContainsAll(view, "左家: 电脑1", "右家: 电脑3") {
		t.Fatalf("left and right seats should share a table row:\n%s", view)
	}
}

func TestRenderTableUsesReferenceInspiredTabletop(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable
	model.Width = 120

	view := renderTable(model)

	for _, want := range []string{"终端麻将", "电脑2", "电脑1", "电脑3", "牌桌", "手牌托盘"} {
		if !strings.Contains(view, want) {
			t.Fatalf("view missing %q:\n%s", want, view)
		}
	}
	if lineIndexContaining(view, "牌桌") <= lineIndexContaining(view, "电脑2") {
		t.Fatalf("center should appear below opposite seat:\n%s", view)
	}
	if lineIndexContaining(view, "手牌托盘") <= lineIndexContaining(view, "牌桌") {
		t.Fatalf("hand tray should appear below center table:\n%s", view)
	}
}

func TestRenderTableCentersMainBoardWhenWide(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable
	model.Width = 120

	view := renderTable(model)

	line := firstLineContaining(view, "终端麻将")
	if !strings.HasPrefix(line, " ") {
		t.Fatalf("wide title should be centered with leading space:\n%s", view)
	}
}

func TestRenderTableKeepsReferenceInspiredWidth(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable
	model.Width = 120

	view := renderTable(model)

	for _, line := range strings.Split(strings.TrimRight(view, "\n"), "\n") {
		if visibleWidth(line) > 120 {
			t.Fatalf("line too wide (%d cells):\n%s", visibleWidth(line), line)
		}
	}
}

func TestRenderTableSplitsFullHandIntoTwoRows(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable

	view := renderTable(model)

	if strings.Count(view, "手牌：") != 1 {
		t.Fatalf("view should keep one stable hand row:\n%s", view)
	}
}

func lineContainsAll(text string, parts ...string) bool {
	for _, line := range strings.Split(text, "\n") {
		matches := true
		for _, part := range parts {
			if !strings.Contains(line, part) {
				matches = false
				break
			}
		}
		if matches {
			return true
		}
	}
	return false
}

func TestHandHitBoxesFindTileIndex(t *testing.T) {
	boxes := handHitBoxes(3, 2, 4)
	index, ok := tileIndexAt(boxes, boxes[1].X1, boxes[1].Y)
	if !ok {
		t.Fatal("expected hit")
	}
	if index != 1 {
		t.Fatalf("index = %d, want 1", index)
	}
}

func TestHandHitBoxesCoverWholeCard(t *testing.T) {
	boxes := handHitBoxes(3, 2, 4)

	for _, y := range []int{boxes[1].Y1, boxes[1].Y, boxes[1].Y2} {
		index, ok := tileIndexAt(boxes, boxes[1].X1, y)
		if !ok || index != 1 {
			t.Fatalf("hit at y=%d = %d,%v; want tile 1", y, index, ok)
		}
	}
}

func TestDefaultHandHitBoxesMatchRenderedHandRows(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	boxes := currentHandHitBoxes(model)

	if boxes[0].Y != handRowY {
		t.Fatalf("first row y = %d, want %d", boxes[0].Y, handRowY)
	}
	if boxes[7].Y != handRowY {
		t.Fatalf("eighth tile y = %d, want same hand row %d", boxes[7].Y, handRowY)
	}
}

func TestDefaultHandHitBoxesMatchRenderedHandStartLine(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable

	view := renderTable(model)

	if got := lineIndexContaining(view, "手牌："); got != handRowY {
		t.Fatalf("rendered hand row = %d, want hitbox row %d:\n%s", got, handRowY, view)
	}
}

func lineIndexContaining(text string, part string) int {
	for index, line := range strings.Split(strings.TrimRight(text, "\n"), "\n") {
		if strings.Contains(line, part) {
			return index
		}
	}
	return -1
}

func firstLineContaining(text string, part string) string {
	for _, line := range strings.Split(strings.TrimRight(text, "\n"), "\n") {
		if strings.Contains(line, part) {
			return line
		}
	}
	return ""
}

func containsTileFrame(text string) bool {
	return strings.ContainsAny(text, "╭╮╰╯│┏┓┗┛┃━")
}
