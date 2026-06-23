# Phase 12B Complete Chinese Official Mahjong Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the temporary compatibility mechanics for `ModeMCR` with a complete, source-cited Chinese Official Mahjong round and 16-hand match implementation, including flowers, legal calls, all 81 scoring elements, the eight-point minimum, and zero-sum settlement.

**Architecture:** Keep shared transport and snapshot types in `internal/game`, but isolate Chinese Official behavior in focused `mcr_*.go` files. The MCR scorer enumerates legal hand decompositions, evaluates typed fan detectors, applies explicit exclusion/counting rules, and returns a readable `MCRScoreBreakdown`; game flow consumes that result through `MCRRuleSet`, so TUI and WebSocket code do not contain scoring branches.

**Tech Stack:** Go standard library, existing `game.Match`/`RuleSet`, JSON golden fixtures, standard `testing`, race detector.

---

## Source And Scope Contract

- Normative source: `GB/T 34708-2017 麻将竞赛规则`, already frozen in `docs/rules/conformance.md`.
- Secondary architecture study only: [dchen327/mahjong-calc](https://github.com/dchen327/mahjong-calc), MIT. Its point-band file organization and grouping-first evaluation are useful references; its code and expected values are not normative and are not copied.
- Project match default: four rounds of four hands, dealer advances after every completed hand, 16 hands total.
- One discarded tile produces at most one winner; among simultaneous winning claims, the nearest eligible player after the discarder has priority.
- Claim priority is win, exposed kong/pong, then chow by the next seat.
- Flower replacements are automatic from the back of the live wall. Flower points are added to settlement after a hand already reaches the eight-point non-flower minimum.
- Self-draw: every opponent pays `8 + total fan`. Discard win: the discarder pays `8 + total fan`, and each other opponent pays `8`.
- No regional options, dealer repeats, limit caps, or unofficial patterns are added in 12B.

## File Map

- Modify `docs/rules/conformance.md`: mark 12A complete and record the exact MCR flow/settlement decisions above.
- Create `docs/rules/mcr-fan-catalog.md`: all 81 standard scoring elements by point band, Chinese/English names, exclusions, and source section.
- Create `testdata/rules/mcr/*.json`: source-cited wall, call, score, minimum, and settlement fixtures.
- Modify `internal/game/tile.go` and tests: represent eight unique flowers without changing the 34 suited/honor count arrays.
- Modify `internal/game/player.go` and snapshots: store/reveal flower tiles and hand counts correctly.
- Modify `internal/game/rules.go`, `game.go`, and tests: let a `RuleSet` build/deal its own wall while preserving compatibility behavior.
- Create `internal/game/mcr_types.go`: typed fan IDs, score context, breakdown, settlement, and MCR round metadata.
- Create `internal/game/mcr_wall.go`: 144-tile wall, dealing, flower replacement, replacement draws, and tile-conservation helpers.
- Create `internal/game/mcr_decompose.go`: standard and special-hand decomposition candidates.
- Create `internal/game/mcr_fans_1_8.go`: 1/2/4/6/8 point detectors.
- Create `internal/game/mcr_fans_12_88.go`: 12/16/24/32/48/64/88 point detectors.
- Create `internal/game/mcr_score.go`: exclusion/counting principles, best-decomposition selection, minimum check, and readable breakdown.
- Create `internal/game/mcr_rules.go`: MCR legal actions, call priority, kong/robbing flow, and win validation.
- Create `internal/game/mcr_settlement.go`: hand deltas and 16-hand match progression.
- Create matching `*_test.go` files beside each component.
- Modify `internal/game/match.go`, `snapshot.go`, `replay.go`: preserve MCR round/match metadata and score breakdown across snapshots/replay.
- Modify `internal/online/server.go`: construct `MCRRuleSet` for MCR rooms; keep compatibility rooms unchanged.
- Modify `docs/workflow.md`: record Phase 12B acceptance evidence and remaining 12C risk.

## Public Types Added In 12B

```go
type FanID string

type FanMatch struct {
	ID       FanID `json:"id"`
	NameZH   string `json:"name_zh"`
	NameEN   string `json:"name_en"`
	Points   int    `json:"points"`
	Count    int    `json:"count"`
}

type MCRScoreContext struct {
	Winner          int
	Discarder       int
	WinningTile     Tile
	WinType         WinType
	SeatWind        Tile
	PrevalentWind   Tile
	LastTileDraw    bool
	LastTileClaim   bool
	ReplacementDraw bool
	RobbingKong     bool
}

type MCRScoreBreakdown struct {
	Fans             []FanMatch `json:"fans"`
	NonFlowerPoints  int        `json:"non_flower_points"`
	FlowerPoints     int        `json:"flower_points"`
	TotalPoints      int        `json:"total_points"`
	MeetsMinimum     bool       `json:"meets_minimum"`
	WinningGrouping  []Meld     `json:"winning_grouping"`
}

type MCRSettlement struct {
	Winner    int               `json:"winner"`
	Discarder int               `json:"discarder"`
	Deltas    [4]int            `json:"deltas"`
	Score     MCRScoreBreakdown `json:"score"`
}
```

## Task 1: Freeze The MCR Catalog And Golden Fixture Envelope

**Files:**
- Modify: `docs/rules/conformance.md`
- Create: `docs/rules/mcr-fan-catalog.md`
- Create: `testdata/rules/mcr/catalog.json`
- Create: `internal/game/mcr_fixture_test.go`

- [ ] **Step 1: Write the complete catalog table**

Create point-band sections for `88, 64, 48, 32, 24, 16, 12, 8, 6, 4, 2, 1`. Include all 81 standard elements, including Chicken Hand and Flower Tiles. Every row has stable `FanID`, Chinese name, English name, points, explicit excluded IDs, repeatability, and a `GB/T 34708-2017` section/table citation.

- [ ] **Step 2: Write the failing catalog completeness test**

```go
func TestMCRFanCatalogIsComplete(t *testing.T) {
	catalog := loadMCRCatalog(t, "../../testdata/rules/mcr/catalog.json")
	if len(catalog) != 81 {
		t.Fatalf("fan count = %d, want 81", len(catalog))
	}
	wantBands := map[int]int{88: 7, 64: 6, 48: 2, 32: 3, 24: 9, 16: 6, 12: 5, 8: 10, 6: 6, 4: 4, 2: 10, 1: 13}
	assertMCRPointBands(t, catalog, wantBands)
	assertUniqueFanIDsAndNames(t, catalog)
	assertCatalogExclusionsResolve(t, catalog)
}
```

- [ ] **Step 3: Verify RED**

Run `go test ./internal/game -run MCRFanCatalog -count=1`.

Expected: fail because the fixture loader and catalog do not exist.

- [ ] **Step 4: Add the fixture loader and catalog**

Use `encoding/json`; reject unknown fields, duplicate IDs, unsupported point values, unresolved exclusions, and empty source citations. Do not create scoring detectors in this task.

- [ ] **Step 5: Verify and commit**

Run `go test ./internal/game -run MCRFanCatalog -count=1` and validate every JSON file with `ConvertFrom-Json`.

```powershell
git add docs/rules/conformance.md docs/rules/mcr-fan-catalog.md testdata/rules/mcr/catalog.json internal/game/mcr_fixture_test.go
git commit -m "docs: freeze complete MCR fan catalog"
```

Step review: every later score assertion now names a stable fan ID and normative source instead of relying on prose or an external calculator.

## Task 2: Add Flower Tiles And Rule-Owned Round Setup

**Files:**
- Modify: `internal/game/tile.go`
- Modify: `internal/game/tile_test.go`
- Modify: `internal/game/player.go`
- Modify: `internal/game/rules.go`
- Modify: `internal/game/game.go`
- Modify: `internal/game/rules_test.go`

- [ ] **Step 1: Write failing tile and setup tests**

```go
func TestBuildMCRWallContains144Tiles(t *testing.T) {
	wall := BuildMCRWall()
	if len(wall) != 144 { t.Fatalf("wall = %d", len(wall)) }
	for tile := Tile(0); tile < BaseTileTypeCount; tile++ {
		if countTile(wall, tile) != 4 { t.Fatalf("%s copies != 4", tile) }
	}
	for tile := FlowerSpring; tile <= FlowerWinter; tile++ {
		if countTile(wall, tile) != 1 { t.Fatalf("%s copies != 1", tile) }
	}
}

func TestCompatibilitySetupRemainsDeterministic(t *testing.T) {
	before := NewGame(31)
	after, err := NewGameWithRules(31, NewCompatibilityRuleSet(ModeCompatibility, RuleConfig{}))
	if err != nil { t.Fatal(err) }
	if before.ShuffleProof != after.ShuffleProof || FormatTiles(before.Players[0].Hand) != FormatTiles(after.Players[0].Hand) {
		t.Fatal("compatibility setup changed")
	}
}
```

- [ ] **Step 2: Verify RED**

Run `go test ./internal/game -run "MCRWall|CompatibilitySetup" -count=1`.

- [ ] **Step 3: Implement the tile boundary**

Keep `BaseTileTypeCount = 34` for count arrays. Add flower values `FlowerPlum` through `FlowerWinter` after the base tiles, `IsFlower`, stable text forms `P1..P4` and `S1..S4`, and Unicode/localized labels. `TileCounts` continues to count only base tiles.

- [ ] **Step 4: Extend RuleSet setup**

Add these methods and implement compatibility behavior with the current 136-tile wall and 13-tile round-robin deal:

```go
type RuleSet interface {
	// existing methods remain
	BuildWall() []Tile
	Deal(round *Game) error
}
```

`NewGameWithRules` shuffles `rules.BuildWall()`, creates the proof from the complete shuffled wall, then calls `rules.Deal(round)`.

- [ ] **Step 5: Verify and commit**

Run focused tests, `go test ./internal/game ./internal/online -count=1`, and fixed-seed replay tests.

```powershell
git add internal/game/tile.go internal/game/tile_test.go internal/game/player.go internal/game/rules.go internal/game/game.go internal/game/rules_test.go
git commit -m "feat: support rule-owned walls and flower tiles"
```

Step review: MCR can own a 144-tile setup without changing compatibility replay hashes or introducing flower values into 34-tile hand algorithms.

## Task 3: Implement MCR Deal And Automatic Flower Replacement

**Files:**
- Create: `internal/game/mcr_types.go`
- Create: `internal/game/mcr_wall.go`
- Create: `internal/game/mcr_wall_test.go`
- Modify: `internal/game/player.go`
- Modify: `internal/game/snapshot.go`
- Modify: `internal/game/event.go`

- [ ] **Step 1: Add source-cited wall fixtures**

Create `testdata/rules/mcr/wall_flow.json` with fixed walls covering no initial flowers, flowers in every seat, consecutive replacement flowers, replacement-wall exhaustion, and a flower drawn during play.

- [ ] **Step 2: Write failing fixture tests**

```go
func TestMCRDealReplacesEveryFlower(t *testing.T) {
	round := newMCRRoundFromFixture(t, "initial-consecutive-flowers")
	for seat := range round.Players {
		if len(round.Players[seat].Hand) != 13 { t.Fatalf("seat %d hand=%d", seat, len(round.Players[seat].Hand)) }
		if containsFlower(round.Players[seat].Hand) { t.Fatalf("seat %d kept flower", seat) }
	}
	assertTileConservation(t, round, 144)
}
```

- [ ] **Step 3: Verify RED**

Run `go test ./internal/game -run "MCRDeal|MCRFlower|MCRReplacement" -count=1`.

- [ ] **Step 4: Implement automatic replacement**

Add `Player.Flowers []Tile`, `EventFlower`, and `EventReplacementDraw`. Normal draws take the front; replacement draws take the back. A flower is exposed, recorded, and replaced repeatedly until a base tile arrives or the wall is exhausted. Snapshot privacy always reveals exposed flowers but redacts a private replacement draw tile like any other draw.

- [ ] **Step 5: Verify and commit**

Run fixture tests, private snapshot tests, and tile-conservation tests 20 times.

```powershell
git add testdata/rules/mcr/wall_flow.json internal/game/mcr_types.go internal/game/mcr_wall.go internal/game/mcr_wall_test.go internal/game/player.go internal/game/snapshot.go internal/game/event.go
git commit -m "feat: deal MCR flowers and replacements"
```

Step review: every MCR draw path preserves 144 tiles and never leaves a flower in a concealed hand.

## Task 4: Enumerate Standard And Special Winning Shapes

**Files:**
- Create: `internal/game/mcr_decompose.go`
- Create: `internal/game/mcr_decompose_test.go`
- Create: `testdata/rules/mcr/decompositions.json`

- [ ] **Step 1: Write decomposition fixtures**

Cover ambiguous standard hands, seven pairs, seven shifted pairs, thirteen orphans, greater/lesser honors and knitted tiles, knitted straight, and nine gates. Each fixture provides concealed tiles, declared melds, winning tile, and every legal grouping expected.

- [ ] **Step 2: Write failing tests**

```go
func TestMCRDecomposeGoldenFixtures(t *testing.T) {
	for _, fixture := range loadMCRDecompositionFixtures(t) {
		got := MCRDecompose(fixture.Hand, fixture.Melds, fixture.WinningTile)
		assertCanonicalGroupings(t, got, fixture.ExpectedGroupings)
	}
}
```

- [ ] **Step 3: Verify RED**

Run `go test ./internal/game -run MCRDecompose -count=1`.

- [ ] **Step 4: Implement exhaustive decomposition**

Use count-array recursion for pair/chow/pung candidates and separate recognizers for the six special shape families. Return canonical immutable candidates; deduplicate equivalent groupings; include open melds in every candidate; never select a score in this layer.

- [ ] **Step 5: Add invariants and commit**

Test permutation invariance, no tile loss, no duplicate grouping, and maximum execution time for worst-case ambiguous hands.

```powershell
git add internal/game/mcr_decompose.go internal/game/mcr_decompose_test.go testdata/rules/mcr/decompositions.json
git commit -m "feat: enumerate MCR winning shapes"
```

Step review: scoring receives all legal interpretations and can choose the maximum valid breakdown without hiding decomposition shortcuts.

## Task 5: Implement All 81 Fan Detectors By Point Band

**Files:**
- Create: `internal/game/mcr_fans_1_8.go`
- Create: `internal/game/mcr_fans_12_88.go`
- Create: `internal/game/mcr_fans_test.go`
- Create: `testdata/rules/mcr/fans/*.json`

- [ ] **Step 1: Add one positive and one near-miss fixture per fan**

Every catalog ID gets a positive fixture. Shape-sensitive fans also get a one-tile near miss; context-sensitive fans get opposite-context cases. Flower Tiles uses exposed flowers and is marked as post-minimum only.

- [ ] **Step 2: Write the failing catalog-driven test**

```go
func TestEveryMCRFanHasGoldenCoverage(t *testing.T) {
	catalog := loadMCRCatalog(t, catalogPath)
	fixtures := loadAllMCRFanFixtures(t)
	for _, fan := range catalog {
		if !hasPositiveFixture(fixtures, fan.ID) { t.Errorf("%s has no positive fixture", fan.ID) }
	}
}
```

- [ ] **Step 3: Verify RED**

Run `go test ./internal/game -run "EveryMCRFan|MCRFan" -count=1`.

- [ ] **Step 4: Implement typed detectors**

Each detector has this shape and returns occurrence/group usage data, not points alone:

```go
type mcrFanDetector struct {
	ID       FanID
	Points   int
	Detect   func(MCRFanContext) []MCRFanOccurrence
}
```

Keep low and high point bands in separate files. Detector IDs and points must match the catalog test. Event-only fans read `MCRScoreContext`; shape detectors read one decomposition; no detector suppresses another detector directly.

- [ ] **Step 5: Verify bands incrementally**

Run one test command per point band, then `go test ./internal/game -run MCRFan -count=1`.

- [ ] **Step 6: Commit**

```powershell
git add internal/game/mcr_fans_1_8.go internal/game/mcr_fans_12_88.go internal/game/mcr_fans_test.go testdata/rules/mcr/fans
git commit -m "feat: detect all MCR scoring elements"
```

Step review: all 81 elements are independently detectable and source-cited before exclusion or minimum logic can mask mistakes.

## Task 6: Apply Exclusions, Counting Principles, And Eight-Point Minimum

**Files:**
- Create: `internal/game/mcr_score.go`
- Create: `internal/game/mcr_score_test.go`
- Create: `testdata/rules/mcr/scoring.json`

- [ ] **Step 1: Add combined-score fixtures**

Include higher-fan exclusions, account-once, non-identical, identical-set limits, ambiguous grouping choosing the maximum, Chicken Hand, exactly 7/8/9 non-flower points, and 7 points plus one or more flowers remaining ineligible.

- [ ] **Step 2: Write failing score tests**

```go
func TestMCRScoringGoldenFixtures(t *testing.T) {
	for _, fixture := range loadMCRScoringFixtures(t) {
		got := ScoreMCR(fixture.Hand, fixture.Melds, fixture.Context)
		assertFanIDs(t, got.Fans, fixture.ExpectedFans)
		if got.NonFlowerPoints != fixture.NonFlower || got.FlowerPoints != fixture.Flowers || got.MeetsMinimum != fixture.Eligible {
			t.Fatalf("%s score = %#v", fixture.ID, got)
		}
	}
}
```

- [ ] **Step 3: Verify RED**

Run `go test ./internal/game -run MCRScoring -count=1`.

- [ ] **Step 4: Implement the scorer**

For every decomposition: collect occurrences, apply catalog exclusions from highest value downward, enforce standard counting principles using occurrence group IDs, add Chicken Hand only when no other non-flower fan remains, and choose the highest legal non-flower total. Add flower points afterward. `MeetsMinimum` is `NonFlowerPoints >= config.MinimumPoints`.

- [ ] **Step 5: Verify and commit**

Run scoring fixtures, permutation tests, and `go test ./internal/game -run MCR -count=20`.

```powershell
git add internal/game/mcr_score.go internal/game/mcr_score_test.go testdata/rules/mcr/scoring.json
git commit -m "feat: score complete MCR fan breakdowns"
```

Step review: every accepted win now has a readable, reproducible breakdown and the minimum cannot be reached using flowers alone.

## Task 7: Implement MCR Legal Actions, Calls, And Kong Windows

**Files:**
- Create: `internal/game/mcr_rules.go`
- Create: `internal/game/mcr_rules_test.go`
- Modify: `internal/game/claim_state.go`
- Modify: `internal/game/snapshot.go`
- Create: `testdata/rules/mcr/legal_actions.json`

- [ ] **Step 1: Add call-priority fixtures**

Cover win over kong/pong/chow, nearest winner among simultaneous wins, exposed kong over chow, pong over chow, chow only by the next seat, concealed kong, added kong, robbing an added kong, and all-pass continuation.

- [ ] **Step 2: Write failing legal-action tests**

```go
func TestMCRLegalActionFixtures(t *testing.T) {
	for _, fixture := range loadMCRLegalActionFixtures(t) {
		round := mcrRoundFromFixture(t, fixture.Initial)
		got := round.rules.LegalActions(round, fixture.PlayerID)
		assertLegalActions(t, got, fixture.Expected)
	}
}
```

- [ ] **Step 3: Verify RED**

Run `go test ./internal/game -run "MCRLegal|MCRPriority|MCRKong" -count=1`.

- [ ] **Step 4: Implement MCRRuleSet**

`LegalActions` exposes win only when `ScoreMCR(...).MeetsMinimum`. Add typed kong selectors for concealed, exposed, and added kong; added kong creates a robbing window before tiles move. Resolve claim groups by standard priority and turn distance. Automatic flower replacement occurs before discard actions are exposed.

- [ ] **Step 5: Verify and commit**

Run focused tests, bot legal-action closure, and snapshot/reconnect tests with pending added-kong claims.

```powershell
git add internal/game/mcr_rules.go internal/game/mcr_rules_test.go internal/game/claim_state.go internal/game/snapshot.go testdata/rules/mcr/legal_actions.json
git commit -m "feat: enforce MCR actions and claim priority"
```

Step review: no MCR command path falls back to compatibility win checks or implicit claim ordering.

## Task 8: Settle Hands And Advance A Sixteen-Hand Match

**Files:**
- Create: `internal/game/mcr_settlement.go`
- Create: `internal/game/mcr_settlement_test.go`
- Modify: `internal/game/match.go`
- Modify: `internal/game/match_test.go`
- Create: `testdata/rules/mcr/settlement.json`

- [ ] **Step 1: Write settlement fixtures**

Cover self-draw, each discard-source seat, flower additions, minimum boundary, wall exhaustion, dealer rotation, fourth-to-fifth hand transition, and the final sixteenth hand.

- [ ] **Step 2: Write failing zero-sum tests**

```go
func TestMCRSettlementGoldenFixtures(t *testing.T) {
	for _, fixture := range loadMCRSettlementFixtures(t) {
		got := SettleMCR(fixture.Score, fixture.Winner, fixture.Discarder, fixture.WinType)
		if got.Deltas != fixture.ExpectedDeltas { t.Fatalf("%s deltas=%v", fixture.ID, got.Deltas) }
		if sum4(got.Deltas) != 0 { t.Fatalf("%s is not zero-sum", fixture.ID) }
	}
}
```

- [ ] **Step 3: Verify RED**

Run `go test ./internal/game -run "MCRSettlement|MCRMatch" -count=1`.

- [ ] **Step 4: Implement settlement and progression**

Apply the payment formulas in the source contract, append settlement to match history, update points, rotate dealer, increment round/hand indexes, create the next fixed-seed-derived round, and mark complete after hand 16. A draw advances without transfer. Never reset accumulated points on round creation.

- [ ] **Step 5: Verify and commit**

Run settlement fixtures, deterministic 16-hand simulations, and zero-sum property tests.

```powershell
git add internal/game/mcr_settlement.go internal/game/mcr_settlement_test.go internal/game/match.go internal/game/match_test.go testdata/rules/mcr/settlement.json
git commit -m "feat: settle complete MCR matches"
```

Step review: a complete MCR match now advances and settles through shared Match APIs with no simplified score fallback.

## Task 9: Preserve MCR State In Snapshots, Reconnect, And Replay

**Files:**
- Modify: `internal/game/snapshot.go`
- Modify: `internal/game/snapshot_test.go`
- Modify: `internal/game/replay.go`
- Modify: `internal/game/replay_test.go`
- Modify: `internal/online/server.go`
- Modify: `internal/online/server_test.go`

- [ ] **Step 1: Write failing integration tests**

Assert private snapshots retain flower counts, seat/prevalent winds, hand number, points, pending added-kong/robbing state, and the viewer's legal actions. Assert replay contains the final MCR breakdown and every hand settlement. Reconnect must reproduce the same private snapshot byte-for-byte after canonical JSON encoding.

- [ ] **Step 2: Verify RED**

Run `go test ./internal/game ./internal/online -run "MCR.*Snapshot|MCR.*Reconnect|MCR.*Replay" -count=1`.

- [ ] **Step 3: Wire MCR room construction and metadata**

For `ModeMCR`, server room creation uses `NewMCRRuleSet(config.MCR)`; compatibility remains `NewCompatibilityRuleSet`. Add typed MCR round/match fields to snapshots rather than unstructured maps. Apply the existing recipient redaction after those fields are copied.

- [ ] **Step 4: Verify and commit**

Run focused tests and `go test ./internal/online -count=20`.

```powershell
git add internal/game/snapshot.go internal/game/snapshot_test.go internal/game/replay.go internal/game/replay_test.go internal/online/server.go internal/online/server_test.go
git commit -m "feat: synchronize complete MCR state"
```

Step review: live and resumed MCR play share one authoritative state, while concealed information remains private and completed matches remain auditable.

## Task 10: Phase 12B Acceptance And Review

**Files:**
- Modify: `docs/rules/conformance.md`
- Modify: `docs/workflow.md`

- [ ] **Step 1: Run catalog and fixture validation**

```powershell
Get-ChildItem testdata/rules/mcr -Recurse -Filter *.json | ForEach-Object { Get-Content -Raw $_.FullName | ConvertFrom-Json | Out-Null }
go test ./internal/game -run "MCRFanCatalog|EveryMCRFan|MCRScoring|MCRSettlement" -count=20
```

- [ ] **Step 2: Run static, race, and full tests**

```powershell
$files = Get-ChildItem internal,cmd -Recurse -Filter *.go | ForEach-Object FullName
gofmt -w $files
git diff --check
go vet ./...
go test ./... -count=1
go test -race ./... -count=1
go build ./cmd/mahjong ./cmd/server ./cmd/client
```

- [ ] **Step 3: Run generated invariants**

For at least 1,000 fixed seeds, assert 144-tile conservation, no flower in concealed hands after replacement, no negative counts, every returned action is accepted from the same state, every settlement is zero-sum, and a 16-hand match terminates.

- [ ] **Step 4: Run real local and WebSocket smoke matches**

Use an MCR fixture sequence to complete one local match and one LAN match, including a flower replacement, one chow/pong/kong, a rejected sub-eight-point win, an accepted win, settlement, disconnect, and reconnect. Verify live clients never print concealed opponent tiles or seeds.

- [ ] **Step 5: Update reviews and commit**

Mark 12B complete and 12C not started in `docs/rules/conformance.md`. Append exact command evidence and remaining Riichi risk to `docs/workflow.md`.

```powershell
git add docs/rules/conformance.md docs/workflow.md
git commit -m "test: record phase 12b acceptance"
```

Phase review: compare every 12B fixture and invariant against the total selectable-rules goal. Begin the separate 12C plan only if MCR rooms use `MCRRuleSet`, all 81 fan IDs have positive coverage, no compatibility scoring path is reachable, and fresh full/race tests pass.

## Plan Self-Review

- Spec coverage: wall/flowers, legal actions and priority, 81 fan detection, exclusions/counting, eight-point minimum, settlement, 16-hand totals, fixtures, invariants, snapshots, reconnect, replay, and acceptance each have a dedicated task.
- Scope: Riichi mechanics, final mode-selection screens, Phase 13 table redesign, and Phase 14 replay persistence are excluded.
- Type consistency: `FanID`, `FanMatch`, `MCRScoreContext`, `MCRScoreBreakdown`, `MCRSettlement`, `BuildMCRWall`, `MCRDecompose`, `ScoreMCR`, `MCRRuleSet`, and `SettleMCR` are introduced before later use.
- Dependency decision: no third-party scoring engine is added; the MIT TypeScript project is an architecture reference only.
- Placeholder scan command:

```powershell
$patterns = @("TB"+"D", "TO"+"DO", "implement"+" later", "fill in"+" details", "appropriate"+" error", "Similar"+" to")
foreach ($pattern in $patterns) { rg -n $pattern docs/superpowers/plans/2026-06-23-phase-12b-complete-mcr.md }
```

Expected: no matches.
