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
