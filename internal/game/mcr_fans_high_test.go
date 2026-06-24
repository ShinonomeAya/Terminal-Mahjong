package game

import "testing"

func TestMCRTwelveToTwentyFourPointFanDetectors(t *testing.T) {
	tests := []struct {
		name    string
		context MCRFanContext
		fan     FanID
	}{
		{name: "lesser honors and knitted tiles", context: MCRFanContext{Decomposition: MCRDecomposition{Kind: MCRShapeLesserHonorsKnitted}}, fan: "mcr_44"},
		{name: "knitted straight", context: MCRFanContext{Decomposition: MCRDecomposition{Kind: MCRShapeKnittedStraight}}, fan: "mcr_45"},
		{name: "upper four", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupChow, false, "6m", "7m", "8m"),
			mcrTestGroup(MCRGroupPung, false, "9p", "9p", "9p"),
		), fan: "mcr_46"},
		{name: "lower four", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupChow, false, "2m", "3m", "4m"),
			mcrTestGroup(MCRGroupPung, false, "1s", "1s", "1s"),
		), fan: "mcr_47"},
		{name: "big three winds", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupPung, false, "E", "E", "E"),
			mcrTestGroup(MCRGroupPung, true, "S", "S", "S"),
			mcrTestGroup(MCRGroupKong, false, "W", "W", "W", "W"),
		), fan: "mcr_48"},
		{name: "pure straight", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupChow, false, "1m", "2m", "3m"),
			mcrTestGroup(MCRGroupChow, false, "4m", "5m", "6m"),
			mcrTestGroup(MCRGroupChow, false, "7m", "8m", "9m"),
		), fan: "mcr_49"},
		{name: "three suited terminal chows", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupChow, false, "1m", "2m", "3m"),
			mcrTestGroup(MCRGroupChow, false, "7m", "8m", "9m"),
			mcrTestGroup(MCRGroupChow, false, "1p", "2p", "3p"),
			mcrTestGroup(MCRGroupChow, false, "7p", "8p", "9p"),
			mcrTestGroup(MCRGroupPair, false, "5s", "5s"),
		), fan: "mcr_50"},
		{name: "pure shifted chows by one", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupChow, false, "1s", "2s", "3s"),
			mcrTestGroup(MCRGroupChow, false, "2s", "3s", "4s"),
			mcrTestGroup(MCRGroupChow, false, "3s", "4s", "5s"),
		), fan: "mcr_51"},
		{name: "all fives", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupChow, false, "3m", "4m", "5m"),
			mcrTestGroup(MCRGroupChow, false, "4p", "5p", "6p"),
			mcrTestGroup(MCRGroupPung, false, "5s", "5s", "5s"),
			mcrTestGroup(MCRGroupPair, false, "5m", "5m"),
		), fan: "mcr_52"},
		{name: "triple pung", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupPung, false, "4m", "4m", "4m"),
			mcrTestGroup(MCRGroupPung, true, "4p", "4p", "4p"),
			mcrTestGroup(MCRGroupKong, false, "4s", "4s", "4s", "4s"),
		), fan: "mcr_53"},
		{name: "three concealed pungs", context: MCRFanContext{Decomposition: decompositionWithGroups(
			mcrTestGroup(MCRGroupPung, false, "2m", "2m", "2m"),
			mcrTestGroup(MCRGroupPung, false, "5p", "5p", "5p"),
			mcrTestGroup(MCRGroupKong, false, "8s", "8s", "8s", "8s"),
		), WinType: WinSelfDraw}, fan: "mcr_54"},
		{name: "seven pairs", context: MCRFanContext{Decomposition: MCRDecomposition{Kind: MCRShapeSevenPairs}}, fan: "mcr_55"},
		{name: "greater honors and knitted tiles", context: MCRFanContext{Decomposition: MCRDecomposition{Kind: MCRShapeGreaterHonorsKnitted}}, fan: "mcr_56"},
		{name: "all even pungs", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupPung, false, "2m", "2m", "2m"),
			mcrTestGroup(MCRGroupPung, true, "4p", "4p", "4p"),
			mcrTestGroup(MCRGroupKong, false, "6s", "6s", "6s", "6s"),
			mcrTestGroup(MCRGroupPung, false, "8m", "8m", "8m"),
			mcrTestGroup(MCRGroupPair, false, "2p", "2p"),
		), fan: "mcr_57"},
		{name: "full flush", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupChow, false, "1p", "2p", "3p"),
			mcrTestGroup(MCRGroupPung, false, "8p", "8p", "8p"),
		), fan: "mcr_58"},
		{name: "pure triple chow", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupChow, false, "3m", "4m", "5m"),
			mcrTestGroup(MCRGroupChow, true, "3m", "4m", "5m"),
			mcrTestGroup(MCRGroupChow, false, "3m", "4m", "5m"),
		), fan: "mcr_59"},
		{name: "pure shifted pungs", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupPung, false, "3s", "3s", "3s"),
			mcrTestGroup(MCRGroupPung, true, "4s", "4s", "4s"),
			mcrTestGroup(MCRGroupKong, false, "5s", "5s", "5s", "5s"),
		), fan: "mcr_60"},
		{name: "upper tiles", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupChow, false, "7m", "8m", "9m"),
			mcrTestGroup(MCRGroupPung, false, "8p", "8p", "8p"),
		), fan: "mcr_61"},
		{name: "middle tiles", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupChow, false, "4m", "5m", "6m"),
			mcrTestGroup(MCRGroupPung, false, "5p", "5p", "5p"),
		), fan: "mcr_62"},
		{name: "lower tiles", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupChow, false, "1s", "2s", "3s"),
			mcrTestGroup(MCRGroupPung, false, "2p", "2p", "2p"),
		), fan: "mcr_63"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := mcrFanOccurrenceCount(DetectMCRFans(test.context), test.fan); got != 1 {
				t.Fatalf("%s count = %d, want 1", test.fan, got)
			}
		})
	}
}

func TestMCRTwelveToTwentyFourPointFanNearMisses(t *testing.T) {
	tests := []struct {
		name    string
		context MCRFanContext
		fan     FanID
	}{
		{name: "upper four contains five", context: fanContextWithGroups(mcrTestGroup(MCRGroupChow, false, "5m", "6m", "7m")), fan: "mcr_46"},
		{name: "lower four contains five", context: fanContextWithGroups(mcrTestGroup(MCRGroupChow, false, "3m", "4m", "5m")), fan: "mcr_47"},
		{name: "only two wind pungs", context: fanContextWithGroups(mcrTestGroup(MCRGroupPung, false, "E", "E", "E"), mcrTestGroup(MCRGroupPung, false, "S", "S", "S")), fan: "mcr_48"},
		{name: "pure straight mixed suits", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupChow, false, "1m", "2m", "3m"),
			mcrTestGroup(MCRGroupChow, false, "4p", "5p", "6p"),
			mcrTestGroup(MCRGroupChow, false, "7m", "8m", "9m"),
		), fan: "mcr_49"},
		{name: "terminal chows pair uses terminal suit", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupChow, false, "1m", "2m", "3m"),
			mcrTestGroup(MCRGroupChow, false, "7m", "8m", "9m"),
			mcrTestGroup(MCRGroupChow, false, "1p", "2p", "3p"),
			mcrTestGroup(MCRGroupChow, false, "7p", "8p", "9p"),
			mcrTestGroup(MCRGroupPair, false, "5m", "5m"),
		), fan: "mcr_50"},
		{name: "shifted chows unequal steps", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupChow, false, "1s", "2s", "3s"),
			mcrTestGroup(MCRGroupChow, false, "2s", "3s", "4s"),
			mcrTestGroup(MCRGroupChow, false, "4s", "5s", "6s"),
		), fan: "mcr_51"},
		{name: "group without five", context: fanContextWithGroups(mcrTestGroup(MCRGroupChow, false, "1m", "2m", "3m")), fan: "mcr_52"},
		{name: "triple pung mixed ranks", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupPung, false, "4m", "4m", "4m"),
			mcrTestGroup(MCRGroupPung, false, "4p", "4p", "4p"),
			mcrTestGroup(MCRGroupPung, false, "5s", "5s", "5s"),
		), fan: "mcr_53"},
		{name: "third pung open", context: MCRFanContext{Decomposition: decompositionWithGroups(
			mcrTestGroup(MCRGroupPung, false, "2m", "2m", "2m"),
			mcrTestGroup(MCRGroupPung, false, "5p", "5p", "5p"),
			mcrTestGroup(MCRGroupPung, true, "8s", "8s", "8s"),
		), WinType: WinSelfDraw}, fan: "mcr_54"},
		{name: "even pungs includes odd pair", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupPung, false, "2m", "2m", "2m"),
			mcrTestGroup(MCRGroupPung, false, "4p", "4p", "4p"),
			mcrTestGroup(MCRGroupPung, false, "6s", "6s", "6s"),
			mcrTestGroup(MCRGroupPung, false, "8m", "8m", "8m"),
			mcrTestGroup(MCRGroupPair, false, "3p", "3p"),
		), fan: "mcr_57"},
		{name: "full flush has honor", context: fanContextWithGroups(mcrTestGroup(MCRGroupChow, false, "1p", "2p", "3p"), mcrTestGroup(MCRGroupPair, false, "E", "E")), fan: "mcr_58"},
		{name: "pure triple chow only twice", context: fanContextWithGroups(mcrTestGroup(MCRGroupChow, false, "3m", "4m", "5m"), mcrTestGroup(MCRGroupChow, false, "3m", "4m", "5m")), fan: "mcr_59"},
		{name: "pure shifted pungs mixed suits", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupPung, false, "3s", "3s", "3s"),
			mcrTestGroup(MCRGroupPung, false, "4s", "4s", "4s"),
			mcrTestGroup(MCRGroupPung, false, "5p", "5p", "5p"),
		), fan: "mcr_60"},
		{name: "upper tiles includes six", context: fanContextWithGroups(mcrTestGroup(MCRGroupPung, false, "6m", "6m", "6m")), fan: "mcr_61"},
		{name: "middle tiles includes seven", context: fanContextWithGroups(mcrTestGroup(MCRGroupPung, false, "7m", "7m", "7m")), fan: "mcr_62"},
		{name: "lower tiles includes four", context: fanContextWithGroups(mcrTestGroup(MCRGroupPung, false, "4m", "4m", "4m")), fan: "mcr_63"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := mcrFanOccurrenceCount(DetectMCRFans(test.context), test.fan); got != 0 {
				t.Fatalf("%s count = %d, want 0", test.fan, got)
			}
		})
	}
}

func TestMCRThirtyTwoToEightyEightPointFanDetectors(t *testing.T) {
	tests := []struct {
		name    string
		context MCRFanContext
		fan     FanID
	}{
		{name: "four shifted chows", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupChow, false, "1m", "2m", "3m"),
			mcrTestGroup(MCRGroupChow, false, "2m", "3m", "4m"),
			mcrTestGroup(MCRGroupChow, false, "3m", "4m", "5m"),
			mcrTestGroup(MCRGroupChow, false, "4m", "5m", "6m"),
		), fan: "mcr_64"},
		{name: "three kongs", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupKong, true, "2m", "2m", "2m", "2m"),
			mcrTestGroup(MCRGroupKong, false, "5p", "5p", "5p", "5p"),
			mcrTestGroup(MCRGroupKong, true, "E", "E", "E", "E"),
		), fan: "mcr_65"},
		{name: "all terminals and honors", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupPung, false, "1m", "1m", "1m"),
			mcrTestGroup(MCRGroupKong, true, "9p", "9p", "9p", "9p"),
			mcrTestGroup(MCRGroupPung, false, "E", "E", "E"),
			mcrTestGroup(MCRGroupPair, false, "Z", "Z"),
		), fan: "mcr_66"},
		{name: "quadruple chow", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupChow, false, "3p", "4p", "5p"),
			mcrTestGroup(MCRGroupChow, false, "3p", "4p", "5p"),
			mcrTestGroup(MCRGroupChow, true, "3p", "4p", "5p"),
			mcrTestGroup(MCRGroupChow, false, "3p", "4p", "5p"),
		), fan: "mcr_67"},
		{name: "four pure shifted pungs", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupPung, false, "2s", "2s", "2s"),
			mcrTestGroup(MCRGroupPung, true, "3s", "3s", "3s"),
			mcrTestGroup(MCRGroupKong, false, "4s", "4s", "4s", "4s"),
			mcrTestGroup(MCRGroupPung, false, "5s", "5s", "5s"),
		), fan: "mcr_68"},
		{name: "all terminals", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupPung, false, "1m", "1m", "1m"),
			mcrTestGroup(MCRGroupKong, true, "9p", "9p", "9p", "9p"),
			mcrTestGroup(MCRGroupPair, false, "1s", "1s"),
		), fan: "mcr_69"},
		{name: "all honors", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupPung, false, "E", "E", "E"),
			mcrTestGroup(MCRGroupKong, true, "S", "S", "S", "S"),
			mcrTestGroup(MCRGroupPair, false, "Z", "Z"),
		), fan: "mcr_70"},
		{name: "little four winds", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupPung, false, "E", "E", "E"),
			mcrTestGroup(MCRGroupPung, true, "S", "S", "S"),
			mcrTestGroup(MCRGroupKong, false, "W", "W", "W", "W"),
			mcrTestGroup(MCRGroupPair, false, "N", "N"),
		), fan: "mcr_71"},
		{name: "little three dragons", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupPung, false, "Z", "Z", "Z"),
			mcrTestGroup(MCRGroupKong, true, "F", "F", "F", "F"),
			mcrTestGroup(MCRGroupPair, false, "B", "B"),
		), fan: "mcr_72"},
		{name: "four concealed pungs", context: MCRFanContext{Decomposition: decompositionWithGroups(
			mcrTestGroup(MCRGroupPung, false, "2m", "2m", "2m"),
			mcrTestGroup(MCRGroupPung, false, "4p", "4p", "4p"),
			mcrTestGroup(MCRGroupKong, false, "6s", "6s", "6s", "6s"),
			mcrTestGroup(MCRGroupPung, false, "8m", "8m", "8m"),
			mcrTestGroup(MCRGroupPair, false, "E", "E"),
		), WinType: WinSelfDraw}, fan: "mcr_73"},
		{name: "pure terminal chows", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupChow, false, "1s", "2s", "3s"),
			mcrTestGroup(MCRGroupChow, false, "1s", "2s", "3s"),
			mcrTestGroup(MCRGroupChow, false, "7s", "8s", "9s"),
			mcrTestGroup(MCRGroupChow, false, "7s", "8s", "9s"),
			mcrTestGroup(MCRGroupPair, false, "5s", "5s"),
		), fan: "mcr_74"},
		{name: "big four winds", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupPung, false, "E", "E", "E"),
			mcrTestGroup(MCRGroupPung, true, "S", "S", "S"),
			mcrTestGroup(MCRGroupKong, false, "W", "W", "W", "W"),
			mcrTestGroup(MCRGroupPung, false, "N", "N", "N"),
		), fan: "mcr_75"},
		{name: "big three dragons", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupPung, false, "Z", "Z", "Z"),
			mcrTestGroup(MCRGroupPung, true, "F", "F", "F"),
			mcrTestGroup(MCRGroupKong, false, "B", "B", "B", "B"),
		), fan: "mcr_76"},
		{name: "all green", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupChow, false, "2s", "3s", "4s"),
			mcrTestGroup(MCRGroupPung, false, "6s", "6s", "6s"),
			mcrTestGroup(MCRGroupPair, false, "F", "F"),
		), fan: "mcr_77"},
		{name: "nine gates", context: MCRFanContext{Decomposition: MCRDecomposition{Kind: MCRShapeNineGates}}, fan: "mcr_78"},
		{name: "four kongs", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupKong, true, "2m", "2m", "2m", "2m"),
			mcrTestGroup(MCRGroupKong, false, "5p", "5p", "5p", "5p"),
			mcrTestGroup(MCRGroupKong, true, "8s", "8s", "8s", "8s"),
			mcrTestGroup(MCRGroupKong, false, "E", "E", "E", "E"),
		), fan: "mcr_79"},
		{name: "seven shifted pairs", context: MCRFanContext{Decomposition: MCRDecomposition{Kind: MCRShapeSevenShiftedPairs}}, fan: "mcr_80"},
		{name: "thirteen orphans", context: MCRFanContext{Decomposition: MCRDecomposition{Kind: MCRShapeThirteenOrphans}}, fan: "mcr_81"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := mcrFanOccurrenceCount(DetectMCRFans(test.context), test.fan); got != 1 {
				t.Fatalf("%s count = %d, want 1", test.fan, got)
			}
		})
	}
}

func TestMCRThirtyTwoToEightyEightPointFanNearMisses(t *testing.T) {
	tests := []struct {
		name    string
		context MCRFanContext
		fan     FanID
	}{
		{name: "four shifted chows unequal steps", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupChow, false, "1m", "2m", "3m"),
			mcrTestGroup(MCRGroupChow, false, "2m", "3m", "4m"),
			mcrTestGroup(MCRGroupChow, false, "4m", "5m", "6m"),
			mcrTestGroup(MCRGroupChow, false, "5m", "6m", "7m"),
		), fan: "mcr_64"},
		{name: "only two kongs", context: fanContextWithGroups(mcrTestGroup(MCRGroupKong, true, "2m", "2m", "2m", "2m"), mcrTestGroup(MCRGroupKong, false, "5p", "5p", "5p", "5p")), fan: "mcr_65"},
		{name: "terminals without honors", context: fanContextWithGroups(mcrTestGroup(MCRGroupPung, false, "1m", "1m", "1m"), mcrTestGroup(MCRGroupPair, false, "9p", "9p")), fan: "mcr_66"},
		{name: "quadruple chow only three", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupChow, false, "3p", "4p", "5p"),
			mcrTestGroup(MCRGroupChow, false, "3p", "4p", "5p"),
			mcrTestGroup(MCRGroupChow, false, "3p", "4p", "5p"),
		), fan: "mcr_67"},
		{name: "four shifted pungs skips rank", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupPung, false, "2s", "2s", "2s"),
			mcrTestGroup(MCRGroupPung, false, "3s", "3s", "3s"),
			mcrTestGroup(MCRGroupPung, false, "5s", "5s", "5s"),
			mcrTestGroup(MCRGroupPung, false, "6s", "6s", "6s"),
		), fan: "mcr_68"},
		{name: "all terminals has honor", context: fanContextWithGroups(mcrTestGroup(MCRGroupPung, false, "1m", "1m", "1m"), mcrTestGroup(MCRGroupPair, false, "E", "E")), fan: "mcr_69"},
		{name: "all honors has terminal", context: fanContextWithGroups(mcrTestGroup(MCRGroupPung, false, "E", "E", "E"), mcrTestGroup(MCRGroupPair, false, "1m", "1m")), fan: "mcr_70"},
		{name: "little four winds missing pair", context: fanContextWithGroups(mcrTestGroup(MCRGroupPung, false, "E", "E", "E"), mcrTestGroup(MCRGroupPung, false, "S", "S", "S"), mcrTestGroup(MCRGroupPung, false, "W", "W", "W"), mcrTestGroup(MCRGroupPair, false, "Z", "Z")), fan: "mcr_71"},
		{name: "little three dragons missing pair", context: fanContextWithGroups(mcrTestGroup(MCRGroupPung, false, "Z", "Z", "Z"), mcrTestGroup(MCRGroupPung, false, "F", "F", "F"), mcrTestGroup(MCRGroupPair, false, "E", "E")), fan: "mcr_72"},
		{name: "fourth concealed pung completed by discard", context: MCRFanContext{Decomposition: decompositionWithGroups(
			mcrTestGroup(MCRGroupPung, false, "2m", "2m", "2m"),
			mcrTestGroup(MCRGroupPung, false, "4p", "4p", "4p"),
			mcrTestGroup(MCRGroupPung, false, "6s", "6s", "6s"),
			mcrTestGroup(MCRGroupPung, false, "8m", "8m", "8m"),
		), WinType: WinDiscard, WinningTile: mustFanTiles(t, "8m")[0]}, fan: "mcr_73"},
		{name: "pure terminal chows mixed suit", context: fanContextWithGroups(
			mcrTestGroup(MCRGroupChow, false, "1s", "2s", "3s"),
			mcrTestGroup(MCRGroupChow, false, "1s", "2s", "3s"),
			mcrTestGroup(MCRGroupChow, false, "7s", "8s", "9s"),
			mcrTestGroup(MCRGroupChow, false, "7p", "8p", "9p"),
			mcrTestGroup(MCRGroupPair, false, "5s", "5s"),
		), fan: "mcr_74"},
		{name: "big four winds only three", context: fanContextWithGroups(mcrTestGroup(MCRGroupPung, false, "E", "E", "E"), mcrTestGroup(MCRGroupPung, false, "S", "S", "S"), mcrTestGroup(MCRGroupPung, false, "W", "W", "W")), fan: "mcr_75"},
		{name: "big three dragons only two", context: fanContextWithGroups(mcrTestGroup(MCRGroupPung, false, "Z", "Z", "Z"), mcrTestGroup(MCRGroupPung, false, "F", "F", "F")), fan: "mcr_76"},
		{name: "all green contains five bamboo", context: fanContextWithGroups(mcrTestGroup(MCRGroupPung, false, "5s", "5s", "5s")), fan: "mcr_77"},
		{name: "four kongs only three", context: fanContextWithGroups(mcrTestGroup(MCRGroupKong, true, "2m", "2m", "2m", "2m"), mcrTestGroup(MCRGroupKong, false, "5p", "5p", "5p", "5p"), mcrTestGroup(MCRGroupKong, true, "E", "E", "E", "E")), fan: "mcr_79"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := mcrFanOccurrenceCount(DetectMCRFans(test.context), test.fan); got != 0 {
				t.Fatalf("%s count = %d, want 0", test.fan, got)
			}
		})
	}
}

func TestMCRHighFanDetectorsDoNotSuppressLowerFans(t *testing.T) {
	tests := []struct {
		name    string
		context MCRFanContext
		fan     FanID
	}{
		{name: "greater honors also matches lesser honors", context: MCRFanContext{Decomposition: MCRDecomposition{Kind: MCRShapeGreaterHonorsKnitted}}, fan: "mcr_44"},
		{name: "seven shifted pairs also matches seven pairs", context: MCRFanContext{Decomposition: MCRDecomposition{Kind: MCRShapeSevenShiftedPairs}}, fan: "mcr_55"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := mcrFanOccurrenceCount(DetectMCRFans(test.context), test.fan); got != 1 {
				t.Fatalf("%s count = %d, want 1", test.fan, got)
			}
		})
	}
}
