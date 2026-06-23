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
| 12B complete Chinese Official rules | In progress |
| 12C complete Riichi rules | Not started |
| 12D client/server/bot integration | Not started |
| 12E dual-mode acceptance | Not started |
