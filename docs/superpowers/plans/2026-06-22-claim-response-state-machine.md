# Claim Response State Machine Implementation Plan

> **For Codex:** Execute this plan task by task with `superpowers:executing-plans`. For behavior changes, follow red-green-refactor and record the failing test before implementation.

**Goal:** Let human players respond to opponent discards with win, pong, chow, or pass in solo and online TUI play without regressing existing game, reconnect, or Unicode rendering behavior.

**Architecture:** Move discard claims from synchronous input prompts into an explicit `Game` phase and serializable pending-claim value. Reuse the same command path for local TUI, bots, and the WebSocket server. Keep the existing simplified rule priority and avoid unrelated rules or persistence work.

**Tech Stack:** Go, Bubble Tea, Lip Gloss, Gorilla WebSocket, standard `testing` and `httptest`.

---

## Task 1: Core Claim Options And Snapshot

**Files:**
- Create: `internal/game/claim_state.go`
- Modify: `internal/game/game.go`
- Modify: `internal/game/snapshot.go`
- Modify: `internal/game/snapshot_test.go`

1. Write failing tests for legal win/pong/chow options, priority order, snapshot phase, and non-aliasing copied options.
2. Run `go test ./internal/game -run "Claim|Snapshot" -count=1` and confirm failure is caused by missing claim-state types.
3. Implement the smallest claim types, option builder, phase fields, and deep-copy snapshot logic.
4. Re-run the focused tests and `go test ./internal/game -count=1`.
5. Step review: legal decisions are now deterministic data and survive snapshot/reconnect boundaries.
6. Commit: `feat: model pending discard claims`.

## Task 2: Command State Machine

**Files:**
- Modify: `internal/game/snapshot.go`
- Modify: `internal/game/turn.go`
- Modify: `internal/game/claim_state.go`
- Modify: `internal/game/turn_test.go`
- Modify: `internal/game/snapshot_test.go`

1. Write failing tests for discard entering claim state, pass progression, all-pass continuation, accepted discard-win, pong, chow, and phase/player rejection.
2. Run the focused tests and confirm expected red failures.
3. Add `pass`, `claim_win`, `pong`, and `chow` commands. Route discard through one transition helper and resolve claims without reading terminal input.
4. Update `AdvanceAIUntilHumanTurn` to stop when player 0 must answer a claim and to continue after local commands.
5. Keep the legacy prompt loop operational; adapt it only where compilation or shared transition behavior requires it.
6. Run `go test ./internal/game -count=1`.
7. Step review: the game can pause and resume at every human decision without hidden auto-decline.
8. Commit: `feat: resolve claims through game commands`.

## Task 3: Bot Claim Decisions

**Files:**
- Modify: `internal/bot/heuristic.go`
- Modify: `internal/bot/heuristic_test.go`

1. Write failing tests for accepting a legal claim-win and returning only one of the active player's legal pong/chow/pass commands.
2. Run `go test ./internal/bot -count=1` and confirm red.
3. Add phase-aware claim decisions, reusing current Mahjong analysis and no new configuration.
4. Run bot and game package tests.
5. Step review: unoccupied seats cannot stall the new state machine or submit an illegal action.
6. Commit: `feat: teach bots to answer discard claims`.

## Task 4: Online Claim Protocol And Reconnect

**Files:**
- Modify: `internal/online/server.go`
- Modify: `internal/online/server_test.go`
- Modify: `internal/online/client_test.go`

1. Write WebSocket tests for a human claim prompt snapshot, wrong-seat rejection, accepted/pass broadcast, bot continuation, and reconnect preserving the same pending claim.
2. Run `go test ./internal/online -run "Claim|Reconnect" -count=1` and confirm red.
3. Adjust the server bot loop to recognize claim phases and stop only at an occupied eligible seat. No new wire message type is needed because commands and snapshots are already JSON values.
4. Run `go test ./internal/online ./internal/protocol -count=1` and then `go test ./...`.
5. Step review: LAN clients share one authoritative claim decision and reconnect cannot skip it.
6. Commit: `feat: synchronize discard claims online`.

## Task 5: Local And Online TUI Claim Controls

**Files:**
- Modify: `internal/tui/input.go`
- Modify: `internal/tui/online.go`
- Modify: `internal/tui/layout.go`
- Modify: `internal/tui/i18n.go`
- Modify: `internal/tui/model_test.go`
- Modify: `internal/tui/layout_test.go`

1. Write failing tests for local and online `H/P/C/Space/Esc` claim controls, chow-option selection, disabled discard during claim response, and fully localized Chinese/English prompts.
2. Run `go test ./internal/tui -run "Claim|Language" -count=1` and confirm red.
3. Add contextual input routing and command senders. Render the pending tile and active options in the existing table/action area, preserving bare Unicode hand tiles and mouse hitboxes.
4. Run focused TUI tests, then all TUI tests.
5. Step review: the pending decision is visible, bounded, localized, and playable without typed commands.
6. Commit: `feat: add tui discard claim controls`.

## Task 6: Related Cleanup And Documentation

**Files:**
- Modify: `internal/tui/layout.go`
- Modify: `README.md`
- Modify: `docs/workflow.md`

1. Confirm the four legacy render helpers are unreferenced with `rg`.
2. Remove only those dead helpers and imports made unused by their removal.
3. Document claim controls, simplified priority, and online reconnect behavior.
4. Run formatting and all package tests.
5. Step review: the changed area is easier to maintain without broad visual refactoring.
6. Commit: `docs: record interactive claim workflow`.

## Task 7: Stage Acceptance

1. Run `gofmt -w` on changed Go files and confirm `git diff --check` is clean.
2. Run `go vet ./...`.
3. Run `go test ./... -count=1`.
4. Run `go test -race ./... -count=1`.
5. Run `go build ./cmd/mahjong ./cmd/server ./cmd/client`.
6. Launch the server and exercise create/join/ready/command/reconnect smoke behavior.
7. Launch the TUI in Chinese and English, capture terminal screenshots, and verify claim prompts do not overlap at normal and compact widths.
8. Stage review: compare results against the design acceptance list and the overall terminal-first, table-like, keyboard-and-mouse playable goal.
9. Commit any acceptance ledger update separately if files change.

