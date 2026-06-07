# Terminal Client Polish Phase 9 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move the Unicode terminal client from "usable" to "client-like": clearer visual hierarchy, stronger selected-tile feedback, more reliable mouse affordance, and consistent menu/game-over presentation.

**Architecture:** Keep game rules in `internal/game`. Limit Phase 9 to `internal/tui` rendering, input feedback, screen state, and documentation. Avoid adding regional rules, replay viewer, networking, or a GUI.

**Tech Stack:** Go 1.23, Bubble Tea v0.27.1, Unicode Mahjong tile glyphs, standard-library tests.

---

## Workflow Contract

Use the existing project loop:

1. Define the stage goal.
2. Complete one step.
3. Verify that step.
4. Write a step review.
5. Repeat until the stage passes.
6. Write a stage review against the total goal.

Step review format:

```text
Step review:
- Stage goal:
- Step completed:
- Evidence:
- Next step:
```

Stage review format:

```text
Stage review:
- Total goal:
- Stage completed:
- Evidence:
- Remaining risk:
```

Total goal:

```text
Build a terminal-first Mahjong client that is readable, attractive, and playable without command typing.
```

## Current Problems To Avoid Regressing

- Unicode width can break decorative borders in Windows Terminal.
- A selected tile must be obvious in the hand row and in a dedicated status line.
- Mouse clicks must map to rendered tile rows, not stale hard-coded coordinates.
- Game Over must offer restart, main menu, and quit choices.
- Layout must remain understandable at normal PowerShell/Windows Terminal sizes.

## Planned File Scope

- Modify: `internal/tui/layout.go`
  - Add small reusable rendering helpers, compact panels, and consistent tile cells.
- Modify: `internal/tui/input.go`
  - Add clearer hover/click/selection feedback where feasible.
- Modify: `internal/tui/model.go`
  - Add UI-only fields if needed, such as last action feedback or transient status.
- Modify: `internal/tui/layout_test.go`
  - Add visual regression tests for line width, labels, panels, and selected cells.
- Modify: `internal/tui/model_test.go`
  - Add interaction tests for menus, game-over choices, and selected-tile state.
- Modify: `README.md`
  - Document polished controls and fallback limitations if behavior changes.
- Modify: `docs/workflow.md`
  - Add Phase 9 review checklist.

## Stage 1: Visual Tokens And Layout Helpers

Goal: create small rendering helpers so the table can become more polished without fragile ad hoc string formatting.

Step review question: did this step make future UI changes more predictable without changing game behavior?

Stage review question: can the TUI use consistent labels, separators, and tile cells in tests?

- [ ] **Step 1: Add tests for fixed tile cell rendering**

Add tests that assert:

- Unselected Unicode cell contains `[02]` and the tile glyph.
- Selected Unicode cell contains `▶ [02]` and `◀`.
- Fallback cell contains internal notation such as `2m`.
- Cell rendering stays under a known rune width.

Run:

```powershell
go test ./internal/tui -run "TestTileCell" -v
```

Expected first result: FAIL because tile cell helpers do not exist.

- [ ] **Step 2: Implement tile cell helpers**

Add focused helpers in `internal/tui/layout.go`, such as:

- `renderTileCell(index int, tile game.Tile, selected bool, unicode bool) string`
- `runeWidth(text string) int`
- `padRightRunes(text string, width int) string`

Keep them package-private and tested. Do not introduce lipgloss styling yet unless plain strings cannot satisfy the tests.

Run:

```powershell
go test ./internal/tui -run "TestTileCell" -v
```

Expected: PASS.

- [ ] **Step 3: Stage 1 verification**

Run:

```powershell
go test ./internal/tui -v
go test ./...
```

Expected: PASS.

Commit:

```powershell
git add internal/tui/layout.go internal/tui/layout_test.go
git commit -m "feat: add tui rendering helpers"
```

Write the Stage 1 review.

## Stage 2: Client-Like Table Composition

Goal: upgrade the table into a clearer client screen while preserving the stable two-row hand layout.

Step review question: did this step make the game easier to scan during play?

Stage review question: can a player understand opponents, center action, hand, selected tile, and controls at a glance?

- [ ] **Step 1: Add table composition regression tests**

Add tests that assert the table view contains:

- A title/header line.
- A status line with wall, events, and turn.
- Three opponent lines with clear names.
- A center/action area with `Last:` and `Tips:`.
- A hand area split across no more than two hand rows for 14 tiles.
- A selected-tile status line.
- A compact controls line.
- No rendered line exceeds 96 runes.

Run:

```powershell
go test ./internal/tui -run "TestRenderTable.*Client|TestRenderTableKeepsReadableLineWidth" -v
```

Expected first result: FAIL for newly added expectations.

- [ ] **Step 2: Refine table rendering**

Refine `renderTable`, `renderOpponent`, `renderSidePlayers`, `renderCenter`, and `renderHand` to produce a consistent screen:

```text
TERMINAL MAHJONG
Wall:67  Events:33  Turn:You

Opponents
AI-2  hand:13  melds:-  discards: ...
AI-1  hand:13  melds:-  discards: ...
AI-3  hand:13  melds:-  discards: ...

Table
Last: ...
Tips: ...

You
Melds: ...
Discards: ...
Hand: ...
      ...
Selected: ...

Keys: ...
```

Avoid heavy box-drawing borders around long Unicode content.

Run:

```powershell
go test ./internal/tui -run "TestRenderTable.*Client|TestRenderTableKeepsReadableLineWidth" -v
```

Expected: PASS.

- [ ] **Step 3: Stage 2 verification**

Run:

```powershell
go test ./internal/tui -v
go test ./...
```

Expected: PASS.

Commit:

```powershell
git add internal/tui/layout.go internal/tui/layout_test.go
git commit -m "feat: polish terminal table layout"
```

Write the Stage 2 review.

## Stage 3: Selection And Mouse Feedback

Goal: make keyboard and mouse selection feel deliberate, visible, and hard to misunderstand.

Step review question: did this step make the selected tile or click result clearer?

Stage review question: can a user always tell which tile is selected and what click happened?

- [ ] **Step 1: Add selection feedback tests**

Add tests that assert:

- Left at first tile remains at first tile.
- Right at last tile remains at last tile.
- Clicking a different tile changes `SelectedIndex`.
- Clicking the same selected tile discards.
- After click selection, view includes the clicked tile in `Selected:`.
- After discard, selected index remains valid for the new hand.

Run:

```powershell
go test ./internal/tui -run "Test.*Selected|TestMouse|TestTable.*Tile" -v
```

Expected: PASS for existing behavior, FAIL only for newly added feedback state if added.

- [ ] **Step 2: Add optional last feedback line**

If the tests require clearer feedback, add a UI-only field to `Model`, for example:

```go
StatusLine string
```

Use it for messages such as:

- `Selected [04] 🀊 (4m)`
- `Discarded [04] 🀊 (4m)`
- `Mouse selected [09] 🀙 (1p)`

Keep this state in `internal/tui`; do not add it to `internal/game`.

Run:

```powershell
go test ./internal/tui -run "Test.*Selected|TestMouse|TestTable.*Tile" -v
```

Expected: PASS.

- [ ] **Step 3: Stage 3 verification**

Run:

```powershell
go test ./internal/tui -v
go test ./...
```

Expected: PASS.

Commit:

```powershell
git add internal/tui/input.go internal/tui/layout.go internal/tui/model.go internal/tui/model_test.go internal/tui/layout_test.go
git commit -m "feat: improve tile selection feedback"
```

Write the Stage 3 review.

## Stage 4: Start And Game Over Screen Polish

Goal: make menu and game-over screens visually consistent with the table screen and easy to operate.

Step review question: did this step make screen transitions clearer without adding scope?

Stage review question: can the player start, learn controls, restart, return to menu, or quit without confusion?

- [ ] **Step 1: Add screen consistency tests**

Add tests that assert:

- Menu view contains title, options, and controls.
- Help view contains keyboard and mouse controls.
- Game Over view contains result, events, replay-ready status, and all choices.
- `Restart`, `Main Menu`, and `Quit` are keyboard selectable.

Run:

```powershell
go test ./internal/tui -run "TestMenu|TestHelp|TestGameOver" -v
```

Expected: PASS for existing basics, FAIL for any new consistency expectations.

- [ ] **Step 2: Polish menu/help/game-over rendering**

Keep the same plain-string approach as the table. Prefer readable sections over fragile box borders.

Run:

```powershell
go test ./internal/tui -run "TestMenu|TestHelp|TestGameOver" -v
```

Expected: PASS.

- [ ] **Step 3: Stage 4 verification**

Run:

```powershell
go test ./internal/tui -v
go test ./...
```

Expected: PASS.

Commit:

```powershell
git add internal/tui/menu.go internal/tui/layout.go internal/tui/model_test.go
git commit -m "feat: polish terminal client screens"
```

Write the Stage 4 review.

## Stage 5: Documentation And Acceptance

Goal: document the polished client and verify Phase 9.

Step review question: did this step make the polished client easier to run or evaluate?

Stage review question: does Phase 9 move the project closer to the target terminal-client style?

- [ ] **Step 1: Update docs**

Update `README.md` if controls or display expectations changed.

Append to `docs/workflow.md`:

```markdown
## Phase 9 Review Checklist

- Step reviews confirm rendering helpers, table composition, selection feedback, mouse feedback, and screen polish.
- Stage reviews compare each visual/interaction improvement against the total goal of a readable terminal-first Mahjong client.
- Verification commands: `go test ./...`, `go test ./... -cover`, `go build ./cmd/mahjong`, and one manual TUI smoke run.
```

- [ ] **Step 2: Final automated verification**

Run:

```powershell
go test ./...
go test ./... -cover
go build ./cmd/mahjong
```

Expected: PASS.

Clean generated binary:

```powershell
if (Test-Path .\mahjong.exe) { Remove-Item -LiteralPath .\mahjong.exe }
```

- [ ] **Step 3: Manual smoke**

Run:

```powershell
go run ./cmd/mahjong
```

Manual acceptance:

- Menu is readable.
- Table does not visually collapse in Windows Terminal.
- Selected tile is obvious before and after movement.
- Mouse click selection is reflected in `Selected:` or status feedback.
- Second click discards the selected tile.
- Game Over offers Restart, Main Menu, and Quit.

- [ ] **Step 4: Commit docs**

Run:

```powershell
git add README.md docs/workflow.md
git commit -m "docs: document terminal client polish"
```

Write the final Phase 9 stage review.

## Final Acceptance Gate

Phase 9 passes only if:

- The table is readable at normal Windows Terminal size.
- No rendered table line exceeds the tested width budget.
- Selected tile state is visible in both hand row and status text.
- Mouse selection and second-click discard are covered by tests.
- Game Over has restart/menu/quit options.
- `go test ./...`, `go test ./... -cover`, and `go build ./cmd/mahjong` pass.
