# Dual Rules, Tactical Table, And Replay Design

## Summary

Terminal Mahjong will evolve in three ordered stages:

1. Phase 12 adds selectable, complete Chinese Official Mahjong and four-player Riichi rule modes.
2. Phase 13 replaces the current table with a wide competitive layout plus a persistent tactical rail.
3. Phase 14 adds automatic, full-information post-game replays and a TUI replay browser.

Every implementation step must review its current phase goal. Every completed phase must review the total goal: a terminal-first Mahjong client with trustworthy rules, LAN play, reconnect, readable table state, and auditable completed matches.

## Confirmed Decisions

- The two modes change complete game behavior, not only scoring labels.
- Shared match, networking, snapshot, bot, TUI, and replay infrastructure remains mode-neutral.
- Chinese Official Mahjong uses an eight-point minimum, its complete fan table, and the standard's winner-priority behavior.
- Riichi is four-player Japanese Mahjong with riichi, furiten, yaku, fu, dora, exhaustive and abortive draws, dealer continuation, honba, riichi sticks, and point settlement.
- Red fives are a room option for Riichi.
- The primary table design is the approved wide competitive table with a persistent tactical rail.
- Narrow rendering is only a compatibility fallback, not a separate visual direction.
- Online replays reveal all hands and draws only after the game ends.

## Rules Baselines

Phase 12 begins by recording a conformance matrix from these baselines before implementing individual rules:

- Chinese Official mode: `GB/T 34708-2017 麻将竞赛规则`, including the eight-point minimum.
- Riichi mode: European Mahjong Association Riichi Rules 2016 as the stable project baseline, with project options recorded explicitly in `RiichiConfig` rather than inherited from an unnamed online platform.

Project options are explicit. `MCRConfig` initially has no platform-derived variants. `RiichiConfig` defaults to an East-South match, 25,000 starting points, open tanyao enabled, and three red fives enabled; red fives can be disabled at room creation. Every additional variant requires a named configuration field and fixture. Implementation must not silently combine rules from different platforms.

## Planning Boundary

This document is the approved master design, not one giant implementation batch. Detailed plans are produced and executed in this order:

1. Phase 12A foundation;
2. Phase 12B Chinese Official rules;
3. Phase 12C Riichi rules;
4. Phase 12D integration;
5. Phase 12E acceptance;
6. Phase 13 table design;
7. Phase 14 replay system.

Each item receives its own plan, tests, commits, step reviews, and phase review. Work does not begin on a later item while the current item has unmet acceptance criteria.

## Architecture

### Shared Match Coordinator

The current `Game` combines round state, simplified rules, and presentation-era assumptions. It will be separated into:

- `Match`: mode, players, cumulative points, round sequence, dealer, completion state, and mode configuration.
- `Round`: wall, dead wall when applicable, hands, melds, discards, flowers, current phase, pending responses, and accepted action history.
- `RuleSet`: creates mode-specific round state, derives legal actions, validates actions, evaluates wins, resolves draws, and settles points.
- `MatchSnapshot`: mode-neutral public state plus a typed mode-specific status payload.
- `PlayerSnapshot`: recipient-filtered live view. Opponent concealed hands are never serialized during an online round.

The match coordinator owns command ordering, response priority, bot scheduling, reconnect, and event publication. A rules package owns only the rules needed to determine legal state transitions and settlement.

### Rule Packages

Rule code is separated by ownership:

- `internal/rules/common`: tile counts, hand decomposition, wait calculation, shared action primitives.
- `internal/rules/mcr`: flowers, legal win threshold, fan detection, exclusions, combination rules, winner priority, and settlement.
- `internal/rules/riichi`: wall/dead-wall layout, dora indicators, riichi, furiten, kan variants, yaku, fu, limit hands, draw rules, and point transfer.

Rules return structured results such as `LegalAction`, `WinEvaluation`, `ScoreBreakdown`, and `Settlement`. The TUI must not detect yaku, fan, furiten, or legal calls itself.

### Commands And Events

Existing discard, win, kong, pong, chow, and pass commands remain recognizable. The shared command model gains rule actions such as:

- `declare_riichi`
- `added_kong`
- `open_kong`
- `flower_replacement`

Commands are validated against a generated legal-action list. Accepted commands produce typed events. Rejected commands do not mutate state. Replay recording and bots consume the same accepted-command and event stream.

## Phase 12: Complete Selectable Rules

### Phase Goal

A player can choose Chinese Official or Riichi at game creation, finish a match under that mode's complete gameplay and settlement rules, play it locally or over LAN, reconnect without losing rule state, and face bots that only choose legal actions.

### Phase 12A: Dual-Mode Foundation

1. Write the rule conformance matrix and golden fixture format.
2. Introduce `RuleMode`, `Match`, `Round`, `RuleSet`, structured legal actions, and recipient-filtered snapshots.
3. Adapt the current simplified behavior behind a temporary compatibility rule set while migration proceeds.
4. Add mode and configuration to room creation, reconnect tokens, snapshots, and replay metadata.
5. Preserve the current single-game entry point until both complete rule sets pass acceptance.

Step review: each step must make rule behavior more explicit without changing unrelated TUI behavior.

Phase review: mode selection and private snapshots must work before either large scoring table is added.

### Phase 12B: Chinese Official Mahjong

1. Build the 144-tile wall and flower replacement flow.
2. Implement legal actions, response ordering, and winner priority exactly as recorded in the conformance matrix.
3. Implement fan detection in value bands with exclusions and non-combinable fan rules.
4. Enforce the eight-point minimum from structured fan results.
5. Implement settlement and match totals.
6. Add published-example fixtures and generated invariant tests.

Step review: each rule step must cite its fixture and return a readable breakdown.

Phase review: a full Chinese Official match must complete without simplified-rule fallbacks.

### Phase 12C: Four-Player Riichi

1. Implement 136-tile wall, dead wall, rinshan draws, dora, and configurable red fives.
2. Implement chi, pon, open/closed/added kan, call priority, and ippatsu cancellation.
3. Implement tenpai, riichi declaration, permanent/temporary/riichi furiten, ron, and tsumo legality.
4. Implement yaku and yakuman evaluation, openness changes, fu, limits, and point rounding.
5. Implement exhaustive and abortive draws, noten payments, honba, riichi sticks, dealer continuation, and match end.
6. Add golden hands and settlement fixtures for dealer/non-dealer ron and tsumo.

Step review: every rule primitive must be testable without TUI or WebSocket setup.

Phase review: a full Riichi match must settle correctly with no Chinese Official branches in the package.

### Phase 12D: Client, Server, And Bots

1. Add local mode selection and mode-specific options to the start flow.
2. Add room mode/configuration to create, browse, join, ready, and reconnect messages.
3. Render only legal actions supplied by the current rules snapshot.
4. Extend bots to rank legal actions instead of constructing commands from implicit assumptions.
5. Localize rule names, score breakdowns, draw reasons, and errors in Chinese and English.

Step review: each integration step must use the shared interfaces and avoid mode checks in layout code.

Phase review: local and online behavior must agree for the same fixed fixture and command sequence.

### Phase 12E: Acceptance

- Golden fixtures for all implemented fan/yaku and settlement boundaries.
- Property tests for tile conservation, non-negative tile counts, legal command closure, and zero-sum Riichi transfers.
- Privacy tests proving live online snapshots hide concealed opponent tiles.
- Reconnect tests for riichi state, furiten, flowers, dead wall, dealer, honba, and pending calls.
- Complete local and LAN smoke matches in both modes.
- `go vet ./...`, full tests, race tests, builds, and fixed-seed replay checks.

Phase review: compare evidence against the total dual-mode terminal Mahjong goal before beginning visual restructuring.

## Phase 13: Wide Competitive Table And Tactical Rail

### Phase Goal

Both rule modes present a clear, client-like terminal table where the spatial relationship of four players, discard rivers, current action, hand, and tactical information can be understood without reading a scrolling log.

### Approved Wide Layout

- Header: mode badge, round, dealer/honba or Chinese Official match status, wall count, and key rule indicators.
- Table: four fixed seats around a central discard-river area.
- Seat panels: wind/seat, points, hand count, melds, flowers where applicable, riichi marker, and active-turn emphasis.
- Center: organized discard rivers, latest discard, pending response, dora indicators, and round state.
- Bottom: one-row bare Unicode hand, selected tile, melds, and contextual legal actions.
- Right tactical rail: shanten, effective tiles, improvements, mode-specific legal status, and recent typed events.

The tactical rail is read-only assistance. It must derive information from snapshots and rule analysis APIs, never mutate game state.

### Phase 13 Workflow

1. **13A Table skeleton:** split the renderer into header, seats, center table, hand, action prompt, and tactical rail components.
2. **13B Table state:** add fixed discard rivers, seat winds/status, meld/flower/riichi markers, and clear current-action emphasis.
3. **13C Tactical rail:** add shanten, effective tiles, improvement tiles, mode status, and bounded recent events.
4. **13D Compatibility fallback:** at medium widths move the rail below the table; below the compact threshold hide it behind `Tab`. The wide layout remains the design target.
5. **13E Interaction and QA:** preserve keyboard and mouse hitboxes, contextual commands, language consistency, ANSI width budgets, and screenshot comparison at wide and compact sizes.

Step review: each step must improve table comprehension without weakening controls or rule correctness.

Phase review: compare wide screenshots in both modes with the approved A-plus-C visual direction and the total terminal-client goal.

## Phase 14: Full Post-Game Replay

### Phase Goal

Completed local and online matches are saved reliably and can be replayed step by step in the same table UI, with full hidden information revealed only after completion.

### Replay Format

`ReplayFile` version 2 contains:

- schema version and application version;
- mode and complete rule configuration;
- shuffle proof, participants, and initial match state;
- accepted commands and typed events;
- compact state frames required for stable viewing;
- round settlements and final standings;
- completion flag and checksum.

State frames make old replays viewable after internal rule implementations evolve. Commands and proofs retain audit value. Replays never contain reconnect tokens or network addresses.

### Persistence And Distribution

- Save to `replays/<timestamp>-<mode>-<id>.json` through a temporary file and atomic rename.
- Do not save incomplete or checksum-invalid files as valid replays.
- The online server records authoritative full state. At match end it sends a replay-ready message and the completed replay payload to each connected player.
- Clients save the received replay locally. Reconnect within the retention window can request it again.

### Replay TUI

- Start menu entry opens a replay browser sorted newest first.
- The viewer reuses the Phase 13 table renderer in a read-only mode.
- Left/right steps between frames; Home/End jumps; Space toggles timed playback; Tab switches table and event/settlement detail.
- The viewer shows mode, round, frame count, playback state, full hands, score breakdown, and final standings.

### Phase 14 Workflow

1. **14A Schema:** add versioned replay types, validation, checksum, and compatibility errors.
2. **14B Recorder:** capture authoritative frames and save atomically for local matches.
3. **14C Online privacy and delivery:** filter live snapshots, reveal full completed replay, and support reconnect retrieval.
4. **14D Browser and viewer:** list files, load safely, navigate frames, and reuse the table renderer read-only.
5. **14E Acceptance:** test deterministic fixtures, corruption handling, unsupported versions, both rule modes, online completeness, privacy, and screenshots.

Step review: each step must improve replay reliability or usability without becoming a second game engine.

Phase review: compare saved and replayed results with the original authoritative match and the total auditable-client goal.

## Error Handling

- Unknown modes and unsupported options fail before a room or local match starts.
- Illegal commands return stable error codes plus localized presentation text.
- Rule evaluation never partially mutates match state; settlement applies only after validation succeeds.
- Unsupported replay versions remain untouched and show a clear compatibility message.
- Corrupt replay files are skipped in the browser and reported without preventing other files from loading.
- Online replay delivery failure does not invalidate the completed match; reconnect can retry retrieval.

## Testing Strategy

- Tests are fixture-first and production code follows red-green-refactor.
- Rule packages use table-driven golden examples and mutation-invariant tests.
- Shared coordinator tests run the same command scenarios against both rule sets where behavior should agree.
- Online tests compare recipient-filtered snapshots and authoritative server state.
- TUI tests verify visible width, contextual actions, mouse hitboxes, bilingual labels, and rule-specific fields.
- Screenshot QA covers wide Chinese Official, wide Riichi, and compact fallback views.
- Replay tests compare every restored frame, settlement, and final standing against the recorded source.

## Git And Review Workflow

- Each subphase gets its own detailed implementation plan.
- Each behavior starts with a failing test and is committed as a reviewable unit.
- After every step, record: phase goal, completed work, evidence, and next step.
- After every subphase, record: total goal, achieved outcome, evidence, and remaining risk.
- Do not begin the next major phase until the current phase passes its acceptance checklist.
- Keep the current branch until the user explicitly chooses merge or PR handling.

## Non-Goals

- Three-player Riichi is not included.
- Regional Chinese variants are not included.
- Public internet matchmaking, accounts, rankings, Redis, and databases are not included.
- Replay editing, branching, and video export are not included.
- The project remains a terminal application and does not become a desktop GUI.
