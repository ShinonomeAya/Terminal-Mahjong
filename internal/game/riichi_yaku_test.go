package game

import (
	"encoding/json"
	"os"
	"testing"
)

type riichiYakuCoverageEntry struct {
	ID       string   `json:"id"`
	Positive []string `json:"positive"`
	NearMiss []string `json:"near_miss"`
}

func TestEveryRiichiYakuHasGoldenCoverage(t *testing.T) {
	catalog := loadRiichiYakuCatalog(t)
	coverage := loadRiichiYakuCoverage(t)
	for _, entry := range catalog {
		if !entry.IsYaku {
			continue
		}
		value, ok := coverage[entry.ID]
		if !ok || len(value.Positive) == 0 || len(value.NearMiss) == 0 {
			t.Errorf("%s lacks positive and near-miss coverage", entry.ID)
		}
	}
}

func TestRiichiYakuPositiveExamples(t *testing.T) {
	tests := []struct {
		id      string
		context RiichiYakuContext
	}{
		{"riichi", yakuCtx("1m", riichiFlag(RiichiAccepted))},
		{"ippatsu", yakuCtx("1m", riichiFlag(RiichiAccepted), ippatsuFlag())},
		{"double_riichi", yakuCtx("1m", doubleRiichiFlag())},
		{"menzen_tsumo", yakuCtx("1m", winTypeFlag(WinSelfDraw))},
		{"pinfu", yakuCtx("4m", handFlag("2m", "3m", "3p", "4p", "5p", "3s", "4s", "5s", "6s", "7s", "8s", "5p", "5p"))},
		{"iipeikou", yakuCtx("3m", handFlag("1m", "1m", "2m", "2m", "3m", "4p", "5p", "6p", "7s", "8s", "9s", "E", "E"))},
		{"tanyao", yakuCtx("4m", handFlag("2m", "3m", "3p", "4p", "5p", "4s", "5s", "6s", "6p", "7p", "8p", "5s", "5s"), openFlag())},
		{"sanshoku_doujun", yakuCtx("3s", handFlag("1m", "2m", "3m", "1p", "2p", "3p", "1s", "2s", "4m", "5m", "6m", "E", "E"))},
		{"ittsu", yakuCtx("9m", handFlag("1m", "2m", "3m", "4m", "5m", "6m", "7m", "8m", "2p", "3p", "4p", "E", "E"))},
		{"yakuhai_dragon", yakuCtx("Z", handFlag("Z", "Z", "1m", "2m", "3m", "4p", "5p", "6p", "7s", "8s", "9s", "E", "E"), openFlag())},
		{"yakuhai_seat", yakuCtx("E", handFlag("E", "E", "1m", "2m", "3m", "4p", "5p", "6p", "7s", "8s", "9s", "2m", "2m"), openFlag(), seatWindFlag("E"))},
		{"yakuhai_prevalent", yakuCtx("S", handFlag("S", "S", "1m", "2m", "3m", "4p", "5p", "6p", "7s", "8s", "9s", "2m", "2m"), openFlag(), prevalentWindFlag("S"))},
		{"chanta", yakuCtx("3m", handFlag("1m", "2m", "7p", "8p", "9p", "1s", "1s", "1s", "E", "E", "E", "9m", "9m"), openFlag())},
		{"rinshan", yakuCtx("1m", rinshanFlag())},
		{"chankan", yakuCtx("1m", chankanFlag())},
		{"haitei", yakuCtx("1m", haiteiFlag(), winTypeFlag(WinSelfDraw))},
		{"houtei", yakuCtx("1m", houteiFlag(), winTypeFlag(WinDiscard))},
		{"chiitoitsu", yakuCtx("Z", handFlag("1m", "1m", "2m", "2m", "3p", "3p", "4p", "4p", "5s", "5s", "E", "E", "Z"))},
		{"sanshoku_doukou", yakuCtx("5s", handFlag("5m", "5m", "5m", "5p", "5p", "5p", "5s", "5s", "1m", "2m", "3m", "E", "E"), openFlag())},
		{"sanankou", yakuCtx("3s", handFlag("1m", "1m", "1m", "2p", "2p", "2p", "3s", "3s", "4m", "5m", "6m", "E", "E"), winTypeFlag(WinSelfDraw))},
		{"sankantsu", yakuCtx("E", meldFlag(kongMeld("1m"), kongMeld("2p"), kongMeld("3s")), handFlag("4m", "5m", "6m", "E"), openFlag())},
		{"toitoi", yakuCtx("3s", handFlag("1m", "1m", "1m", "2p", "2p", "2p", "3s", "3s", "4m", "4m", "4m", "E", "E"), openFlag())},
		{"honitsu", yakuCtx("3m", handFlag("1m", "1m", "1m", "2m", "2m", "2m", "3m", "3m", "E", "E", "E", "Z", "Z"), openFlag())},
		{"shousangen", yakuCtx("3m", handFlag("Z", "Z", "Z", "B", "B", "B", "F", "F", "1m", "2m", "E", "E", "E"), openFlag())},
		{"honroutou", yakuCtx("9s", handFlag("1m", "1m", "1m", "9p", "9p", "9p", "9s", "9s", "E", "E", "E", "Z", "Z"), openFlag())},
		{"junchan", yakuCtx("3m", handFlag("1m", "2m", "7m", "8m", "9m", "1p", "2p", "3p", "7s", "8s", "9s", "9p", "9p"), openFlag())},
		{"ryanpeikou", yakuCtx("3m", handFlag("1m", "1m", "2m", "2m", "3m", "4p", "4p", "5p", "5p", "6p", "6p", "9s", "9s"))},
		{"chinitsu", yakuCtx("3m", handFlag("1m", "1m", "1m", "2m", "2m", "2m", "3m", "3m", "4m", "4m", "4m", "5m", "5m"), openFlag())},
		{"renhou", yakuCtx("1m", renhouFlag())},
		{"kokushi", yakuCtx("E", handFlag("1m", "9m", "1p", "9p", "1s", "9s", "E", "S", "W", "N", "Z", "F", "B"))},
		{"chuuren", yakuCtx("5m", handFlag("1m", "1m", "1m", "2m", "3m", "4m", "5m", "6m", "7m", "8m", "9m", "9m", "9m"))},
		{"tenhou", yakuCtx("1m", tenhouFlag())},
		{"chiihou", yakuCtx("1m", chiihouFlag())},
		{"suuankou", yakuCtx("4s", handFlag("1m", "1m", "1m", "2p", "2p", "2p", "3s", "3s", "3s", "4s", "4s", "E", "E"), winTypeFlag(WinSelfDraw))},
		{"suukantsu", yakuCtx("E", meldFlag(kongMeld("1m"), kongMeld("2p"), kongMeld("3s"), kongMeld("4m")), handFlag("E"), openFlag())},
		{"ryuuiisou", yakuCtx("8s", handFlag("2s", "2s", "2s", "3s", "3s", "3s", "4s", "4s", "4s", "6s", "6s", "F", "F"), openFlag())},
		{"chinroutou", yakuCtx("9s", handFlag("1m", "1m", "1m", "9p", "9p", "9p", "9s", "9s", "1s", "1s", "1s", "9m", "9m"), openFlag())},
		{"tsuuiisou", yakuCtx("Z", handFlag("E", "E", "E", "S", "S", "S", "W", "W", "W", "Z", "Z", "F", "F"), openFlag())},
		{"daisangen", yakuCtx("B", handFlag("Z", "Z", "Z", "F", "F", "F", "B", "B", "1m", "2m", "3m", "E", "E"), openFlag())},
		{"shousuushii", yakuCtx("3m", handFlag("E", "E", "E", "S", "S", "S", "W", "W", "W", "N", "N", "1m", "2m"), openFlag())},
		{"daisuushii", yakuCtx("1m", handFlag("E", "E", "E", "S", "S", "S", "W", "W", "W", "N", "N", "N", "1m"), openFlag())},
	}
	for _, test := range tests {
		t.Run(test.id, func(t *testing.T) {
			if !hasRiichiYakuID(DetectRiichiYaku(test.context), test.id) {
				t.Fatalf("DetectRiichiYaku missing %s in %#v", test.id, DetectRiichiYaku(test.context))
			}
		})
	}
}

func loadRiichiYakuCatalog(t *testing.T) []riichiCatalogEntry {
	t.Helper()
	file, err := os.Open("../../testdata/rules/riichi/catalog.json")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	var catalog []riichiCatalogEntry
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&catalog); err != nil {
		t.Fatal(err)
	}
	return catalog
}

func loadRiichiYakuCoverage(t *testing.T) map[string]riichiYakuCoverageEntry {
	t.Helper()
	file, err := os.Open("../../testdata/rules/riichi/yaku/coverage.json")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	var entries []riichiYakuCoverageEntry
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&entries); err != nil {
		t.Fatal(err)
	}
	out := make(map[string]riichiYakuCoverageEntry, len(entries))
	for _, entry := range entries {
		if entry.ID == "" || out[entry.ID].ID != "" {
			t.Fatalf("invalid yaku coverage: %#v", entry)
		}
		out[entry.ID] = entry
	}
	return out
}

func hasRiichiYakuID(values []RiichiYakuMatch, id string) bool {
	for _, value := range values {
		if value.ID == id {
			return true
		}
	}
	return false
}

type yakuContextOption func(*riichiYakuTestConfig)

type riichiYakuTestConfig struct {
	hand          []Tile
	melds         []Meld
	riichi        RiichiDeclarationState
	winType       WinType
	closed        bool
	seatWind      Tile
	prevalentWind Tile
	ippatsu       bool
	doubleRiichi  bool
	rinshan       bool
	chankan       bool
	haitei        bool
	houtei        bool
	renhou        bool
	tenhou        bool
	chiihou       bool
}

func yakuCtx(winning string, options ...yakuContextOption) RiichiYakuContext {
	cfg := riichiYakuTestConfig{
		hand:          mustYakuTiles("1m", "2m", "3m", "4m", "5m", "6m", "7p", "8p", "9p", "2s", "2s", "3s", "4s"),
		winType:       WinDiscard,
		closed:        true,
		seatWind:      mustYakuTile("E"),
		prevalentWind: mustYakuTile("S"),
	}
	for _, option := range options {
		option(&cfg)
	}
	win := mustYakuTile(winning)
	decompositions := RiichiDecompose(cfg.hand, cfg.melds, win)
	decomposition := RiichiDecomposition{Tiles: append(append([]Tile(nil), cfg.hand...), win)}
	if len(decompositions) > 0 {
		decomposition = decompositions[0]
		for _, candidate := range decompositions {
			if candidate.Kind == RiichiShapeStandard {
				decomposition = candidate
				break
			}
		}
	}
	return RiichiYakuContext{
		Decomposition: decomposition,
		WinningTile:   win,
		WinType:       cfg.winType,
		Closed:        cfg.closed,
		SeatWind:      cfg.seatWind,
		PrevalentWind: cfg.prevalentWind,
		Riichi:        cfg.riichi,
		Ippatsu:       cfg.ippatsu,
		DoubleRiichi:  cfg.doubleRiichi,
		Rinshan:       cfg.rinshan,
		Chankan:       cfg.chankan,
		Haitei:        cfg.haitei,
		Houtei:        cfg.houtei,
		Renhou:        cfg.renhou,
		Tenhou:        cfg.tenhou,
		Chiihou:       cfg.chiihou,
	}
}

func handFlag(texts ...string) yakuContextOption {
	return func(cfg *riichiYakuTestConfig) { cfg.hand = mustYakuTiles(texts...) }
}

func meldFlag(melds ...Meld) yakuContextOption {
	return func(cfg *riichiYakuTestConfig) { cfg.melds = melds }
}

func openFlag() yakuContextOption {
	return func(cfg *riichiYakuTestConfig) { cfg.closed = false }
}

func riichiFlag(state RiichiDeclarationState) yakuContextOption {
	return func(cfg *riichiYakuTestConfig) { cfg.riichi = state }
}

func doubleRiichiFlag() yakuContextOption {
	return func(cfg *riichiYakuTestConfig) { cfg.doubleRiichi = true; cfg.riichi = RiichiAccepted }
}

func ippatsuFlag() yakuContextOption {
	return func(cfg *riichiYakuTestConfig) { cfg.ippatsu = true }
}

func winTypeFlag(winType WinType) yakuContextOption {
	return func(cfg *riichiYakuTestConfig) { cfg.winType = winType }
}

func seatWindFlag(tile string) yakuContextOption {
	return func(cfg *riichiYakuTestConfig) { cfg.seatWind = mustYakuTile(tile) }
}

func prevalentWindFlag(tile string) yakuContextOption {
	return func(cfg *riichiYakuTestConfig) { cfg.prevalentWind = mustYakuTile(tile) }
}

func rinshanFlag() yakuContextOption {
	return func(cfg *riichiYakuTestConfig) { cfg.rinshan = true }
}

func chankanFlag() yakuContextOption {
	return func(cfg *riichiYakuTestConfig) { cfg.chankan = true }
}

func haiteiFlag() yakuContextOption {
	return func(cfg *riichiYakuTestConfig) { cfg.haitei = true }
}

func houteiFlag() yakuContextOption {
	return func(cfg *riichiYakuTestConfig) { cfg.houtei = true }
}

func renhouFlag() yakuContextOption {
	return func(cfg *riichiYakuTestConfig) { cfg.renhou = true }
}

func tenhouFlag() yakuContextOption {
	return func(cfg *riichiYakuTestConfig) { cfg.tenhou = true }
}

func chiihouFlag() yakuContextOption {
	return func(cfg *riichiYakuTestConfig) { cfg.chiihou = true }
}

func kongMeld(tile string) Meld {
	value := mustYakuTile(tile)
	return Meld{Kind: MeldKong, Tiles: []Tile{value, value, value, value}}
}

func mustYakuTiles(texts ...string) []Tile {
	tiles := make([]Tile, 0, len(texts))
	for _, text := range texts {
		tiles = append(tiles, mustYakuTile(text))
	}
	return tiles
}

func mustYakuTile(text string) Tile {
	tile, ok := ParseTile(text)
	if !ok {
		panic("test fixture parse failed: " + text)
	}
	return tile
}
