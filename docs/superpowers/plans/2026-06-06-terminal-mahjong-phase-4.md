# Terminal Mahjong Phase 4 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add standard-hand shanten, waits, player tips, and a shanten-aware AI discard heuristic while keeping the game terminal-first.

**Architecture:** Keep hand analysis in a new `internal/game/analysis.go` file so the rules engine and renderer can consume the same result. Use the existing simplified `4 melds + 1 pair` win shape only; do not add seven pairs, special hands, or regional scoring in this phase.

**Tech Stack:** Go 1.23, standard library only, `go test ./...`, `go build ./cmd/mahjong`, and scripted terminal smoke runs.

---

## Scope

Phase 4 teaches the terminal game to explain the player's hand. It does not add a GUI, a full TUI library, seven pairs, thirteen orphans, riichi rules, regional fan tables, or full defensive AI.

Acceptance criteria:

- `ShantenStandard` returns `-1` for a winning standard hand, `0` for tenpai, and positive values for incomplete hands.
- `WinningTiles` returns the tiles that complete a 13-tile tenpai hand under the existing standard win shape.
- The terminal table shows a compact `Tips:` line for the human player.
- AI discard choice prefers discards that lower shanten, falling back to the existing tile-usefulness heuristic.
- Existing Phase 3 event logging and Phase 2 claim behavior continue to pass tests.
- Documentation explains Phase 4 as terminal-first player assistance.

## File Map

- Create `internal/game/analysis.go`: shanten search, waits, discard recommendation.
- Create `internal/game/analysis_test.go`: focused hand-analysis tests.
- Modify `internal/game/ai.go`: use `BestDiscardIndex` before old usefulness fallback.
- Modify `internal/game/render.go`: show human tips.
- Modify `README.md`: add Phase 4 direction and tips description.
- Modify `docs/workflow.md`: append Phase 4 review checklist.

## Task 1: Standard Shanten Core

**Files:**
- Create: `internal/game/analysis.go`
- Create: `internal/game/analysis_test.go`

- [ ] **Step 1: Add shanten tests**

Add tests:

```go
func TestShantenStandardWinningHand(t *testing.T) {
	hand := mustTiles(t,
		"1m", "2m", "3m",
		"4m", "5m", "6m",
		"2p", "3p", "4p",
		"7s", "7s", "7s",
		"E", "E",
	)
	if got := ShantenStandard(hand); got != -1 {
		t.Fatalf("shanten = %d, want -1", got)
	}
}

func TestShantenStandardTenpaiHand(t *testing.T) {
	hand := mustTiles(t,
		"1m", "2m", "3m",
		"4m", "5m", "6m",
		"2p", "3p", "4p",
		"7s", "7s", "7s",
		"E",
	)
	if got := ShantenStandard(hand); got != 0 {
		t.Fatalf("shanten = %d, want 0", got)
	}
}
```

Run: `go test ./internal/game -run TestShantenStandard -v`

Expected first result: build fails because `ShantenStandard` does not exist.

- [ ] **Step 2: Implement bounded standard-hand shanten search**

Implement `ShantenStandard(tiles []Tile) int`.

Rules:

- Return `-1` immediately when `CanWin(tiles)` is true.
- For 13-tile hands, return `0` when any legal tile added to the hand makes `CanWin` true.
- For other non-winning hands, compute a conservative standard-hand shanten number by recursively counting complete melds, pairs, and partial sequences.
- Cap returned values at `6`; exactness beyond early-stage hand quality is not required in Phase 4.

Run: `go test ./internal/game -run TestShantenStandard -v`

Expected: shanten tests pass.

Step review:

```text
Step review:
- Stage goal: add hand explanation while preserving simplified rules.
- Step completed: standard winning and tenpai hands can be classified.
- Evidence: shanten tests pass.
- Next step: expose waits for tenpai hands.
```

## Task 2: Waiting Tiles and Tips Text

**Files:**
- Modify: `internal/game/analysis.go`
- Modify: `internal/game/analysis_test.go`

- [ ] **Step 1: Add waiting-tile tests**

Add tests:

```go
func TestWinningTilesFindsSinglePairWait(t *testing.T) {
	hand := mustTiles(t,
		"1m", "2m", "3m",
		"4m", "5m", "6m",
		"2p", "3p", "4p",
		"7s", "7s", "7s",
		"E",
	)
	waits := WinningTiles(hand)
	if FormatTiles(waits) != "E" {
		t.Fatalf("waits = %s, want E", FormatTiles(waits))
	}
}

func TestHandTipsShowsTenpaiWaits(t *testing.T) {
	tips := HandTips(mustTiles(t,
		"1m", "2m", "3m",
		"4m", "5m", "6m",
		"2p", "3p", "4p",
		"7s", "7s", "7s",
		"E",
	))
	if !strings.Contains(tips, "tenpai") || !strings.Contains(tips, "E") {
		t.Fatalf("tips = %q", tips)
	}
}
```

Run: `go test ./internal/game -run 'TestWinningTiles|TestHandTips' -v`

Expected first result: build fails because waits/tips functions do not exist.

- [ ] **Step 2: Implement `WinningTiles` and `HandTips`**

Implement:

- `WinningTiles(tiles []Tile) []Tile`: for each tile type with fewer than 4 copies in the hand, add it and return tiles that make `CanWin` true.
- `HandTips(tiles []Tile) string`: return one of:
  - `winning hand`
  - `tenpai: waits E 3m`
  - `shanten: N`

Run: `go test ./internal/game -run 'TestWinningTiles|TestHandTips' -v`

Expected: waits/tips tests pass.

Step review:

```text
Step review:
- Stage goal: make terminal play more informative without changing UI medium.
- Step completed: tenpai waits and compact tips text exist.
- Evidence: waits/tips tests pass.
- Next step: render tips in the terminal table.
```

## Task 3: Terminal Tips Rendering

**Files:**
- Modify: `internal/game/render.go`
- Modify: `internal/game/game_test.go`

- [ ] **Step 1: Add render test**

Add test:

```go
func TestPrintTableIncludesHumanTips(t *testing.T) {
	game := NewGame(1)
	game.Players[0].Hand = mustTiles(t,
		"1m", "2m", "3m",
		"4m", "5m", "6m",
		"2p", "3p", "4p",
		"7s", "7s", "7s",
		"E",
	)
	var out strings.Builder
	game.printTable(&out)
	if !strings.Contains(out.String(), "Tips: tenpai") {
		t.Fatalf("table output:\n%s", out.String())
	}
}
```

Run: `go test ./internal/game -run TestPrintTableIncludesHumanTips -v`

Expected first result: test fails because table does not show tips.

- [ ] **Step 2: Render tips**

Update `printTable` so after the human hand line it prints:

```text
Tips: tenpai: waits E
```

Use `HandTips(g.Players[0].Hand)`.

Run: `go test ./internal/game -run TestPrintTableIncludesHumanTips -v`

Expected: render test passes.

Step review:

```text
Step review:
- Stage goal: keep the terminal identity but make the interface more helpful.
- Step completed: terminal table now shows hand tips.
- Evidence: render tips test passes.
- Next step: improve AI discard with shanten analysis.
```

## Task 4: Shanten-Aware AI Discard

**Files:**
- Modify: `internal/game/analysis.go`
- Modify: `internal/game/ai.go`
- Modify: `internal/game/analysis_test.go`

- [ ] **Step 1: Add best-discard test**

Add test:

```go
func TestBestDiscardIndexKeepsCompleteMelds(t *testing.T) {
	hand := mustTiles(t,
		"1m", "2m", "3m",
		"4m", "5m", "6m",
		"2p", "3p", "4p",
		"7s", "7s", "7s",
		"E", "9m",
	)
	index := BestDiscardIndex(hand)
	if index < 0 || hand[index] != mustTile(t, "9m") {
		t.Fatalf("discard = %d:%s, want 9m", index, hand[index])
	}
}
```

Run: `go test ./internal/game -run TestBestDiscardIndex -v`

Expected first result: build fails because `BestDiscardIndex` does not exist.

- [ ] **Step 2: Implement best discard**

Implement `BestDiscardIndex(hand []Tile) int`.

Decision order:

1. For each discard candidate, remove that tile and compute `ShantenStandard`.
2. Prefer the lowest shanten.
3. Break ties by discarding the tile with the lowest `tileUsefulness`.
4. Return `ChooseAIDiscard(hand)` when the hand is empty or no candidate improves analysis.

Update `ChooseAIDiscard` to call `BestDiscardIndex` for hands with at least 2 tiles. Avoid recursion by keeping the old heuristic in a new helper named `chooseAIDiscardByUsefulness`.

Run: `go test ./internal/game -run 'TestBestDiscardIndex|TestChooseAIDiscard' -v`

Expected: AI discard tests pass.

Step review:

```text
Step review:
- Stage goal: use hand analysis for better play, not just display.
- Step completed: AI discard now prefers lower-shanten hands.
- Evidence: best-discard and AI tests pass.
- Next step: update docs and verify the phase.
```

## Task 5: Documentation, Verification, and Commit

**Files:**
- Modify: `README.md`
- Modify: `docs/workflow.md`

- [ ] **Step 1: Update README**

Add a `Phase 4 Direction` section:

```markdown
## Phase 4 Direction

The game remains terminal-first. Phase 4 adds standard-hand shanten, tenpai waits, and compact table tips so the terminal game becomes easier to understand without becoming a GUI.
```

Run: `rg -n "Phase 4 Direction|shanten|tenpai|terminal-first" README.md`

Expected: terms appear.

- [ ] **Step 2: Append workflow review checklist**

Append:

```markdown
## Phase 4 Review Checklist

- Step reviews confirm shanten, waits, terminal tips, and AI discard behavior.
- Stage review compares player assistance against the total terminal Mahjong goal.
- Verification commands: `go test ./...`, `go build ./cmd/mahjong`, one scripted smoke run.
```

Run: `rg -n "Phase 4 Review Checklist|player assistance|terminal Mahjong" docs/workflow.md`

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
- Smoke run shows terminal table with a `Tips:` line.
- `mahjong.exe` is removed.
- Only intended source and documentation files are modified before commit.

- [ ] **Step 4: Commit Phase 4**

Run:

```powershell
git add README.md docs/workflow.md docs/superpowers/plans/2026-06-06-terminal-mahjong-phase-4.md internal/game
git commit -m "feat: add terminal mahjong hand tips"
```

Expected: one commit containing Phase 4 implementation.

Stage review:

```text
Stage review:
- Total goal: build a small but complete terminal Mahjong game in Go.
- Stage completed: Phase 4 adds standard-hand shanten, waits, table tips, and shanten-aware AI discard.
- Evidence: go test ./..., go test ./... -cover, go build ./cmd/mahjong, scripted terminal smoke run.
- Remaining risk: shanten is limited to the standard hand shape; special hands and defensive AI remain future phases.
```

