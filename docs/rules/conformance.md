# Rule Conformance Matrix

This matrix freezes the rule sources and project defaults before complete rule implementation begins. Every golden fixture must name a source section or published example. Platform-specific behavior is not inherited implicitly.

## Baselines

| Area | Chinese Official (`mcr`) | Four-player Riichi (`riichi`) |
| --- | --- | --- |
| Primary source | `GB/T 34708-2017 麻将竞赛规则` | European Mahjong Association, `Riichi Rules 2016` |
| Wall | 144 tiles including 8 flowers | 136 tiles with a 14-tile dead wall |
| Win requirement | Standard-defined winning shape and at least 8 points | At least one yaku; dora alone is not a yaku |
| Response/winner priority | Follow the standard's documented order | Follow EMA call and winning priority |
| Match length | Standard competition round sequence | East-South match |
| Initial points | Match total starts at 0 | 25,000 per player |
| Open tanyao | Not applicable | Enabled |
| Red fives | Not applicable | Three by default; room option allows zero |
| Hidden live information | Opponent concealed tiles, private draw tiles, and shuffle seed are hidden | Opponent concealed tiles, private draw tiles, and shuffle seed are hidden |
| Post-game replay | Complete hands and draws are revealed | Complete hands and draws are revealed |

## Configuration Contract

- `MCRConfig.MinimumPoints` is fixed at `8` for the standard mode.
- `RiichiConfig.MatchLength` is `east_south` in the first complete implementation.
- `RiichiConfig.StartingPoints` defaults to `25000`.
- `RiichiConfig.OpenTanyao` defaults to `true`.
- `RiichiConfig.RedFives` accepts `0` or `3`, defaulting to `3`.
- A new variant requires a named configuration field, source note, validation rule, and at least one golden fixture.

## Riichi Source Contract

- The normative EMA PDF, retrieved on 2026-06-24, has SHA-256 `1BCFE2A0B50FC89DA10CD24C89225D5D8EC57313A790B02388C699D394ED6530`.
- Exact page/section decisions and the complete yaku/bonus catalog are recorded in `docs/rules/riichi-source-notes.md` and `testdata/rules/riichi/catalog.json`.
- The approved project specification intentionally differs from EMA 2016 on starting points (25,000), optional red fives (0 or 3), and five fixed abortive draws. These are labeled project overrides in every affected fixture.
- EMA behavior retained by the project includes multiple ron, no 4-han-30-fu kiriage mangan, 13+ normal han as sanbaiman, single non-cumulative yakuman, mangan renhou, no nagashi mangan, no bankruptcy end, and no agari-yame.

## Chinese Official Project Decisions

- A match is four rounds of four hands. The dealer advances after every completed hand; there are no dealer repeats.
- The 144-tile wall contains the 136 suited/honor tiles plus eight unique flowers.
- Flowers are exposed and replaced automatically from the back of the live wall until a non-flower tile is drawn.
- Flower points are added after the hand reaches the eight-point non-flower minimum; flowers cannot make a sub-eight-point hand legal.
- Claim priority is win, exposed kong/pong, then chow by the next seat. Among simultaneous winning claims, the nearest eligible player after the discarder wins.
- On self-draw, each opponent pays `8 + total fan`. On discard win, the discarder pays `8 + total fan` and each other opponent pays `8`.
- Drawn hands advance the dealer without a point transfer.

## Fixture Rules

- `source` identifies a standard section, official example, or published governing-body example.
- `initial` contains all state needed to reproduce the case without random dealing.
- `commands` contains only player-visible commands accepted by the shared match coordinator.
- `expected` contains legal actions, events, score breakdown, settlement, or terminal state as appropriate.
- Fixtures must not use prose-only expected results.

## Implementation Status

| Subphase | Status |
| --- | --- |
| 12A shared match/rule/privacy foundation | Complete |
| 12B complete Chinese Official rules | Complete |
| 12C complete Riichi rules | Complete |
| 12D client/server/bot integration | Complete |
| 12E dual-mode acceptance | Complete |

### Phase 12B Acceptance

- The MCR room path uses `MCRRuleSet` with a 144-tile wall, automatic flower replacement, complete legal actions, and recipient-private snapshots.
- All 81 catalog IDs have positive and near-miss coverage; scoring applies catalog exclusions, group-use counting, Chicken Hand fallback, and the eight-point non-flower minimum.
- Matches settle zero-sum, rotate the dealer after every hand, retain all settlement history, and complete after 16 hands.
- Replay schema v3 stores the typed MCR score and settlement history without using compatibility scoring.
- Acceptance on 2026-06-24 passed catalog/fixture validation, 1,000 fixed-seed invariants, full tests, race tests, vet, all command builds, and online tests repeated 20 times.

### Phase 12C Acceptance

- Riichi room creation now uses `RiichiRuleSet` with a 136-tile wall, 14-tile dead wall, public dora indicators, hidden ura indicators during play, honba, riichi sticks, declarations, and recipient-private furiten state.
- Red fives, dora/ura/kan-dora, rinshan draws, riichi, ippatsu cancellation, all furiten states, chi/pon/kan/ron windows, yaku detection, fu/han/limit scoring, exhaustive draw payments, and East-South settlement are covered by focused fixtures and generated invariants.
- Bots consume authoritative `LegalActions`; Riichi command validation no longer depends on clients reconstructing hidden rule state.
- Replay schema v3 stores typed Riichi scores, dora, post-game ura, and Riichi settlement history without using compatibility scoring.
- Acceptance on 2026-06-25 passed Riichi fixture JSON parsing, catalog/yaku/scoring/settlement tests repeated 20 times, 1,000 fixed-seed invariants for both red-five modes, WebSocket ready/discard/reconnect smoke, full tests, race tests, vet, all command builds, and online tests repeated 20 times.

### Phase 12D Acceptance

- The TUI start menu selects Riichi, Chinese Official, or compatibility rules and exposes the Riichi red-five option without adding rule logic to layout code.
- Local TUI startup and online room startup both use the shared `Match` coordinator and the same validated `RuleConfig`.
- The CLI creates typed MCR/Riichi rooms with `-mode` and `-red-fives`, and room listings expose mode and option metadata.
- Win, kong, riichi, and claim presentation now follows authoritative `LegalActions`; the TUI no longer reconstructs win or kong legality from concealed hand shape.
- Acceptance on 2026-06-28 passed local/online rule-mode parity tests, TUI/CLI/online tests repeated 20 times, full tests, race tests, and all command builds.

### Phase 12E Acceptance

- Fixed-seed MCR and Riichi matches produce canonical-equal snapshots and replay logs, and every returned initial legal action is accepted against a fresh copy of the same state.
- Representative MCR ron/tsumo and Riichi ron/tsumo/exhaustive-draw transfers are zero-sum.
- The dual-mode WebSocket matrix verifies recipient-private hands, hidden shuffle seeds, canonical reconnect JSON, public MCR flowers, public Riichi dora/dead-wall counts, and hidden live ura indicators.
- Acceptance on 2026-06-28 passed all rule JSON parsing, focused catalog/yaku/fan/scoring/settlement/generated tests, dual-mode online privacy/reconnect tests repeated 20 times, TUI/CLI tests repeated 20 times, formatting, diff, vet, full tests, race tests, and all command builds.
- Phase 12 selectable complete rules are accepted; Phase 13 may begin without reopening rule mechanics.
