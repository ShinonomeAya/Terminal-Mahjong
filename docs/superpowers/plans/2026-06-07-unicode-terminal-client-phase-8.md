# Unicode Terminal Client Phase 8 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Unicode Mahjong terminal client with a start menu, table-like TUI layout, keyboard tile selection, and mouse tile selection while preserving the existing rules engine.

**Architecture:** Keep `internal/game` as the rules source of truth. Add small game-facing APIs only where the TUI needs non-blocking actions. Put Bubble Tea state, rendering, keyboard handling, mouse hit testing, and screen routing in `internal/tui`.

**Tech Stack:** Go 1.23, standard library, `github.com/charmbracelet/bubbletea` v0.27.1, Unicode Mahjong tile glyphs.

---

## Workflow Contract

Use this phase loop for every stage:

1. State the stage goal before starting the stage.
2. Complete one checked step.
3. Run the step verification command.
4. Write the step review:

```text
Step review:
- Stage goal:
- Step completed:
- Evidence:
- Next step:
```

5. Repeat until the stage is complete.
6. Write the stage review:

```text
Stage review:
- Total goal:
- Stage completed:
- Evidence:
- Remaining risk:
```

Total goal for all stage reviews:

```text
Build a polished terminal-first Mahjong client that feels playable without command typing while preserving the tested rules engine.
```

## Planned File Structure

- Modify: `go.mod`
  - Add Bubble Tea dependency during Stage 3.
- Create: `internal/game/tile_display.go`
  - Map internal `Tile` values to Unicode glyphs and fallback labels.
- Create: `internal/game/tile_display_test.go`
  - Verify glyphs and fallback labels.
- Create: `internal/game/turn.go`
  - Add focused non-blocking action helpers for TUI discard/draw flow.
- Create: `internal/game/turn_test.go`
  - Verify helpers without a terminal.
- Create: `internal/tui/model.go`
  - Bubble Tea model, screen enum, selected tile index, game state.
- Create: `internal/tui/input.go`
  - Keyboard and mouse event handling.
- Create: `internal/tui/layout.go`
  - Table rendering and hand hit box generation.
- Create: `internal/tui/menu.go`
  - Start menu and help screen rendering/navigation.
- Create: `internal/tui/model_test.go`
  - Start menu, selected tile movement, and action tests.
- Create: `internal/tui/layout_test.go`
  - Render and mouse hit-test tests.
- Modify: `cmd/mahjong/main.go`
  - Start the Bubble Tea app by default.
- Modify: `README.md`
  - Document Unicode TUI controls and fallback behavior.
- Modify: `docs/workflow.md`
  - Add Phase 8 review checklist.

## Stage 1: Unicode Tile Skin

Goal: render every internal tile as a Unicode Mahjong glyph with a text fallback.

Step review question: did this step make tile display prettier without changing tile rules?

Stage review question: can the project render Mahjong glyphs and fallback labels independently of the TUI?

- [ ] **Step 1: Write glyph mapping tests**

Create `internal/game/tile_display_test.go`:

```go
package game

import "testing"

func TestTileGlyphMapsSuits(t *testing.T) {
	cases := map[string]string{
		"1m": "🀇",
		"9m": "🀏",
		"1s": "🀐",
		"9s": "🀘",
		"1p": "🀙",
		"9p": "🀡",
	}
	for text, want := range cases {
		tile, ok := ParseTile(text)
		if !ok {
			t.Fatalf("ParseTile(%q) failed", text)
		}
		if got := TileGlyph(tile); got != want {
			t.Fatalf("TileGlyph(%q) = %q, want %q", text, got, want)
		}
	}
}

func TestTileGlyphMapsHonors(t *testing.T) {
	cases := map[string]string{
		"E": "🀀",
		"S": "🀁",
		"W": "🀂",
		"N": "🀃",
		"Z": "🀄",
		"F": "🀅",
		"B": "🀆",
	}
	for text, want := range cases {
		tile, ok := ParseTile(text)
		if !ok {
			t.Fatalf("ParseTile(%q) failed", text)
		}
		if got := TileGlyph(tile); got != want {
			t.Fatalf("TileGlyph(%q) = %q, want %q", text, got, want)
		}
	}
}

func TestTileLabelFallbackUsesExistingNotation(t *testing.T) {
	tile, ok := ParseTile("5p")
	if !ok {
		t.Fatal("ParseTile failed")
	}
	if got := TileLabel(tile, false); got != "5p" {
		t.Fatalf("TileLabel fallback = %q, want 5p", got)
	}
}

func TestFormatTileLabelsSupportsUnicodeAndFallback(t *testing.T) {
	tiles := mustTiles(t, "1m", "2m", "E")
	if got := FormatTileLabels(tiles, true); got != "🀇 🀈 🀀" {
		t.Fatalf("unicode labels = %q", got)
	}
	if got := FormatTileLabels(tiles, false); got != "1m 2m E" {
		t.Fatalf("fallback labels = %q", got)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run:

```powershell
go test ./internal/game -run "TestTileGlyph|TestTileLabel|TestFormatTileLabels" -v
```

Expected: FAIL because `TileGlyph`, `TileLabel`, and `FormatTileLabels` do not exist.

- [ ] **Step 3: Implement tile display helpers**

Create `internal/game/tile_display.go`:

```go
package game

import "strings"

var tileGlyphs = [TileTypeCount]string{
	"🀇", "🀈", "🀉", "🀊", "🀋", "🀌", "🀍", "🀎", "🀏",
	"🀙", "🀚", "🀛", "🀜", "🀝", "🀞", "🀟", "🀠", "🀡",
	"🀐", "🀑", "🀒", "🀓", "🀔", "🀕", "🀖", "🀗", "🀘",
	"🀀", "🀁", "🀂", "🀃", "🀄", "🀅", "🀆",
}

func TileGlyph(tile Tile) string {
	if tile < 0 || int(tile) >= TileTypeCount {
		return "?"
	}
	return tileGlyphs[int(tile)]
}

func TileLabel(tile Tile, unicode bool) string {
	if unicode {
		return TileGlyph(tile)
	}
	return tile.String()
}

func FormatTileLabels(tiles []Tile, unicode bool) string {
	if len(tiles) == 0 {
		return "-"
	}
	parts := make([]string, len(tiles))
	for i, tile := range tiles {
		parts[i] = TileLabel(tile, unicode)
	}
	return strings.Join(parts, " ")
}
```

- [ ] **Step 4: Run Stage 1 tests**

Run:

```powershell
go test ./internal/game -run "TestTileGlyph|TestTileLabel|TestFormatTileLabels" -v
```

Expected: PASS.

- [ ] **Step 5: Run all game tests**

Run:

```powershell
go test ./internal/game -v
```

Expected: PASS.

- [ ] **Step 6: Commit Stage 1**

Run:

```powershell
git add internal/game/tile_display.go internal/game/tile_display_test.go
git commit -m "feat: add unicode mahjong tile display"
```

Write the Stage 1 review before moving on.

## Stage 2: Non-Blocking Game Action Helpers

Goal: expose small game actions the TUI can call without using the blocking `Play(in, out)` loop.

Step review question: did this step move turn state forward without mixing in Bubble Tea?

Stage review question: can the TUI drive a human discard through tested game helpers?

- [ ] **Step 1: Write helper tests for starting a human turn and discarding**

Create `internal/game/turn_test.go`:

```go
package game

import "testing"

func TestStartHumanTurnDrawsTile(t *testing.T) {
	game := NewGame(1)
	startWall := len(game.Wall)
	startHand := len(game.Players[0].Hand)

	tile, ok := game.StartHumanTurn()

	if !ok {
		t.Fatal("StartHumanTurn returned false")
	}
	if tile < 0 || int(tile) >= TileTypeCount {
		t.Fatalf("drawn tile = %v", tile)
	}
	if len(game.Wall) != startWall-1 {
		t.Fatalf("wall = %d, want %d", len(game.Wall), startWall-1)
	}
	if len(game.Players[0].Hand) != startHand+1 {
		t.Fatalf("hand = %d, want %d", len(game.Players[0].Hand), startHand+1)
	}
}

func TestHumanDiscardSelectedRemovesTileAndRecordsEvent(t *testing.T) {
	game := NewGame(1)
	game.Players[0].Hand = mustTiles(t, "1m", "2m", "3m")

	discard, err := game.HumanDiscardSelected(1)

	if err != nil {
		t.Fatal(err)
	}
	if discard.String() != "2m" {
		t.Fatalf("discard = %s, want 2m", discard)
	}
	if FormatTiles(game.Players[0].Discards) != "2m" {
		t.Fatalf("discards = %s, want 2m", FormatTiles(game.Players[0].Discards))
	}
	if len(game.Events) == 0 || game.Events[len(game.Events)-1].Kind != EventDiscard {
		t.Fatalf("last event = %#v, want discard", game.Events)
	}
	if game.Current != 1 {
		t.Fatalf("current = %d, want next player 1", game.Current)
	}
}

func TestHumanDiscardSelectedRejectsInvalidIndex(t *testing.T) {
	game := NewGame(1)
	game.Players[0].Hand = mustTiles(t, "1m")

	if _, err := game.HumanDiscardSelected(3); err == nil {
		t.Fatal("expected invalid index error")
	}
}

func TestAdvanceAIUntilHumanTurnReturnsToHumanWithDraw(t *testing.T) {
	game := NewGame(1)
	game.StartHumanTurn()
	if _, err := game.HumanDiscardSelected(0); err != nil {
		t.Fatal(err)
	}
	startEvents := len(game.Events)

	game.AdvanceAIUntilHumanTurn()

	if game.Over {
		t.Fatalf("game ended early: %s", game.Reason)
	}
	if game.Current != 0 {
		t.Fatalf("current = %d, want human player 0", game.Current)
	}
	if len(game.Players[0].Hand) != 14 {
		t.Fatalf("human hand = %d, want 14 after next draw", len(game.Players[0].Hand))
	}
	if len(game.Events) <= startEvents {
		t.Fatalf("events did not advance: %d <= %d", len(game.Events), startEvents)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run:

```powershell
go test ./internal/game -run "TestStartHumanTurn|TestHumanDiscardSelected|TestAdvanceAIUntilHumanTurn" -v
```

Expected: FAIL because the helper methods do not exist.

- [ ] **Step 3: Implement minimal game helpers**

Create `internal/game/turn.go`:

```go
package game

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

func (g *Game) StartHumanTurn() (Tile, bool) {
	if g.Over || g.Current != 0 || len(g.Wall) == 0 {
		return -1, false
	}
	return g.draw(0), true
}

func (g *Game) HumanDiscardSelected(index int) (Tile, error) {
	if g.Over {
		return -1, fmt.Errorf("game is over")
	}
	if g.Current != 0 {
		return -1, fmt.Errorf("not the human turn")
	}
	discard, err := g.Players[0].RemoveAt(index)
	if err != nil {
		return -1, err
	}
	g.Players[0].Discards = append(g.Players[0].Discards, discard)
	g.RecordEvent(EventDiscard, 0, discard, "")
	g.Current = 1
	return discard, nil
}

func (g *Game) Quit(reason string) {
	g.Over = true
	g.Reason = reason
	g.RecordEvent(EventQuit, 0, -1, reason)
}

func (g *Game) AdvanceAIUntilHumanTurn() {
	declineReader := bufio.NewReader(strings.NewReader(""))
	for !g.Over && g.Current != 0 {
		if len(g.Wall) == 0 {
			g.Over = true
			g.Reason = "draw: wall exhausted"
			g.RecordEvent(EventWallExhausted, g.Current, -1, g.Reason)
			return
		}
		g.draw(g.Current)
		if CanWin(g.Players[g.Current].Hand) {
			g.finish(g.Current, "self-draw", WinSelfDraw)
			return
		}
		g.resolveAIKongs(io.Discard, g.Current)
		if g.Over {
			return
		}
		discard, ok := g.takeDiscardTurn(declineReader, io.Discard, g.Current)
		if !ok {
			return
		}
		if g.resolveDiscardClaims(declineReader, io.Discard, g.Current, discard) {
			continue
		}
		g.Current = (g.Current + 1) % len(g.Players)
	}
	if !g.Over && g.Current == 0 && len(g.Players[0].Hand)%3 == 1 {
		g.StartHumanTurn()
	}
}
```

- [ ] **Step 4: Run Stage 2 tests**

Run:

```powershell
go test ./internal/game -run "TestStartHumanTurn|TestHumanDiscardSelected|TestAdvanceAIUntilHumanTurn" -v
```

Expected: PASS.

- [ ] **Step 5: Run all game tests**

Run:

```powershell
go test ./internal/game -v
```

Expected: PASS.

- [ ] **Step 6: Commit Stage 2**

Run:

```powershell
git add internal/game/turn.go internal/game/turn_test.go
git commit -m "feat: add tui game action helpers"
```

Write the Stage 2 review before moving on.

## Stage 3: Bubble Tea Start Menu

Goal: introduce Bubble Tea with a tested start menu before touching table play.

Step review question: did this step add TUI structure without changing game rules?

Stage review question: can the app start in a navigable terminal menu?

- [ ] **Step 1: Add Bubble Tea dependency**

Run:

```powershell
go get github.com/charmbracelet/bubbletea@v0.27.1
```

Expected: `go.mod` and `go.sum` include Bubble Tea v0.27.1 dependencies. This version preserves Go 1.23 compatibility for this project.

- [ ] **Step 2: Write menu tests**

Create `internal/tui/model_test.go`:

```go
package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewModelStartsAtMenu(t *testing.T) {
	model := NewModel()
	if model.Screen != ScreenMenu {
		t.Fatalf("screen = %v, want menu", model.Screen)
	}
	if model.MenuIndex != 0 {
		t.Fatalf("menu index = %d, want 0", model.MenuIndex)
	}
}

func TestMenuDownMovesSelection(t *testing.T) {
	model := NewModel()
	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	updated := next.(Model)
	if updated.MenuIndex != 1 {
		t.Fatalf("menu index = %d, want 1", updated.MenuIndex)
	}
}

func TestMenuEnterStartsSoloGame(t *testing.T) {
	model := NewModel()
	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := next.(Model)
	if updated.Screen != ScreenTable {
		t.Fatalf("screen = %v, want table", updated.Screen)
	}
	if updated.Game == nil {
		t.Fatal("expected game to be created")
	}
}

func TestMenuViewContainsOptions(t *testing.T) {
	view := NewModel().View()
	for _, text := range []string{"TERMINAL MAHJONG", "Solo Game", "How to Play", "Quit"} {
		if !strings.Contains(view, text) {
			t.Fatalf("view missing %q:\n%s", text, view)
		}
	}
}
```

- [ ] **Step 3: Run tests to verify they fail**

Run:

```powershell
go test ./internal/tui -run "TestNewModel|TestMenu" -v
```

Expected: FAIL because `internal/tui` does not exist.

- [ ] **Step 4: Implement menu model**

Create `internal/tui/model.go`:

```go
package tui

import (
	"mahjong/internal/game"

	tea "github.com/charmbracelet/bubbletea"
)

type Screen int

const (
	ScreenMenu Screen = iota
	ScreenHelp
	ScreenTable
	ScreenGameOver
)

type Model struct {
	Screen        Screen
	MenuIndex     int
	SelectedIndex int
	UnicodeTiles  bool
	Game          *game.Game
}

func NewModel() Model {
	return Model{Screen: ScreenMenu, UnicodeTiles: true}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch m.Screen {
	case ScreenMenu:
		return updateMenu(m, key)
	case ScreenHelp:
		return updateHelp(m, key)
	default:
		return m, nil
	}
}

func (m Model) View() string {
	switch m.Screen {
	case ScreenMenu:
		return renderMenu(m)
	case ScreenHelp:
		return renderHelp()
	case ScreenTable:
		return renderTable(m)
	case ScreenGameOver:
		return renderGameOver(m)
	default:
		return ""
	}
}
```

Create `internal/tui/menu.go`:

```go
package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

var menuItems = []string{"Solo Game", "How to Play", "Quit"}

func updateMenu(m Model, key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.Type {
	case tea.KeyDown:
		m.MenuIndex = (m.MenuIndex + 1) % len(menuItems)
	case tea.KeyUp:
		m.MenuIndex = (m.MenuIndex + len(menuItems) - 1) % len(menuItems)
	case tea.KeyEnter:
		switch m.MenuIndex {
		case 0:
			m.Game = newStartedGame()
			m.Screen = ScreenTable
		case 1:
			m.Screen = ScreenHelp
		case 2:
			return m, tea.Quit
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
	out.WriteString("╔════════════════ TERMINAL MAHJONG ════════════════╗\n")
	out.WriteString("║                                                  ║\n")
	for i, item := range menuItems {
		prefix := "  "
		if i == m.MenuIndex {
			prefix = "> "
		}
		out.WriteString("║              " + prefix + item + strings.Repeat(" ", 28-len(item)) + "║\n")
	}
	out.WriteString("║                                                  ║\n")
	out.WriteString("║        ↑/↓ choose   Enter confirm   Q quit       ║\n")
	out.WriteString("╚══════════════════════════════════════════════════╝\n")
	return out.String()
}

func renderHelp() string {
	return "TERMINAL MAHJONG HELP\n\n←/→ select tile\nEnter/Space discard\nMouse click selects a tile\nSecond click discards selected tile\nEsc returns\n"
}
```

Create `internal/tui/layout.go` with temporary table/game-over renderers:

```go
package tui

func renderTable(m Model) string {
	return "TABLE\n"
}

func renderGameOver(m Model) string {
	return "GAME OVER\n"
}
```

Create `internal/tui/game_flow.go`:

```go
package tui

import "mahjong/internal/game"

func newStartedGame() *game.Game {
	g := game.NewGame(0)
	g.StartHumanTurn()
	return g
}
```

- [ ] **Step 5: Run Stage 3 tests**

Run:

```powershell
go test ./internal/tui -run "TestNewModel|TestMenu" -v
```

Expected: PASS.

- [ ] **Step 6: Run all tests**

Run:

```powershell
go test ./...
```

Expected: PASS.

- [ ] **Step 7: Commit Stage 3**

Run:

```powershell
git add go.mod go.sum internal/tui
git commit -m "feat: add terminal client start menu"
```

Write the Stage 3 review before moving on.

## Stage 4: Keyboard Tile Selection and Discard

Goal: make normal play possible with arrows and Enter/Space instead of typed commands.

Step review question: did this step make play smoother without removing tested fallback behavior?

Stage review question: can a human discard from the TUI model through keyboard selection?

- [ ] **Step 1: Write selection and discard tests**

Append to `internal/tui/model_test.go`:

```go
func TestTableRightMovesSelectedTile(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyRight})
	updated := next.(Model)

	if updated.SelectedIndex != 1 {
		t.Fatalf("selected index = %d, want 1", updated.SelectedIndex)
	}
}

func TestTableLeftWrapsSelectedTile(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyLeft})
	updated := next.(Model)

	if updated.SelectedIndex != len(updated.Game.Players[0].Hand)-1 {
		t.Fatalf("selected index = %d, want last hand index", updated.SelectedIndex)
	}
}

func TestTableEnterDiscardsSelectedTile(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable
	model.SelectedIndex = 0
	startEvents := len(model.Game.Events)

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := next.(Model)

	if len(updated.Game.Players[0].Discards) != 1 {
		t.Fatalf("discards = %d, want 1", len(updated.Game.Players[0].Discards))
	}
	if len(updated.Game.Events) <= startEvents+1 {
		t.Fatalf("events = %d, want AI turns after human discard", len(updated.Game.Events))
	}
	if !updated.Game.Over && updated.Game.Current != 0 {
		t.Fatalf("current = %d, want human turn after AI advance", updated.Game.Current)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run:

```powershell
go test ./internal/tui -run "TestTable.*Selected|TestTableEnter" -v
```

Expected: FAIL because table key handling is not implemented.

- [ ] **Step 3: Implement table key handling**

Create `internal/tui/input.go`:

```go
package tui

import tea "github.com/charmbracelet/bubbletea"

func updateTable(m Model, key tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.Game == nil {
		return m, nil
	}
	handLen := len(m.Game.Players[0].Hand)
	switch key.Type {
	case tea.KeyLeft:
		if handLen > 0 {
			m.SelectedIndex = (m.SelectedIndex + handLen - 1) % handLen
		}
	case tea.KeyRight:
		if handLen > 0 {
			m.SelectedIndex = (m.SelectedIndex + 1) % handLen
		}
	case tea.KeyEnter:
		return discardSelected(m)
	}
	switch key.String() {
	case " ":
		return discardSelected(m)
	case "q":
		m.Game.Quit("quit")
		m.Screen = ScreenGameOver
	}
	return m, nil
}

func discardSelected(m Model) (tea.Model, tea.Cmd) {
	if m.Game == nil || len(m.Game.Players[0].Hand) == 0 {
		return m, nil
	}
	if m.SelectedIndex >= len(m.Game.Players[0].Hand) {
		m.SelectedIndex = len(m.Game.Players[0].Hand) - 1
	}
	if _, err := m.Game.HumanDiscardSelected(m.SelectedIndex); err != nil {
		return m, nil
	}
	m.Game.AdvanceAIUntilHumanTurn()
	if m.SelectedIndex >= len(m.Game.Players[0].Hand) {
		m.SelectedIndex = len(m.Game.Players[0].Hand) - 1
	}
	if m.Game.Over {
		m.Screen = ScreenGameOver
	}
	return m, nil
}
```

Update `internal/tui/model.go` table case:

```go
case ScreenTable:
	return updateTable(m, key)
```

- [ ] **Step 4: Run Stage 4 tests**

Run:

```powershell
go test ./internal/tui -run "TestTable.*Selected|TestTableEnter" -v
```

Expected: PASS.

- [ ] **Step 5: Run all tests**

Run:

```powershell
go test ./...
```

Expected: PASS.

- [ ] **Step 6: Commit Stage 4**

Run:

```powershell
git add internal/tui
git commit -m "feat: add keyboard tile selection"
```

Write the Stage 4 review before moving on.

## Stage 5: Table-Like Unicode Layout

Goal: replace the temporary table view with a table-like Unicode Mahjong layout.

Step review question: did this step improve the client feel without coupling layout to game rules?

Stage review question: does the table screen look like Mahjong in a modern terminal?

- [ ] **Step 1: Write layout tests**

Create `internal/tui/layout_test.go`:

```go
package tui

import (
	"strings"
	"testing"
)

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

func TestRenderTableMarksSelectedTile(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable
	model.SelectedIndex = 2

	view := renderTable(model)

	if !strings.Contains(view, "selected") && !strings.Contains(view, "▲") {
		t.Fatalf("view missing selected marker:\n%s", view)
	}
}
```

Create `internal/tui/test_helpers_test.go`:

```go
package tui

import (
	"testing"

	"mahjong/internal/game"
)

func mustUITiles(t *testing.T, texts ...string) []game.Tile {
	t.Helper()
	tiles := make([]game.Tile, 0, len(texts))
	for _, text := range texts {
		tile, ok := game.ParseTile(text)
		if !ok {
			t.Fatalf("ParseTile(%q) failed", text)
		}
		tiles = append(tiles, tile)
	}
	game.SortTiles(tiles)
	return tiles
}
```

- [ ] **Step 2: Run layout tests to verify they fail**

Run:

```powershell
go test ./internal/tui -run "TestRenderTable" -v
```

Expected: FAIL because `renderTable` is still temporary.

- [ ] **Step 3: Implement table renderer**

Replace `internal/tui/layout.go` with:

```go
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

func renderTable(m Model) string {
	if m.Game == nil {
		return "No game\n"
	}
	g := m.Game
	var out strings.Builder
	out.WriteString("╔════════════════════════ TERMINAL MAHJONG ════════════════════════╗\n")
	out.WriteString(fmt.Sprintf("║ Wall %-3d  Events %-3d  Turn %-10s  Replay ready                 ║\n", len(g.Wall), len(g.Events), g.Players[g.Current].Name))
	out.WriteString("╠════════════════════════════ AI-2 ═════════════════════════════════╣\n")
	out.WriteString(renderOpponent(g.Players[2], m.UnicodeTiles))
	out.WriteString("╠══════════════ AI-1 ═══════════╦════════ CENTER ════════╦════════ AI-3 ════════╣\n")
	out.WriteString(renderSidePlayers(g, m.UnicodeTiles))
	out.WriteString("╠══════════════════════════════ YOU ════════════════════════════════╣\n")
	out.WriteString(fmt.Sprintf("║ Melds: %-24s Discards: %-24s ║\n", g.Players[0].MeldSummary(), game.FormatTileLabels(g.Players[0].Discards, m.UnicodeTiles)))
	out.WriteString(renderHand(g.Players[0].Hand, m.SelectedIndex, m.UnicodeTiles))
	out.WriteString("╚══ ←/→ select  Enter/Space discard  mouse click tile  H win  K kong  Q quit ═╝\n")
	return out.String()
}

func renderOpponent(player game.Player, unicode bool) string {
	return fmt.Sprintf("║ %-8s hand:%2d  melds:%-12s discards:%-24s ║\n", player.Name, len(player.Hand), player.MeldSummary(), game.FormatTileLabels(player.Discards, unicode))
}

func renderSidePlayers(g *game.Game, unicode bool) string {
	left := g.Players[1]
	right := g.Players[3]
	recent := game.RecentEvents(g.Events, 3)
	center := "No events yet"
	if len(recent) > 0 {
		center = recent[len(recent)-1].String()
	}
	return fmt.Sprintf("║ %-25s ║ %-20s ║ %-20s ║\n", left.Name+" discards "+game.FormatTileLabels(left.Discards, unicode), center, right.Name+" discards "+game.FormatTileLabels(right.Discards, unicode))
}

func renderHand(hand []game.Tile, selected int, unicode bool) string {
	var tiles strings.Builder
	var markers strings.Builder
	tiles.WriteString("║ ")
	markers.WriteString("║ ")
	for i, tile := range hand {
		label := game.TileLabel(tile, unicode)
		cell := fmt.Sprintf("[%2d]%s ", i+1, label)
		tiles.WriteString(cell)
		if i == selected {
			markers.WriteString(strings.Repeat(" ", len(cell)/2))
			markers.WriteString("▲ selected ")
		} else {
			markers.WriteString(strings.Repeat(" ", len(cell)))
		}
	}
	tiles.WriteString("\n")
	markers.WriteString("\n")
	return tiles.String() + markers.String()
}

func renderGameOver(m Model) string {
	if m.Game == nil {
		return "GAME OVER\n"
	}
	return fmt.Sprintf("GAME OVER\nResult: %s\nEvents: %d\nReplay-ready event log: yes\n", m.Game.Reason, len(m.Game.Events))
}
```

- [ ] **Step 4: Run Stage 5 tests**

Run:

```powershell
go test ./internal/tui -run "TestRenderTable" -v
```

Expected: PASS.

- [ ] **Step 5: Run all tests**

Run:

```powershell
go test ./...
```

Expected: PASS.

- [ ] **Step 6: Commit Stage 5**

Run:

```powershell
git add internal/tui
git commit -m "feat: render unicode mahjong table"
```

Write the Stage 5 review before moving on.

## Stage 6: Mouse Tile Selection

Goal: support mouse click selection and second-click discard through deterministic hit testing.

Step review question: did this step make the UI more client-like while keeping keyboard fallback?

Stage review question: can mouse events drive the same tested discard path as keyboard events?

- [ ] **Step 1: Write hit-test and mouse tests**

Append to `internal/tui/layout_test.go`:

```go
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
```

Append to `internal/tui/model_test.go`:

```go
func TestMouseClickSelectsTile(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable
	model.HandHitBoxes = handHitBoxes(len(model.Game.Players[0].Hand), 2, 10)

	next, _ := model.Update(tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
		X:      model.HandHitBoxes[2].X1,
		Y:      model.HandHitBoxes[2].Y,
	})
	updated := next.(Model)

	if updated.SelectedIndex != 2 {
		t.Fatalf("selected index = %d, want 2", updated.SelectedIndex)
	}
}

func TestSecondMouseClickDiscardsSelectedTile(t *testing.T) {
	model := NewModel()
	model.Game = newStartedGame()
	model.Screen = ScreenTable
	model.SelectedIndex = 2
	model.HandHitBoxes = handHitBoxes(len(model.Game.Players[0].Hand), 2, 10)

	next, _ := model.Update(tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
		X:      model.HandHitBoxes[2].X1,
		Y:      model.HandHitBoxes[2].Y,
	})
	updated := next.(Model)

	if len(updated.Game.Players[0].Discards) != 1 {
		t.Fatalf("discards = %d, want 1", len(updated.Game.Players[0].Discards))
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run:

```powershell
go test ./internal/tui -run "TestHandHitBoxes|TestMouse" -v
```

Expected: FAIL because hit boxes and mouse handling do not exist.

- [ ] **Step 3: Add hit boxes to model and layout**

Update `internal/tui/model.go`:

```go
HandHitBoxes []TileHitBox
```

Add to `internal/tui/layout.go`:

```go
func handHitBoxes(count int, startX int, y int) []TileHitBox {
	boxes := make([]TileHitBox, count)
	x := startX
	for i := 0; i < count; i++ {
		boxes[i] = TileHitBox{Index: i, X1: x, X2: x + 5, Y: y}
		x += 6
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
```

Add to `internal/tui/layout.go`:

```go
func currentHandHitBoxes(m Model) []TileHitBox {
	if len(m.HandHitBoxes) > 0 {
		return m.HandHitBoxes
	}
	if m.Game == nil {
		return nil
	}
	return handHitBoxes(len(m.Game.Players[0].Hand), 2, 10)
}
```

- [ ] **Step 4: Implement mouse handling**

Replace `internal/tui/model.go` `Update` with:

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.Screen {
		case ScreenMenu:
			return updateMenu(m, msg)
		case ScreenHelp:
			return updateHelp(m, msg)
		case ScreenTable:
			return updateTable(m, msg)
		default:
			return m, nil
		}
	case tea.MouseMsg:
		if m.Screen == ScreenTable {
			return updateTableMouse(m, msg)
		}
		return m, nil
	default:
		return m, nil
	}
}
```

Add to `internal/tui/input.go`:

```go
func updateTableMouse(m Model, msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if msg.Action != tea.MouseActionPress || msg.Button != tea.MouseButtonLeft || m.Game == nil {
		return m, nil
	}
	boxes := currentHandHitBoxes(m)
	index, ok := tileIndexAt(boxes, msg.X, msg.Y)
	if !ok {
		return m, nil
	}
	if index == m.SelectedIndex {
		return discardSelected(m)
	}
	m.SelectedIndex = index
	return m, nil
}
```

- [ ] **Step 5: Run Stage 6 tests**

Run:

```powershell
go test ./internal/tui -run "TestHandHitBoxes|TestMouse" -v
```

Expected: PASS.

- [ ] **Step 6: Run all tests**

Run:

```powershell
go test ./...
```

Expected: PASS.

- [ ] **Step 7: Commit Stage 6**

Run:

```powershell
git add internal/tui
git commit -m "feat: add mouse tile selection"
```

Write the Stage 6 review before moving on.

## Stage 7: Wire TUI Into the CLI

Goal: make `go run ./cmd/mahjong` start the Unicode terminal client.

Step review question: did this step make the real app use the new client without hiding old rules tests?

Stage review question: can a user launch the game and reach the Unicode table from the start menu?

- [ ] **Step 1: Add TUI run function**

Create `internal/tui/run.go`:

```go
package tui

import tea "github.com/charmbracelet/bubbletea"

func Run() error {
	program := tea.NewProgram(NewModel(), tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := program.Run()
	return err
}
```

- [ ] **Step 2: Update command entrypoint**

Replace `cmd/mahjong/main.go` with:

```go
package main

import (
	"fmt"
	"os"

	"mahjong/internal/tui"
)

func main() {
	if err := tui.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

- [ ] **Step 3: Run tests and build**

Run:

```powershell
go test ./...
go build ./cmd/mahjong
```

Expected: both commands PASS.

- [ ] **Step 4: Manual smoke**

Run:

```powershell
go run ./cmd/mahjong
```

Expected:

- Start menu appears.
- Arrow keys move the selected menu item.
- Enter on `Solo Game` opens the table.
- Hand tiles use Unicode glyphs by default.
- Left/right arrows move selected tile.
- Enter discards the selected tile.
- Mouse click selects a tile in a terminal with mouse support.
- Second click discards the selected tile.
- `Q` exits to a game-over summary or quits the client.

- [ ] **Step 5: Remove local binary if build created one**

Run:

```powershell
if (Test-Path .\mahjong.exe) { Remove-Item -LiteralPath .\mahjong.exe }
```

Expected: no local binary remains.

- [ ] **Step 6: Commit Stage 7**

Run:

```powershell
git add cmd/mahjong/main.go internal/tui/run.go
git commit -m "feat: start unicode terminal client"
```

Write the Stage 7 review before moving on.

## Stage 8: Documentation and Final Acceptance

Goal: document the new client controls and prove Phase 8 meets acceptance.

Step review question: did this step make the finished client easier to run, test, or understand?

Stage review question: does Phase 8 make the total project goal more true?

- [ ] **Step 1: Update README controls**

Modify `README.md` command section to describe:

```markdown
## Controls

- `↑` / `↓`: move in menus.
- `Enter`: confirm menu item or discard the selected tile.
- `←` / `→`: select a hand tile.
- Mouse click: select a hand tile when the terminal supports mouse events.
- Second click on the selected hand tile: discard it.
- `Space`: discard the selected tile.
- `H`: win when available.
- `K`: declare a concealed kong when available.
- `Esc`: cancel or decline a prompt.
- `Q`: quit.
```

Also mention:

```markdown
The client renders Mahjong tiles with Unicode glyphs by default and keeps text labels available for fallback rendering in tests and future configuration.
```

- [ ] **Step 2: Add Phase 8 workflow checklist**

Append to `docs/workflow.md`:

```markdown
## Phase 8 Review Checklist

- Step reviews confirm Unicode tile rendering, non-blocking game helpers, start menu, keyboard selection, table layout, mouse selection, and CLI wiring.
- Stage reviews compare each TUI addition against the total goal of a polished terminal-first Mahjong client.
- Verification commands: `go test ./...`, `go build ./cmd/mahjong`, and one manual TUI smoke run.
```

- [ ] **Step 3: Run final automated verification**

Run:

```powershell
go test ./...
go test ./... -cover
go build ./cmd/mahjong
```

Expected: all PASS.

- [ ] **Step 4: Run final manual acceptance smoke**

Run:

```powershell
go run ./cmd/mahjong
```

Expected:

- TUI start menu appears.
- `Solo Game` enters a Unicode table screen.
- Keyboard selection and discard work.
- Mouse selection works in a mouse-capable terminal.
- Quit produces a clear game-over or exit path.

- [ ] **Step 5: Clean generated binary**

Run:

```powershell
if (Test-Path .\mahjong.exe) { Remove-Item -LiteralPath .\mahjong.exe }
git status --short
```

Expected: only intended source/doc changes remain.

- [ ] **Step 6: Commit Stage 8**

Run:

```powershell
git add README.md docs/workflow.md
git commit -m "docs: document unicode terminal client"
```

Write the final Phase 8 stage review before reporting acceptance.

## Final Phase 8 Acceptance Gate

Run these commands before saying Phase 8 is complete:

```powershell
go test ./...
go test ./... -cover
go build ./cmd/mahjong
git status --short
```

Manual check:

```powershell
go run ./cmd/mahjong
```

Phase 8 passes only if:

- Start menu exists.
- Unicode Mahjong glyphs render on the table.
- Text fallback is covered by tests.
- Keyboard selection and discard work.
- Mouse selection and second-click discard work where terminal support exists.
- The existing rules tests remain green.
- The app remains terminal-first and runnable from `go run ./cmd/mahjong`.
