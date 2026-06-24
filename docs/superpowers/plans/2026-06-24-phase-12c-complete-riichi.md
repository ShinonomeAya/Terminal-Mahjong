# Phase 12C Complete Four-Player Riichi Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the Riichi compatibility label with a complete, source-cited four-player EMA Riichi 2016 implementation covering the wall, calls, riichi and furiten, yaku/fu/han scoring, draws, settlement, and East-South match progression.

**Architecture:** Keep shared transport, `Game`, `Match`, and recipient-private snapshot APIs in `internal/game`, while isolating Japanese rules in focused `riichi_*.go` files. `RiichiRuleSet` owns dead-wall state and legal actions; decomposition, yaku, fu, scoring, and settlement remain pure layers so neither TUI nor WebSocket code branches on scoring details.

**Tech Stack:** Go standard library, existing Bubble Tea client and Gorilla WebSocket server, JSON golden fixtures, standard `testing`, race detector.

---

## Source And Scope Contract

- Normative source: European Mahjong Association, `Riichi Rules 2016`, already frozen in `docs/rules/conformance.md`. Task 1 records exact section/page references before mechanics are implemented.
- Four players only. Three-player rules, local platform house rules, and unnamed online-client behavior are excluded.
- Project options remain exactly those already validated by `RiichiConfig`: East-South match, 25,000 starting points, open tanyao enabled, and either three red fives or none.
- Dora, ura-dora, kan-dora, and red fives add han but never satisfy the one-yaku win requirement.
- Every ambiguous baseline decision, including multiple ron, abortive draws, kazoe limits, double yakuman, bankruptcy, all-last extension, and liability payments, is copied from the normative source into `docs/rules/riichi-source-notes.md`; implementation cannot begin until each row has a source citation and project value.
- Existing Chinese Official and compatibility modes must remain byte-for-byte stable unless a shared tile normalization test explicitly requires a compatible representation change.

## File Map

- Create `docs/rules/riichi-source-notes.md`: source pages, yaku table, fu table, limits, draw rules, priority, and match-end decisions.
- Create `testdata/rules/riichi/*.json`: wall, waits, calls, furiten, yaku, scoring, draws, settlement, and match fixtures.
- Modify `internal/game/tile.go`: represent red fives while normalizing them to the existing 34 base tile indexes.
- Create `internal/game/riichi_types.go`: typed round state, declarations, yaku matches, score breakdown, and settlement.
- Create `internal/game/riichi_wall.go`: 136-tile wall, dead wall, dora/ura indicators, rinshan draws, and kan-dora reveal.
- Create `internal/game/riichi_decompose.go`: standard, seven-pairs, and thirteen-orphans candidates plus wait classification.
- Create `internal/game/riichi_actions.go`: calls, kan variants, turn/call priority, and first-turn state.
- Create `internal/game/riichi_furiten.go`: tenpai, permanent/temporary/riichi furiten, riichi declaration, ippatsu, and legality.
- Create `internal/game/riichi_yaku.go`: normal yaku and yakuman detectors with open/closed han.
- Create `internal/game/riichi_score.go`: fu, dora, limits, best grouping, and readable breakdown.
- Create `internal/game/riichi_draw.go`: exhaustive and abortive draw resolution.
- Create `internal/game/riichi_settlement.go`: ron/tsumo transfers, honba, sticks, noten payments, dealer continuation, and match end.
- Modify `internal/game/match.go`, `snapshot.go`, and `replay.go`: preserve Riichi state and history.
- Modify `internal/online/server.go`: construct `RiichiRuleSet` for Riichi rooms.
- Modify `internal/bot`: consume only legal Riichi actions and never bypass furiten or yaku checks.
- Modify `docs/workflow.md` and `docs/rules/conformance.md`: record Phase 12C acceptance.

## Public Types Added In 12C

```go
type RiichiDeclarationState string

const (
	RiichiNone     RiichiDeclarationState = "none"
	RiichiDeclared RiichiDeclarationState = "declared"
	RiichiAccepted RiichiDeclarationState = "accepted"
)

type RiichiRoundState struct {
	DeadWall         []Tile                    `json:"dead_wall"`
	DoraIndicators   []Tile                    `json:"dora_indicators"`
	UraIndicators    []Tile                    `json:"ura_indicators,omitempty"`
	RinshanDraws     int                       `json:"rinshan_draws"`
	KanCount         int                       `json:"kan_count"`
	Honba            int                       `json:"honba"`
	RiichiSticks     int                       `json:"riichi_sticks"`
	Declarations     [4]RiichiDeclarationState `json:"declarations"`
	Ippatsu          [4]bool                   `json:"ippatsu"`
	TemporaryFuriten [4]bool                   `json:"temporary_furiten"`
	RiichiFuriten    [4]bool                   `json:"riichi_furiten"`
}

type RiichiShapeKind string
type RiichiWaitKind string

type RiichiDecomposition struct {
	Kind   RiichiShapeKind `json:"kind"`
	Groups []MCRGroup      `json:"groups"`
	Tiles  []Tile          `json:"tiles"`
	Wait   RiichiWaitKind  `json:"wait"`
}

type RiichiYakuMatch struct {
	ID       string `json:"id"`
	NameZH   string `json:"name_zh"`
	NameEN   string `json:"name_en"`
	Han      int    `json:"han"`
	Yakuman  int    `json:"yakuman,omitempty"`
}

type RiichiScoreBreakdown struct {
	Yaku          []RiichiYakuMatch `json:"yaku"`
	Fu            int               `json:"fu"`
	YakuHan       int               `json:"yaku_han"`
	BonusHan      int               `json:"bonus_han"`
	Yakuman       int               `json:"yakuman"`
	BasePoints    int               `json:"base_points"`
	LimitName     string            `json:"limit_name,omitempty"`
	HasYaku       bool              `json:"has_yaku"`
	WinningGroups []Meld            `json:"winning_groups"`
}

type RiichiSettlement struct {
	Winners     []int                  `json:"winners"`
	Discarder   int                    `json:"discarder"`
	Deltas      [4]int                 `json:"deltas"`
	Scores      []RiichiScoreBreakdown `json:"scores"`
	HonbaAfter  int                    `json:"honba_after"`
	SticksAfter int                    `json:"sticks_after"`
}

type RiichiSettlementInput struct {
	Winners      []int
	Discarder    int
	Dealer       int
	WinType      WinType
	Scores       []RiichiScoreBreakdown
	Honba        int
	RiichiSticks int
}

func ScoreRiichi(hand []Tile, melds []Meld, context RiichiYakuContext) RiichiScoreBreakdown
func SettleRiichi(input RiichiSettlementInput) RiichiSettlement
```

## Task 1: Freeze EMA Decisions And Golden Fixture Schemas

**Files:**
- Create: `docs/rules/riichi-source-notes.md`
- Create: `testdata/rules/riichi/catalog.json`
- Create: `internal/game/riichi_fixture_test.go`
- Modify: `docs/rules/conformance.md`

- [ ] **Step 1: Record normative decisions**

Create tables for tile/wall layout, calls, riichi, all furiten forms, every normal yaku/yakuman, fu, limits, exhaustive/abortive draws, multiple wins, honba/sticks, dealer continuation, and match end. Every row contains an EMA page/section and a concrete project value; no platform names are accepted as sources.

- [ ] **Step 2: Write the failing catalog test**

```go
func TestRiichiCatalogMatchesSourceNotes(t *testing.T) {
	catalog := loadRiichiCatalog(t, "../../testdata/rules/riichi/catalog.json")
	if len(catalog) == 0 { t.Fatal("empty riichi catalog") }
	assertUniqueRiichiIDs(t, catalog)
	assertEveryRiichiEntryHasSource(t, catalog)
	assertNoDoraEntryIsYaku(t, catalog)
}
```

- [ ] **Step 3: Verify RED**

Run `go test ./internal/game -run RiichiCatalog -count=1`. Expected: fail because the source notes, loader, and catalog do not exist.

- [ ] **Step 4: Add strict JSON loading and catalog data**

Use `json.Decoder.DisallowUnknownFields`; reject empty names, duplicate IDs, negative han, unsupported open/closed values, unresolved exclusions, and empty source citations.

- [ ] **Step 5: Verify and commit**

Run the catalog test and parse every Riichi JSON file with `ConvertFrom-Json`.

```powershell
git add docs/rules/riichi-source-notes.md docs/rules/conformance.md testdata/rules/riichi/catalog.json internal/game/riichi_fixture_test.go
git commit -m "docs: freeze EMA riichi rules"
```

Step review: every later expected value now points to the frozen EMA baseline instead of remembered platform behavior.

## Task 2: Add Red Fives, Dead Wall, Dora, And Rinshan

**Files:**
- Modify: `internal/game/tile.go`
- Modify: `internal/game/tile_test.go`
- Create: `internal/game/riichi_types.go`
- Create: `internal/game/riichi_wall.go`
- Create: `internal/game/riichi_wall_test.go`
- Create: `testdata/rules/riichi/wall.json`

- [ ] **Step 1: Write failing normalization and wall tests**

```go
func TestRedFivesNormalizeToBaseTiles(t *testing.T) {
	for red, base := range map[Tile]Tile{RedFiveMan: 4, RedFivePin: 13, RedFiveSou: 22} {
		if red.Base() != base || red.Rank() != 5 { t.Fatalf("red %v base=%v", red, red.Base()) }
	}
}

func TestRiichiWallPreservesDeadWallAndIndicators(t *testing.T) {
	round := newRiichiRoundFromFixture(t, "three-red-fives")
	assertRiichiTileConservation(t, round, 136)
	if len(round.Riichi.DeadWall) != 14 || len(round.Riichi.DoraIndicators) != 1 { t.Fatal(round.Riichi) }
}
```

- [ ] **Step 2: Verify RED**

Run `go test ./internal/game -run "RedFive|RiichiWall|Rinshan|DoraIndicator" -count=1`.

- [ ] **Step 3: Implement base-tile normalization**

Add `RedFiveMan`, `RedFivePin`, and `RedFiveSou`; implement `Base`, `IsRed`, and red text forms. Make `TileCounts`, suit/rank checks, sorting, decomposition, and discard matching use `Base()` while preserving the red identity in hands, melds, events, and replay.

- [ ] **Step 4: Implement `RiichiRuleSet` setup**

```go
type RiichiRuleSet struct { config RiichiConfig }

func (rules *RiichiRuleSet) BuildWall() []Tile
func (rules *RiichiRuleSet) Deal(round *Game) error
func (rules *RiichiRuleSet) Draw(round *Game, player int, source DrawSource) (Tile, bool)
```

Split the shuffled wall into a 122-tile live portion and a 14-tile dead wall before dealing. Normal draws reduce the live wall; kan draws consume one of four rinshan positions and reveal the next kan-dora indicator according to the frozen source layout.

- [ ] **Step 5: Verify compatibility and commit**

Run red-five modes `0` and `3`, 1,000 wall seeds, MCR conservation, compatibility replay hashes, and `go test ./internal/game -run "Wall|Tile" -count=20`.

```powershell
git add internal/game/tile.go internal/game/tile_test.go internal/game/riichi_types.go internal/game/riichi_wall.go internal/game/riichi_wall_test.go testdata/rules/riichi/wall.json
git commit -m "feat: build riichi wall and dead wall"
```

Step review: Japanese wall state is explicit and reconnectable without changing MCR or compatibility tile counts.

## Task 3: Enumerate Riichi Hands, Waits, And Tenpai

**Files:**
- Create: `internal/game/riichi_decompose.go`
- Create: `internal/game/riichi_decompose_test.go`
- Create: `testdata/rules/riichi/decompositions.json`

- [ ] **Step 1: Add source-cited decomposition fixtures**

Cover standard ambiguity, open melds, seven pairs, thirteen orphans, tanki, kanchan, penchan, ryanmen, shanpon, red-five normalization, and invalid duplicate/flower inputs.

- [ ] **Step 2: Write failing tests**

```go
func TestRiichiWaitsAreTileIdentityIndependent(t *testing.T) {
	waits := RiichiWaits(parseTiles(t, "123m456m789p22s34s"), nil)
	assertTiles(t, waits, "2s", "5s")
}

func TestRiichiTenpaiIncludesSevenPairsAndKokushi(t *testing.T) {
	for _, fixture := range loadRiichiDecompositionFixtures(t) {
		if got := RiichiTenpai(fixture.Hand, fixture.Melds); got != fixture.Tenpai { t.Fatal(fixture.ID) }
	}
}
```

- [ ] **Step 3: Verify RED**

Run `go test ./internal/game -run "RiichiDecompose|RiichiWait|RiichiTenpai" -count=1`.

- [ ] **Step 4: Implement pure decomposition**

Normalize red tiles for count recursion but copy original identities into winning candidates. Return every canonical standard grouping plus seven-pairs and thirteen-orphans candidates; classify the winning wait per candidate and never choose score in this layer.

- [ ] **Step 5: Verify invariants and commit**

Test permutation invariance, candidate deduplication, no tile loss, and bounded execution on ambiguous hands.

```powershell
git add internal/game/riichi_decompose.go internal/game/riichi_decompose_test.go testdata/rules/riichi/decompositions.json
git commit -m "feat: enumerate riichi hands and waits"
```

Step review: yaku and furiten can consume exhaustive waits without duplicating hand parsing.

## Task 4: Implement Calls, Kan Variants, And Round Windows

**Files:**
- Create: `internal/game/riichi_actions.go`
- Create: `internal/game/riichi_actions_test.go`
- Modify: `internal/game/claim_state.go`
- Modify: `internal/game/snapshot.go`
- Create: `testdata/rules/riichi/actions.json`

- [ ] **Step 1: Add legal-action fixtures**

Cover ron priority, pon/daiminkan over chi, chi by next seat only, ankan, daiminkan, shouminkan and chankan, rinshan exhaustion, four-kan abort rules, first-turn interruption, and kuikae restrictions exactly as frozen in Task 1.

- [ ] **Step 2: Write failing command-closure tests**

```go
func TestRiichiLegalActionFixtures(t *testing.T) {
	for _, fixture := range loadRiichiActionFixtures(t) {
		round := riichiRoundFromFixture(t, fixture.Initial)
		actions := round.rules.LegalActions(round, fixture.PlayerID)
		assertRiichiActions(t, actions, fixture.Expected)
		assertEveryActionAppliesToFreshCopy(t, round, fixture.PlayerID, actions)
	}
}
```

- [ ] **Step 3: Verify RED**

Run `go test ./internal/game -run "RiichiLegal|RiichiCall|RiichiKan" -count=1`.

- [ ] **Step 4: Implement typed call windows**

Extend `PendingClaim` only with fields shared snapshots need. `RiichiRuleSet` constructs ordered ron/pon/kan/chi options; shouminkan remains uncommitted during chankan; accepted kan updates dead-wall state before exposing the next legal discard or win.

- [ ] **Step 5: Verify and commit**

Run action fixtures 20 times, compatibility/MCR claim tests, bot command closure, and pending-window reconnect tests.

```powershell
git add internal/game/riichi_actions.go internal/game/riichi_actions_test.go internal/game/claim_state.go internal/game/snapshot.go testdata/rules/riichi/actions.json
git commit -m "feat: enforce riichi calls and kan windows"
```

Step review: all call and kan mutations are delayed until their response windows resolve and remain reproducible after reconnect.

## Task 5: Implement Riichi Declaration, Ippatsu, And Furiten

**Files:**
- Create: `internal/game/riichi_furiten.go`
- Create: `internal/game/riichi_furiten_test.go`
- Modify: `internal/game/snapshot.go`
- Create: `testdata/rules/riichi/furiten.json`

- [ ] **Step 1: Add declaration/furiten fixtures**

Cover closed tenpai, four-tile wall threshold, 1,000-point payment, declaration discard, accepted riichi, double riichi, ippatsu cancellation, own-discard furiten, temporary pass furiten, riichi pass furiten, turn reset, and tsumo while furiten.

- [ ] **Step 2: Write failing state-transition tests**

```go
func TestRiichiFuritenFixtures(t *testing.T) {
	for _, fixture := range loadRiichiFuritenFixtures(t) {
		round := riichiRoundFromFixture(t, fixture.Initial)
		for _, command := range fixture.Commands { mustApplyRiichi(t, round, command) }
		assertRiichiDeclarationState(t, round.Riichi, fixture.Expected)
	}
}
```

- [ ] **Step 3: Verify RED**

Run `go test ./internal/game -run "RiichiDeclaration|Ippatsu|Furiten" -count=1`.

- [ ] **Step 4: Implement declaration and furiten state**

Add `CommandRiichi` with the declaration discard index. Validate closed tenpai, remaining live tiles, and points before accepting; lock post-riichi discards to tsumogiri except legal kan decisions. Recompute permanent furiten from waits/discards, set temporary furiten on passed ron, make it persistent after riichi, and clear only the source-sanctioned states.

- [ ] **Step 5: Verify and commit**

Run fixtures, snapshot-copy tests, reconnect canonical JSON, and random legal-action closure.

```powershell
git add internal/game/riichi_furiten.go internal/game/riichi_furiten_test.go internal/game/snapshot.go testdata/rules/riichi/furiten.json
git commit -m "feat: enforce riichi and furiten"
```

Step review: ron legality now comes from one authoritative wait/furiten state and cannot be bypassed by clients or bots.

## Task 6: Detect Every EMA Yaku And Yakuman

**Files:**
- Create: `internal/game/riichi_yaku.go`
- Create: `internal/game/riichi_yaku_test.go`
- Create: `testdata/rules/riichi/yaku/*.json`

- [ ] **Step 1: Add one positive and one near-miss per catalog ID**

Fixtures include open/closed variants, event-only yaku, pair/wait-sensitive yaku, yakuman, and source-specific double-yakuman treatment. Bonus indicators are fixture context but never expected as yaku.

- [ ] **Step 2: Write failing catalog-driven coverage**

```go
func TestEveryRiichiYakuHasGoldenCoverage(t *testing.T) {
	for _, entry := range loadRiichiCatalog(t, riichiCatalogPath) {
		if !hasRiichiPositiveAndNearMiss(entry.ID) { t.Errorf("%s lacks coverage", entry.ID) }
	}
}
```

- [ ] **Step 3: Verify RED**

Run `go test ./internal/game -run "EveryRiichiYaku|RiichiYaku" -count=1`.

- [ ] **Step 4: Implement typed detectors**

```go
type RiichiYakuContext struct {
	Decomposition RiichiDecomposition
	WinningTile Tile
	WinType WinType
	Closed bool
	SeatWind Tile
	PrevalentWind Tile
	Riichi RiichiDeclarationState
	Ippatsu bool
	Rinshan bool
	Chankan bool
	Haitei bool
	Houtei bool
}
```

Detectors return independent yaku matches; they do not count dora and do not suppress another detector. Yakuman multiplicity follows only Task 1's source table.

- [ ] **Step 5: Verify bands and commit**

Run normal-yaku and yakuman tests separately, then all yaku tests 20 times.

```powershell
git add internal/game/riichi_yaku.go internal/game/riichi_yaku_test.go testdata/rules/riichi/yaku
git commit -m "feat: detect complete EMA riichi yaku"
```

Step review: the one-yaku gate and every source-defined limit hand are independently testable before score arithmetic.

## Task 7: Calculate Fu, Han, Dora, Limits, And Best Score

**Files:**
- Create: `internal/game/riichi_score.go`
- Create: `internal/game/riichi_score_test.go`
- Create: `testdata/rules/riichi/scoring.json`

- [ ] **Step 1: Add golden score fixtures**

Cover every fu component, pinfu tsumo, seven-pairs 25 fu, open minimum fu, rounding, red/normal/kan/ura dora, dora-only rejection, dealer/non-dealer ron and tsumo, mangan through yakuman, multiple yakuman, and ambiguous grouping choosing the highest payment.

- [ ] **Step 2: Write failing score tests**

```go
func TestRiichiScoringGoldenFixtures(t *testing.T) {
	for _, fixture := range loadRiichiScoringFixtures(t) {
		got := ScoreRiichi(fixture.Hand, fixture.Melds, fixture.Context)
		assertRiichiScore(t, got, fixture.Expected)
	}
}
```

- [ ] **Step 3: Verify RED**

Run `go test ./internal/game -run "RiichiFu|RiichiScoring|RiichiDora|RiichiLimit" -count=1`.

- [ ] **Step 4: Implement score pipeline**

For each decomposition: detect yaku, reject dora-only candidates, calculate fu, count visible dora/red fives and eligible ura/kan indicators, apply source-defined limits, and retain the candidate producing the highest actual payment. Return base points and payer-specific rounded values without mutating match totals.

- [ ] **Step 5: Verify and commit**

Run score fixtures, permutation tests, indicator wrap tests, and `go test ./internal/game -run Riichi -count=20`.

```powershell
git add internal/game/riichi_score.go internal/game/riichi_score_test.go testdata/rules/riichi/scoring.json
git commit -m "feat: score complete EMA riichi hands"
```

Step review: every accepted win has a reproducible yaku/fu/han breakdown and bonus tiles cannot create legality.

## Task 8: Resolve Draws, Settlement, Dealer Continuation, And Match End

**Files:**
- Create: `internal/game/riichi_draw.go`
- Create: `internal/game/riichi_draw_test.go`
- Create: `internal/game/riichi_settlement.go`
- Create: `internal/game/riichi_settlement_test.go`
- Modify: `internal/game/match.go`
- Create: `testdata/rules/riichi/draws.json`
- Create: `testdata/rules/riichi/settlement.json`

- [ ] **Step 1: Add draw and settlement fixtures**

Cover exhaustive draws with 0/1/2/3/4 tenpai players, every frozen abortive draw, dealer/non-dealer ron and tsumo, honba, riichi-stick collection, dealer win/draw continuation, round advancement, bankruptcy, South 4, source-defined extension, and final ranking ties.

- [ ] **Step 2: Write failing zero-sum and progression tests**

```go
func TestRiichiSettlementGoldenFixtures(t *testing.T) {
	for _, fixture := range loadRiichiSettlementFixtures(t) {
		got := SettleRiichi(fixture.Input)
		if got.Deltas != fixture.Deltas || sum4(got.Deltas) != 0 { t.Fatal(fixture.ID, got) }
	}
}
```

- [ ] **Step 3: Verify RED**

Run `go test ./internal/game -run "RiichiDraw|RiichiSettlement|RiichiMatch" -count=1`.

- [ ] **Step 4: Implement state transitions**

Apply rounded payments, honba, sticks, and noten transfers as pure settlements, then let `Match` update points once. Continue or rotate dealer exactly as the source table requires; reveal ura indicators only for eligible riichi winners; preserve all hand settlements and terminate only under frozen match-end conditions.

- [ ] **Step 5: Verify and commit**

Run deterministic East-South simulations, zero-sum properties, and 1,000 fixed-seed terminating matches.

```powershell
git add internal/game/riichi_draw.go internal/game/riichi_draw_test.go internal/game/riichi_settlement.go internal/game/riichi_settlement_test.go internal/game/match.go testdata/rules/riichi/draws.json testdata/rules/riichi/settlement.json
git commit -m "feat: settle complete riichi matches"
```

Step review: a full East-South match advances, repeats, and terminates through shared `Match` APIs without MCR branches.

## Task 9: Synchronize Riichi State Across Server, Bots, Snapshots, And Replay

**Files:**
- Modify: `internal/game/snapshot.go`
- Modify: `internal/game/replay.go`
- Modify: `internal/online/server.go`
- Modify: `internal/online/server_test.go`
- Modify: `internal/bot/heuristic.go`
- Modify: `internal/bot/heuristic_test.go`

- [ ] **Step 1: Write failing integration tests**

Assert recipient-private snapshots preserve dead-wall counts, public dora, dealer, honba, sticks, declarations, each viewer's own furiten/legal actions, and pending kan/ron windows. Canonical reconnect JSON must match before disconnect. Replay contains concealed ura only after game end and all settlements.

- [ ] **Step 2: Verify RED**

Run `go test ./internal/game ./internal/online ./internal/bot -run "Riichi.*Snapshot|Riichi.*Reconnect|Riichi.*Replay|Riichi.*Bot" -count=1`.

- [ ] **Step 3: Wire authoritative state**

For `ModeRiichi`, `roomRuleSet` returns `NewRiichiRuleSet(config.Riichi)`. Snapshot redaction hides dead-wall identities, ura indicators, opponent hands, and private waits/furiten while retaining public dora and declarations. Bots select only from `LegalActions`; add no private server-only information to `GameSnapshot`.

- [ ] **Step 4: Verify and commit**

Run focused reconnect tests, every bot action against a fresh state, and online tests 20 times under both red-five options.

```powershell
git add internal/game/snapshot.go internal/game/replay.go internal/online/server.go internal/online/server_test.go internal/bot/heuristic.go internal/bot/heuristic_test.go
git commit -m "feat: synchronize complete riichi state"
```

Step review: live, resumed, bot, and replay consumers share one authoritative rules state without leaking hidden indicators or concealed tiles.

## Task 10: Phase 12C Acceptance And Review

**Files:**
- Modify: `docs/rules/conformance.md`
- Modify: `docs/workflow.md`
- Create: `internal/game/riichi_invariant_test.go`

- [ ] **Step 1: Validate source/catalog/fixtures**

```powershell
Get-ChildItem testdata/rules/riichi -Recurse -Filter *.json | ForEach-Object { Get-Content -Raw $_.FullName | ConvertFrom-Json | Out-Null }
go test ./internal/game -run "RiichiCatalog|EveryRiichiYaku|RiichiScoring|RiichiSettlement" -count=20
```

- [ ] **Step 2: Run generated invariants**

Across at least 1,000 fixed seeds and both red-five modes, assert 136-tile conservation, 14-tile dead-wall accounting, at most four rinshan draws, indicator bounds, no negative counts, every returned action accepted from the same state, furiten ron never exposed, all transfers zero-sum, and deterministic match termination.

- [ ] **Step 3: Run static, race, full, and build checks**

```powershell
$files = Get-ChildItem internal,cmd -Recurse -Filter *.go | ForEach-Object FullName
if (gofmt -l $files) { throw "unformatted Go files" }
git diff --check
go vet ./...
go test ./... -count=1
go test -race ./... -count=1
go build ./cmd/mahjong ./cmd/server ./cmd/client
go test ./internal/online -count=20
```

- [ ] **Step 4: Run local and WebSocket smoke matches**

Complete one deterministic local East-South match and one WebSocket match containing riichi, ippatsu cancellation, furiten rejection, chi, pon, all three kan forms, dora/ura/red scoring, exhaustive draw, honba, sticks, disconnect, reconnect, and final settlement. Verify clients never receive opponent hands, dead-wall identities, ura indicators before a win, or shuffle seeds.

- [ ] **Step 5: Update review and commit**

Mark 12C complete and 12D not started. Record exact evidence and remaining client-integration risk.

```powershell
git add internal/game/riichi_invariant_test.go docs/rules/conformance.md docs/workflow.md
git commit -m "test: record phase 12c acceptance"
```

Phase review: begin Phase 12D only if Riichi rooms use `RiichiRuleSet`, all catalog yaku have positive/near-miss coverage, no MCR or compatibility scoring path is reachable, hidden dead-wall state remains private, and fresh full/race tests pass.

## Plan Self-Review

- Spec coverage: wall/dead wall, red fives, dora, calls/kan, tenpai, riichi, all furiten states, yaku/yakuman, fu/han/limits, exhaustive/abortive draws, honba/sticks, dealer continuation, match end, bots, snapshots, reconnect, replay, and acceptance each have a dedicated task.
- Scope: three-player rules, final mode-selection UI, Phase 13 table redesign, and Phase 14 replay controls remain excluded.
- Type consistency: `RiichiRoundState`, `RiichiYakuContext`, `RiichiScoreBreakdown`, `RiichiSettlement`, `RiichiRuleSet`, `ScoreRiichi`, and `SettleRiichi` are introduced before use.
- Isolation: MCR types and fixtures remain untouched except shared red-tile normalization tests; Riichi scoring does not call MCR fan detection or compatibility scoring.
- Placeholder scan command:

```powershell
$patterns = @("TB"+"D", "TO"+"DO", "implement"+" later", "fill in"+" details", "appropriate"+" error", "Similar"+" to")
foreach ($pattern in $patterns) { rg -n $pattern docs/superpowers/plans/2026-06-24-phase-12c-complete-riichi.md }
```

Expected: no matches.
