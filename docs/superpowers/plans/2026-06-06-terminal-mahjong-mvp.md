# Terminal Mahjong MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a small but complete Go terminal Mahjong game with one human player, three simple AI players, and a verified rules core.

**Architecture:** Keep rules in `internal/game` so they can be tested without terminal I/O. Put only process startup and stdin/stdout wiring in `cmd/mahjong`. Model a single round with simplified pushdown rules and avoid scoring until the MVP is stable.

**Tech Stack:** Go 1.23, standard library only, `go test ./...` for verification.

---

## Files

- `README.md`: run commands, MVP scope, and workflow summary.
- `docs/workflow.md`: stage loop, step review, and stage review process.
- `cmd/mahjong/main.go`: terminal entrypoint.
- `internal/game/*.go`: tile model, player state, win detection, AI, round loop.
- `internal/game/*_test.go`: focused tests for deck and win behavior.

## Tasks

### Task 1: Foundation

- [ ] Create `go.mod` with module name `mahjong`.
- [ ] Add README and workflow docs with MVP scope and review loop.
- [ ] Add `.gitignore` for local build artifacts.
- [ ] Step review: foundation files make the project runnable and understandable.

### Task 2: Rules Core

- [ ] Implement tile constants, names, deck construction, sorting, and parsing.
- [ ] Implement player hand and meld operations.
- [ ] Implement `CanWin` for standard `4 melds + 1 pair` hands.
- [ ] Add tests for winning hands, non-winning hands, and deck composition.
- [ ] Step review: rules can be tested without terminal I/O.

### Task 3: Round Engine

- [ ] Implement round setup, draw, discard, pong, concealed kong, self-win, discard-win, and wall exhaustion.
- [ ] Add simple AI discard choice based on duplicate count and local tile usefulness.
- [ ] Keep scoring out of scope; winner and win reason are enough.
- [ ] Step review: one round can advance from deal to a terminal state.

### Task 4: Terminal UI

- [ ] Implement a text prompt that shows wall count, discards, melds, and the human hand.
- [ ] Accept `h`, `k <tile>`, `d <index>`, `<index>`, and `q` commands.
- [ ] Prompt for human discard-win and pong opportunities after AI discards.
- [ ] Step review: one user can play without knowing internal indexes beyond hand numbers.

### Task 5: Verification

- [ ] Run `go test ./...`.
- [ ] Run `go run ./cmd/mahjong` long enough to verify startup, initial deal, and prompt rendering.
- [ ] Stage review: compare current behavior with the total MVP goal and list remaining risks.

