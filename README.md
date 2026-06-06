# Terminal Mahjong

A small, complete terminal Mahjong game written in Go.

## MVP Scope

- Four-player single game: one human player and three simple AI players.
- Simplified pushdown rules.
- Supported actions: draw, discard, self-win, discard-win, pong, chow, concealed kong.
- Win check: standard `4 melds + 1 pair` hand shape.
- No scoring, flowers, seat wind rounds, riichi rules, or network play in the first version.

## Commands

- `<number>` or `d <number>`: discard the numbered tile in your hand.
- `h`: claim self-draw when your hand can win.
- `k <tile>`: declare a concealed kong, for example `k 1m`.
- `q`: quit the current game.
- When the game offers win, pong, or chow claims, answer `y` to take the claim or press Enter to decline.

## Simplified Scoring

- Self-draw: 2 points.
- Discard-win: 1 point.
- Each pong or kong meld: +1 point.
- Chow has no point bonus.
- Full regional scoring, flowers, riichi, and round wind scoring remain outside Phase 2.

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
