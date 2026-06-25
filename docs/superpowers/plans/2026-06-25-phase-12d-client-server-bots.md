# Phase 12D Client, Server, And Bots Integration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the completed MCR and Riichi rules selectable and understandable from the terminal client and command-line client, with online/local behavior using the same rule snapshots and legal actions.

**Architecture:** Keep rule execution in `internal/game` and room state in `internal/online`; TUI and CLI clients only choose modes/options and render snapshots. Do not add new rule branches to layout code except display labels; commands must come from `GameSnapshot.LegalActions` where available. This phase finishes integration polish before the Phase 13 wide table redesign.

**Tech Stack:** Go, Bubble Tea TUI, existing Gorilla WebSocket JSON protocol, standard `testing`.

---

## Current State Audit

- Already complete: protocol messages carry `Mode` and `RuleConfig`; room listing exposes mode/config; server creates `MCRRuleSet` and `RiichiRuleSet`; reconnect snapshots preserve mode and private rule state; bots prefer `LegalActions`.
- Remaining gaps: TUI start flow does not let the user select MCR/Riichi or red-five options; command-line client cannot create MCR/Riichi rooms from flags; room list does not render mode/options clearly; TUI action labels still infer availability in places; score and error/status text is not consistently localized.

## File Map

- Modify `internal/tui/model.go`: add selected rule mode and red-five option fields.
- Modify `internal/tui/menu.go`: add mode/options menu states and rendering.
- Modify `internal/tui/network.go`: create online rooms with selected mode/config.
- Modify `internal/tui/layout.go` and `internal/tui/i18n.go`: render mode names, legal action states, score summaries, draw reasons, and errors in Chinese/English.
- Modify `internal/tui/model_test.go` and `internal/tui/layout_test.go`: cover local mode selection, online mode creation, room list labels, legal-action controls, and localization.
- Modify `cmd/client/main.go`: add `-mode` and `-red-fives` flags and print room mode/config in summaries.
- Modify `cmd/client/main_test.go`: cover CLI creation/listing for MCR and Riichi.
- Modify `internal/online/client.go` and `internal/online/client_test.go` only if a typed join/create helper is missing.
- Modify `docs/workflow.md` and `docs/rules/conformance.md`: record Phase 12D acceptance.

## Task 1: TUI Local Mode And Options Selection

**Files:**
- Modify: `internal/tui/model.go`
- Modify: `internal/tui/menu.go`
- Modify: `internal/tui/model_test.go`

- [ ] **Step 1: Write failing menu tests**

Add tests that press a mode/options key from the menu and then start a local game:

```go
func TestMenuCanStartLocalRiichiWithRedFivesDisabled(t *testing.T) {
	m := NewModel()
	m.SelectedMode = game.ModeRiichi
	m.SelectedRiichiRedFives = 0
	updated, _ := updateMenu(m, tea.KeyMsg{Type: tea.KeyEnter})
	got := updated.(Model)
	if got.Game == nil || got.Game.Mode != game.ModeRiichi || got.Game.RuleConfig.Riichi.RedFives != 0 {
		t.Fatalf("local mode = %#v", got.Game)
	}
}

func TestMenuCanStartLocalMCR(t *testing.T) {
	m := NewModel()
	m.SelectedMode = game.ModeMCR
	updated, _ := updateMenu(m, tea.KeyMsg{Type: tea.KeyEnter})
	got := updated.(Model)
	if got.Game == nil || got.Game.Mode != game.ModeMCR {
		t.Fatalf("local mode = %#v", got.Game)
	}
}
```

- [ ] **Step 2: Verify RED**

Run `go test ./internal/tui -run "MenuCanStartLocal.*Mode|MenuCanStartLocalRiichi" -count=1`. Expected: fail because selected mode fields and local rule creation do not exist.

- [ ] **Step 3: Implement selected mode fields and local creation**

Add to `Model`:

```go
SelectedMode           game.RuleMode
SelectedRiichiRedFives int
```

Initialize `SelectedMode: game.ModeRiichi` and `SelectedRiichiRedFives: 3`. Replace `newStartedGame()` in the solo start branch with `newStartedGameWithRules(m.SelectedMode, selectedRuleConfig(m))`.

- [ ] **Step 4: Add simple menu toggles**

Add two menu entries before Help:

```text
规则：日麻 / 中庸 / 经典
红五：三张 / 关闭
```

The rule toggle cycles `riichi -> mcr -> compatibility`; the red-five toggle only changes `0/3` and is rendered disabled/muted when mode is not Riichi.

- [ ] **Step 5: Verify and commit**

Run `go test ./internal/tui -run "MenuCanStartLocal|Menu.*Language|RenderMenu" -count=1` and `go test ./internal/tui -count=1`.

```powershell
git add internal/tui/model.go internal/tui/menu.go internal/tui/model_test.go
git commit -m "feat: select local rule mode in tui"
```

Step review: local play can start the same MCR/Riichi rules implemented in 12B/12C without changing game internals.

## Task 2: Online And CLI Mode/Option Creation

**Files:**
- Modify: `internal/tui/network.go`
- Modify: `internal/tui/model_test.go`
- Modify: `cmd/client/main.go`
- Modify: `cmd/client/main_test.go`

- [ ] **Step 1: Write failing online and CLI tests**

```go
func TestCreateOnlineRoomUsesSelectedRiichiOptions(t *testing.T) {
	m := NewModel()
	m.SelectedMode = game.ModeRiichi
	m.SelectedRiichiRedFives = 0
	msg := protocol.Message{Type: protocol.MsgRoomCreated, Mode: game.ModeRiichi, RuleConfig: selectedRuleConfig(m)}
	updated := applyOnlineConnected(m, onlineConnectedMsg{Message: msg})
	if updated.OnlineMatch.RuleConfig.Riichi.RedFives != 0 {
		t.Fatalf("online config = %#v", updated.OnlineMatch.RuleConfig)
	}
}
```

For `cmd/client`, add a test that runs `run(ctx, []string{"-mode", "riichi", "-red-fives", "0"}, out)` against the existing test server and asserts the output contains `mode=riichi`.

- [ ] **Step 2: Verify RED**

Run `go test ./internal/tui ./cmd/client -run "SelectedRiichi|mode|red" -count=1`. Expected: fail because CLI flags and TUI create command do not pass selected config.

- [ ] **Step 3: Wire TUI create room config**

Change `createOnlineRoomCmd(m)` to call `client.CreateRoomWithRules(ctx, m.SelectedMode, selectedRuleConfig(m))`.

- [ ] **Step 4: Add CLI flags**

Add:

```go
modeText := flags.String("mode", "compatibility", "room rule mode: compatibility, mcr, riichi")
redFives := flags.Int("red-fives", 3, "riichi red fives: 0 or 3")
```

Parse with `game.ParseRuleMode`, build `RuleConfig`, validate it, and pass it to `CreateRoomWithRules` when creating a room. Joining and reconnecting keep server state authoritative.

- [ ] **Step 5: Print room mode/options**

Update `printMessage` and `printRoomList` to include mode and red-five count when present.

- [ ] **Step 6: Verify and commit**

Run `go test ./internal/tui ./cmd/client ./internal/online -run "Rules|Room|mode|red|Create" -count=1` and `go test ./cmd/client ./internal/tui -count=1`.

```powershell
git add internal/tui/network.go internal/tui/model_test.go cmd/client/main.go cmd/client/main_test.go
git commit -m "feat: create online rooms with rule options"
```

Step review: local TUI, online TUI, and CLI room creation all choose the same validated `RuleConfig`.

## Task 3: Legal Action Rendering And Localization Polish

**Files:**
- Modify: `internal/tui/layout.go`
- Modify: `internal/tui/i18n.go`
- Modify: `internal/tui/input.go`
- Modify: `internal/tui/layout_test.go`
- Modify: `internal/tui/model_test.go`

- [ ] **Step 1: Write failing legal-action rendering tests**

Create tests where an online snapshot has only `CommandDiscard` and verify `[H] Win` and `[K] Kong` render as off in both languages. Create another snapshot with `CommandRiichi` and verify the controls show Riichi/立直.

- [ ] **Step 2: Verify RED**

Run `go test ./internal/tui -run "LegalAction|Riichi.*Control|Localized.*Score|Localized.*Error" -count=1`.

- [ ] **Step 3: Centralize legal action checks**

Make `canOnlineWin`, `canOnlineKong`, and new `canOnlineRiichi` read `OnlineSnapshot.LegalActions` only. Local mode uses `Game.Snapshot().LegalActions`.

- [ ] **Step 4: Localize mode, score, draw, and error text**

Add small mapping functions in `i18n.go`:

```go
func localizedModeName(m Model, mode game.RuleMode) string
func localizedCommandName(m Model, kind game.CommandKind) string
func localizedReason(m Model, reason string) string
func localizedScoreSummary(m Model, snapshot game.GameSnapshot) string
```

Keep unknown strings unchanged.

- [ ] **Step 5: Verify and commit**

Run `go test ./internal/tui -run "LegalAction|Localized|OnlineTable|RenderTable" -count=1` and `go test ./internal/tui -count=1`.

```powershell
git add internal/tui/layout.go internal/tui/i18n.go internal/tui/input.go internal/tui/layout_test.go internal/tui/model_test.go
git commit -m "feat: localize rule actions in tui"
```

Step review: the TUI no longer advertises actions that the authoritative snapshot does not allow, and mixed Chinese/English rule text is reduced.

## Task 4: Local/Online Parity And Phase Review

**Files:**
- Create: `internal/tui/rule_mode_integration_test.go`
- Modify: `docs/workflow.md`
- Modify: `docs/rules/conformance.md`

- [ ] **Step 1: Write parity tests**

Build one local match and one WebSocket-created match for each mode using seed/config fixtures, compare `Mode`, `RuleConfig`, initial points, wall count, and first legal action kinds.

- [ ] **Step 2: Verify RED/GREEN**

Run `go test ./internal/tui ./internal/online ./internal/game -run "RuleModeParity|Phase12D|LegalAction" -count=1`.

- [ ] **Step 3: Run acceptance checks**

```powershell
go test ./internal/tui ./cmd/client ./internal/online -count=20
go test ./... -count=1
go test -race ./... -count=1
go build ./cmd/mahjong ./cmd/server ./cmd/client
```

- [ ] **Step 4: Update docs and commit**

Mark Phase 12D complete and 12E not started. Record exact evidence and remaining Phase 12E risk.

```powershell
git add internal/tui/rule_mode_integration_test.go docs/workflow.md docs/rules/conformance.md
git commit -m "test: record phase 12d integration acceptance"
```

Phase review: local and online clients select the same rule configs, render authoritative legal actions, and remain ready for Phase 13 table redesign.

## Plan Self-Review

- Spec coverage: local mode selection, online mode/config messages, legal action rendering, bot legal-action integration, and Chinese/English display polish each have a task.
- Placeholder scan: no TODO/TBD placeholders remain.
- Type consistency: tasks use existing `game.RuleMode`, `game.RuleConfig`, `game.CommandKind`, `GameSnapshot.LegalActions`, and `protocol.Message` fields.
- Scope: no new rule mechanics, databases, Redis, external AI, or GUI are added in this phase.
