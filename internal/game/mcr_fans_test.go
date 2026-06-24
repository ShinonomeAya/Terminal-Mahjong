package game

import "testing"

func TestMCROnePointFanDetectors(t *testing.T) {
	tests := []struct {
		name    string
		context MCRFanContext
		fan     FanID
		count   int
	}{
		{name: "pure double chow", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupChow, true, "1m", "2m", "3m"),
			mcrTestGroup(MCRGroupChow, false, "1m", "2m", "3m"),
		), fan: "mcr_01", count: 1},
		{name: "mixed double chow", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupChow, true, "2m", "3m", "4m"),
			mcrTestGroup(MCRGroupChow, false, "2p", "3p", "4p"),
		), fan: "mcr_02", count: 1},
		{name: "short straight", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupChow, false, "1m", "2m", "3m"),
			mcrTestGroup(MCRGroupChow, false, "4m", "5m", "6m"),
		), fan: "mcr_03", count: 1},
		{name: "two terminal chows", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupChow, false, "1s", "2s", "3s"),
			mcrTestGroup(MCRGroupChow, false, "7s", "8s", "9s"),
		), fan: "mcr_04", count: 1},
		{name: "terminal and honor pungs", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupPung, false, "1m", "1m", "1m"),
			mcrTestGroup(MCRGroupPung, true, "E", "E", "E"),
		), fan: "mcr_05", count: 2},
		{name: "melded kong", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupKong, true, "5p", "5p", "5p", "5p"),
		), fan: "mcr_06", count: 1},
		{name: "one voided suit", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupChow, false, "1m", "2m", "3m"),
			mcrTestGroup(MCRGroupPung, false, "6p", "6p", "6p"),
		), fan: "mcr_07", count: 1},
		{name: "no honors", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupChow, false, "1m", "2m", "3m"),
			mcrTestGroup(MCRGroupPair, false, "5p", "5p"),
		), fan: "mcr_08", count: 1},
		{name: "self drawn", context: MCRFanContext{WinType: WinSelfDraw}, fan: "mcr_09", count: 1},
		{name: "flower tiles", context: MCRFanContext{Flowers: mustFanTiles(t, "P1", "P2", "S1")}, fan: "mcr_10", count: 3},
		{name: "edge wait", context: MCRFanContext{
			Decomposition: decompositionWithGroups(mcrTestGroup(MCRGroupChow, false, "1m", "2m", "3m")),
			WinningTile:   mustFanTiles(t, "3m")[0],
		}, fan: "mcr_11", count: 1},
		{name: "closed wait", context: MCRFanContext{
			Decomposition: decompositionWithGroups(mcrTestGroup(MCRGroupChow, false, "4p", "5p", "6p")),
			WinningTile:   mustFanTiles(t, "5p")[0],
		}, fan: "mcr_12", count: 1},
		{name: "pair wait", context: MCRFanContext{
			Decomposition: decompositionWithGroups(mcrTestGroup(MCRGroupPair, false, "E", "E")),
			WinningTile:   mustFanTiles(t, "E")[0],
		}, fan: "mcr_13", count: 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			occurrences := DetectMCRFans(test.context)
			if got := mcrFanOccurrenceCount(occurrences, test.fan); got != test.count {
				t.Fatalf("%s count = %d, want %d; occurrences=%#v", test.fan, got, test.count, occurrences)
			}
		})
	}
}

func TestMCROnePointFanNearMisses(t *testing.T) {
	tests := []struct {
		name    string
		context MCRFanContext
		fan     FanID
	}{
		{name: "different pure chows", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupChow, false, "1m", "2m", "3m"),
			mcrTestGroup(MCRGroupChow, false, "2m", "3m", "4m"),
		), fan: "mcr_01"},
		{name: "all three suits present", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupChow, false, "1m", "2m", "3m"),
			mcrTestGroup(MCRGroupChow, false, "1p", "2p", "3p"),
			mcrTestGroup(MCRGroupChow, false, "1s", "2s", "3s"),
		), fan: "mcr_07"},
		{name: "honor present", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupPung, false, "E", "E", "E"),
		), fan: "mcr_08"},
		{name: "open wait", context: MCRFanContext{
			Decomposition: decompositionWithGroups(mcrTestGroup(MCRGroupChow, false, "2m", "3m", "4m")),
			WinningTile:   mustFanTiles(t, "2m")[0],
		}, fan: "mcr_11"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := mcrFanOccurrenceCount(DetectMCRFans(test.context), test.fan); got != 0 {
				t.Fatalf("%s count = %d, want 0", test.fan, got)
			}
		})
	}
}

func TestMCRTwoPointFanDetectors(t *testing.T) {
	closedChows := []MCRGroup{
		mcrTestGroup(MCRGroupChow, false, "1m", "2m", "3m"),
		mcrTestGroup(MCRGroupChow, false, "4m", "5m", "6m"),
		mcrTestGroup(MCRGroupChow, false, "2p", "3p", "4p"),
		mcrTestGroup(MCRGroupChow, false, "6s", "7s", "8s"),
		mcrTestGroup(MCRGroupPair, false, "5p", "5p"),
	}
	tests := []struct {
		name    string
		context MCRFanContext
		fan     FanID
	}{
		{name: "dragon pung", context: fanContextWithGroups(mcrTestGroup(MCRGroupPung, false, "Z", "Z", "Z")), fan: "mcr_14"},
		{name: "prevalent wind", context: MCRFanContext{Decomposition: decompositionWithGroups(mcrTestGroup(MCRGroupPung, true, "E", "E", "E")), PrevalentWind: mustFanTiles(t, "E")[0]}, fan: "mcr_15"},
		{name: "seat wind", context: MCRFanContext{Decomposition: decompositionWithGroups(mcrTestGroup(MCRGroupPung, false, "S", "S", "S")), SeatWind: mustFanTiles(t, "S")[0]}, fan: "mcr_16"},
		{name: "concealed hand", context: MCRFanContext{Decomposition: decompositionWithGroups(closedChows...), WinType: WinDiscard}, fan: "mcr_17"},
		{name: "all chows", context: fanContextWithGroups(closedChows...), fan: "mcr_18"},
		{name: "tile hog", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupPung, false, "5m", "5m", "5m"),
			mcrTestGroup(MCRGroupChow, false, "4m", "5m", "6m"),
		), fan: "mcr_19"},
		{name: "double pung", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupPung, false, "2m", "2m", "2m"),
			mcrTestGroup(MCRGroupPung, true, "2p", "2p", "2p"),
		), fan: "mcr_20"},
		{name: "two concealed pungs", context: MCRFanContext{Decomposition: decompositionWithGroups(
			mcrTestGroup(MCRGroupPung, false, "3m", "3m", "3m"),
			mcrTestGroup(MCRGroupPung, false, "7p", "7p", "7p"),
		), WinType: WinSelfDraw}, fan: "mcr_21"},
		{name: "concealed kong", context: fanContextWithGroups(mcrTestGroup(MCRGroupKong, false, "4s", "4s", "4s", "4s")), fan: "mcr_22"},
		{name: "all simples", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupChow, false, "2m", "3m", "4m"),
			mcrTestGroup(MCRGroupPung, true, "6p", "6p", "6p"),
			mcrTestGroup(MCRGroupPair, false, "8s", "8s"),
		), fan: "mcr_23"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := mcrFanOccurrenceCount(DetectMCRFans(test.context), test.fan); got != 1 {
				t.Fatalf("%s count = %d, want 1", test.fan, got)
			}
		})
	}
}

func TestMCRTwoPointFanNearMisses(t *testing.T) {
	tests := []struct {
		name    string
		context MCRFanContext
		fan     FanID
	}{
		{name: "dragon pair", context: fanContextWithGroups(mcrTestGroup(MCRGroupPair, false, "Z", "Z")), fan: "mcr_14"},
		{name: "open concealed hand", context: MCRFanContext{Decomposition: decompositionWithGroups(mcrTestGroup(MCRGroupChow, true, "1m", "2m", "3m")), WinType: WinDiscard}, fan: "mcr_17"},
		{name: "all chows honor pair", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupChow, false, "1m", "2m", "3m"),
			mcrTestGroup(MCRGroupChow, false, "4m", "5m", "6m"),
			mcrTestGroup(MCRGroupChow, false, "2p", "3p", "4p"),
			mcrTestGroup(MCRGroupChow, false, "6s", "7s", "8s"),
			mcrTestGroup(MCRGroupPair, false, "E", "E"),
		), fan: "mcr_18"},
		{name: "kong is not tile hog", context: fanContextWithGroups(mcrTestGroup(MCRGroupKong, false, "5m", "5m", "5m", "5m")), fan: "mcr_19"},
		{name: "discard completes second pung", context: MCRFanContext{Decomposition: decompositionWithGroups(
			mcrTestGroup(MCRGroupPung, false, "3m", "3m", "3m"),
			mcrTestGroup(MCRGroupPung, false, "7p", "7p", "7p"),
		), WinType: WinDiscard, WinningTile: mustFanTiles(t, "7p")[0]}, fan: "mcr_21"},
		{name: "terminal present", context: fanContextWithGroups(mcrTestGroup(MCRGroupPung, false, "1m", "1m", "1m")), fan: "mcr_23"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := mcrFanOccurrenceCount(DetectMCRFans(test.context), test.fan); got != 0 {
				t.Fatalf("%s count = %d, want 0", test.fan, got)
			}
		})
	}
}

func TestMCRFourSixEightPointFanDetectors(t *testing.T) {
	allPungs := []MCRGroup{
		mcrTestGroup(MCRGroupPung, false, "2m", "2m", "2m"),
		mcrTestGroup(MCRGroupPung, true, "4p", "4p", "4p"),
		mcrTestGroup(MCRGroupPung, false, "6s", "6s", "6s"),
		mcrTestGroup(MCRGroupKong, true, "E", "E", "E", "E"),
		mcrTestGroup(MCRGroupPair, false, "Z", "Z"),
	}
	tests := []struct {
		name    string
		context MCRFanContext
		fan     FanID
	}{
		{name: "outside hand", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupChow, false, "1m", "2m", "3m"),
			mcrTestGroup(MCRGroupPung, true, "9p", "9p", "9p"),
			mcrTestGroup(MCRGroupPair, false, "E", "E"),
		), fan: "mcr_24"},
		{name: "fully concealed hand", context: MCRFanContext{Decomposition: decompositionWithGroups(mcrTestGroup(MCRGroupChow, false, "2m", "3m", "4m")), WinType: WinSelfDraw}, fan: "mcr_25"},
		{name: "two melded kongs", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupKong, true, "3m", "3m", "3m", "3m"),
			mcrTestGroup(MCRGroupKong, true, "7p", "7p", "7p", "7p"),
		), fan: "mcr_26"},
		{name: "last of kind", context: MCRFanContext{LastOfKind: true}, fan: "mcr_27"},
		{name: "all pungs", context: fanContextWithGroups(allPungs...), fan: "mcr_28"},
		{name: "half flush", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupChow, false, "1m", "2m", "3m"),
			mcrTestGroup(MCRGroupPung, false, "E", "E", "E"),
		), fan: "mcr_29"},
		{name: "mixed shifted chows", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupChow, false, "1m", "2m", "3m"),
			mcrTestGroup(MCRGroupChow, false, "2p", "3p", "4p"),
			mcrTestGroup(MCRGroupChow, false, "3s", "4s", "5s"),
		), fan: "mcr_30"},
		{name: "all types", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupChow, false, "1m", "2m", "3m"),
			mcrTestGroup(MCRGroupPung, false, "4p", "4p", "4p"),
			mcrTestGroup(MCRGroupPung, false, "6s", "6s", "6s"),
			mcrTestGroup(MCRGroupPung, false, "E", "E", "E"),
			mcrTestGroup(MCRGroupPair, false, "Z", "Z"),
		), fan: "mcr_31"},
		{name: "melded hand", context: MCRFanContext{Decomposition: decompositionWithGroups(
			mcrTestGroup(MCRGroupChow, true, "1m", "2m", "3m"),
			mcrTestGroup(MCRGroupPung, true, "4p", "4p", "4p"),
			mcrTestGroup(MCRGroupPung, true, "6s", "6s", "6s"),
			mcrTestGroup(MCRGroupPung, true, "E", "E", "E"),
			mcrTestGroup(MCRGroupPair, false, "Z", "Z"),
		), WinType: WinDiscard}, fan: "mcr_32"},
		{name: "two dragon pungs", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupPung, false, "Z", "Z", "Z"),
			mcrTestGroup(MCRGroupPung, true, "F", "F", "F"),
		), fan: "mcr_33"},
		{name: "mixed straight", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupChow, false, "1m", "2m", "3m"),
			mcrTestGroup(MCRGroupChow, false, "4p", "5p", "6p"),
			mcrTestGroup(MCRGroupChow, false, "7s", "8s", "9s"),
		), fan: "mcr_34"},
		{name: "reversible tiles", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupPung, false, "1p", "1p", "1p"),
			mcrTestGroup(MCRGroupChow, false, "4s", "5s", "6s"),
			mcrTestGroup(MCRGroupPair, false, "B", "B"),
		), fan: "mcr_35"},
		{name: "mixed triple chow", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupChow, false, "3m", "4m", "5m"),
			mcrTestGroup(MCRGroupChow, false, "3p", "4p", "5p"),
			mcrTestGroup(MCRGroupChow, false, "3s", "4s", "5s"),
		), fan: "mcr_36"},
		{name: "mixed shifted pungs", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupPung, false, "3m", "3m", "3m"),
			mcrTestGroup(MCRGroupPung, false, "4p", "4p", "4p"),
			mcrTestGroup(MCRGroupPung, false, "5s", "5s", "5s"),
		), fan: "mcr_37"},
		{name: "two concealed kongs", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupKong, false, "3m", "3m", "3m", "3m"),
			mcrTestGroup(MCRGroupKong, false, "5p", "5p", "5p", "5p"),
		), fan: "mcr_38"},
		{name: "last tile draw", context: MCRFanContext{LastTileDraw: true}, fan: "mcr_39"},
		{name: "last tile claim", context: MCRFanContext{LastTileClaim: true}, fan: "mcr_40"},
		{name: "replacement draw", context: MCRFanContext{ReplacementDraw: true}, fan: "mcr_41"},
		{name: "robbing kong", context: MCRFanContext{RobbingKong: true}, fan: "mcr_42"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := mcrFanOccurrenceCount(DetectMCRFans(test.context), test.fan); got != 1 {
				t.Fatalf("%s count = %d, want 1", test.fan, got)
			}
		})
	}
}

func TestMCRFourSixEightPointFanNearMisses(t *testing.T) {
	tests := []struct {
		name    string
		context MCRFanContext
		fan     FanID
	}{
		{name: "outside group without terminal", context: fanContextWithGroups(mcrTestGroup(MCRGroupChow, false, "2m", "3m", "4m")), fan: "mcr_24"},
		{name: "full flush is not half flush", context: fanContextWithGroups(mcrTestGroup(MCRGroupChow, false, "1m", "2m", "3m")), fan: "mcr_29"},
		{name: "all types missing dragon", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupChow, false, "1m", "2m", "3m"),
			mcrTestGroup(MCRGroupPung, false, "4p", "4p", "4p"),
			mcrTestGroup(MCRGroupPung, false, "6s", "6s", "6s"),
			mcrTestGroup(MCRGroupPung, false, "E", "E", "E"),
		), fan: "mcr_31"},
		{name: "reversible contains seven dots", context: fanContextWithGroups(mcrTestGroup(MCRGroupPung, false, "7p", "7p", "7p")), fan: "mcr_35"},
		{name: "mixed straight repeats suit", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupChow, false, "1m", "2m", "3m"),
			mcrTestGroup(MCRGroupChow, false, "4m", "5m", "6m"),
			mcrTestGroup(MCRGroupChow, false, "7s", "8s", "9s"),
		), fan: "mcr_34"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := mcrFanOccurrenceCount(DetectMCRFans(test.context), test.fan); got != 0 {
				t.Fatalf("%s count = %d, want 0", test.fan, got)
			}
		})
	}
}

func fanContextWithGroups(groups ...MCRGroup) MCRFanContext {
	return MCRFanContext{Decomposition: decompositionWithGroups(groups...)}
}

func decompositionWithGroups(groups ...MCRGroup) MCRDecomposition {
	var tiles []Tile
	for _, group := range groups {
		tiles = append(tiles, group.Tiles...)
	}
	return MCRDecomposition{Kind: MCRShapeStandard, Groups: groups, Tiles: tiles}
}

func mcrTestGroup(kind MCRGroupKind, open bool, values ...string) MCRGroup {
	tiles := make([]Tile, len(values))
	for i, value := range values {
		tile, ok := ParseTile(value)
		if !ok {
			panic("invalid test tile " + value)
		}
		tiles[i] = tile
	}
	return MCRGroup{Kind: kind, Tiles: tiles, Open: open}
}

func mustFanTiles(t *testing.T, values ...string) []Tile {
	t.Helper()
	return parseFixtureTiles(t, values)
}

func mcrFanOccurrenceCount(values []MCRFanOccurrence, fan FanID) int {
	count := 0
	for _, value := range values {
		if value.ID == fan {
			count += value.Count
		}
	}
	return count
}
