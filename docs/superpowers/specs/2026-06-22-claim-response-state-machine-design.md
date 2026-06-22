# Claim Response State Machine Design

## Context

The legacy line-oriented game loop can ask a human whether to claim an opponent's discard. The Bubble Tea and online paths cannot: they advance immediately and implicitly decline every human claim. This stage makes claim response an explicit, serializable game state shared by solo TUI, bots, and the WebSocket server.

The implementation keeps the project's simplified Mahjong rules. It does not add Riichi scoring, furiten, multiple simultaneous wins, call timers, or persistent rooms.

## Stage Goal

A human player can see and answer every legal `win`, `pong`, or `chow` opportunity with keyboard controls in both local and online play. `pass` is explicit, bots remain legal, reconnect restores the pending decision, and existing discard/self-draw/kong behavior continues to work.

## State Model

`Game` gains a `Phase` and optional `PendingClaim`:

- `awaiting_discard`: `Current` must draw if needed, then may discard, self-win, or declare a concealed kong.
- `awaiting_claim`: the last discard has generated one or more ordered legal claim options.
- `round_over`: no further commands are accepted.

`PendingClaim` contains the discarder, discarded tile, ordered eligible players, and their legal options. Options are data, not callbacks, so snapshots and reconnect can reproduce the decision exactly.

Claim priority remains compatible with the existing rules:

1. discard win, in turn order after the discarder;
2. pong, in turn order;
3. chow, only for the next player.

The state machine asks eligible players in that order. A pass advances to the next eligible response. An accepted claim resolves immediately. If everyone passes, play advances to the player after the discarder.

## Public Data

Add:

```go
type TurnPhase string
type ClaimKind string
type ClaimOption struct {
    Kind     ClaimKind
    Player   int
    Consumed []Tile
}
type PendingClaim struct {
    Discarder int
    Tile      Tile
    Options   []ClaimOption
    Active    int
}
```

`GameSnapshot` exposes `Phase` and a copied `PendingClaim`. `GameCommand` adds `pass`, `claim_win`, `pong`, and `chow`; `TileIndex` selects among multiple chow options for the active player.

## Core Transitions

- Discard records the event and builds legal claim options.
- No options: set the next player and return to `awaiting_discard`.
- Options exist: enter `awaiting_claim`, with `Current` set to the active claimant.
- Pass: advance to the next option group or complete the discard transition.
- Win: finish with `WinDiscard`.
- Pong/chow: remove the consumed tiles, add the meld, remove the claimed discard from the discarder, and give the claimant an `awaiting_discard` turn without drawing.

Commands are rejected when their phase, player, or option is invalid.

## Bot And Server

`HeuristicBot.Decide` first checks the phase. During a claim it accepts a legal win, uses the existing pong/chow heuristics, otherwise passes. The server's existing bot loop continues until it reaches an occupied human seat or the round ends. Because pending claims live in `GameSnapshot`, reconnect requires no separate recovery protocol.

## TUI Interaction

When the local or online human is the active claimant, the action line changes contextually:

- `H`: win on the discard
- `P`: pong
- `C`: chow; left/right selects among multiple chow combinations
- `Space` or `Esc`: pass

The pending discard and selected chow combination are rendered next to the action prompt. Normal hand selection and discard are disabled while a claim response is pending. Chinese and English strings are added together so the screen never mixes languages.

## Compatibility

The legacy `Play` loop keeps its prompt-based behavior for now. New state-machine helpers are used by `ApplyCommand`, the TUI, and online server. Existing exported APIs remain available, and fixed-seed replay behavior is preserved.

## Acceptance

- Core tests prove option generation, priority, pass progression, each accepted claim, invalid command rejection, and snapshot copying.
- Bot tests prove every claim decision is legal.
- WebSocket tests prove claim broadcast, wrong-player rejection, and reconnect snapshot restoration.
- TUI tests prove localized prompts and keyboard behavior in local and online modes.
- `go test ./...`, `go test -race ./...`, `go vet ./...`, and all command builds pass.

