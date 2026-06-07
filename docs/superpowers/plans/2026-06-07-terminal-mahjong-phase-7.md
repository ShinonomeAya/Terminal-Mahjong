# Terminal Mahjong Phase 7 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add replay-friendly event log export and summary support so completed terminal games can be inspected outside the live session.

**Architecture:** Build on Phase 3 `GameEvent` history and keep export logic inside `internal/game/replay.go`. Use JSON from the Go standard library only. Do not add networking, a replay viewer UI, or persistent save-file management in Phase 7.

**Tech Stack:** Go 1.23, standard library only, `go test ./...`, `go build ./cmd/mahjong`, and scripted terminal smoke runs.

---

## Scope

Phase 7 turns the in-memory event log into a reusable artifact. It does not add multiplayer, sockets, AI-vs-AI tournaments, a GUI replay viewer, or automatic file saving.

Acceptance criteria:

- `ReplayLog` captures seed, result, winner, score label, and event list.
- `Game.ReplayLog()` returns a replay log for the current game.
- `ReplayLog.ToJSON()` produces deterministic JSON using standard library encoding.
- `ReplaySummary` formats exported logs into a compact human-readable summary.
- Terminal result output mentions that the event log is replay-ready.
- Existing Phase 6 rules and Phase 5 terminal layout continue to pass tests.

## File Map

- Create `internal/game/replay.go`: replay DTO, JSON export, replay summary.
- Create `internal/game/replay_test.go`: replay log, JSON, and summary tests.
- Modify `internal/game/game.go`: store game seed for replay metadata.
- Modify `internal/game/render.go`: mention replay-ready event log in result output.
- Modify `README.md`: document Phase 7 replay/export direction.
- Modify `docs/workflow.md`: append Phase 7 review checklist.

## Task 1: Replay Log Model

**Files:**
- Create: `internal/game/replay.go`
- Create: `internal/game/replay_test.go`
- Modify: `internal/game/game.go`

- [ ] **Step 1: Add replay model tests**

Add tests:

```go
func TestReplayLogCapturesGameMetadata(t *testing.T) {
	game := NewGame(7)
	game.RecordEvent(EventDraw, 0, mustTile(t, "1m"), "")
	game.finish(0, "self-draw", WinSelfDraw)
	log := game.ReplayLog()
	if log.Seed != 7 || log.Winner != "You" || log.Result == "" || len(log.Events) != 2 {
		t.Fatalf("log = %#v", log)
	}
}
```

Run: `go test ./internal/game -run TestReplayLogCapturesGameMetadata -v`

Expected first result: build fails because `ReplayLog` does not exist and `Game` does not store seed.

- [ ] **Step 2: Store seed and implement replay DTO**

Add `Seed int64` to `Game` and set it in `NewGame`.

Create `replay.go` with:

- `type ReplayLog struct`
- `func (g *Game) ReplayLog() ReplayLog`

Include:

- `Seed int64`
- `Winner string`
- `Result string`
- `Score string`
- `Events []GameEvent`

Run: `go test ./internal/game -run TestReplayLogCapturesGameMetadata -v`

Expected: replay model test passes.

Step review:

```text
Step review:
- Stage goal: make completed games inspectable after the terminal session.
- Step completed: game state can produce a replay DTO.
- Evidence: replay model test passes.
- Next step: add deterministic JSON export.
```

## Task 2: JSON Export

**Files:**
- Modify: `internal/game/replay.go`
- Modify: `internal/game/replay_test.go`

- [ ] **Step 1: Add JSON export test**

Add test:

```go
func TestReplayLogToJSON(t *testing.T) {
	game := NewGame(7)
	game.RecordEvent(EventDiscard, 0, mustTile(t, "1m"), "")
	data, err := game.ReplayLog().ToJSON()
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, want := range []string{`"seed":7`, `"events"`, `"discard"`} {
		if !strings.Contains(text, want) {
			t.Fatalf("json missing %s:\n%s", want, text)
		}
	}
}
```

Run: `go test ./internal/game -run TestReplayLogToJSON -v`

Expected first result: build fails because `ToJSON` does not exist.

- [ ] **Step 2: Implement JSON export**

Implement:

```go
func (r ReplayLog) ToJSON() ([]byte, error)
```

Use `json.MarshalIndent(r, "", "  ")`.

Run: `go test ./internal/game -run TestReplayLogToJSON -v`

Expected: JSON export test passes.

Step review:

```text
Step review:
- Stage goal: make event history reusable outside the live game.
- Step completed: replay logs export to deterministic JSON.
- Evidence: JSON export test passes.
- Next step: add human-readable replay summary.
```

## Task 3: Replay Summary and Result Hint

**Files:**
- Modify: `internal/game/replay.go`
- Modify: `internal/game/replay_test.go`
- Modify: `internal/game/render.go`

- [ ] **Step 1: Add summary tests**

Add tests:

```go
func TestReplaySummaryIncludesResultAndEvents(t *testing.T) {
	game := NewGame(7)
	game.RecordEvent(EventDiscard, 0, mustTile(t, "1m"), "")
	game.finish(0, "self-draw", WinSelfDraw)
	summary := ReplaySummary(game.ReplayLog())
	for _, want := range []string{"Replay", "Winner: You", "Events: 2"} {
		if !strings.Contains(summary, want) {
			t.Fatalf("summary missing %q:\n%s", want, summary)
		}
	}
}

func TestPrintResultMentionsReplayReadyLog(t *testing.T) {
	game := NewGame(7)
	game.finish(0, "self-draw", WinSelfDraw)
	var out strings.Builder
	game.printResult(&out)
	if !strings.Contains(out.String(), "Replay-ready event log") {
		t.Fatalf("result output:\n%s", out.String())
	}
}
```

Run: `go test ./internal/game -run 'TestReplaySummary|TestPrintResultMentionsReplayReadyLog' -v`

Expected first result: summary test fails because `ReplaySummary` does not exist; result test fails until output is updated.

- [ ] **Step 2: Implement summary and result hint**

Implement:

```go
func ReplaySummary(log ReplayLog) string
```

Include seed, winner/result, score, event count, and the last five events.

Update `printResult` to include:

```text
Replay-ready event log: yes
```

Run: `go test ./internal/game -run 'TestReplaySummary|TestPrintResultMentionsReplayReadyLog' -v`

Expected: summary and result hint tests pass.

Step review:

```text
Step review:
- Stage goal: make replay artifacts understandable without building a viewer.
- Step completed: replay logs have text summaries and the terminal result advertises readiness.
- Evidence: replay summary and result hint tests pass.
- Next step: update docs and verify.
```

## Task 4: Documentation, Verification, and Commit

**Files:**
- Modify: `README.md`
- Modify: `docs/workflow.md`

- [ ] **Step 1: Update README**

Add:

```markdown
## Phase 7 Direction

The game remains terminal-first. Phase 7 adds replay-ready event log export and summaries using standard-library JSON, without adding networking or a replay GUI.
```

Run: `rg -n "Phase 7 Direction|replay-ready|standard-library JSON|networking" README.md`

Expected: terms appear.

- [ ] **Step 2: Append workflow checklist**

Add:

```markdown
## Phase 7 Review Checklist

- Step reviews confirm replay metadata, JSON export, replay summaries, and result hints.
- Stage review compares replay support against the total terminal Mahjong goal.
- Verification commands: `go test ./...`, `go build ./cmd/mahjong`, one scripted smoke run.
```

Run: `rg -n "Phase 7 Review Checklist|replay support|JSON export" docs/workflow.md`

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
- Smoke run keeps the terminal layout and mentions replay-ready log.
- `mahjong.exe` is removed.
- No external dependency appears in `go.mod`.

- [ ] **Step 4: Commit Phase 7**

Run:

```powershell
git add README.md docs/workflow.md docs/superpowers/plans/2026-06-07-terminal-mahjong-phase-7.md internal/game
git commit -m "feat: add replay log export"
```

Expected: one commit containing Phase 7 implementation.

Stage review:

```text
Stage review:
- Total goal: build a small but complete terminal Mahjong game in Go.
- Stage completed: Phase 7 adds replay-ready logs, JSON export, summaries, and terminal result hints.
- Evidence: go test ./..., go test ./... -cover, go build ./cmd/mahjong, scripted terminal smoke run.
- Remaining risk: replay logs are exportable data only; file writing and interactive replay viewer remain future phases.
```

