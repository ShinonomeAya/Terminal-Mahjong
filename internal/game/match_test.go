package game

import "testing"

func TestNewMatchCarriesModeConfigAndInitialPoints(t *testing.T) {
	tests := []struct {
		name   string
		mode   RuleMode
		config RuleConfig
		points [4]int
	}{
		{name: "compatibility", mode: ModeCompatibility, config: RuleConfig{}},
		{name: "MCR", mode: ModeMCR, config: DefaultRuleConfig(ModeMCR)},
		{name: "riichi", mode: ModeRiichi, config: DefaultRuleConfig(ModeRiichi), points: [4]int{25000, 25000, 25000, 25000}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			match, err := NewMatch(9, NewCompatibilityRuleSet(test.mode, test.config))
			if err != nil {
				t.Fatalf("NewMatch() error = %v", err)
			}
			if match.Mode != test.mode || match.RuleConfig != test.config || match.Points != test.points {
				t.Fatalf("match = %#v, want mode=%q config=%#v points=%v", match, test.mode, test.config, test.points)
			}
			if match.Round == nil || match.Round.Mode != test.mode {
				t.Fatalf("round = %#v, want mode %q", match.Round, test.mode)
			}
		})
	}
}

func TestMatchDelegatesDrawAndCommand(t *testing.T) {
	match, err := NewMatch(9, NewCompatibilityRuleSet(ModeCompatibility, RuleConfig{}))
	if err != nil {
		t.Fatal(err)
	}
	startWall := len(match.Round.Wall)

	if _, ok := match.EnsureCurrentTurnDraw(); !ok {
		t.Fatal("expected delegated draw")
	}
	result := match.ApplyCommand(GameCommand{PlayerID: "0", Kind: CommandDiscard, TileIndex: 0})

	if !result.OK || len(match.Round.Wall) != startWall-1 || len(match.Round.Players[0].Discards) != 1 {
		t.Fatalf("result=%#v wall=%d discards=%v", result, len(match.Round.Wall), match.Round.Players[0].Discards)
	}
}

func TestMatchMarksCompleteAfterRoundEnds(t *testing.T) {
	match, err := NewMatch(9, NewCompatibilityRuleSet(ModeCompatibility, RuleConfig{}))
	if err != nil {
		t.Fatal(err)
	}
	match.Round.Players[0].Hand = mustTiles(t,
		"1m", "2m", "3m",
		"4m", "5m", "6m",
		"2p", "3p", "4p",
		"7s", "7s", "7s",
		"E", "E",
	)

	result := match.ApplyCommand(GameCommand{PlayerID: "0", Kind: CommandWin})

	if !result.OK || !match.Complete {
		t.Fatalf("result=%#v complete=%v", result, match.Complete)
	}
}

func TestMatchSnapshotIsCopied(t *testing.T) {
	config := DefaultRuleConfig(ModeRiichi)
	match, err := NewMatch(9, NewCompatibilityRuleSet(ModeRiichi, config))
	if err != nil {
		t.Fatal(err)
	}

	snapshot := match.Snapshot()
	snapshot.RuleConfig.Riichi.StartingPoints = 1000
	snapshot.Round.Players[0].Hand[0] = mustTile(t, "E")

	if match.RuleConfig.Riichi.StartingPoints != 25000 {
		t.Fatal("snapshot config should not mutate match")
	}
	if match.Round.Players[0].Hand[0] == mustTile(t, "E") {
		t.Fatal("snapshot round should not mutate match")
	}
}
