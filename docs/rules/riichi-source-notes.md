# EMA Riichi 2016 Source Notes

## Source Identity

- Normative document: European Mahjong Association, *Riichi Rules 2016*, April 2016.
- Official URL: `https://mahjong-europe.org/portal/images/docs/Riichi-rules-2016-EN.pdf`.
- Retrieved: 2026-06-24.
- SHA-256: `1BCFE2A0B50FC89DA10CD24C89225D5D8EC57313A790B02388C699D394ED6530`.
- The PDF is 32 pages. It is not vendored into this repository; fixtures cite its printed page and section numbers.

## Baseline Decisions

| Area | Project value | Source |
| --- | --- | --- |
| Players and wall | Four players, 136 tiles, no flowers | EMA pp. 6-7, sections 1 and 2 |
| Dead wall | 14 tiles remain outside normal draws; replacement tiles and indicators come from the dead wall | EMA p. 8, sections 2.6-2.7 |
| Calls | Ron has priority; pon/kan has priority over chi; chi is from the preceding player's discard | EMA pp. 11-13, section 3.3 |
| Kuikae | Swap-calling is forbidden | EMA p. 5 revision notes and p. 12, section 3.3.3 |
| Multiple winners | More than one ron winner is allowed; the discarder settles each hand | EMA p. 5 revision notes and p. 18, section 4.1 |
| Fourth kan | Play continues after the fourth kan; a fifth kan is forbidden | EMA p. 13, section 3.3.10 |
| Riichi | Closed tenpai, 1,000-point stake, and at least four live-wall tiles | EMA pp. 14-15, section 3.3.14 |
| Post-riichi kan | Ankan only when the wait shape is unchanged and the three tiles have only a pung interpretation | EMA p. 14, section 3.3.14 |
| Ippatsu | Ends on chi, pon, open/added/closed kan | EMA p. 15, section 3.3.14 |
| Furiten | Any own discard in the complete wait forbids ron; a passed completion causes temporary furiten; after riichi it lasts for the hand; tsumo remains legal | EMA pp. 15-16, section 3.4.5 |
| Exhaustive draw | 3,000-point noten transfer; riichi players must reveal tenpai | EMA p. 15, section 3.4.2 |
| Dealer continuation | East repeats after an East win or East tenpai exhaustive draw | EMA p. 16, section 3.4.11 |
| Match length | East-South; no agari-yame and no bankruptcy termination | EMA pp. 16-17, sections 3.5-3.6 |
| Honba | 300 on ron; 100 from each payer on tsumo; increases after East win and exhaustive draw | EMA p. 16, section 3.4.10 |
| Riichi sticks | Winner collects; with multiple winners the nearest winner after the discarder collects unclaimed sticks | EMA pp. 14 and 18, sections 3.3.14 and 4.1 |
| Fu | 20 base/open/tsumo, 30 closed ron, seven pairs fixed 25, set/pair/wait additions, round up to 10 | EMA pp. 18-19, section 4.1.1 |
| Limits | 5 mangan, 6-7 haneman, 8-10 baiman, 11+ sanbaiman, yakuman separate | EMA pp. 19-20 and 30, sections 4.1.2-4.1.3 |
| Kiriage | 4 han 30 fu is not mangan | EMA p. 5 revision notes and p. 30 table |
| Kazoe | 13 or more normal han remains sanbaiman, not yakuman | EMA p. 5 revision notes |
| Yakuman | Yakuman are not cumulative; Big Four Winds is one yakuman | EMA pp. 20 and 22-23, sections 4.2 and 4.2.5 |
| Renhou | Fixed mangan; not cumulative with yaku or dora | EMA pp. 5 and 22, section 4.2.4 |
| Nagashi mangan | Not used | EMA pp. 5 and 23, section 4.2.6 |
| Responsibility | Pao applies to the third open dragon set and fourth open wind set | EMA p. 13, section 3.3.9 |
| Final ranking | Ties allowed; remaining sticks go to the winner and split on a tie; EMA uma is 15k/5k/-5k/-15k | EMA p. 17, sections 3.6-3.6.1 |

## Explicit Project Overrides

These values come from the approved dual-rules project specification and are not attributed to EMA 2016.

| Area | Project value | Difference from EMA 2016 |
| --- | --- | --- |
| Starting points | 25,000 | EMA p. 6 starts at 30,000 |
| Red fives | Room option: one red five in each suit, or none | EMA p. 6 removed red fives |
| Open tanyao | Enabled | Matches EMA p. 5 |
| Abortive draws | Kyuushu kyuuhai, suufon renda, suucha riichi, suukaikan except four kans by one player, and sanchahou | EMA p. 5 and section 3.4.3 removed abortive draws; required by the approved project specification |
| Triple ron | Sanchahou abortive draw; one or two ron winners otherwise | Fixed together with the abortive-draw project extension |
| Uma | Report the EMA 15k/5k/-5k/-15k ranking adjustment separately; do not mutate in-match point totals | Keeps gameplay totals and tournament ranking output distinct |
| Chombo and etiquette penalties | Not exposed as player commands; illegal commands are rejected without mutation | Tournament adjudication is outside the terminal game command model |

## Dead-Wall Array Orientation

The in-memory `DeadWall[14]` is ordered from the replacement end toward the indicator end. Indexes `0..3` are the four rinshan positions. Index pairs `4/5`, `6/7`, `8/9`, `10/11`, and `12/13` are dora/ura indicators. A rinshan draw replaces its consumed slot with the last live-wall tile, keeping the dead-wall array at 14 while reducing the live wall. `testdata/rules/riichi/wall.json` freezes this representation; it is an implementation orientation of EMA sections 2.6-2.7, not a change to play order.

## Catalog Contract

- `testdata/rules/riichi/catalog.json` contains 29 ordinary/fixed-limit yaku, 12 single non-cumulative yakuman, and four bonus categories.
- `closed_han` and `open_han` are zero when that openness is illegal or when the entry is a fixed limit/yakuman.
- `requires_closed` resolves yakuman and fixed-limit openness that cannot be expressed by han fields.
- Dora, ura-dora, kan-dora, and red fives have `is_yaku=false`; they can add bonus han only after at least one yaku is present.
- Every fixture cites this source note plus the printed EMA page/section or the explicit project-override row.
