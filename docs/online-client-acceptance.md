# Online Client Acceptance

This document is the acceptance checklist for the terminal-first online Mahjong client.

## Reference Boundary

The project learns architecture ideas from `palemoky/fight-the-landlord`: Bubble Tea phase separation, WebSocket room flow, reconnect progress notifications, room/session tests, and bot decision boundaries.

No GPL implementation code from that repository is copied into this repository.

## Required Behaviors

- `cmd/server` starts an in-memory WebSocket room server.
- `cmd/client -list` prints joinable waiting rooms.
- `cmd/client` creates a room and saves a reconnect session.
- `cmd/client -join <room>` joins a waiting room.
- `cmd/client -ready` broadcasts ready room state.
- `cmd/client -watch` survives one dropped socket by reconnecting with the saved token.
- The TUI main menu can create, browse, join, reconnect, return to menu, and quit.
- The TUI shows one of these network states: local, waiting, your turn, reconnecting, reconnected, offline.
- The server rejects joins after a room has started.
- The server rejects non-current-player commands.
- The server keeps disconnected sessions reconnectable within the configured reconnect window.
- Expired disconnected sessions cannot reconnect.
- Expired idle rooms are not listed in room discovery.

## Verification Commands

Run these commands from the repository root:

```powershell
go test ./...
go test ./internal/online ./cmd/client ./internal/tui
go run ./cmd/server
go run ./cmd/client -list
```

Manual smoke flow:

```powershell
go run ./cmd/server
go run ./cmd/client -name Host -session .host-session.json
go run ./cmd/client -name Guest -join <printed-room-code> -session .guest-session.json
go run ./cmd/client -session .host-session.json -reconnect -ready -watch
```
