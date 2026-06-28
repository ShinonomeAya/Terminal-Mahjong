# Phase 13 Wide Competitive Table And Tactical Rail Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the current stacked terminal table with the approved wide A-layout plus C tactical rail so four seats, discard rivers, the hand, legal actions, and tactical state are understandable at a glance.

**Architecture:** Normalize local and online state into one read-only `tableViewState`, then render independent header, seat, center-river, hand/action, and tactical-rail components. The game and protocol remain authoritative; tactical analysis is pure and read-only. Wide layout is the target, medium width moves the rail below, and compact width hides the rail behind `Tab`.

**Tech Stack:** Go, Bubble Tea, Lip Gloss, existing Unicode Mahjong tiles, standard `testing`, PTY screenshot capture.

---

## Approved Visual Contract

- Keep terminal-first presentation; do not create a GUI.
- Use the wide competitive table shown in the approved left-side mockup as the primary layout.
- Do not make the narrow mockup a second primary style; it is only a compatibility fallback.
- Header contains mode, round/hand, dealer/honba or MCR status, wall, dora/flowers, and network state.
- Four fixed seats surround organized discard rivers.
- Bottom hand remains a one-row bare/half-bare Unicode row with visible selected state and contextual actions.
- Right rail contains read-only shanten, effective tiles, improvements, rule status, and bounded recent events.
- Preserve Chinese/English consistency, keyboard/mouse controls, privacy, ANSI width budgets, and Phase 12 contracts.

## File Map

- Create `internal/tui/table_state.go`: normalize local/online snapshot, viewer seat, mode/config, match points, and legal actions.
- Create `internal/tui/table_components.go`: header, seat, river, center, hand/action, and layout composition.
- Create `internal/tui/tactical.go`: tactical rail view model and renderer.
- Modify `internal/tui/layout.go`: delegate table rendering to the new components while retaining game-over/menu renderers and shared tile helpers.
- Modify `internal/tui/model.go`: add tactical rail visibility state.
- Modify `internal/tui/input.go`: toggle compact tactical rail with `Tab`.
- Modify `internal/tui/style.go`: add restrained active-seat, river, and rail styles using existing Lip Gloss conventions.
- Modify `internal/game/analysis.go`: add pure effective/improvement tile analysis.
- Modify `internal/game/analysis_test.go`: verify tactical analysis.
- Create `internal/tui/table_components_test.go`: structural, state, width, and localization tests.
- Create `internal/tui/tactical_test.go`: tactical content and privacy tests.
- Modify `internal/tui/layout_test.go` and `internal/tui/model_test.go`: preserve existing controls, hitboxes, and fallback behavior.
- Modify `docs/workflow.md`: record 13A-E reviews and screenshot evidence.

## Task 1: 13A Shared Table State And Component Skeleton

**Files:**
- Create: `internal/tui/table_state.go`
- Create: `internal/tui/table_components.go`
- Create: `internal/tui/table_components_test.go`
- Modify: `internal/tui/layout.go`

- [x] **Step 1: Write failing structural tests**

Create local and online models at width 140. Assert the rendered view contains exactly one header, four seat labels, one center table, one hand row, one action row, and one tactical rail placeholder.

- [x] **Step 2: Verify RED**

Run:

```powershell
go test ./internal/tui -run "WideTableSkeleton|SharedTableState" -count=1
```

Expected: fail because component state and the tactical rail region do not exist.

- [x] **Step 3: Add normalized state**

Define:

```go
type tableViewState struct {
	Snapshot   game.GameSnapshot
	Match      game.MatchSnapshot
	ViewerSeat int
	Mode       game.RuleMode
	Online     bool
	Started    bool
	RoomCode   string
}

func tableStateFor(m Model) tableViewState
```

Local state uses `m.Game.SnapshotFor("0")`; online state uses `m.OnlineSnapshot` and `m.OnlineMatch`. No component reads `Game` internals directly after normalization.

- [x] **Step 4: Add component skeleton**

Create `renderWideTable`, `renderTableHeader`, `renderSeatBlock`, `renderTableCenter`, `renderHandAndActions`, and `renderTacticalPlaceholder`. Compose with `lipgloss.JoinHorizontal` and `lipgloss.JoinVertical`.

- [x] **Step 5: Preserve existing view entry points**

Make `renderTable` and `renderOnlineTable` delegate to `renderWideTable` when width is at least 110. Keep existing compact rendering temporarily for narrower widths.

- [x] **Step 6: Verify and commit**

Run `go test ./internal/tui -run "WideTableSkeleton|SharedTableState|RenderTable|OnlineTable" -count=1` and `go test ./internal/tui -count=1`.

```powershell
git add internal/tui/table_state.go internal/tui/table_components.go internal/tui/table_components_test.go internal/tui/layout.go
git commit -m "feat: add wide table component skeleton"
```

Step review: local and online tables now share one component boundary without changing game commands or privacy.

## Task 2: 13B Four Seats And Fixed Discard Rivers

**Files:**
- Modify: `internal/tui/table_components.go`
- Modify: `internal/tui/table_components_test.go`
- Modify: `internal/tui/style.go`

- [x] **Step 1: Write failing seat/river tests**

Assert each seat shows seat direction/wind, points when present, hand count, meld count, flowers for MCR, riichi marker for Riichi, and active-turn emphasis. Assert each discard river is rendered in stable rows with a fixed cell budget and latest-discard emphasis.

- [x] **Step 2: Verify RED**

Run `go test ./internal/tui -run "WideTableSeat|DiscardRiver|ActiveSeat" -count=1`.

- [x] **Step 3: Implement seat view data**

Define `seatView` with seat index, localized label, wind, points, hand count, melds, flowers, riichi state, active flag, and discards. Build four seat views from `tableViewState`.

- [x] **Step 4: Implement fixed rivers**

Render discards in deterministic six-tile rows. Use stable cell widths and truncate only beyond the bounded river capacity with an explicit count marker; do not wrap based on glyph width.

- [x] **Step 5: Add mode-specific public markers**

MCR seats show flowers; Riichi seats show accepted riichi and public dora. Never render opponent concealed tiles or live ura indicators.

- [x] **Step 6: Verify and commit**

Run focused seat/river tests, privacy tests, and `go test ./internal/tui -count=1`.

```powershell
git add internal/tui/table_components.go internal/tui/table_components_test.go internal/tui/style.go
git commit -m "feat: render fixed seats and discard rivers"
```

Step review: a player can identify all four seats, turn ownership, and discard history without reading the event log.

## Task 3: 13C Tactical Analysis Rail

**Files:**
- Modify: `internal/game/analysis.go`
- Modify: `internal/game/analysis_test.go`
- Create: `internal/tui/tactical.go`
- Create: `internal/tui/tactical_test.go`
- Modify: `internal/tui/table_components.go`

- [x] **Step 1: Write failing effective-tile tests**

Add:

```go
func TestEffectiveTilesReduceShanten(t *testing.T) {
	hand := mustAnalysisTiles(t, "123m456m789p22s34s")
	got := EffectiveTiles(hand)
	assertAnalysisTiles(t, got, "2s", "5s")
}
```

Add an improvement test for a 14-tile hand that returns discard/effective-tile candidates without mutating the input.

- [x] **Step 2: Verify RED**

Run `go test ./internal/game -run "EffectiveTiles|ImprovementTiles" -count=1`.

- [x] **Step 3: Implement pure analysis**

Add:

```go
type TileImprovement struct {
	Discard   Tile
	Effective []Tile
}

func EffectiveTiles(hand []Tile) []Tile
func ImprovementTiles(hand []Tile) []TileImprovement
```

Normalize red tiles for counting, skip fifth copies, sort/deduplicate output, and never mutate input.

- [x] **Step 4: Build tactical view model**

Define `tacticalView` with shanten, effective tiles, improvements, mode status, legal actions, and at most five recent typed events. Use only viewer-visible hand and snapshot fields.

- [x] **Step 5: Render the C rail**

Render a fixed-width right rail with localized headings and bare Unicode tiles. Keep it read-only and visually separate from the central table without nesting cards.

- [x] **Step 6: Verify and commit**

Run game analysis tests, tactical privacy/localization tests, and full TUI tests.

```powershell
git add internal/game/analysis.go internal/game/analysis_test.go internal/tui/tactical.go internal/tui/tactical_test.go internal/tui/table_components.go
git commit -m "feat: add tactical analysis rail"
```

Step review: the C rail explains the viewer's current tactical state without exposing hidden server information or changing commands.

## Task 4: 13D Medium And Compact Fallback

**Files:**
- Modify: `internal/tui/model.go`
- Modify: `internal/tui/input.go`
- Modify: `internal/tui/table_components.go`
- Modify: `internal/tui/table_components_test.go`
- Modify: `internal/tui/model_test.go`

- [x] **Step 1: Write failing responsive tests**

At width 140 assert the rail is right-aligned; at width 90 assert it moves below the table; at width 64 assert it is hidden by default and appears after `Tab`.

- [x] **Step 2: Verify RED**

Run `go test ./internal/tui -run "TacticalFallback|TabTogglesTactical" -count=1`.

- [x] **Step 3: Add visibility state**

Add `ShowTactical bool` to `Model`, default false. In table input, `Tab` toggles it without changing selected tile or claim state.

- [x] **Step 4: Compose three width modes**

Use constants:

```go
const wideTableMinWidth = 110
const mediumTableMinWidth = 80
```

Wide joins rail right; medium joins rail below; compact hides rail unless toggled. The compact table remains playable and is not a second visual theme.

- [x] **Step 5: Verify and commit**

Run responsive, line-width, hitbox, and full TUI tests.

```powershell
git add internal/tui/model.go internal/tui/input.go internal/tui/table_components.go internal/tui/table_components_test.go internal/tui/model_test.go
git commit -m "feat: add tactical rail fallback modes"
```

Step review: narrow terminals remain usable while the approved wide table stays the primary design.

## Task 5: 13E Interaction, Screenshot QA, And Review

**Files:**
- Modify: `internal/tui/layout_test.go`
- Modify: `internal/tui/model_test.go`
- Modify: `docs/workflow.md`
- Create: `artifacts/phase13/` screenshots generated by the existing PTY/manual capture workflow.

- [x] **Step 1: Preserve interaction contracts**

Run and extend tests for arrow selection, Enter/Space discard, H/K/L actions, claims, mouse hitboxes, menu return, language switching, online ready/reconnect, and game-over navigation.

- [x] **Step 2: Verify ANSI width budgets**

Require no wide-layout line to exceed the requested viewport and no compact control line to exceed its budget. Assert Mahjong Unicode cells keep stable hitboxes after selection.

- [x] **Step 3: Capture deterministic visual evidence**

Generate fixed-seed Chinese and English 140x42 HTML renders for MCR and Riichi, plus one 80x42 fallback. Inspect the real PTY and rendered artifacts for overlap, selected-tile visibility, dora/flowers, action state, and tactical headings. Direct PNG capture was unavailable because the desktop browser rejected local-file navigation; retain the deterministic HTML renders as the reviewable evidence instead of bypassing that policy.

- [x] **Step 4: Run full acceptance**

```powershell
go test ./internal/tui -count=20
go test ./... -count=1
go test -race ./... -count=1
go build ./cmd/mahjong ./cmd/server ./cmd/client
```

- [x] **Step 5: Update review and commit**

Record exact visual-artifact paths and verification commands in `docs/workflow.md`, mark 13A-E complete, and commit:

```powershell
git add internal/tui/layout_test.go internal/tui/model_test.go docs/workflow.md artifacts/phase13
git commit -m "test: record phase 13 wide table acceptance"
```

Phase review: compare wide MCR and Riichi screenshots with the approved A-plus-C direction; begin Phase 14 only if table comprehension improved without weakening controls, privacy, width, or rule contracts.

## Plan Self-Review

- Spec coverage: 13A skeleton, 13B table state, 13C tactical rail, 13D fallback, and 13E interaction/screenshot QA each have a dedicated task.
- Placeholder scan: no TODO/TBD or unspecified implementation steps remain.
- Type consistency: all new TUI types consume existing `GameSnapshot`, `MatchSnapshot`, `LegalAction`, and analysis APIs.
- Scope: no GUI, replay persistence, new rules, database, or external AI work is included.
- Visual consistency: wide A plus right C is the only target style; narrow layouts are compatibility fallbacks.
