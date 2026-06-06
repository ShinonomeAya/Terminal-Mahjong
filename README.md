# Terminal Mahjong

A small, complete terminal Mahjong game written in Go.

## MVP Scope

- Four-player single game: one human player and three simple AI players.
- Simplified pushdown rules.
- Supported actions: draw, discard, self-win, discard-win, pong, concealed kong.
- Win check: standard `4 melds + 1 pair` hand shape.
- No scoring, flowers, seat wind rounds, riichi rules, or network play in the first version.

## Run

```powershell
go run ./cmd/mahjong
```

## Test

```powershell
go test ./...
```

## Development Workflow

This project uses a phase-and-step workflow:

1. Pick one stage goal.
2. Break it into small steps.
3. After each step, review whether the step moved the current stage forward.
4. After each stage, review whether the whole project goal is still becoming more true.
5. Keep changes surgical: no speculative rules, UI, or abstractions.

The detailed workflow lives in `docs/workflow.md`.

