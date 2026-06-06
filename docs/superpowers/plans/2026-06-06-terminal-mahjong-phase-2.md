# Terminal Mahjong Phase 2 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Upgrade the MVP into a more Mahjong-like terminal game by adding chow, basic settlement, round summary, and stronger acceptance tests without expanding into full regional rules.

**Architecture:** Keep rule decisions inside `internal/game` and keep terminal I/O inside the existing `Game.Play` flow. Add the smallest new types needed for actions and settlement, then drive changes with tests before touching the interactive prompt.

**Tech Stack:** Go 1.23, standard library only, `go test ./...`, `go build ./cmd/mahjong`, and scripted terminal smoke runs.

---

## Scope

Phase 2 is about playability and a clearer one-round result. It does not add networking, save files, full scoring tables, flowers, riichi, seat wind rounds, or multiple-rule configuration.

Acceptance criteria:

- Human can chow only from the previous player's discard.
- AI can optionally chow when it improves local tile usefulness.
- Pong still outranks chow when multiple claims are possible.
- Win claims still outrank pong and chow.
- Game result shows winner, win type, source discard when relevant, and basic points.
- Tests cover chow eligibility, claim priority, settlement, and terminal command parsing.
- Each task includes a step review; the finished phase includes a stage review against the total goal.

## File Map

- Modify `internal/game/player.go`: add `MeldChow` and keep meld formatting stable.
- Modify `internal/game/game.go`: add chow claim flow, claim priority, and settlement output.
- Create `internal/game/score.go`: basic point calculation for self-draw, discard-win, pong/kong bonuses.
- Create `internal/game/action.go`: parse human commands into typed actions.
- Modify `internal/game/game_test.go`: add chow, priority, action parsing, and settlement tests.
- Modify `README.md`: document Phase 2 commands and simplified scoring.
- Modify `docs/workflow.md`: append Phase 2 review notes after implementation.

## Task 1: Typed Human Actions

**Files:**
- Create: `internal/game/action.go`
- Test: `internal/game/game_test.go`

- [ ] **Step 1: Add tests for command parsing**

Add tests that prove these inputs parse correctly:

```go
func TestParseActionDiscardByNumber(t *testing.T) {
	action, err := ParseAction("3")
	if err != nil {
		t.Fatal(err)
	}
	if action.Kind != ActionDiscard || action.Index != 2 {
		t.Fatalf("action = %#v, want discard index 2", action)
	}
}

func TestParseActionChow(t *testing.T) {
	action, err := ParseAction("c 2m 3m")
	if err != nil {
		t.Fatal(err)
	}
	if action.Kind != ActionChow || len(action.Tiles) != 2 {
		t.Fatalf("action = %#v, want chow with two tiles", action)
	}
}
```

Run: `go test ./internal/game -run TestParseAction -v`

Expected first result: build fails because `ParseAction` and action types do not exist.

- [ ] **Step 2: Implement minimal action parser**

Create `internal/game/action.go` with:

```go
package game

import (
	"fmt"
	"strconv"
	"strings"
)

type ActionKind int

const (
	ActionUnknown ActionKind = iota
	ActionDiscard
	ActionWin
	ActionKong
	ActionChow
	ActionQuit
)

type Action struct {
	Kind  ActionKind
	Index int
	Tiles []Tile
}

func ParseAction(line string) (Action, error) {
	text := strings.TrimSpace(strings.ToLower(line))
	if text == "q" {
		return Action{Kind: ActionQuit}, nil
	}
	if text == "h" {
		return Action{Kind: ActionWin}, nil
	}
	if strings.HasPrefix(text, "k ") {
		tile, ok := ParseTile(text[2:])
		if !ok {
			return Action{}, fmt.Errorf("unknown kong tile")
		}
		return Action{Kind: ActionKong, Tiles: []Tile{tile}}, nil
	}
	if strings.HasPrefix(text, "c ") {
		parts := strings.Fields(text)
		if len(parts) != 3 {
			return Action{}, fmt.Errorf("chow needs two hand tiles")
		}
		left, okLeft := ParseTile(parts[1])
		right, okRight := ParseTile(parts[2])
		if !okLeft || !okRight {
			return Action{}, fmt.Errorf("unknown chow tile")
		}
		return Action{Kind: ActionChow, Tiles: []Tile{left, right}}, nil
	}
	indexText := strings.TrimPrefix(text, "d ")
	index, err := strconv.Atoi(strings.TrimSpace(indexText))
	if err != nil {
		return Action{}, fmt.Errorf("unknown action")
	}
	return Action{Kind: ActionDiscard, Index: index - 1}, nil
}
```

Run: `go test ./internal/game -run TestParseAction -v`

Expected: parsing tests pass.

- [ ] **Step 3: Refactor human discard command handling to use `ParseAction`**

Replace direct string parsing in `humanDiscard` with `ParseAction`. Preserve existing commands: `h`, `k <tile>`, `d <index>`, `<index>`, `q`.

Run: `go test ./...`

Expected: all tests pass.

Step review:

```text
Step review:
- Stage goal: improve terminal playability without expanding rules.
- Step completed: human commands now parse through a typed action boundary.
- Evidence: go test ./...
- Next step: add chow eligibility rules.
```

## Task 2: Chow Eligibility

**Files:**
- Modify: `internal/game/player.go`
- Modify: `internal/game/game.go`
- Test: `internal/game/game_test.go`

- [ ] **Step 1: Add tests for chow eligibility**

Add tests:

```go
func TestCanChowPreviousDiscardWithTwoHandTiles(t *testing.T) {
	player := Player{Hand: mustTiles(t, "2m", "4m", "E")}
	options := ChowOptions(player, mustTile(t, "3m"))
	if len(options) != 1 || FormatTiles(options[0]) != "2m 3m 4m" {
		t.Fatalf("options = %#v, want 2m 3m 4m", options)
	}
}

func TestCannotChowHonorTile(t *testing.T) {
	player := Player{Hand: mustTiles(t, "E", "E")}
	options := ChowOptions(player, mustTile(t, "E"))
	if len(options) != 0 {
		t.Fatalf("options = %#v, want none", options)
	}
}
```

Run: `go test ./internal/game -run TestCanChow -v`

Expected first result: build fails because `ChowOptions` does not exist.

- [ ] **Step 2: Implement chow options**

Add `MeldChow` to `player.go`:

```go
const (
	MeldChow MeldKind = "chow"
	MeldPong MeldKind = "pong"
	MeldKong MeldKind = "kong"
)
```

Add `ChowOptions(player Player, discard Tile) [][]Tile` in `game.go` or a new small `claim.go` if `game.go` becomes hard to scan. It must:

- Return no options for honor tiles.
- Return only three-tile sequences containing the discard.
- Require the player to hold the other two tiles.
- Return sorted meld tiles.

Run: `go test ./internal/game -run TestCanChow -v`

Expected: chow eligibility tests pass.

- [ ] **Step 3: Add human chow claim after previous player discard**

Modify `resolveDiscardClaims` so chow is only offered to `(discarder + 1) % 4`, and only after win and pong claims are declined or unavailable.

Prompt text:

```text
Chow 3m with 2m 4m? [y/N]
```

If the player answers `y`, remove the two hand tiles, add a `chow(...)` meld, then prompt the claimer to discard. If multiple chow options exist, show them one at a time and accept the first option the player confirms.

Run: `go test ./...`

Expected: all tests pass.

- [ ] **Step 4: Add conservative AI chow**

Add `shouldAIChow(player Player, discard Tile, options [][]Tile) (int, bool)`.

Decision rule:

- Return `false` when no options exist.
- Score each option by summing `tileUsefulness` for the two consumed hand tiles before removal.
- Choose the lowest-scoring option only if the total is `<= 12`.
- This keeps AI chow conservative and avoids turning every sequence into a forced open meld.

Add a test:

```go
func TestAIChowChoosesLowUsefulnessTiles(t *testing.T) {
	player := Player{Hand: mustTiles(t, "2m", "4m", "2p", "3p", "4p", "E", "E")}
	options := ChowOptions(player, mustTile(t, "3m"))
	index, ok := shouldAIChow(player, mustTile(t, "3m"), options)
	if !ok || index != 0 {
		t.Fatalf("index=%d ok=%v, want option 0", index, ok)
	}
}
```

After AI chow succeeds, remove the two hand tiles from the selected option, add a `chow(...)` meld, and force the AI to discard immediately.

Run: `go test ./internal/game -run TestAIChow -v`

Expected: AI chow tests pass.

Step review:

```text
Step review:
- Stage goal: make the game more Mahjong-like while keeping simplified rules.
- Step completed: human and AI chow exist with correct eligibility and turn restriction.
- Evidence: chow tests, AI chow tests, and go test ./...
- Next step: enforce claim priority.
```

## Task 3: Claim Priority

**Files:**
- Modify: `internal/game/game.go`
- Test: `internal/game/game_test.go`

- [ ] **Step 1: Add priority tests**

Add tests proving:

- Discard-win beats pong.
- Pong beats chow.
- Chow is not offered to non-next players.

Use direct calls to `resolveDiscardClaims` with scripted input. Each test should assert final melds, winner, or current player.

Run: `go test ./internal/game -run TestClaimPriority -v`

Expected first result: at least one test fails until priority is explicit.

- [ ] **Step 2: Make claim priority explicit**

Structure `resolveDiscardClaims` into three ordered passes:

1. Win claims from all other players.
2. Pong claims from all other players.
3. Chow claim from next player only.

Do not add robbing kong or multi-winner rules in Phase 2.

Run: `go test ./...`

Expected: all tests pass.

Step review:

```text
Step review:
- Stage goal: improve rule correctness without introducing full regional complexity.
- Step completed: claim priority is deterministic and tested.
- Evidence: claim priority tests and go test ./...
- Next step: add basic settlement.
```

## Task 4: Basic Settlement

**Files:**
- Create: `internal/game/score.go`
- Modify: `internal/game/game.go`
- Test: `internal/game/game_test.go`

- [ ] **Step 1: Add settlement tests**

Add tests:

```go
func TestScoreSelfDraw(t *testing.T) {
	result := ScoreRound(WinContext{WinType: WinSelfDraw, Melds: []Meld{}})
	if result.Points != 2 || result.Label != "self-draw +2" {
		t.Fatalf("result = %#v", result)
	}
}

func TestScoreKongBonus(t *testing.T) {
	result := ScoreRound(WinContext{
		WinType: WinDiscard,
		Melds: []Meld{{Kind: MeldKong, Tiles: mustTiles(t, "1m", "1m", "1m", "1m")}},
	})
	if result.Points != 2 {
		t.Fatalf("points = %d, want 2", result.Points)
	}
}
```

Run: `go test ./internal/game -run TestScore -v`

Expected first result: build fails because score types do not exist.

- [ ] **Step 2: Implement simple scoring**

Create `score.go`:

- `WinSelfDraw`: base 2 points.
- `WinDiscard`: base 1 point.
- Each pong: +1 point.
- Each kong: +1 point.
- No other fan, no limit hands.

Run: `go test ./internal/game -run TestScore -v`

Expected: scoring tests pass.

- [ ] **Step 3: Include settlement in game result**

Track winner melds and win type when calling `finish`. Update final output:

```text
Winner: AI-2
Win: discard-win on 5s from AI-1
Score: discard-win +1, meld bonus +2 = 3
```

Run: `go test ./...`

Expected: all tests pass.

Step review:

```text
Step review:
- Stage goal: make completed rounds more satisfying and easier to inspect.
- Step completed: round result now includes basic settlement.
- Evidence: scoring tests and go test ./...
- Next step: document and smoke test the Phase 2 game.
```

## Task 5: Documentation, Smoke Run, and Stage Review

**Files:**
- Modify: `README.md`
- Modify: `docs/workflow.md`
- Test: terminal smoke command

- [ ] **Step 1: Update README**

Document:

- Commands: discard, win, kong, chow, quit.
- Simplified scoring.
- Phase 2 exclusions.

Run: `rg -n "chow|Score|Phase 2|go run" README.md`

Expected: all terms appear.

- [ ] **Step 2: Add Phase 2 stage review to workflow**

Append a dated section to `docs/workflow.md`:

```text
## Phase 2 Review Checklist

- Step review is written after each completed implementation task.
- Stage review compares chow, scoring, and terminal play against the total goal.
- Verification commands: go test ./..., go build ./cmd/mahjong, one scripted smoke run.
```

Run: `rg -n "Phase 2 Review Checklist|go build ./cmd/mahjong" docs/workflow.md`

Expected: both terms appear.

- [ ] **Step 3: Run final verification**

Run:

```powershell
go test ./...
go build ./cmd/mahjong
cmd /c "(echo 1& echo.& echo.& echo q) | mahjong.exe"
if (Test-Path 'mahjong.exe') { Remove-Item -LiteralPath 'mahjong.exe' }
```

Expected:

- Tests pass.
- Build exits 0.
- Smoke run shows initial table, one human discard, AI turns, and quit.
- Local `mahjong.exe` is removed after verification.

Stage review:

```text
Stage review:
- Total goal: build a small but complete terminal Mahjong game in Go.
- Stage completed: Phase 2 adds chow, deterministic claim priority, and basic settlement.
- Evidence: go test ./..., go build ./cmd/mahjong, scripted terminal smoke run.
- Remaining risk: scoring remains intentionally simplified and not a full regional rule set.
```
