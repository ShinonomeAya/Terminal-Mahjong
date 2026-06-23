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

## Fixture Rules

- `source` identifies a standard section, official example, or published governing-body example.
- `initial` contains all state needed to reproduce the case without random dealing.
- `commands` contains only player-visible commands accepted by the shared match coordinator.
- `expected` contains legal actions, events, score breakdown, settlement, or terminal state as appropriate.
- Fixtures must not use prose-only expected results.

## Implementation Status

| Subphase | Status |
| --- | --- |
| 12A shared match/rule/privacy foundation | In progress |
| 12B complete Chinese Official rules | Not started |
| 12C complete Riichi rules | Not started |
| 12D client/server/bot integration | Not started |
| 12E dual-mode acceptance | Not started |

