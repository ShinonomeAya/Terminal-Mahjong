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
