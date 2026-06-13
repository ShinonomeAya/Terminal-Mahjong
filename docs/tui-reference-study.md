# TUI Reference Study

Reference project: `palemoky/fight-the-landlord`

Local reference clone used for study:
`C:\Users\wenwen\AppData\Local\Temp\fight-the-landlord-reference`

## Boundary

This project studies the reference screenshots and layout architecture only.
No GPL implementation code is copied.

## Screenshot Observations

`docs/lobby.png`:
- Main content is centered in the terminal.
- A primary menu panel sits beside a secondary info/chat panel.
- The bottom control hint is muted and visually separate.

`docs/in-game.png`:
- The game screen is centered vertically and horizontally.
- The top area shows compact status/counter information.
- The middle area shows opponent boxes and the latest play.
- The bottom area is a framed hand tray.
- The prompt/action hint sits below the hand, not mixed into the table.

## Mahjong Translation

- Lobby: keep terminal menu, but make it centered and panel-based.
- Table: render four seats around a central discard/event area.
- Hand: render player's hand as a stable tray with clear selected tile focus.
- Width: keep important lines within 96 visible cells on normal terminals and 80 cells in compact mode.
- Mouse: hitboxes must continue to match the rendered hand row.
