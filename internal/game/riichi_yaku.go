package game

import "sort"

type riichiYakuValue struct {
	zh      string
	en      string
	closed  int
	open    int
	yakuman int
}

var riichiYakuValues = map[string]riichiYakuValue{
	"riichi":            {zh: "立直", en: "Riichi", closed: 1},
	"ippatsu":           {zh: "一发", en: "Ippatsu", closed: 1},
	"double_riichi":     {zh: "两立直", en: "Double Riichi", closed: 2},
	"menzen_tsumo":      {zh: "门前清自摸和", en: "Fully Concealed Hand", closed: 1},
	"pinfu":             {zh: "平和", en: "Pinfu", closed: 1},
	"iipeikou":          {zh: "一杯口", en: "Pure Double Chow", closed: 1},
	"tanyao":            {zh: "断幺九", en: "All Simples", closed: 1, open: 1},
	"sanshoku_doujun":   {zh: "三色同顺", en: "Mixed Triple Chow", closed: 2, open: 1},
	"ittsu":             {zh: "一气通贯", en: "Pure Straight", closed: 2, open: 1},
	"yakuhai_dragon":    {zh: "役牌-三元牌", en: "Dragon Pung", closed: 1, open: 1},
	"yakuhai_seat":      {zh: "役牌-门风牌", en: "Seat Wind", closed: 1, open: 1},
	"yakuhai_prevalent": {zh: "役牌-场风牌", en: "Prevalent Wind", closed: 1, open: 1},
	"chanta":            {zh: "混全带幺九", en: "Outside Hand", closed: 2, open: 1},
	"rinshan":           {zh: "岭上开花", en: "After a Kong", closed: 1, open: 1},
	"chankan":           {zh: "抢杠", en: "Robbing the Kong", closed: 1, open: 1},
	"haitei":            {zh: "海底摸月", en: "Under the Sea - Draw", closed: 1, open: 1},
	"houtei":            {zh: "河底捞鱼", en: "Under the Sea - Discard", closed: 1, open: 1},
	"chiitoitsu":        {zh: "七对子", en: "Seven Pairs", closed: 2},
	"sanshoku_doukou":   {zh: "三色同刻", en: "Triple Pung", closed: 2, open: 2},
	"sanankou":          {zh: "三暗刻", en: "Three Concealed Pungs", closed: 2, open: 2},
	"sankantsu":         {zh: "三杠子", en: "Three Kongs", closed: 2, open: 2},
	"toitoi":            {zh: "对对和", en: "All Pungs", closed: 2, open: 2},
	"honitsu":           {zh: "混一色", en: "Half Flush", closed: 3, open: 2},
	"shousangen":        {zh: "小三元", en: "Little Three Dragons", closed: 2, open: 2},
	"honroutou":         {zh: "混老头", en: "All Terminals and Honours", closed: 2, open: 2},
	"junchan":           {zh: "纯全带幺九", en: "Terminals in All Sets", closed: 3, open: 2},
	"ryanpeikou":        {zh: "二杯口", en: "Twice Pure Double Chows", closed: 3},
	"chinitsu":          {zh: "清一色", en: "Full Flush", closed: 6, open: 5},
	"renhou":            {zh: "人和", en: "Blessing of Man"},
	"kokushi":           {zh: "国士无双", en: "Thirteen Orphans", yakuman: 1},
	"chuuren":           {zh: "九莲宝灯", en: "Nine Gates", yakuman: 1},
	"tenhou":            {zh: "天和", en: "Blessing of Heaven", yakuman: 1},
	"chiihou":           {zh: "地和", en: "Blessing of Earth", yakuman: 1},
	"suuankou":          {zh: "四暗刻", en: "Four Concealed Pungs", yakuman: 1},
	"suukantsu":         {zh: "四杠子", en: "Four Kongs", yakuman: 1},
	"ryuuiisou":         {zh: "绿一色", en: "All Green", yakuman: 1},
	"chinroutou":        {zh: "清老头", en: "All Terminals", yakuman: 1},
	"tsuuiisou":         {zh: "字一色", en: "All Honours", yakuman: 1},
	"daisangen":         {zh: "大三元", en: "Big Three Dragons", yakuman: 1},
	"shousuushii":       {zh: "小四喜", en: "Little Four Winds", yakuman: 1},
	"daisuushii":        {zh: "大四喜", en: "Big Four Winds", yakuman: 1},
}

func DetectRiichiYaku(context RiichiYakuContext) []RiichiYakuMatch {
	var matches []RiichiYakuMatch
	add := func(id string) {
		if match, ok := riichiYakuMatchFor(id, context.Closed); ok {
			matches = append(matches, match)
		}
	}

	if context.Riichi == RiichiAccepted || context.Riichi == RiichiDeclared {
		add("riichi")
	}
	if context.Ippatsu {
		add("ippatsu")
	}
	if context.DoubleRiichi {
		add("double_riichi")
	}
	if context.Closed && context.WinType == WinSelfDraw {
		add("menzen_tsumo")
	}
	if context.Rinshan {
		add("rinshan")
	}
	if context.Chankan {
		add("chankan")
	}
	if context.Haitei && context.WinType == WinSelfDraw {
		add("haitei")
	}
	if context.Houtei && context.WinType == WinDiscard {
		add("houtei")
	}
	if context.Renhou {
		add("renhou")
	}
	if context.Tenhou {
		add("tenhou")
	}
	if context.Chiihou {
		add("chiihou")
	}

	if riichiIsPinfu(context) {
		add("pinfu")
	}
	if context.Closed && riichiIdenticalChowPairs(context.Decomposition.Groups) >= 1 {
		add("iipeikou")
	}
	if riichiAllTiles(context, riichiIsSimple) {
		add("tanyao")
	}
	if riichiHasSanshokuDoujun(context.Decomposition.Groups) {
		add("sanshoku_doujun")
	}
	if riichiHasIttsu(context.Decomposition.Groups) {
		add("ittsu")
	}
	if riichiHasDragonPung(context.Decomposition.Groups) {
		add("yakuhai_dragon")
	}
	if riichiHasPungOf(context.Decomposition.Groups, context.SeatWind) {
		add("yakuhai_seat")
	}
	if riichiHasPungOf(context.Decomposition.Groups, context.PrevalentWind) {
		add("yakuhai_prevalent")
	}
	if riichiIsChanta(context.Decomposition.Groups) {
		add("chanta")
	}
	if context.Decomposition.Kind == RiichiShapeSevenPairs {
		add("chiitoitsu")
	}
	if riichiHasSanshokuDoukou(context.Decomposition.Groups) {
		add("sanshoku_doukou")
	}
	if riichiConcealedPungCount(context) >= 3 {
		add("sanankou")
	}
	if riichiKongCount(context.Decomposition.Groups) >= 3 {
		add("sankantsu")
	}
	if riichiIsToitoi(context.Decomposition.Groups) {
		add("toitoi")
	}
	if riichiFlushKind(context.Decomposition.Tiles) == "honitsu" {
		add("honitsu")
	}
	if riichiIsShousangen(context.Decomposition.Groups) {
		add("shousangen")
	}
	if riichiAllTiles(context, riichiIsTerminalOrHonor) {
		add("honroutou")
	}
	if riichiIsJunchan(context.Decomposition.Groups) {
		add("junchan")
	}
	if context.Closed && riichiIdenticalChowPairs(context.Decomposition.Groups) >= 2 {
		add("ryanpeikou")
	}
	if riichiFlushKind(context.Decomposition.Tiles) == "chinitsu" {
		add("chinitsu")
	}

	if context.Decomposition.Kind == RiichiShapeThirteenOrphans {
		add("kokushi")
	}
	if context.Closed && riichiIsChuuren(context.Decomposition.Tiles) {
		add("chuuren")
	}
	if context.Closed && riichiConcealedPungCount(context) >= 4 {
		add("suuankou")
	}
	if riichiKongCount(context.Decomposition.Groups) >= 4 {
		add("suukantsu")
	}
	if riichiAllTiles(context, riichiIsGreen) {
		add("ryuuiisou")
	}
	if riichiAllTiles(context, riichiIsTerminal) {
		add("chinroutou")
	}
	if riichiAllTiles(context, riichiIsHonor) {
		add("tsuuiisou")
	}
	if riichiDragonPungCount(context.Decomposition.Groups) >= 3 {
		add("daisangen")
	}
	if riichiWindPungCount(context.Decomposition.Groups) >= 3 && riichiHasWindPair(context.Decomposition.Groups) {
		add("shousuushii")
	}
	if riichiWindPungCount(context.Decomposition.Groups) >= 4 {
		add("daisuushii")
	}

	sort.Slice(matches, func(i, j int) bool { return matches[i].ID < matches[j].ID })
	return matches
}

func riichiYakuMatchFor(id string, closed bool) (RiichiYakuMatch, bool) {
	value, ok := riichiYakuValues[id]
	if !ok {
		return RiichiYakuMatch{}, false
	}
	han := value.open
	if closed {
		han = value.closed
	}
	if value.yakuman == 0 && han == 0 && id != "renhou" {
		return RiichiYakuMatch{}, false
	}
	return RiichiYakuMatch{ID: id, NameZH: value.zh, NameEN: value.en, Han: han, Yakuman: value.yakuman}, true
}

func riichiIsPinfu(context RiichiYakuContext) bool {
	if !context.Closed || context.Decomposition.Kind != RiichiShapeStandard {
		return false
	}
	hasPair := false
	for _, group := range context.Decomposition.Groups {
		switch group.Kind {
		case MCRGroupChow:
		case MCRGroupPair:
			hasPair = true
			if len(group.Tiles) == 0 || riichiIsDragon(group.Tiles[0]) || group.Tiles[0].Base() == context.SeatWind.Base() || group.Tiles[0].Base() == context.PrevalentWind.Base() {
				return false
			}
		default:
			return false
		}
	}
	return hasPair
}

func riichiIdenticalChowPairs(groups []MCRGroup) int {
	counts := make(map[string]int)
	for _, group := range groups {
		if group.Kind != MCRGroupChow {
			continue
		}
		tiles := append([]Tile(nil), group.Tiles...)
		SortTiles(tiles)
		counts[FormatTiles(tiles)]++
	}
	pairs := 0
	for _, count := range counts {
		pairs += count / 2
	}
	return pairs
}

func riichiHasSanshokuDoujun(groups []MCRGroup) bool {
	seen := make(map[int]map[int]bool)
	for _, group := range groups {
		if group.Kind != MCRGroupChow || len(group.Tiles) != 3 {
			continue
		}
		tiles := append([]Tile(nil), group.Tiles...)
		SortTiles(tiles)
		suit := int(tiles[0].Base()) / 9
		start := tiles[0].Rank()
		if seen[start] == nil {
			seen[start] = make(map[int]bool)
		}
		seen[start][suit] = true
	}
	for _, suits := range seen {
		if len(suits) == 3 {
			return true
		}
	}
	return false
}

func riichiHasIttsu(groups []MCRGroup) bool {
	seen := make(map[int]map[int]bool)
	for _, group := range groups {
		if group.Kind != MCRGroupChow || len(group.Tiles) != 3 {
			continue
		}
		tiles := append([]Tile(nil), group.Tiles...)
		SortTiles(tiles)
		suit := int(tiles[0].Base()) / 9
		start := tiles[0].Rank()
		if seen[suit] == nil {
			seen[suit] = make(map[int]bool)
		}
		seen[suit][start] = true
	}
	for _, starts := range seen {
		if starts[1] && starts[4] && starts[7] {
			return true
		}
	}
	return false
}

func riichiHasDragonPung(groups []MCRGroup) bool {
	return riichiDragonPungCount(groups) > 0
}

func riichiDragonPungCount(groups []MCRGroup) int {
	count := 0
	for _, group := range groups {
		if riichiIsPungLike(group) && len(group.Tiles) > 0 && riichiIsDragon(group.Tiles[0]) {
			count++
		}
	}
	return count
}

func riichiHasPungOf(groups []MCRGroup, tile Tile) bool {
	for _, group := range groups {
		if riichiIsPungLike(group) && len(group.Tiles) > 0 && group.Tiles[0].Base() == tile.Base() {
			return true
		}
	}
	return false
}

func riichiIsChanta(groups []MCRGroup) bool {
	hasChow := false
	for _, group := range groups {
		if group.Kind == MCRGroupChow {
			hasChow = true
		}
		if !riichiGroupHas(group, riichiIsTerminalOrHonor) {
			return false
		}
	}
	return hasChow
}

func riichiHasSanshokuDoukou(groups []MCRGroup) bool {
	seen := make(map[int]map[int]bool)
	for _, group := range groups {
		if !riichiIsPungLike(group) || len(group.Tiles) == 0 || !group.Tiles[0].IsSuit() {
			continue
		}
		rank := group.Tiles[0].Rank()
		suit := int(group.Tiles[0].Base()) / 9
		if seen[rank] == nil {
			seen[rank] = make(map[int]bool)
		}
		seen[rank][suit] = true
	}
	for _, suits := range seen {
		if len(suits) == 3 {
			return true
		}
	}
	return false
}

func riichiConcealedPungCount(context RiichiYakuContext) int {
	count := 0
	for _, group := range context.Decomposition.Groups {
		if !riichiIsPungLike(group) || group.Open {
			continue
		}
		if context.WinType == WinDiscard && group.Kind == MCRGroupPung && context.Decomposition.Wait == RiichiWaitShanpon && riichiGroupHasTile(group, context.WinningTile) {
			continue
		}
		count++
	}
	return count
}

func riichiKongCount(groups []MCRGroup) int {
	count := 0
	for _, group := range groups {
		if group.Kind == MCRGroupKong {
			count++
		}
	}
	return count
}

func riichiIsToitoi(groups []MCRGroup) bool {
	pungs := 0
	for _, group := range groups {
		switch group.Kind {
		case MCRGroupPair:
		case MCRGroupPung, MCRGroupKong:
			pungs++
		default:
			return false
		}
	}
	return pungs >= 4
}

func riichiFlushKind(tiles []Tile) string {
	suit := -1
	hasHonor := false
	for _, tile := range tiles {
		base := tile.Base()
		if base >= 27 {
			hasHonor = true
			continue
		}
		if base < 0 || base >= 27 {
			continue
		}
		tileSuit := int(base) / 9
		if suit == -1 {
			suit = tileSuit
		} else if suit != tileSuit {
			return ""
		}
	}
	if suit == -1 {
		return ""
	}
	if hasHonor {
		return "honitsu"
	}
	return "chinitsu"
}

func riichiIsShousangen(groups []MCRGroup) bool {
	dragonPungs := 0
	dragonPair := false
	for _, group := range groups {
		if len(group.Tiles) == 0 {
			continue
		}
		if riichiIsPungLike(group) && riichiIsDragon(group.Tiles[0]) {
			dragonPungs++
		}
		if group.Kind == MCRGroupPair && riichiIsDragon(group.Tiles[0]) {
			dragonPair = true
		}
	}
	return dragonPungs == 2 && dragonPair
}

func riichiIsJunchan(groups []MCRGroup) bool {
	hasChow := false
	for _, group := range groups {
		if group.Kind == MCRGroupChow {
			hasChow = true
		}
		if riichiGroupHas(group, riichiIsHonor) || !riichiGroupHas(group, riichiIsTerminal) {
			return false
		}
	}
	return hasChow
}

func riichiIsChuuren(tiles []Tile) bool {
	if len(tiles) != 14 {
		return false
	}
	counts := TileCounts(tiles)
	suit := -1
	for index, count := range counts {
		if count == 0 {
			continue
		}
		tile := Tile(index)
		if !tile.IsSuit() {
			return false
		}
		tileSuit := index / 9
		if suit == -1 {
			suit = tileSuit
		} else if suit != tileSuit {
			return false
		}
	}
	if suit < 0 {
		return false
	}
	base := suit * 9
	if counts[base] < 3 || counts[base+8] < 3 {
		return false
	}
	for offset := 1; offset <= 7; offset++ {
		if counts[base+offset] < 1 {
			return false
		}
	}
	return true
}

func riichiWindPungCount(groups []MCRGroup) int {
	count := 0
	for _, group := range groups {
		if riichiIsPungLike(group) && len(group.Tiles) > 0 && riichiIsWind(group.Tiles[0]) {
			count++
		}
	}
	return count
}

func riichiHasWindPair(groups []MCRGroup) bool {
	for _, group := range groups {
		if group.Kind == MCRGroupPair && len(group.Tiles) > 0 && riichiIsWind(group.Tiles[0]) {
			return true
		}
	}
	return false
}

func riichiAllTiles(context RiichiYakuContext, predicate func(Tile) bool) bool {
	if len(context.Decomposition.Tiles) == 0 {
		return false
	}
	for _, tile := range context.Decomposition.Tiles {
		if !predicate(tile) {
			return false
		}
	}
	return true
}

func riichiGroupHas(group MCRGroup, predicate func(Tile) bool) bool {
	for _, tile := range group.Tiles {
		if predicate(tile) {
			return true
		}
	}
	return false
}

func riichiGroupHasTile(group MCRGroup, tile Tile) bool {
	for _, groupTile := range group.Tiles {
		if groupTile.Base() == tile.Base() {
			return true
		}
	}
	return false
}

func riichiIsPungLike(group MCRGroup) bool {
	return group.Kind == MCRGroupPung || group.Kind == MCRGroupKong
}

func riichiIsSimple(tile Tile) bool {
	return tile.IsSuit() && tile.Rank() >= 2 && tile.Rank() <= 8
}

func riichiIsTerminal(tile Tile) bool {
	return tile.IsSuit() && (tile.Rank() == 1 || tile.Rank() == 9)
}

func riichiIsHonor(tile Tile) bool {
	base := tile.Base()
	return base >= 27 && base < 34
}

func riichiIsTerminalOrHonor(tile Tile) bool {
	return riichiIsTerminal(tile) || riichiIsHonor(tile)
}

func riichiIsDragon(tile Tile) bool {
	base := tile.Base()
	return base >= 31 && base <= 33
}

func riichiIsWind(tile Tile) bool {
	base := tile.Base()
	return base >= 27 && base <= 30
}

func riichiIsGreen(tile Tile) bool {
	switch tile.Base() {
	case 19, 20, 21, 23, 25, 32:
		return true
	default:
		return false
	}
}
