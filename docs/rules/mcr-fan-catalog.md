# Chinese Official Fan Catalog

This table is the human-readable mirror of `testdata/rules/mcr/catalog.json`. `GB/T 34708-2017` is normative. The MIT project `dchen327/mahjong-calc` was consulted only to cross-check English names, point bands, and exclusion identifiers; its detector code and expected scores are not copied.

`排除` means that scoring the row suppresses the listed lower scoring elements. Standard counting principles still apply after direct exclusions. `花牌` is counted after the non-flower eight-point minimum is met.

| ID | 中文 | English | 分 | 排除 | 可重复 | 来源 |
| --- | --- | --- | ---: | --- | --- | --- |
| mcr_01 | 一般高 | Pure Double Chow | 1 | - | 是 | GB/T item 1 |
| mcr_02 | 喜相逢 | Mixed Double Chow | 1 | - | 是 | GB/T item 2 |
| mcr_03 | 连六 | Short Straight | 1 | - | 是 | GB/T item 3 |
| mcr_04 | 老少副 | Two Terminal Chows | 1 | - | 是 | GB/T item 4 |
| mcr_05 | 幺九刻 | Pung of Terminals or Honors | 1 | - | 是 | GB/T item 5 |
| mcr_06 | 明杠 | Melded Kong | 1 | - | 是 | GB/T item 6 |
| mcr_07 | 缺一门 | One Voided Suit | 1 | - | 否 | GB/T item 7 |
| mcr_08 | 无字 | No Honor Tiles | 1 | - | 否 | GB/T item 8 |
| mcr_09 | 自摸 | Self Drawn | 1 | - | 否 | GB/T item 9 |
| mcr_10 | 花牌 | Flower Tiles | 1 | - | 是 | GB/T item 10 |
| mcr_11 | 边张 | Edge Wait | 1 | mcr_12, mcr_13 | 否 | GB/T item 11 |
| mcr_12 | 坎张 | Closed Wait | 1 | mcr_11, mcr_13 | 否 | GB/T item 12 |
| mcr_13 | 单钓将 | Pair Wait | 1 | mcr_11, mcr_12 | 否 | GB/T item 13 |
| mcr_14 | 箭刻 | Dragon Pung | 2 | - | 是 | GB/T item 14 |
| mcr_15 | 圈风刻 | Prevalent Wind | 2 | - | 否 | GB/T item 15 |
| mcr_16 | 门风刻 | Seat Wind | 2 | - | 否 | GB/T item 16 |
| mcr_17 | 门前清 | Concealed Hand | 2 | - | 否 | GB/T item 17 |
| mcr_18 | 平和 | All Chows | 2 | mcr_08 | 否 | GB/T item 18 |
| mcr_19 | 四归一 | Tile Hog | 2 | - | 是 | GB/T item 19 |
| mcr_20 | 双同刻 | Double Pung | 2 | - | 是 | GB/T item 20 |
| mcr_21 | 双暗刻 | Two Concealed Pungs | 2 | - | 否 | GB/T item 21 |
| mcr_22 | 暗杠 | Concealed Kong | 2 | - | 是 | GB/T item 22 |
| mcr_23 | 断幺 | All Simples | 2 | mcr_08 | 否 | GB/T item 23 |
| mcr_24 | 全带幺 | Outside Hand | 4 | - | 否 | GB/T item 24 |
| mcr_25 | 不求人 | Fully Concealed Hand | 4 | mcr_09 | 否 | GB/T item 25 |
| mcr_26 | 双明杠 | Two Melded Kongs | 4 | mcr_06 | 否 | GB/T item 26 |
| mcr_27 | 和绝张 | Last of its Kind | 4 | - | 否 | GB/T item 27 |
| mcr_28 | 碰碰和 | All Pungs | 6 | - | 否 | GB/T item 28 |
| mcr_29 | 混一色 | Half Flush | 6 | mcr_07 | 否 | GB/T item 29 |
| mcr_30 | 三色三步高 | Mixed Shifted Chows | 6 | - | 否 | GB/T item 30 |
| mcr_31 | 五门齐 | All Types | 6 | - | 否 | GB/T item 31 |
| mcr_32 | 全求人 | Melded Hand | 6 | mcr_13 | 否 | GB/T item 32 |
| mcr_33 | 双箭刻 | Two Dragon Pungs | 6 | mcr_14 | 否 | GB/T item 33 |
| mcr_34 | 花龙 | Mixed Straight | 8 | - | 否 | GB/T item 34 |
| mcr_35 | 推不倒 | Reversible Tiles | 8 | mcr_07 | 否 | GB/T item 35 |
| mcr_36 | 三色三同顺 | Mixed Triple Chow | 8 | mcr_02 | 否 | GB/T item 36 |
| mcr_37 | 三色三节高 | Mixed Shifted Pungs | 8 | - | 否 | GB/T item 37 |
| mcr_38 | 双暗杠 | Two Concealed Kongs | 8 | mcr_21, mcr_22 | 否 | GB/T item 38 |
| mcr_39 | 妙手回春 | Last Tile Draw | 8 | mcr_09 | 否 | GB/T item 39 |
| mcr_40 | 海底捞月 | Last Tile Claim | 8 | - | 否 | GB/T item 40 |
| mcr_41 | 杠上开花 | Out with Replacement Tile | 8 | mcr_09 | 否 | GB/T item 41 |
| mcr_42 | 抢杠和 | Robbing the Kong | 8 | mcr_27 | 否 | GB/T item 42 |
| mcr_43 | 无番和 | Chicken Hand | 8 | - | 否 | GB/T item 43 |
| mcr_44 | 全不靠 | Lesser Honors and Knitted Tiles | 12 | mcr_17, mcr_31 | 否 | GB/T item 44 |
| mcr_45 | 组合龙 | Knitted Straight | 12 | - | 否 | GB/T item 45 |
| mcr_46 | 大于五 | Upper Four | 12 | mcr_08 | 否 | GB/T item 46 |
| mcr_47 | 小于五 | Lower Four | 12 | mcr_08 | 否 | GB/T item 47 |
| mcr_48 | 三风刻 | Big Three Winds | 12 | mcr_05 | 否 | GB/T item 48 |
| mcr_49 | 清龙 | Pure Straight | 16 | mcr_03, mcr_04 | 否 | GB/T item 49 |
| mcr_50 | 三色双龙会 | Three-Suited Terminal Chows | 16 | mcr_02, mcr_04, mcr_08, mcr_18 | 否 | GB/T item 50 |
| mcr_51 | 一色三步高 | Pure Shifted Chows | 16 | - | 否 | GB/T item 51 |
| mcr_52 | 全带五 | All Fives | 16 | mcr_08, mcr_23 | 否 | GB/T item 52 |
| mcr_53 | 三同刻 | Triple Pung | 16 | mcr_20 | 否 | GB/T item 53 |
| mcr_54 | 三暗刻 | Three Concealed Pungs | 16 | mcr_21 | 否 | GB/T item 54 |
| mcr_55 | 七对 | Seven Pairs | 24 | mcr_13, mcr_17 | 否 | GB/T item 55 |
| mcr_56 | 七星不靠 | Greater Honors and Knitted Tiles | 24 | mcr_17, mcr_31, mcr_44 | 否 | GB/T item 56 |
| mcr_57 | 全双刻 | All Even Pungs | 24 | mcr_08, mcr_23, mcr_28 | 否 | GB/T item 57 |
| mcr_58 | 清一色 | Full Flush | 24 | mcr_07, mcr_08, mcr_29 | 否 | GB/T item 58 |
| mcr_59 | 一色三同顺 | Pure Triple Chow | 24 | mcr_01 | 否 | GB/T item 59 |
| mcr_60 | 一色三节高 | Pure Shifted Pungs | 24 | - | 否 | GB/T item 60 |
| mcr_61 | 全大 | Upper Tiles | 24 | mcr_08, mcr_46 | 否 | GB/T item 61 |
| mcr_62 | 全中 | Middle Tiles | 24 | mcr_08, mcr_23 | 否 | GB/T item 62 |
| mcr_63 | 全小 | Lower Tiles | 24 | mcr_08, mcr_47 | 否 | GB/T item 63 |
| mcr_64 | 一色四步高 | Four Shifted Chows | 32 | mcr_03, mcr_51 | 否 | GB/T item 64 |
| mcr_65 | 三杠 | Three Kongs | 32 | mcr_06, mcr_22, mcr_26, mcr_38 | 否 | GB/T item 65 |
| mcr_66 | 混幺九 | All Terminals and Honors | 32 | mcr_05, mcr_24, mcr_28 | 否 | GB/T item 66 |
| mcr_67 | 一色四同顺 | Quadruple Chow | 48 | mcr_01, mcr_19, mcr_59 | 否 | GB/T item 67 |
| mcr_68 | 一色四节高 | Four Pure Shifted Pungs | 48 | mcr_28, mcr_60 | 否 | GB/T item 68 |
| mcr_69 | 清幺九 | All Terminals | 64 | mcr_05, mcr_08, mcr_24, mcr_28, mcr_66 | 否 | GB/T item 69 |
| mcr_70 | 字一色 | All Honors | 64 | mcr_05, mcr_07, mcr_24, mcr_28, mcr_66 | 否 | GB/T item 70 |
| mcr_71 | 小四喜 | Little Four Winds | 64 | mcr_05, mcr_48 | 否 | GB/T item 71 |
| mcr_72 | 小三元 | Little Three Dragons | 64 | mcr_14, mcr_33 | 否 | GB/T item 72 |
| mcr_73 | 四暗刻 | Four Concealed Pungs | 64 | mcr_17, mcr_21, mcr_28, mcr_54 | 否 | GB/T item 73 |
| mcr_74 | 一色双龙会 | Pure Terminal Chows | 64 | mcr_01, mcr_04, mcr_07, mcr_18, mcr_29, mcr_58 | 否 | GB/T item 74 |
| mcr_75 | 大四喜 | Big Four Winds | 88 | mcr_05, mcr_15, mcr_16, mcr_28, mcr_48 | 否 | GB/T item 75 |
| mcr_76 | 大三元 | Big Three Dragons | 88 | mcr_14, mcr_33 | 否 | GB/T item 76 |
| mcr_77 | 绿一色 | All Green | 88 | - | 否 | GB/T item 77 |
| mcr_78 | 九莲宝灯 | Nine Gates | 88 | mcr_05, mcr_07, mcr_08, mcr_17, mcr_29, mcr_58 | 否 | GB/T item 78 |
| mcr_79 | 四杠 | Four Kongs | 88 | mcr_06, mcr_13, mcr_22, mcr_26, mcr_28, mcr_38, mcr_65 | 否 | GB/T item 79 |
| mcr_80 | 连七对 | Seven Shifted Pairs | 88 | mcr_07, mcr_08, mcr_13, mcr_17, mcr_29, mcr_55, mcr_58 | 否 | GB/T item 80 |
| mcr_81 | 十三幺 | Thirteen Orphans | 88 | mcr_17, mcr_24, mcr_31, mcr_66 | 否 | GB/T item 81 |
