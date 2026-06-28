# Phase 12E Dual-Mode Acceptance Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Prove that Chinese Official and four-player Riichi modes satisfy the combined deterministic, legal-action, privacy, reconnect, replay, build, and race gates before Phase 13 visual restructuring.

**Architecture:** Add acceptance-only tests over the existing `RuleSet`, `Match`, recipient-private snapshot, WebSocket room, and replay interfaces. Do not add rule mechanics or client features in this phase; any failure must be fixed at the smallest owning boundary and covered by the failing acceptance test.

**Tech Stack:** Go standard `testing`, Gorilla WebSocket test server, JSON canonical encoding, PowerShell fixture validation, Go race detector.

---

## Acceptance Contract

- Both MCR and Riichi create deterministic state from a fixed seed and validated `RuleConfig`.
- Every legal action returned by the current snapshot is accepted against that same state.
- Live snapshots hide shuffle seeds and opponent concealed hands in both modes.
- Riichi additionally hides dead-wall identities and ura indicators before round end.
- MCR preserves public flowers while hiding replacement draws from other viewers.
- Reconnect returns byte-identical canonical private match JSON.
- Replay from the same fixed seed and command sequence is byte-identical.
- Existing 1,000-seed MCR/Riichi invariants, full tests, race tests, command builds, and online stability loops remain green.

## File Map

- Create `internal/game/dual_mode_acceptance_test.go`: fixed-seed determinism, replay equality, legal-action closure, tile conservation routing, and zero-sum settlement checks across both modes.
- Create `internal/online/dual_mode_acceptance_test.go`: recipient-private create/ready/action/reconnect matrix for MCR and Riichi.
- Modify `docs/rules/conformance.md`: mark Phase 12E complete and record the final rule-mode acceptance evidence.
- Modify `docs/workflow.md`: add Phase 12E step reviews and total-goal review.
- Modify this plan: mark every completed checkbox before the acceptance commit.

## Task 1: Fixed-Seed Dual-Mode Determinism

**Files:**
- Create: `internal/game/dual_mode_acceptance_test.go`

- [ ] **Step 1: Write fixed-seed acceptance tests**

Create a table with MCR and Riichi rules. For each mode, construct two matches with seed `120012`, compare canonical JSON snapshots and replay logs, and verify each initial legal action kind is accepted on a fresh copy:

```go
func TestDualModeFixedSeedReplayAndLegalActionClosure(t *testing.T) {
	for _, fixture := range dualModeFixtures() {
		first := mustAcceptanceMatch(t, 120012, fixture.rules())
		second := mustAcceptanceMatch(t, 120012, fixture.rules())
		first.EnsureCurrentTurnDraw()
		second.EnsureCurrentTurnDraw()
		assertCanonicalJSONEqual(t, first.Snapshot(), second.Snapshot())
		assertCanonicalJSONEqual(t, first.ReplayLog(), second.ReplayLog())
		assertFreshLegalActionsAccepted(t, 120012, fixture.rules, first.Round.Snapshot().LegalActions)
	}
}
```

- [ ] **Step 2: Add settlement conservation checks**

Call representative MCR discard/self-draw settlements and Riichi ron/tsumo/exhaustive-draw settlements. Sum every `[4]int` delta array and require zero.

- [ ] **Step 3: Verify**

Run:

```powershell
go test ./internal/game -run "DualMode|MCRGenerated|RiichiGenerated" -count=1
```

Expected: all deterministic and generated invariant tests pass.

- [ ] **Step 4: Commit**

```powershell
git add internal/game/dual_mode_acceptance_test.go
git commit -m "test: verify dual-mode deterministic rules"
```

Step review: both complete rulesets satisfy the same fixed-seed and legal-command contracts without sharing scoring implementations.

## Task 2: Dual-Mode Privacy And Reconnect Matrix

**Files:**
- Create: `internal/online/dual_mode_acceptance_test.go`

- [ ] **Step 1: Write table-driven WebSocket acceptance**

For MCR and Riichi:

1. create a room with explicit mode/config;
2. join enough clients to observe two private views;
3. ready the room;
4. assert each client sees only its own hand and zero shuffle seed;
5. execute the current player's first legal discard;
6. disconnect and reconnect that player;
7. compare canonical private `MatchSnapshot` JSON before and after reconnect.

Use the existing `startTestServer`, `dialTestClient`, `readUntil`, `assertPrivateSnapshot`, and `firstDiscardAction` helpers from `internal/online/server_test.go`.

- [ ] **Step 2: Add mode-specific privacy assertions**

For Riichi, require `DeadWallCount == 14`, non-empty public dora, and empty live `UraIndicators`. For MCR, require public flower arrays to remain present while opponent hands remain nil.

- [ ] **Step 3: Verify stability**

Run:

```powershell
go test ./internal/online -run "DualMode.*Privacy|DualMode.*Reconnect" -count=20
```

Expected: both mode subtests pass 20 consecutive runs.

- [ ] **Step 4: Commit**

```powershell
git add internal/online/dual_mode_acceptance_test.go
git commit -m "test: verify dual-mode privacy and reconnect"
```

Step review: the same server/session protocol preserves mode-specific public state while enforcing concealed-information privacy in both modes.

## Task 3: Full Acceptance Gate And Review

**Files:**
- Modify: `docs/rules/conformance.md`
- Modify: `docs/workflow.md`
- Modify: `docs/superpowers/plans/2026-06-28-phase-12e-dual-mode-acceptance.md`

- [ ] **Step 1: Validate every rule fixture**

Run:

```powershell
Get-ChildItem testdata/rules -Recurse -Filter *.json | ForEach-Object { Get-Content -Raw $_.FullName | ConvertFrom-Json | Out-Null }
```

Expected: exit code 0 with no JSON parse errors.

- [ ] **Step 2: Run focused rule and online gates**

Run:

```powershell
go test ./internal/game -run "MCRFanCatalog|EveryMCRFan|RiichiCatalog|EveryRiichiYaku|Scoring|Settlement|Generated|DualMode" -count=1
go test ./internal/online -run "Private|Reconnect|Riichi|MCR|DualMode" -count=20
go test ./internal/tui ./cmd/client -count=20
```

Expected: every package exits 0.

- [ ] **Step 3: Run static, full, race, and build gates**

Run:

```powershell
$files = Get-ChildItem internal,cmd -Recurse -Filter *.go | ForEach-Object FullName
if (gofmt -l $files) { throw "unformatted Go files" }
git diff --check
go vet ./...
go test ./... -count=1
go test -race ./... -count=1
go build ./cmd/mahjong ./cmd/server ./cmd/client
```

Expected: no formatting output, no diff errors, and every Go command exits 0.

- [ ] **Step 4: Update stage and total-goal reviews**

Mark 12E complete in `docs/rules/conformance.md`. In `docs/workflow.md`, record the exact commands, the two-mode privacy/reconnect matrix, fixed-seed replay equality, and the decision that Phase 13 may begin.

- [ ] **Step 5: Mark this plan complete and commit**

Change every `- [ ]` in this file to `- [x]`, then run `git diff --check`.

```powershell
git add docs/rules/conformance.md docs/workflow.md docs/superpowers/plans/2026-06-28-phase-12e-dual-mode-acceptance.md
git commit -m "test: record phase 12 dual-mode acceptance"
```

Phase review: Phase 12 is complete only when both rulesets, local/online clients, bots, privacy, reconnect, replay, full tests, race tests, and command builds pass together from a clean worktree.

## Plan Self-Review

- Spec coverage: golden fixtures, generated properties, privacy, reconnect, local/online smoke paths, replay determinism, full tests, race tests, and builds are all represented.
- Placeholder scan: no TODO/TBD or unspecified implementation steps remain.
- Type consistency: the plan uses existing `RuleSet`, `Match`, `GameSnapshot`, `MatchSnapshot`, `ReplayLog`, `LegalAction`, and protocol message types.
- Scope: no new rules, visual redesign, replay persistence, database, or external AI work is included.
