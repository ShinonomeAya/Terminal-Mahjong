# Terminal Client Visual Skin Phase 10 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a restrained terminal visual skin to the Mahjong TUI: color-coded sections, stronger selected tile styling, and accurate visible-width checks that survive ANSI styling.

**Architecture:** Keep all visual styling inside `internal/tui`. Do not change Mahjong rules, scoring, AI, replay, or the CLI flow.

**Tech Stack:** Go 1.23, Bubble Tea v0.27.1, Lip Gloss v0.13.0, Unicode Mahjong glyphs.

---

## Stage 1: Style Helpers

Goal: add theme helpers and accurate visible-width helpers before styling screens.

- [ ] Add tests for `visibleWidth`, `styleSectionTitle`, `styleSelectedTile`, and `styleMuted`.
- [ ] Implement `internal/tui/style.go` using `github.com/charmbracelet/lipgloss`.
- [ ] Verify with `go test ./internal/tui -run "TestStyle|TestVisibleWidth" -v`.
- [ ] Run `go test ./...`.
- [ ] Commit: `feat: add tui visual style helpers`.

Step review:

```text
Step review:
- Stage goal:
- Step completed:
- Evidence:
- Next step:
```

Stage review:

```text
Stage review:
- Total goal:
- Stage completed:
- Evidence:
- Remaining risk:
```

## Stage 2: Styled Table Screen

Goal: apply the theme to the table while preserving readable line widths and existing text affordances.

- [ ] Add table tests that use `visibleWidth` instead of rune count for styled output.
- [ ] Style title, section headings, selected tile cells, status line, and controls.
- [ ] Keep the plain section text present so tests and users can still scan the UI.
- [ ] Run `go test ./internal/tui -v` and `go test ./...`.
- [ ] Commit: `feat: style terminal mahjong table`.

## Stage 3: Styled Menu And Game Over

Goal: apply the same visual language to menu, help, and game-over screens.

- [ ] Add tests that menu/game-over include styled markers while retaining readable labels.
- [ ] Style current menu item and game-over item with the selected tile style.
- [ ] Style help headings and control hints.
- [ ] Run `go test ./internal/tui -v` and `go test ./...`.
- [ ] Commit: `feat: style terminal client screens`.

## Stage 4: Docs And Acceptance

Goal: document the visual skin and verify the app.

- [ ] Update `README.md` and `docs/workflow.md` with Phase 10 notes.
- [ ] Run:

```powershell
go test ./...
go test ./... -cover
go build ./cmd/mahjong
```

- [ ] Remove `mahjong.exe` if build created it.
- [ ] Commit: `docs: document terminal client visual skin`.

## Final Acceptance Gate

Phase 10 passes when:

- Styled output keeps visible line width under the tested budget.
- Selected tile/menu item is visually stronger than normal rows.
- ANSI styling does not break content tests.
- `go test ./...`, coverage, and build pass.
