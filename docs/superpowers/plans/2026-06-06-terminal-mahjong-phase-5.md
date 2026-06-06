# Terminal Mahjong Phase 5 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Upgrade the terminal table layout into a clearer text UI with sections, stable labels, command help, and recent event summaries without changing the game into a GUI.

**Architecture:** Keep using standard-library terminal output. Add small formatting helpers in `internal/game/render.go` and test the rendered text directly. Do not introduce Bubble Tea, tcell, colors, alternate screen mode, or key-by-key input in Phase 5; those can come later once the text layout is stable.

**Tech Stack:** Go 1.23, standard library only, `go test ./...`, `go build ./cmd/mahjong`, and scripted terminal smoke runs.

---

## Scope

Phase 5 is a terminal UI readability upgrade. It preserves the current command model and all Phase 4 hand-analysis behavior.

Acceptance criteria:

- `printTable` renders a bordered or clearly sectioned terminal table.
- Human hand, tips, wall count, opponents, melds, and discards remain visible.
- Command help appears in the table every turn.
- Recent events appear in the table or result output using the Phase 3 event log.
- Existing Phase 4 tips and AI behavior continue to pass tests.
- No GUI, browser UI, desktop window, or heavyweight TUI dependency is introduced.

## File Map

- Modify `internal/game/render.go`: structured table sections, command help, recent events.
- Modify `internal/game/game_test.go`: rendering tests for sections and command help.
- Modify `internal/game/event.go`: optional `RecentEvents` helper if rendering needs it.
- Modify `README.md`: document Phase 5 as terminal UI, not GUI.
- Modify `docs/workflow.md`: append Phase 5 review checklist.

## Task 1: Render Section Helpers

**Files:**
- Modify: `internal/game/render.go`
- Modify: `internal/game/game_test.go`

- [ ] **Step 1: Add section rendering test**

Add test:

```go
func TestPrintTableUsesTerminalSections(t *testing.T) {
	game := NewGame(1)
	var out strings.Builder
	game.printTable(&out)
	text := out.String()
	for _, want := range []string{"Terminal Mahjong", "Opponents", "Your Hand", "Commands"} {
		if !strings.Contains(text, want) {
			t.Fatalf("table output missing %q:\n%s", want, text)
		}
	}
}
```

Run: `go test ./internal/game -run TestPrintTableUsesTerminalSections -v`

Expected first result: test fails because the current table is still plain text.

- [ ] **Step 2: Add small render helpers**

In `render.go`, add helpers:

- `sectionTitle(title string) string`
- `commandHelp() string`
- `formatOpponentLine(player Player) string`

Keep ASCII output only.

Run: `go test ./internal/game -run TestPrintTableUsesTerminalSections -v`

Expected: section rendering test passes.

Step review:

```text
Step review:
- Stage goal: improve terminal readability while staying terminal-first.
- Step completed: table output now has stable section labels.
- Evidence: section rendering test passes.
- Next step: add recent events to the table.
```

## Task 2: Recent Event Display

**Files:**
- Modify: `internal/game/event.go`
- Modify: `internal/game/render.go`
- Modify: `internal/game/game_test.go`

- [ ] **Step 1: Add recent-events tests**

Add tests:

```go
func TestRecentEventsReturnsTail(t *testing.T) {
	game := NewGame(1)
	game.RecordEvent(EventDraw, 0, mustTile(t, "1m"), "")
	game.RecordEvent(EventDiscard, 0, mustTile(t, "1m"), "")
	game.RecordEvent(EventDraw, 1, mustTile(t, "2m"), "")
	recent := RecentEvents(game.Events, 2)
	if len(recent) != 2 || recent[0].Kind != EventDiscard || recent[1].Kind != EventDraw {
		t.Fatalf("recent = %#v", recent)
	}
}

func TestPrintTableIncludesRecentEvents(t *testing.T) {
	game := NewGame(1)
	game.RecordEvent(EventDiscard, 0, mustTile(t, "1m"), "")
	var out strings.Builder
	game.printTable(&out)
	if !strings.Contains(out.String(), "Recent Events") || !strings.Contains(out.String(), "You discard 1m") {
		t.Fatalf("table output:\n%s", out.String())
	}
}
```

Run: `go test ./internal/game -run 'TestRecentEvents|TestPrintTableIncludesRecentEvents' -v`

Expected first result: build or assertion fails because recent event rendering does not exist.

- [ ] **Step 2: Implement recent events**

Implement:

```go
func RecentEvents(events []GameEvent, limit int) []GameEvent
```

Rules:

- `limit <= 0` returns empty.
- If event count is smaller than limit, return all events.
- Preserve chronological order.

Render the last 5 events in `printTable` under `Recent Events`.

Run: `go test ./internal/game -run 'TestRecentEvents|TestPrintTableIncludesRecentEvents' -v`

Expected: recent-event tests pass.

Step review:

```text
Step review:
- Stage goal: make terminal state easier to scan.
- Step completed: recent event history appears on the table.
- Evidence: recent-event tests pass.
- Next step: tighten result output.
```

## Task 3: Result Summary Layout

**Files:**
- Modify: `internal/game/render.go`
- Modify: `internal/game/game_test.go`

- [ ] **Step 1: Add result layout test**

Add test:

```go
func TestPrintResultUsesSummarySections(t *testing.T) {
	game := NewGame(1)
	game.finish(0, "self-draw", WinSelfDraw)
	var out strings.Builder
	game.printResult(&out)
	text := out.String()
	for _, want := range []string{"Game Over", "Summary", "Events"} {
		if !strings.Contains(text, want) {
			t.Fatalf("result output missing %q:\n%s", want, text)
		}
	}
}
```

Run: `go test ./internal/game -run TestPrintResultUsesSummarySections -v`

Expected first result: test fails until result layout is sectioned.

- [ ] **Step 2: Update result layout**

Keep existing winner, win reason, score, and event count. Add stable labels:

- `Game Over`
- `Summary`
- `Recent Events`

Run: `go test ./internal/game -run 'TestPrintResultUsesSummarySections|TestPrintResultIncludesScore|TestPrintResultIncludesEventCount' -v`

Expected: result tests pass.

Step review:

```text
Step review:
- Stage goal: make completed games easier to review in the terminal.
- Step completed: result output has summary sections.
- Evidence: result layout tests pass.
- Next step: update docs and verify.
```

## Task 4: Documentation, Verification, and Commit

**Files:**
- Modify: `README.md`
- Modify: `docs/workflow.md`

- [ ] **Step 1: Update README**

Add:

```markdown
## Phase 5 Direction

The game remains terminal-first. Phase 5 improves the text table with stable sections, command help, and recent event summaries. It is still a terminal game, not a GUI.
```

Run: `rg -n "Phase 5 Direction|terminal-first|not a GUI|Recent Events" README.md`

Expected: terms appear.

- [ ] **Step 2: Append workflow checklist**

Add:

```markdown
## Phase 5 Review Checklist

- Step reviews confirm table sections, command help, recent events, and result summary layout.
- Stage review compares terminal readability against the total terminal Mahjong goal.
- Verification commands: `go test ./...`, `go build ./cmd/mahjong`, one scripted smoke run.
```

Run: `rg -n "Phase 5 Review Checklist|terminal readability|Recent Events" docs/workflow.md`

Expected: terms appear.

- [ ] **Step 3: Final verification**

Run:

```powershell
go test ./...
go test ./... -cover
go build ./cmd/mahjong
cmd /c "(echo 1& echo.& echo.& echo q) | mahjong.exe"
if (Test-Path 'mahjong.exe') { Remove-Item -LiteralPath 'mahjong.exe' }
git status --short
```

Expected:

- Tests pass.
- Build exits 0.
- Smoke run shows section labels, command help, tips, and recent events.
- `mahjong.exe` is removed.
- No external TUI dependency appears in `go.mod`.

- [ ] **Step 4: Commit Phase 5**

Run:

```powershell
git add README.md docs/workflow.md docs/superpowers/plans/2026-06-06-terminal-mahjong-phase-5.md internal/game
git commit -m "feat: improve terminal mahjong table"
```

Expected: one commit containing Phase 5 implementation.

Stage review:

```text
Stage review:
- Total goal: build a small but complete terminal Mahjong game in Go.
- Stage completed: Phase 5 improves terminal table readability with sections, command help, and recent event summaries.
- Evidence: go test ./..., go test ./... -cover, go build ./cmd/mahjong, scripted terminal smoke run.
- Remaining risk: output is still static text; full interactive TUI navigation remains a future phase.
```

