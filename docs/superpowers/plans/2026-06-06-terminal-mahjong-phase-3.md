# Terminal Mahjong Phase 3 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add deterministic round event logging, replay-oriented summaries, and focused game-file boundaries while keeping the project a terminal Mahjong game.

**Architecture:** Introduce a small `GameEvent` log inside `internal/game` and record every meaningful round transition from existing game operations. Split claim handling and terminal rendering out of `game.go` after tests protect behavior, so future shanten, AI, and terminal UI upgrades can build on event history instead of parsing printed text.

**Tech Stack:** Go 1.23, standard library only, `go test ./...`, `go build ./cmd/mahjong`, and scripted terminal smoke runs.

---

## Scope

Phase 3 is an engine-hardening phase, not a visual TUI phase. It preserves the current terminal text interface while adding event history, deterministic smoke coverage, and smaller code boundaries.

Acceptance criteria:

- Every draw, discard, chow, pong, kong, win, quit, and wall-exhausted draw records a typed `GameEvent`.
- A completed or interrupted round can print a compact event summary without parsing terminal output.
- Seeded games with the same scripted input produce the same event sequence.
- `game.go` is reduced by moving claim handling and terminal rendering into focused files.
- Existing commands and Phase 2 behavior continue to pass tests.
- Documentation explains why Phase 3 prepares for later terminal UI and AI work.

## File Map

- Create `internal/game/event.go`: event kinds, `GameEvent`, `RecordEvent`, and event summary formatting.
- Create `internal/game/claim.go`: discard claim resolution, chow/pong helpers, AI claim choices.
- Create `internal/game/render.go`: terminal table and result rendering helpers.
- Modify `internal/game/game.go`: keep round setup, draw loop, human/AI turn orchestration, and state fields.
- Modify `internal/game/game_test.go`: add event log, deterministic replay, and summary tests.
- Modify `README.md`: document Phase 3 event-log direction while preserving terminal identity.
- Modify `docs/workflow.md`: append Phase 3 review checklist.

## Task 1: Event Model

**Files:**
- Create: `internal/game/event.go`
- Modify: `internal/game/game.go`
- Test: `internal/game/game_test.go`

- [ ] **Step 1: Add event model tests**

Add tests:

```go
func TestRecordEventStoresRoundTransition(t *testing.T) {
	game := NewGame(1)
	game.RecordEvent(EventDiscard, 0, mustTile(t, "3m"), "discarded from hand")
	if len(game.Events) != 1 {
		t.Fatalf("events = %d, want 1", len(game.Events))
	}
	event := game.Events[0]
	if event.Kind != EventDiscard || event.Player != 0 || event.Tile != mustTile(t, "3m") {
		t.Fatalf("event = %#v", event)
	}
}

func TestEventSummaryFormatsCompactHistory(t *testing.T) {
	game := NewGame(1)
	game.RecordEvent(EventDraw, 0, mustTile(t, "3m"), "")
	game.RecordEvent(EventDiscard, 0, mustTile(t, "3m"), "")
	summary := EventSummary(game.Events)
	if !strings.Contains(summary, "You draw 3m") || !strings.Contains(summary, "You discard 3m") {
		t.Fatalf("summary:\n%s", summary)
	}
}
```

Run: `go test ./internal/game -run 'TestRecordEvent|TestEventSummary' -v`

Expected first result: build fails because event types do not exist.

- [ ] **Step 2: Implement event types and summary**

Create `event.go`:

```go
package game

import "strings"

type EventKind string

const (
	EventDraw          EventKind = "draw"
	EventDiscard       EventKind = "discard"
	EventChow          EventKind = "chow"
	EventPong          EventKind = "pong"
	EventKong          EventKind = "kong"
	EventWin           EventKind = "win"
	EventQuit          EventKind = "quit"
	EventWallExhausted EventKind = "wall-exhausted"
)

type GameEvent struct {
	Turn   int
	Kind   EventKind
	Player int
	Tile   Tile
	Note   string
}

func (g *Game) RecordEvent(kind EventKind, player int, tile Tile, note string) {
	g.Events = append(g.Events, GameEvent{
		Turn:   len(g.Events) + 1,
		Kind:   kind,
		Player: player,
		Tile:   tile,
		Note:   note,
	})
}

func EventSummary(events []GameEvent) string {
	lines := make([]string, 0, len(events))
	for _, event := range events {
		lines = append(lines, event.String())
	}
	return strings.Join(lines, "\n")
}
```

Add `Events []GameEvent` to `Game`.

Implement `GameEvent.String()` with player labels `You`, `AI-1`, `AI-2`, `AI-3` and concise action text.

Run: `go test ./internal/game -run 'TestRecordEvent|TestEventSummary' -v`

Expected: event model tests pass.

Step review:

```text
Step review:
- Stage goal: make round state observable without parsing terminal output.
- Step completed: typed events can be recorded and summarized.
- Evidence: event model tests pass.
- Next step: record events from actual game operations.
```

## Task 2: Record Existing Game Operations

**Files:**
- Modify: `internal/game/game.go`
- Modify: `internal/game/game_test.go`

- [ ] **Step 1: Add operation event tests**

Add tests:

```go
func TestPlayRecordsQuitEvent(t *testing.T) {
	game := NewGame(1)
	var out strings.Builder
	game.Play(strings.NewReader("q\n"), &out)
	if !hasEvent(game.Events, EventQuit) {
		t.Fatalf("events = %#v", game.Events)
	}
}

func TestPlayRecordsDrawAndDiscardEvents(t *testing.T) {
	game := NewGame(1)
	var out strings.Builder
	game.Play(strings.NewReader("1\nq\n"), &out)
	if !hasEvent(game.Events, EventDraw) || !hasEvent(game.Events, EventDiscard) {
		t.Fatalf("events = %#v", game.Events)
	}
}
```

Add helper:

```go
func hasEvent(events []GameEvent, kind EventKind) bool {
	for _, event := range events {
		if event.Kind == kind {
			return true
		}
	}
	return false
}
```

Run: `go test ./internal/game -run 'TestPlayRecords' -v`

Expected first result: tests fail because existing operations do not record events.

- [ ] **Step 2: Record draw, discard, quit, and wall exhaustion**

Record:

- `EventDraw` in `draw`.
- `EventDiscard` immediately after human and AI discards.
- `EventQuit` when the human chooses `q`.
- `EventWallExhausted` when the wall ends.

Run: `go test ./internal/game -run 'TestPlayRecords' -v`

Expected: operation event tests pass.

- [ ] **Step 3: Record claim and win events**

Record:

- `EventChow` in `claimChow`.
- `EventPong` in `claimPong`.
- `EventKong` in human and AI kong paths.
- `EventWin` in `finish` for all winning paths.

Run: `go test ./...`

Expected: all existing tests pass.

Step review:

```text
Step review:
- Stage goal: make full rounds inspectable through event history.
- Step completed: existing operations now emit typed events.
- Evidence: operation event tests and go test ./...
- Next step: prove deterministic event replay for seeded scripted games.
```

## Task 3: Deterministic Scripted Runs

**Files:**
- Modify: `internal/game/game_test.go`

- [ ] **Step 1: Add deterministic event-sequence test**

Add test:

```go
func TestSeededScriptProducesStableEventSummary(t *testing.T) {
	first := runScriptedSummary(t, 7, "1\n\n\nq\n")
	second := runScriptedSummary(t, 7, "1\n\n\nq\n")
	if first != second {
		t.Fatalf("event summaries differ:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}

func runScriptedSummary(t *testing.T, seed int64, input string) string {
	t.Helper()
	game := NewGame(seed)
	var out strings.Builder
	game.Play(strings.NewReader(input), &out)
	return EventSummary(game.Events)
}
```

Run: `go test ./internal/game -run TestSeededScriptProducesStableEventSummary -v`

Expected: test passes after Task 2 event recording.

- [ ] **Step 2: Add a golden-shape assertion without a fragile full snapshot**

Extend the test to assert the summary contains at least:

- `You draw`
- `You discard`
- either `quit` or a win event

Do not assert the full summary string; tile order can legitimately change if setup changes in later phases.

Run: `go test ./...`

Expected: all tests pass.

Step review:

```text
Step review:
- Stage goal: prepare for replay and future AI testing.
- Step completed: seeded scripted runs produce stable event summaries.
- Evidence: deterministic script test and go test ./...
- Next step: split claim handling out of game.go.
```

## Task 4: Split Claim Handling

**Files:**
- Create: `internal/game/claim.go`
- Modify: `internal/game/game.go`

- [ ] **Step 1: Move claim functions without behavior changes**

Move these functions from `game.go` to `claim.go`:

- `ChowOptions`
- `resolveDiscardClaims`
- `shouldAIChow`
- `chowHandTiles`
- `claimChow`
- `shouldAIPong`
- `claimPong`

Keep function signatures unchanged.

Run: `go test ./...`

Expected: all tests pass.

- [ ] **Step 2: Check file sizes**

Run:

```powershell
Get-ChildItem -Path internal\game -Filter '*.go' | ForEach-Object { "$($_.Name) $((Get-Content $_.FullName).Count)" }
```

Expected:

- `claim.go` contains claim-specific logic.
- `game.go` is smaller than before this phase.

Step review:

```text
Step review:
- Stage goal: keep future terminal UI and AI work from piling into game.go.
- Step completed: claim handling moved behind a focused file boundary.
- Evidence: go test ./... and file-size check.
- Next step: split terminal rendering.
```

## Task 5: Split Terminal Rendering and Add Event Summary Output

**Files:**
- Create: `internal/game/render.go`
- Modify: `internal/game/game.go`
- Modify: `internal/game/game_test.go`

- [ ] **Step 1: Move rendering helpers**

Move these functions from `game.go` to `render.go`:

- `printTable`
- `printResult`
- `drawVerb`

Keep signatures unchanged.

Run: `go test ./...`

Expected: all tests pass.

- [ ] **Step 2: Include event count in result output**

Update `printResult` so the end of a game includes:

```text
Events: 12
```

Add test:

```go
func TestPrintResultIncludesEventCount(t *testing.T) {
	game := NewGame(1)
	game.RecordEvent(EventDraw, 0, mustTile(t, "1m"), "")
	game.finish(0, "self-draw", WinSelfDraw)
	var out strings.Builder
	game.printResult(&out)
	if !strings.Contains(out.String(), "Events: 2") {
		t.Fatalf("result output:\n%s", out.String())
	}
}
```

Run: `go test ./internal/game -run TestPrintResultIncludesEventCount -v`

Expected: result event-count test passes.

Step review:

```text
Step review:
- Stage goal: preserve terminal identity while preparing better TUI rendering later.
- Step completed: rendering is isolated and result output exposes event history size.
- Evidence: result rendering tests and go test ./...
- Next step: update docs and run final verification.
```

## Task 6: Documentation, Verification, and Commit

**Files:**
- Modify: `README.md`
- Modify: `docs/workflow.md`

- [ ] **Step 1: Update README**

Add a `Phase 3 Direction` section:

```markdown
## Phase 3 Direction

The game remains terminal-first. Phase 3 adds typed event logs and deterministic scripted runs so later shanten, AI, replay, and terminal UI upgrades can use game state directly instead of scraping printed text.
```

Run: `rg -n "Phase 3 Direction|terminal-first|event logs" README.md`

Expected: terms appear.

- [ ] **Step 2: Append workflow review checklist**

Append:

```markdown
## Phase 3 Review Checklist

- Step reviews confirm event logging, deterministic scripts, claim split, and render split.
- Stage review compares event history and code boundaries against the total terminal Mahjong goal.
- Verification commands: `go test ./...`, `go build ./cmd/mahjong`, one scripted smoke run.
```

Run: `rg -n "Phase 3 Review Checklist|event history|code boundaries" docs/workflow.md`

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
- Smoke run shows terminal table and game-over result.
- `mahjong.exe` is removed.
- Only intended source and documentation files are modified before commit.

- [ ] **Step 4: Commit Phase 3**

Run:

```powershell
git add README.md docs/workflow.md docs/superpowers/plans/2026-06-06-terminal-mahjong-phase-3.md internal/game
git commit -m "feat: add terminal mahjong event log"
```

Expected: one commit containing Phase 3 implementation.

Stage review:

```text
Stage review:
- Total goal: build a small but complete terminal Mahjong game in Go.
- Stage completed: Phase 3 adds typed event history, deterministic scripted verification, and smaller engine/rendering boundaries.
- Evidence: go test ./..., go test ./... -cover, go build ./cmd/mahjong, scripted terminal smoke run.
- Remaining risk: event logs are in-memory only; file export and replay viewer remain future phases.
```

