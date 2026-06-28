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

### Phase 12C Review

- Stage goal: replace the Riichi compatibility label with complete, source-cited four-player EMA Riichi 2016 mechanics while preserving shared local, online, bot, snapshot, and replay interfaces.
- Completed: EMA source notes and catalog fixtures; 136-tile wall, dead wall, dora/ura/kan-dora, red fives, and rinshan draws; Riichi hand/wait enumeration; chi, pon, ron, daiminkan, ankan, shouminkan, and chankan windows; riichi declarations, ippatsu, and all furiten states; complete yaku, fu, han, dora bonus, limits, settlement, exhaustive draw payments, honba, riichi sticks, and East-South match progression; recipient-private Riichi snapshots, canonical reconnect, typed replay state, and legal-action-bound bots.
- Step evidence: focused RED/GREEN tests were committed for source fixtures, wall/dead wall, decomposition, action windows, furiten/riichi, yaku detection, scoring, settlement, snapshots, replay, reconnect, and bot legal-action behavior.
- Acceptance evidence (2026-06-25):
  - all `testdata/rules/riichi/**/*.json` files parsed with `ConvertFrom-Json`;
  - `go test ./internal/game -run "RiichiCatalog|EveryRiichiYaku|RiichiScoring|RiichiSettlement" -count=20`;
  - `go test ./internal/game -run "RiichiGenerated|RiichiFuritenRonNever" -count=1`;
  - `go test ./internal/game ./internal/online ./internal/bot -run "Riichi.*Snapshot|Riichi.*Reconnect|Riichi.*Replay|Riichi.*Bot|RiichiRoom" -count=1`;
  - `go test ./internal/online -run TestRiichiWebSocketReadyDiscardAndReconnectSmoke -count=1`;
  - `gofmt -l` over `internal` and `cmd`, `git diff --check`, `go vet ./...`, `go test ./... -count=1`, and `go test -race ./... -count=1`;
  - `go build ./cmd/mahjong ./cmd/server ./cmd/client`;
  - `go test ./internal/online -count=20`.
- Total-goal review: both complete rule modes now run through the shared terminal/network architecture with private live information redacted, typed score/replay state, and bots constrained by server-provided legal actions.
- Remaining risk: Phase 12D still needs client-facing mode/options polish, localized score/error presentation, and side-by-side local/online command-sequence agreement checks before the wide table redesign begins.

### Phase 12D Review

- Stage goal: expose completed MCR and Riichi rules through the local TUI, online TUI, CLI client, shared protocol, and bots without duplicating rule decisions in presentation code.
- Completed: TUI rule-mode and red-five selection; local startup through `Match`; typed online room creation; CLI `-mode` and `-red-fives` options; mode-aware room listings; LegalActions-driven win, kong, riichi, and claim controls; Chinese/English Riichi action labels; and local/online startup parity coverage.
- Step evidence:
  - Task 1 added local rule selection and verified all TUI menu tests;
  - Task 2 added online/CLI room configuration and verified TUI, CLI, and online packages;
  - Task 3 replaced hand-derived Win/Kong availability with authoritative `LegalActions`, added playable Riichi controls, and preserved compact/wide line-width tests;
  - Task 4 found and removed a real local/online divergence by routing local games through the shared `Match` coordinator.
- Acceptance evidence (2026-06-28):
  - `go test ./internal/tui -run TestRuleModeParityBetweenLocalAndOnlineCreation -count=1`;
  - `go test ./internal/tui ./cmd/client ./internal/online -count=20`;
  - `go test ./... -count=1`;
  - `go test -race ./... -count=1`;
  - `go build ./cmd/mahjong ./cmd/server ./cmd/client`.
- Total-goal review: rule selection, room configuration, reconnect state, bot commands, and TUI controls now share the same typed mode/configuration/legal-action contracts in local and LAN play.
- Remaining risk: Phase 12E must rerun the combined dual-mode property, privacy, reconnect, smoke, static, race, and build gates before Phase 13 visual restructuring begins.

### Phase 12E Review

- Stage goal: prove MCR and Riichi satisfy the combined deterministic, legal-action, privacy, reconnect, replay, static, race, and build contracts before visual restructuring.
- Completed: fixed-seed canonical snapshot/replay equality; fresh-state legal action closure; representative zero-sum settlement checks; table-driven MCR/Riichi WebSocket privacy and reconnect coverage; and the complete repository verification gate.
- Step evidence:
  - Task 1 built two matches per mode from seed `120012`, compared canonical snapshots/replays, replayed every initial legal action on a fresh match, and checked representative settlements;
  - Task 2 created, readied, played, disconnected, and reconnected MCR and Riichi rooms while checking both recipient-private views;
  - the online 20-run gate found and corrected an overly strict MCR flower-count assertion, then passed the same command cleanly.
- Acceptance evidence (2026-06-28):
  - all `testdata/rules/**/*.json` files parsed with `ConvertFrom-Json`;
  - `go test ./internal/game -run "MCRFanCatalog|EveryMCRFan|RiichiCatalog|EveryRiichiYaku|Scoring|Settlement|Generated|DualMode" -count=1`;
  - `go test ./internal/online -run "Private|Reconnect|Riichi|MCR|DualMode" -count=20`;
  - `go test ./internal/tui ./cmd/client -count=20`;
  - `gofmt -l` over `internal` and `cmd`, `git diff --check`, and `go vet ./...`;
  - `go test ./... -count=1` and `go test -race ./... -count=1`;
  - `go build ./cmd/mahjong ./cmd/server ./cmd/client`.
- Total-goal review: Phase 12 now delivers selectable complete Chinese Official and four-player Riichi rules through local and LAN terminal clients, with deterministic replay metadata, authoritative legal actions, legal bots, private live state, and canonical reconnect.
- Remaining risk: Phase 13 is presentation work. It must preserve these accepted rule, protocol, privacy, replay, and width-test contracts while restructuring the wide competitive table and tactical rail.

### Phase 13 Review

- Stage goal: replace the stacked table with the approved wide competitive layout plus a read-only tactical rail, while keeping 80-column terminals playable.
- Completed:
  - 13A normalized local and online snapshots into one table view state and introduced component-based table rendering;
  - 13B added four fixed seats, public mode markers, active-turn emphasis, and deterministic six-tile discard rivers;
  - 13C added pure effective/improvement analysis and a bounded viewer-private tactical rail;
  - 13D added right-side, below-table, and `Tab`-controlled fallback behavior;
  - 13E bounded long tactical lists, removed random acceptance inputs, and made the 80x42 fallback use vertical seats, a narrow hand panel, compact controls, and height-aware tactical visibility.
- Visual evidence (fixed seed `1313`):
  - `artifacts/phase13/riichi-wide-zh.html`;
  - `artifacts/phase13/riichi-wide-en.html`;
  - `artifacts/phase13/mcr-wide-zh.html`;
  - `artifacts/phase13/mcr-wide-en.html`;
  - `artifacts/phase13/riichi-fallback-80.html`.
- Visual review: a real PTY run exposed an over-tall tactical rail, and the 80x42 capture test exposed both a 49-line composition and a 96-cell legacy middle row. The final generator asserts every artifact stays within its requested width and 42-line height. Automated PNG capture was not archived because local-file browser navigation was blocked; the deterministic HTML files remain directly reviewable.
- Acceptance evidence (2026-06-28):
  - `MAHJONG_PHASE13_CAPTURE_DIR=artifacts/phase13 go test ./internal/tui -run TestGeneratePhase13Snapshots -count=1`;
  - `go test ./internal/tui -count=20`;
  - `go test ./... -count=1`;
  - `go test -race ./... -count=1`;
  - `go vet ./...`;
  - `go build ./cmd/mahjong ./cmd/server ./cmd/client`;
  - `gofmt -l` over all changed Go files and `git diff --check`.
- Total-goal review: the wide A-layout plus C tactical rail now improves seat, river, hand, action, and tactical comprehension without changing game authority, hidden-information rules, reconnect state, or terminal-first controls.
- Remaining risk: Mahjong glyph size and color still depend on the user's terminal font and renderer; future visual acceptance should archive native terminal PNGs when the desktop capture path is available.
