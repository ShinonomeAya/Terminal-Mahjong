# Terminal Mahjong Phase 6 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add one carefully bounded rules extension, seven pairs, with matching waits, tips, and simple scoring while keeping all existing terminal behavior stable.

**Architecture:** Keep special-hand detection in `win.go` next to standard win detection, and expose the hand pattern through a small `WinPattern` helper for scoring and tips. Do not build a full fan system, rule configuration system, or regional scoring table in Phase 6.

**Tech Stack:** Go 1.23, standard library only, `go test ./...`, `go build ./cmd/mahjong`, and scripted terminal smoke runs.

---

## Scope

Phase 6 extends simplified rules with seven pairs only. It does not add thirteen orphans, all-pongs scoring, flowers, riichi, regional fan tables, exposed kong variants, or rule presets.

Acceptance criteria:

- `CanWin` returns true for seven pairs.
- `WinPattern` identifies standard hands and seven-pairs hands.
- `WinningTiles` and `HandTips` include seven-pairs waits.
- `ScoreRound` adds a small seven-pairs bonus and labels it clearly.
- Terminal result output shows the score label without layout regression.
- Existing Phase 5 terminal sections, recent events, shanten tips, and claim behavior continue to pass tests.

## File Map

- Modify `internal/game/win.go`: seven-pairs detection and `WinPattern`.
- Modify `internal/game/win_test.go`: seven-pairs win and pattern tests.
- Modify `internal/game/analysis.go`: wait detection already uses `CanWin`; add a seven-pairs tips test.
- Modify `internal/game/score.go`: scoring context includes win pattern bonus.
- Modify `internal/game/game.go`: `printResult` score context receives winner hand pattern.
- Modify `README.md`: document Phase 6 as seven-pairs support, not full scoring.
- Modify `docs/workflow.md`: append Phase 6 review checklist.

## Task 1: Seven Pairs Win Detection

**Files:**
- Modify: `internal/game/win.go`
- Modify: `internal/game/win_test.go`

- [ ] **Step 1: Add seven-pairs tests**

Add tests:

```go
func TestCanWinWithSevenPairs(t *testing.T) {
	hand := mustTiles(t, "1m", "1m", "2m", "2m", "3m", "3m", "4p", "4p", "5p", "5p", "E", "E", "B", "B")
	if !CanWin(hand) {
		t.Fatal("expected seven-pairs hand to win")
	}
}

func TestWinPatternSevenPairs(t *testing.T) {
	hand := mustTiles(t, "1m", "1m", "2m", "2m", "3m", "3m", "4p", "4p", "5p", "5p", "E", "E", "B", "B")
	if got := WinPatternOf(hand); got != WinPatternSevenPairs {
		t.Fatalf("pattern = %v, want seven pairs", got)
	}
}
```

Run: `go test ./internal/game -run 'TestCanWinWithSevenPairs|TestWinPatternSevenPairs' -v`

Expected first result: at least `WinPatternOf` fails to compile, and `CanWin` may reject seven pairs.

- [ ] **Step 2: Implement seven-pairs detection and pattern enum**

Add:

- `type WinPattern int`
- `WinPatternNone`
- `WinPatternStandard`
- `WinPatternSevenPairs`
- `WinPatternOf(tiles []Tile) WinPattern`
- `CanWinSevenPairs(tiles []Tile) bool`

Update `CanWin` to return true when either `CanWinSevenPairs` or the existing standard shape succeeds.

Run: `go test ./internal/game -run 'TestCanWinWithSevenPairs|TestWinPatternSevenPairs|TestCanWin' -v`

Expected: win tests pass.

Step review:

```text
Step review:
- Stage goal: add one bounded rules extension without broad rule configuration.
- Step completed: seven pairs is a recognized winning pattern.
- Evidence: seven-pairs win and pattern tests pass.
- Next step: support seven-pairs waits and tips.
```

## Task 2: Seven Pairs Waits and Tips

**Files:**
- Modify: `internal/game/analysis_test.go`

- [ ] **Step 1: Add waits/tips tests**

Add tests:

```go
func TestWinningTilesIncludesSevenPairsWait(t *testing.T) {
	hand := mustTiles(t, "1m", "1m", "2m", "2m", "3m", "3m", "4p", "4p", "5p", "5p", "E", "E", "B")
	waits := WinningTiles(hand)
	if FormatTiles(waits) != "B" {
		t.Fatalf("waits = %s, want B", FormatTiles(waits))
	}
}

func TestHandTipsShowsSevenPairsWait(t *testing.T) {
	tips := HandTips(mustTiles(t, "1m", "1m", "2m", "2m", "3m", "3m", "4p", "4p", "5p", "5p", "E", "E", "B"))
	if !strings.Contains(tips, "tenpai") || !strings.Contains(tips, "B") {
		t.Fatalf("tips = %q", tips)
	}
}
```

Run: `go test ./internal/game -run 'TestWinningTilesIncludesSevenPairsWait|TestHandTipsShowsSevenPairsWait' -v`

Expected: tests pass after Task 1 because `WinningTiles` already delegates to `CanWin`.

Step review:

```text
Step review:
- Stage goal: keep player assistance aligned with expanded rules.
- Step completed: seven-pairs waits flow through existing tips.
- Evidence: waits/tips tests pass.
- Next step: add simple scoring label.
```

## Task 3: Seven Pairs Scoring Label

**Files:**
- Modify: `internal/game/score.go`
- Modify: `internal/game/game.go`
- Modify: `internal/game/game_test.go`

- [ ] **Step 1: Add scoring tests**

Add test:

```go
func TestScoreSevenPairsBonus(t *testing.T) {
	result := ScoreRound(WinContext{WinType: WinDiscard, Pattern: WinPatternSevenPairs})
	if result.Points != 3 || !strings.Contains(result.Label, "seven pairs +2") {
		t.Fatalf("result = %#v", result)
	}
}
```

Run: `go test ./internal/game -run TestScoreSevenPairsBonus -v`

Expected first result: build fails because `WinContext.Pattern` does not exist.

- [ ] **Step 2: Add pattern scoring**

Add `Pattern WinPattern` to `WinContext`.

Scoring rule:

- Existing base remains: self-draw +2, discard-win +1.
- Existing meld bonus remains: pong/kong +1 each.
- Seven pairs adds +2.
- Label includes `seven pairs +2`.

Run: `go test ./internal/game -run 'TestScoreSevenPairsBonus|TestScoreSelfDraw|TestScoreKongBonus' -v`

Expected: scoring tests pass.

- [ ] **Step 3: Pass winner pattern to result scoring**

Update `printResult` so the score context uses:

```go
Pattern: WinPatternOf(g.Players[g.Winner].Hand)
```

Run: `go test ./...`

Expected: all tests pass.

Step review:

```text
Step review:
- Stage goal: keep scoring understandable while adding one special hand.
- Step completed: seven-pairs results receive a simple labeled bonus.
- Evidence: scoring tests and go test ./... pass.
- Next step: update docs and verify.
```

## Task 4: Documentation, Verification, and Commit

**Files:**
- Modify: `README.md`
- Modify: `docs/workflow.md`

- [ ] **Step 1: Update README**

Add:

```markdown
## Phase 6 Direction

The game remains simplified and terminal-first. Phase 6 adds seven-pairs win support and a small seven-pairs scoring bonus, but still avoids full regional scoring tables.
```

Run: `rg -n "Phase 6 Direction|seven-pairs|regional scoring" README.md`

Expected: terms appear.

- [ ] **Step 2: Append workflow checklist**

Add:

```markdown
## Phase 6 Review Checklist

- Step reviews confirm seven-pairs detection, waits, tips, and scoring labels.
- Stage review compares the bounded rule extension against the total terminal Mahjong goal.
- Verification commands: `go test ./...`, `go build ./cmd/mahjong`, one scripted smoke run.
```

Run: `rg -n "Phase 6 Review Checklist|bounded rule extension|seven-pairs" docs/workflow.md`

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
- Smoke run keeps the Phase 5 terminal layout.
- `mahjong.exe` is removed.
- No external dependency appears in `go.mod`.

- [ ] **Step 4: Commit Phase 6**

Run:

```powershell
git add README.md docs/workflow.md docs/superpowers/plans/2026-06-06-terminal-mahjong-phase-6.md internal/game
git commit -m "feat: add seven pairs support"
```

Expected: one commit containing Phase 6 implementation.

Stage review:

```text
Stage review:
- Total goal: build a small but complete terminal Mahjong game in Go.
- Stage completed: Phase 6 adds seven-pairs win detection, waits, tips, and a simple scoring bonus.
- Evidence: go test ./..., go test ./... -cover, go build ./cmd/mahjong, scripted terminal smoke run.
- Remaining risk: full fan tables and exposed kong variants remain future phases.
```

