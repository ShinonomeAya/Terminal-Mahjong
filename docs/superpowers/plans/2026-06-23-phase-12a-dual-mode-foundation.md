# Phase 12A Dual-Mode Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Introduce mode-neutral match coordination, explicit rule configuration, structured legal actions, recipient-private online snapshots, and mode-aware protocol/replay metadata while preserving the current playable compatibility rules.

**Architecture:** Keep the existing `game.Game` as the round implementation during migration, wrap it in a new `game.Match`, and put rule ownership behind a `RuleSet` interface implemented first by `CompatibilityRuleSet`. Live WebSocket messages use recipient-filtered snapshots; authoritative snapshots remain available only to the server and bots. Existing `NewGame`, TUI, and CLI paths remain functional until complete MCR and Riichi implementations replace the compatibility rule set in later subphases.

**Tech Stack:** Go, Bubble Tea, Gorilla WebSocket, JSON, standard `testing`/`httptest`, race detector.

---

## File Map

- Create `internal/game/rules.go`: rule modes, configurations, legal actions, `RuleSet`, compatibility implementation.
- Create `internal/game/rules_test.go`: configuration and legal-action behavior.
- Create `internal/game/match.go`: `Match`, match snapshots, round delegation.
- Create `internal/game/match_test.go`: match creation, points, command delegation, copied snapshots.
- Modify `internal/game/game.go`: mode/config/rule-set ownership and compatible constructors.
- Modify `internal/game/snapshot.go`: legal actions, hand counts, recipient-filtered snapshots and redaction.
- Modify `internal/game/snapshot_test.go`: privacy, event, seed, and pending-claim tests.
- Modify `internal/game/replay.go` and `replay_test.go`: schema/mode/config metadata.
- Modify `internal/protocol/message.go`: mode-aware room and match fields.
- Modify `internal/online/client.go` and tests: explicit mode room creation.
- Modify `internal/online/server.go` and tests: room-owned `Match`, personalized broadcasts, reconnect privacy.
- Modify `internal/tui/model.go`, `online.go`, `layout.go`, and tests: retain match metadata and show hidden hand counts correctly.
- Create `docs/rules/conformance.md`: frozen baselines and variant defaults.
- Create `testdata/rules/fixture.schema.json`: common golden-fixture envelope.
- Modify `docs/workflow.md`: Phase 12A review evidence.

## Task 1: Freeze Rule Baselines And Fixture Contract

**Files:**
- Create: `docs/rules/conformance.md`
- Create: `testdata/rules/fixture.schema.json`

- [ ] **Step 1: Write the conformance matrix**

Record both baselines, project defaults, implementation status, and the rule that fixture citations are mandatory. The table must include wall composition, win threshold, response priority, match length, red fives, open tanyao, initial points, and live hidden-information policy.

- [ ] **Step 2: Write the fixture schema**

Create a JSON Schema requiring this envelope:

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "required": ["id", "mode", "source", "initial", "commands", "expected"],
  "properties": {
    "id": {"type": "string", "minLength": 1},
    "mode": {"enum": ["mcr", "riichi"]},
    "source": {"type": "string", "minLength": 1},
    "initial": {"type": "object"},
    "commands": {"type": "array", "items": {"type": "object"}},
    "expected": {"type": "object"}
  },
  "additionalProperties": false
}
```

- [ ] **Step 3: Verify the documents**

Run:

```powershell
Get-Content -Raw testdata/rules/fixture.schema.json | ConvertFrom-Json | Out-Null
rg -n "GB/T 34708-2017|EMA Riichi Rules 2016|8 points|25,000|three red fives|hidden" docs/rules/conformance.md
```

Expected: JSON parsing succeeds and every baseline phrase is found.

- [ ] **Step 4: Step review and commit**

Step review: the phase now has one explicit source-of-truth contract before code can encode rule assumptions.

```powershell
git add docs/rules/conformance.md testdata/rules/fixture.schema.json
git commit -m "docs: freeze dual-mode rule baselines"
```

## Task 2: Rule Modes And Validated Configuration

**Files:**
- Create: `internal/game/rules.go`
- Create: `internal/game/rules_test.go`

- [ ] **Step 1: Write failing configuration tests**

Add tests proving:

```go
func TestDefaultRuleConfig(t *testing.T) {
	if got := DefaultRuleConfig(ModeRiichi).Riichi; got.StartingPoints != 25000 || got.MatchLength != MatchEastSouth || !got.OpenTanyao || got.RedFives != 3 {
		t.Fatalf("riichi defaults = %#v", got)
	}
	if got := DefaultRuleConfig(ModeMCR).MCR.MinimumPoints; got != 8 {
		t.Fatalf("MCR minimum = %d, want 8", got)
	}
}

func TestRuleConfigRejectsWrongPayload(t *testing.T) {
	config := DefaultRuleConfig(ModeRiichi)
	config.MCR.MinimumPoints = 8
	if err := config.Validate(ModeRiichi); err == nil {
		t.Fatal("expected mixed-mode config rejection")
	}
}

func TestParseRuleModeRejectsUnknown(t *testing.T) {
	if _, err := ParseRuleMode("regional"); err == nil {
		t.Fatal("expected unknown mode error")
	}
}
```

- [ ] **Step 2: Verify RED**

Run `go test ./internal/game -run "RuleConfig|RuleMode" -count=1`.

Expected: compile failure because the new types and functions do not exist.

- [ ] **Step 3: Implement the configuration types**

Define:

```go
type RuleMode string

const (
	ModeCompatibility RuleMode = "compatibility"
	ModeMCR           RuleMode = "mcr"
	ModeRiichi        RuleMode = "riichi"
)

type MatchLength string

const MatchEastSouth MatchLength = "east_south"

type MCRConfig struct {
	MinimumPoints int `json:"minimum_points"`
}

type RiichiConfig struct {
	StartingPoints int         `json:"starting_points"`
	MatchLength    MatchLength `json:"match_length"`
	OpenTanyao     bool        `json:"open_tanyao"`
	RedFives       int         `json:"red_fives"`
}

type RuleConfig struct {
	MCR    MCRConfig    `json:"mcr,omitempty"`
	Riichi RiichiConfig `json:"riichi,omitempty"`
}
```

`DefaultRuleConfig`, `ParseRuleMode`, and `Validate` return stable errors for unknown modes, mixed payloads, unsupported match length, starting points outside `1000..100000`, red fives outside `0..3`, and MCR minimum other than eight. Compatibility mode accepts an empty configuration only.

- [ ] **Step 4: Verify GREEN**

Run `go test ./internal/game -run "RuleConfig|RuleMode" -count=1` and `go test ./internal/game -count=1`.

Expected: pass.

- [ ] **Step 5: Step review and commit**

Step review: mode and option values are now explicit, serializable, and rejected before match creation when invalid.

```powershell
git add internal/game/rules.go internal/game/rules_test.go
git commit -m "feat: add validated rule configurations"
```

## Task 3: Structured Legal Actions And Compatibility Rule Set

**Files:**
- Modify: `internal/game/rules.go`
- Modify: `internal/game/rules_test.go`
- Modify: `internal/game/game.go`
- Modify: `internal/game/snapshot.go`

- [ ] **Step 1: Write failing legal-action tests**

Test that a normal current player receives one discard action per hand tile plus available self-win/kong actions, a claimant receives only pass and the active claim type, a non-current player receives no actions, and invalid compatibility commands are rejected without mutation.

Use this public shape:

```go
type LegalAction struct {
	Kind      CommandKind `json:"kind"`
	TileIndex int         `json:"tile_index,omitempty"`
	Tile      string      `json:"tile,omitempty"`
	Consumed  []Tile      `json:"consumed,omitempty"`
}

type RuleSet interface {
	Mode() RuleMode
	Config() RuleConfig
	InitialPoints() [4]int
	LegalActions(round *Game, playerID string) []LegalAction
	Allows(round *Game, command GameCommand) bool
}
```

- [ ] **Step 2: Verify RED**

Run `go test ./internal/game -run "LegalAction|CompatibilityRule" -count=1`.

Expected: compile failure for missing `LegalAction` and `RuleSet` behavior.

- [ ] **Step 3: Implement minimal compatibility behavior**

Add `CompatibilityRuleSet` with a mode/config constructor. Generate actions from current round state:

- awaiting discard: indexed discards, available self-win, each concealed kong, and quit;
- awaiting claim: pass plus active win/pong/chow options;
- round over or non-current player: no actions.

`Allows` compares kind and kind-specific selector (`TileIndex` for discard/chow, `Tile` for kong). Deep-copy consumed tiles.

- [ ] **Step 4: Attach rules to Game**

Add unexported `rules RuleSet`, exported `Mode RuleMode`, and `RuleConfig RuleConfig` fields. Keep:

```go
func NewGame(seed int64) *Game {
	game, err := NewGameWithRules(seed, NewCompatibilityRuleSet(ModeCompatibility, RuleConfig{}))
	if err != nil { panic(err) }
	return game
}
```

`NewGameWithRules` validates configuration, preserves crypto/fixed seeds, and builds the existing wall/deal. `ApplyCommand` rejects commands not returned by `Allows` with `command is not legal` before mutation.

- [ ] **Step 5: Expose legal actions in snapshots**

Add `LegalActions []LegalAction` and populate it for the current player in authoritative snapshots.

- [ ] **Step 6: Verify GREEN**

Run focused rules tests and `go test ./internal/game ./internal/bot -count=1`.

Expected: pass; existing bot commands remain legal.

- [ ] **Step 7: Step review and commit**

Step review: the current game is now expressed through a rule boundary and structured legal actions without claiming complete MCR or Riichi behavior.

```powershell
git add internal/game/rules.go internal/game/rules_test.go internal/game/game.go internal/game/snapshot.go
git commit -m "feat: expose compatibility legal actions"
```

## Task 4: Match Coordinator And Match Snapshot

**Files:**
- Create: `internal/game/match.go`
- Create: `internal/game/match_test.go`

- [ ] **Step 1: Write failing Match tests**

Cover compatibility, MCR, and Riichi construction; initial point arrays; delegated commands; copied configuration; and complete status after the round ends.

Required types:

```go
type Match struct {
	Mode        RuleMode
	RuleConfig  RuleConfig
	Points      [4]int
	Dealer      int
	RoundNumber int
	Complete    bool
	Round       *Game
	rules       RuleSet
}

type MatchSnapshot struct {
	Mode        RuleMode     `json:"mode"`
	RuleConfig  RuleConfig   `json:"rule_config"`
	Points      [4]int       `json:"points"`
	Dealer      int          `json:"dealer"`
	RoundNumber int          `json:"round_number"`
	Complete    bool         `json:"complete"`
	Round       GameSnapshot `json:"round"`
}
```

- [ ] **Step 2: Verify RED**

Run `go test ./internal/game -run "Match" -count=1`.

Expected: compile failure because `Match` is missing.

- [ ] **Step 3: Implement Match**

Add:

```go
func NewMatch(seed int64, rules RuleSet) (*Match, error)
func (m *Match) Snapshot() MatchSnapshot
func (m *Match) SnapshotFor(playerID string) MatchSnapshot
func (m *Match) ApplyCommand(command GameCommand) CommandResult
func (m *Match) EnsureCurrentTurnDraw() (Tile, bool)
```

For 12A, compatibility mechanics power all three mode labels; mode-specific rule behavior remains explicitly incomplete in the conformance matrix. `ApplyCommand` marks `Complete` when the round ends. MCR initial points are zero; Riichi initial points come from config; compatibility points are zero.

- [ ] **Step 4: Verify GREEN**

Run `go test ./internal/game -run "Match" -count=1` and all game tests.

- [ ] **Step 5: Step review and commit**

Step review: round state now lives under a match identity that can carry dealer, points, configuration, and later multi-round settlement.

```powershell
git add internal/game/match.go internal/game/match_test.go
git commit -m "feat: add match coordinator"
```

## Task 5: Recipient-Private Snapshots

**Files:**
- Modify: `internal/game/snapshot.go`
- Modify: `internal/game/snapshot_test.go`

- [ ] **Step 1: Write failing privacy tests**

Tests must prove that during a live round `SnapshotFor("0")`:

- retains player 0's hand and exposes `HandCount` for every player;
- replaces opponent hands with nil;
- redacts opponent draw-event tiles to `-1`;
- redacts seed and shuffle seed;
- hides pending claim options from a non-active viewer;
- includes legal actions only for the viewer when the viewer is current;
- reveals full hands, seed, events, and claim data once the round is over.

- [ ] **Step 2: Verify RED**

Run `go test ./internal/game -run "SnapshotFor|PrivateSnapshot" -count=1`.

Expected: failure because current snapshots reveal all hidden state.

- [ ] **Step 3: Implement redaction**

Add `HandCount int` to `PlayerView`. Make `Snapshot()` authoritative and add:

```go
func (g *Game) SnapshotFor(playerID string) GameSnapshot
```

Build an authoritative deep copy first, then redact only while `!g.Over`. Unknown viewers receive no concealed hands or legal actions. Do not mutate game state or the authoritative snapshot.

- [ ] **Step 4: Verify GREEN**

Run focused privacy tests and all game tests.

- [ ] **Step 5: Step review and commit**

Step review: live state now enforces the privacy boundary required for fair online play and future full-information post-game replay.

```powershell
git add internal/game/snapshot.go internal/game/snapshot_test.go
git commit -m "feat: add recipient-private snapshots"
```

## Task 6: Mode-Aware Protocol And Personalized Server Broadcasts

**Files:**
- Modify: `internal/protocol/message.go`
- Modify: `internal/online/client.go`
- Modify: `internal/online/client_test.go`
- Modify: `internal/online/server.go`
- Modify: `internal/online/server_test.go`

- [ ] **Step 1: Write failing client/protocol tests**

Add `CreateRoomWithRules(ctx, mode, config)` and verify the sent message includes mode/config while existing `CreateRoom` sends compatibility defaults. Verify room summaries and create/join/reconnect replies carry `MatchSnapshot`.

- [ ] **Step 2: Write failing server privacy tests**

With two connected clients, assert each recipient sees only its own concealed hand, both see accurate opponent `HandCount`, neither receives seed during play, and reconnect receives the same private view. Assert a MCR room cannot be joined with a conflicting mode request.

- [ ] **Step 3: Verify RED**

Run `go test ./internal/online -run "Rules|Mode|Private|Reconnect" -count=1`.

Expected: compile or assertion failures for missing protocol fields and shared broadcasts.

- [ ] **Step 4: Extend protocol shapes**

Add `Mode`, `RuleConfig`, and `Match` to `protocol.Message`; add mode/config to `RoomSummary`. Do not remove `Snapshot` yet because TUI/CLI compatibility is required through Phase 12D.

- [ ] **Step 5: Move rooms to Match**

Replace `room.game` with `room.match`. Creation uses requested/default mode and validated config. Existing access sites use `room.match.Round` for command/bot mechanics.

- [ ] **Step 6: Personalize every outbound state**

Create helpers that build messages per `session` using both:

```go
Snapshot: room.match.Round.SnapshotFor(fmt.Sprintf("%d", session.seat))
Match:    room.match.SnapshotFor(fmt.Sprintf("%d", session.seat))
```

Use personalized messages for create, join, room state, accepted-command broadcast, and reconnect. Bots continue using authoritative snapshots.

- [ ] **Step 7: Verify GREEN and repeatability**

Run:

```powershell
go test ./internal/online -count=20
go test ./... -count=1
```

Expected: pass without random privacy or turn-state failures.

- [ ] **Step 8: Step review and commit**

Step review: mode/configuration now survives room discovery and reconnect, while each network recipient receives only permitted live information.

```powershell
git add internal/protocol/message.go internal/online/client.go internal/online/client_test.go internal/online/server.go internal/online/server_test.go
git commit -m "feat: synchronize private match snapshots"
```

## Task 7: TUI Match Metadata And Hidden Hand Counts

**Files:**
- Modify: `internal/tui/model.go`
- Modify: `internal/tui/online.go`
- Modify: `internal/tui/layout.go`
- Modify: `internal/tui/layout_test.go`
- Modify: `internal/tui/model_test.go`

- [ ] **Step 1: Write failing TUI tests**

Prove that an online opponent with `Hand=nil, HandCount=13` renders `13`, the header shows `MCR`/`Riichi`/compatibility from match metadata, and online updates retain the latest `MatchSnapshot`.

- [ ] **Step 2: Verify RED**

Run `go test ./internal/tui -run "HandCount|MatchMode|MatchSnapshot" -count=1`.

Expected: failure because the model only stores `GameSnapshot` and opponent panels use `len(Hand)`.

- [ ] **Step 3: Implement minimal TUI adaptation**

Add `OnlineMatch game.MatchSnapshot` to `Model`. Populate it on connect/reconnect/snapshot messages. Add `playerHandCount` returning `HandCount` when present, otherwise `len(Hand)`. Use it in online seat panels. Add a compact mode label to the existing header without redesigning the table.

- [ ] **Step 4: Verify GREEN**

Run focused tests and `go test ./internal/tui -count=20`.

- [ ] **Step 5: Step review and commit**

Step review: the current TUI remains playable after privacy enforcement and can identify the match mode without beginning the Phase 13 redesign.

```powershell
git add internal/tui/model.go internal/tui/online.go internal/tui/layout.go internal/tui/layout_test.go internal/tui/model_test.go
git commit -m "feat: render private match metadata"
```

## Task 8: Replay Metadata And Compatibility

**Files:**
- Modify: `internal/game/replay.go`
- Modify: `internal/game/replay_test.go`

- [ ] **Step 1: Write failing replay metadata tests**

Assert `ReplayLog` contains `SchemaVersion: 2`, mode, complete rule configuration, and still includes seed/proof/events. Assert mutating returned configuration cannot affect the game.

- [ ] **Step 2: Verify RED**

Run `go test ./internal/game -run "Replay.*Mode|Replay.*Schema|ReplayLog" -count=1`.

Expected: failure for missing metadata.

- [ ] **Step 3: Implement metadata**

Add:

```go
const ReplaySchemaVersion = 2

type ReplayLog struct {
	SchemaVersion int         `json:"schema_version"`
	Mode          RuleMode    `json:"mode"`
	RuleConfig    RuleConfig  `json:"rule_config"`
	// existing fields remain
}
```

Populate fields from `Game` and include mode in `ReplaySummary`. Do not add file persistence or replay frames before Phase 14.

- [ ] **Step 4: Verify GREEN**

Run focused replay tests and all game tests.

- [ ] **Step 5: Step review and commit**

Step review: existing replay-ready logs can identify the rules that produced them without prematurely implementing the Phase 14 file format.

```powershell
git add internal/game/replay.go internal/game/replay_test.go
git commit -m "feat: version replay rule metadata"
```

## Task 9: Phase 12A Acceptance And Review

**Files:**
- Modify: `docs/workflow.md`

- [ ] **Step 1: Run formatting and static analysis**

```powershell
$files = Get-ChildItem internal,cmd -Recurse -Filter *.go | ForEach-Object FullName
gofmt -w $files
git diff --check
go vet ./...
```

Expected: no formatting list, whitespace errors, or vet findings.

- [ ] **Step 2: Run fresh automated acceptance**

```powershell
go test ./... -count=1
go test -race ./... -count=1
go build ./cmd/mahjong ./cmd/server ./cmd/client
```

Expected: all commands exit zero.

- [ ] **Step 3: Run privacy and stability repetitions**

```powershell
go test ./internal/game -run "Rule|Match|SnapshotFor|Private|Replay" -count=20
go test ./internal/online -count=20
go test ./internal/tui -count=20
```

Expected: pass across all repetitions.

- [ ] **Step 4: Run a real WebSocket smoke**

Start `cmd/server` on a spare local port, create a room with the compatibility client, ready, discard, reconnect, and verify the client never prints opponent concealed tiles or a live seed. Stop every process started by the smoke.

- [ ] **Step 5: Record Phase 12A review**

Append to `docs/workflow.md`:

```markdown
### Phase 12A Review

- Stage goal: establish mode-neutral match and privacy foundations before implementing either complete rule set.
- Completed: validated configurations, compatibility RuleSet, Match coordinator, legal actions, private snapshots, mode-aware protocol, TUI hand counts, and replay metadata.
- Evidence: go vet, full tests, race tests, repeated privacy/online/TUI tests, command builds, and WebSocket smoke.
- Total-goal review: the architecture can host MCR and Riichi without duplicating networking or leaking live concealed information.
- Remaining risk: MCR and Riichi still use compatibility mechanics until Phases 12B and 12C replace them.
```

- [ ] **Step 6: Commit the review**

```powershell
git add docs/workflow.md
git commit -m "test: record phase 12a acceptance"
```

- [ ] **Step 7: Phase review**

Phase review: verify every 12A requirement against current files and fresh command output. If complete, automatically begin the separate Phase 12B specification/plan cycle; do not mark the master roadmap complete.

## Plan Self-Review

- Spec coverage: 12A conformance contract, mode/config types, Match/Round migration boundary, legal actions, private snapshots, protocol/reconnect metadata, replay metadata, compatibility entry points, and acceptance all map to explicit tasks.
- Scope: complete MCR, complete Riichi, final mode-selection UI, tactical table redesign, and replay persistence remain in their approved later phases.
- Type consistency: `RuleMode`, `RuleConfig`, `LegalAction`, `RuleSet`, `Match`, `MatchSnapshot`, `SnapshotFor`, `CreateRoomWithRules`, and `OnlineMatch` are introduced before later use.
- Placeholder scan command:

```powershell
$patterns = @("TB"+"D", "TO"+"DO", "implement"+" later", "fill in"+" details", "appropriate"+" error", "Similar"+" to")
foreach ($pattern in $patterns) { rg -n $pattern docs/superpowers/plans/2026-06-23-phase-12a-dual-mode-foundation.md }
```

Expected: no matches.
