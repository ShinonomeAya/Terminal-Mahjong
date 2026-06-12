# Terminal Mahjong

A small, complete terminal Mahjong game written in Go.

## MVP Scope

- Four-player single game: one human player and three simple AI players.
- Simplified pushdown rules.
- Supported actions: draw, discard, self-win, discard-win, pong, chow, concealed kong.
- Win check: standard `4 melds + 1 pair` hand shape.
- No scoring, flowers, seat wind rounds, riichi rules, or network play in the first version.

## Fair Shuffle And Replay Audit

- New default games use a seed generated from `crypto/rand`, then shuffle the wall deterministically from that seed.
- Tests and scripted runs can still pass a fixed seed to reproduce the exact deal and event sequence.
- Replay logs include the seed and a SHA-256 wall hash so a round can be audited without scraping terminal output.
- The game does not use player history, win rate, or any adaptive deal control when shuffling.

## Controls

- `Up` / `Down`: move in menus.
- `Enter`: confirm a menu item or discard the selected tile.
- `Left` / `Right`: select a hand tile.
- Mouse click: select a hand tile when the terminal supports mouse events.
- Second click on the selected hand tile: discard it.
- `Space`: discard the selected tile.
- `Q`: quit the current game.

The client renders Mahjong tiles with Unicode glyphs by default and keeps text labels available for fallback rendering in tests and future configuration.

The table view is split into readable terminal sections:

- `Opponents`: AI hands, melds, and discards.
- `Table`: latest action and hand tips.
- `You`: melds, discards, hand tiles, selected tile, and status feedback.
- `Controls`: available keyboard and mouse actions.

## Simplified Scoring

- Self-draw: 2 points.
- Discard-win: 1 point.
- Each pong or kong meld: +1 point.
- Chow has no point bonus.
- Full regional scoring, flowers, riichi, and round wind scoring remain outside Phase 2.

## Phase 3 Direction

The game remains terminal-first. Phase 3 adds typed event logs and deterministic scripted runs so later shanten, AI, replay, and terminal UI upgrades can use game state directly instead of scraping printed text.

## Phase 4 Direction

The game remains terminal-first. Phase 4 adds standard-hand shanten, tenpai waits, and compact table tips so the terminal game becomes easier to understand without becoming a GUI.

## Phase 5 Direction

The game remains terminal-first. Phase 5 improves the text table with stable sections, command help, and recent event summaries. It is still a terminal game, not a GUI.

## Phase 6 Direction

The game remains simplified and terminal-first. Phase 6 adds seven-pairs win support and a small seven-pairs scoring bonus, but still avoids full regional scoring tables.

## Phase 7 Direction

The game remains terminal-first. Phase 7 adds replay-ready event log export and summaries using standard-library JSON, without adding networking or a replay GUI.

## Phase 8 Direction

The game remains terminal-first. Phase 8 adds a Bubble Tea TUI client with a start menu, Unicode Mahjong tiles, keyboard tile selection, mouse tile selection, and a table-like layout.

## Phase 9 Direction

The game remains terminal-first. Phase 9 polishes the TUI into clearer client sections, stronger tile selection feedback, consistent menu/game-over screens, and safer line-width rendering for Windows Terminal.

## Phase 10 Direction

The game remains terminal-first. Phase 10 adds a restrained Lip Gloss visual skin with colored section headings, selected-tile emphasis, status highlighting, and ANSI-safe visible-width tests.

## Run

```powershell
go run ./cmd/mahjong
```

The TUI start menu supports local play plus online room actions:

- `Create Online Room`: connect to `ws://127.0.0.1:8080/ws`, create a room, and save `.mahjong-session.json`.
- `Join Online Room`: type a numeric room code and join an existing room.
- `Reconnect Online`: restore the saved session after a disconnect.
- Online tables use `R` to ready up; discards are blocked until all occupied seats are ready.

Start the local WebSocket server:

```powershell
go run ./cmd/server
```

Use the minimal online CLI client:

```powershell
# Create a room and save the reconnect session.
go run ./cmd/client -name Alice

# Join from another terminal.
go run ./cmd/client -name Bob -join 000001 -session .mahjong-bob.json

# Reconnect with a saved session.
go run ./cmd/client -reconnect

# Keep reading game snapshots and reconnect up to 5 times if the socket drops.
go run ./cmd/client -reconnect -watch

# Send a one-shot discard after connecting. Indexes are 1-based for the CLI.
go run ./cmd/client -reconnect -discard 1
```

The first online version uses in-memory rooms for local or LAN play. Restarting the server clears rooms and reconnect sessions.

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
