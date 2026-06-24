# Terminal Mahjong Workflow

## Total Goal

Build a small but complete terminal Mahjong game in Go: playable in one terminal session, understandable as a codebase, and easy to extend after the MVP is verified.

## Stage Loop

Each stage uses the same loop:

1. Define the stage goal in one sentence.
2. Split the stage into small steps.
3. Complete one step.
4. Review that single step against the stage goal.
5. Repeat until the stage goal is met.
6. Review the whole stage against the total goal.

## Stages

### Stage 1: Foundation

Goal: create a Go module, project docs, and focused package boundaries.

Step review question: did this step make the project easier to run, test, or understand?

Stage review question: can a new contributor see the MVP scope and start the project?

### Stage 2: Rules Core

Goal: model tiles, wall, hands, melds, and standard hand-shape win detection.

Step review question: did this step add one verified rule primitive?

Stage review question: can rules be tested without the terminal UI?

### Stage 3: Game Loop

Goal: support a complete four-player round from deal to win or wall exhaustion.

Step review question: did this step advance the round state without mixing in UI concerns?

Stage review question: can the game complete without panics or invalid hand counts?

### Stage 4: Terminal Play

Goal: make the human player able to inspect state and choose valid actions.

Step review question: did this step make the player decision clearer or safer?

Stage review question: can a user play one complete game with only terminal prompts?

### Stage 5: Verification

Goal: test core rules and manually verify a playable game start.

Step review question: did this step prove a specific behavior?

Stage review question: does the current state satisfy the total MVP goal?

## Review Format

Use this short format after each development step:

```text
Step review:
- Stage goal:
- Step completed:
- Evidence:
- Next step:
```

Use this short format after each stage:

```text
Stage review:
- Total goal:
- Stage completed:
- Evidence:
- Remaining risk:
```

## Phase 2 Review Checklist

- Step review is written after each completed implementation task.
- Stage review compares chow, scoring, and terminal play against the total goal.
- Verification commands: `go test ./...`, `go build ./cmd/mahjong`, one scripted smoke run.

Stage review:
- Total goal: build a small but complete terminal Mahjong game in Go.
- Stage completed: Phase 2 adds chow, deterministic claim priority, and basic settlement.
- Evidence: run `go test ./...`, `go build ./cmd/mahjong`, and a scripted terminal smoke run before closing the phase.
- Remaining risk: scoring remains intentionally simplified and not a full regional rule set.

## Phase 3 Review Checklist

- Step reviews confirm event logging, deterministic scripts, claim split, and render split.
- Stage review compares event history and code boundaries against the total terminal Mahjong goal.
- Verification commands: `go test ./...`, `go build ./cmd/mahjong`, one scripted smoke run.

## Phase 4 Review Checklist

- Step reviews confirm shanten, waits, terminal tips, and AI discard behavior.
- Stage review compares player assistance against the total terminal Mahjong goal.
- Verification commands: `go test ./...`, `go build ./cmd/mahjong`, one scripted smoke run.

## Phase 5 Review Checklist

- Step reviews confirm table sections, command help, recent events, and result summary layout.
- Stage review compares terminal readability against the total terminal Mahjong goal.
- Verification commands: `go test ./...`, `go build ./cmd/mahjong`, one scripted smoke run.

## Phase 6 Review Checklist

- Step reviews confirm seven-pairs detection, waits, tips, and scoring labels.
- Stage review compares the bounded rule extension against the total terminal Mahjong goal.
- Verification commands: `go test ./...`, `go build ./cmd/mahjong`, one scripted smoke run.

## Phase 7 Review Checklist

- Step reviews confirm replay metadata, JSON export, replay summaries, and result hints.
- Stage review compares replay support against the total terminal Mahjong goal.
- Verification commands: `go test ./...`, `go build ./cmd/mahjong`, one scripted smoke run.

## Phase 8 Review Checklist

- Step reviews confirm Unicode tile rendering, non-blocking game helpers, start menu, keyboard selection, table layout, mouse selection, and CLI wiring.
- Stage reviews compare each TUI addition against the total goal of a polished terminal-first Mahjong client.
- Verification commands: `go test ./...`, `go test ./... -cover`, `go build ./cmd/mahjong`, and one manual TUI smoke run.

## Phase 9 Review Checklist

- Step reviews confirm rendering helpers, table composition, selection feedback, mouse feedback, and screen polish.
- Stage reviews compare each visual/interaction improvement against the total goal of a readable terminal-first Mahjong client.
- Verification commands: `go test ./...`, `go test ./... -cover`, `go build ./cmd/mahjong`, and one manual TUI smoke run.

## Phase 10 Review Checklist

- Step reviews confirm style helpers, ANSI-safe visible-width checks, styled table sections, styled selected tiles, and styled menu/game-over screens.
- Stage reviews compare the visual skin against the total goal of an attractive but readable terminal-first Mahjong client.
- Verification commands: `go test ./...`, `go test ./... -cover`, `go build ./cmd/mahjong`, and one manual TUI smoke run.

## Phase 11 Review Checklist

- Step reviews confirm explicit pending-claim state, deterministic priority, bot legality, online recovery, and localized TUI controls.
- Stage review compares interactive win, pong, chow, and pass responses against the total goal of a complete terminal-first Mahjong client.
- Verification commands: `go vet ./...`, `go test ./...`, `go test -race ./...`, all command builds, one WebSocket reconnect smoke run, and Chinese/English TUI screenshots.

### Phase 12A Review

- Stage goal: establish mode-neutral match and privacy foundations before implementing either complete rule set.
- Completed: validated configurations, compatibility RuleSet, Match coordinator, legal actions, private snapshots, mode-aware protocol, TUI hand counts, and replay metadata.
- Evidence: `go vet ./...`, full tests, race tests, game/online/TUI tests repeated 20 times, all command builds, and a real WebSocket create/ready/discard/reconnect privacy smoke.
- Total-goal review: the architecture can host MCR and Riichi without duplicating networking or leaking live concealed information.
- Remaining risk: MCR and Riichi still use compatibility mechanics until Phases 12B and 12C replace them.

### Phase 12B Review

- Stage goal: replace every compatibility mechanic reachable from an MCR room with complete Chinese Official wall, action, scoring, settlement, match, snapshot, and replay behavior.
- Completed: 144-tile wall and flower replacement; exhaustive standard and special decomposition; all 81 fan IDs; catalog exclusions and counting; eight-point non-flower minimum; chow, pong, three kong forms, and robbing-kong windows; zero-sum settlement; 16-hand progression; private reconnect snapshots; typed replay score and settlement history.
- Step evidence: focused RED/GREEN tests were committed for each point band, combined scoring, action priority, kong windows, settlement, match progression, replay, and reconnect behavior.
- Acceptance evidence (2026-06-24):
  - all `testdata/rules/mcr/**/*.json` files parsed with `ConvertFrom-Json`;
  - `go test ./internal/game -run "MCRFanCatalog|EveryMCRFan|MCRScoring|MCRSettlement" -count=20`;
  - `go test ./internal/game -run MCRGeneratedInvariantsAcrossOneThousandSeeds -count=1`;
  - `go vet ./...`, `go test ./... -count=1`, and `go test -race ./... -count=1`;
  - `go build ./cmd/mahjong ./cmd/server ./cmd/client`;
  - `go test ./internal/online -count=20`;
  - WebSocket MCR creation, private broadcast, pending claim, canonical reconnect, and 144-tile conservation integration tests.
- Total-goal review: Chinese Official games now run through the shared terminal/network architecture without compatibility scoring or action fallback, while concealed tiles and shuffle seeds remain private during play.
- Remaining risk: Phase 12C Riichi mechanics are not started. Dead wall, dora/ura-dora, riichi/furiten, yaku/fu/han, exhaustive draw, honba/riichi sticks, and East-South progression still require their own source-cited plan and acceptance cycle.
