package game

import "testing"

func TestDefaultRuleConfig(t *testing.T) {
	riichi := DefaultRuleConfig(ModeRiichi).Riichi
	if riichi.StartingPoints != 25000 || riichi.MatchLength != MatchEastSouth || !riichi.OpenTanyao || riichi.RedFives != 3 {
		t.Fatalf("riichi defaults = %#v", riichi)
	}
	if got := DefaultRuleConfig(ModeMCR).MCR.MinimumPoints; got != 8 {
		t.Fatalf("MCR minimum = %d, want 8", got)
	}
	if got := DefaultRuleConfig(ModeCompatibility); got != (RuleConfig{}) {
		t.Fatalf("compatibility defaults = %#v, want empty", got)
	}
}

func TestRuleConfigRejectsWrongPayload(t *testing.T) {
	config := DefaultRuleConfig(ModeRiichi)
	config.MCR.MinimumPoints = 8
	if err := config.Validate(ModeRiichi); err == nil {
		t.Fatal("expected mixed-mode config rejection")
	}
}

func TestRuleConfigRejectsUnsupportedRiichiValues(t *testing.T) {
	tests := []struct {
		name   string
		change func(*RuleConfig)
	}{
		{name: "points", change: func(c *RuleConfig) { c.Riichi.StartingPoints = 999 }},
		{name: "length", change: func(c *RuleConfig) { c.Riichi.MatchLength = "east_only" }},
		{name: "red fives", change: func(c *RuleConfig) { c.Riichi.RedFives = 2 }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := DefaultRuleConfig(ModeRiichi)
			test.change(&config)
			if err := config.Validate(ModeRiichi); err == nil {
				t.Fatalf("expected invalid config rejection: %#v", config)
			}
		})
	}
}

func TestRuleConfigRejectsNonStandardMCRMinimum(t *testing.T) {
	config := DefaultRuleConfig(ModeMCR)
	config.MCR.MinimumPoints = 6
	if err := config.Validate(ModeMCR); err == nil {
		t.Fatal("expected MCR minimum rejection")
	}
}

func TestParseRuleMode(t *testing.T) {
	for _, text := range []string{"compatibility", "mcr", "riichi"} {
		mode, err := ParseRuleMode(text)
		if err != nil || string(mode) != text {
			t.Fatalf("ParseRuleMode(%q) = %q, %v", text, mode, err)
		}
	}
	if _, err := ParseRuleMode("regional"); err == nil {
		t.Fatal("expected unknown mode error")
	}
}

func TestCompatibilityRuleLegalActionsIncludeIndexedDiscardsAndQuit(t *testing.T) {
	g := NewGame(9)
	rules := NewCompatibilityRuleSet(ModeCompatibility, RuleConfig{})

	actions := rules.LegalActions(g, "0")

	discards := 0
	quit := false
	for _, action := range actions {
		switch action.Kind {
		case CommandDiscard:
			if action.TileIndex != discards {
				t.Fatalf("discard action %d has index %d", discards, action.TileIndex)
			}
			discards++
		case CommandQuit:
			quit = true
		}
	}
	if discards != len(g.Players[0].Hand) || !quit {
		t.Fatalf("actions = %#v, want %d discards and quit", actions, len(g.Players[0].Hand))
	}
}

func TestCompatibilityRuleLegalActionsIncludeAvailableWinAndKong(t *testing.T) {
	rules := NewCompatibilityRuleSet(ModeCompatibility, RuleConfig{})

	winGame := NewGame(9)
	winGame.Players[0].Hand = mustTiles(t,
		"1m", "2m", "3m",
		"4m", "5m", "6m",
		"2p", "3p", "4p",
		"7s", "7s", "7s",
		"E", "E",
	)
	if !hasLegalAction(rules.LegalActions(winGame, "0"), CommandWin, "") {
		t.Fatal("complete hand should expose win")
	}

	kongGame := NewGame(9)
	kongGame.Players[0].Hand = mustTiles(t, "1m", "1m", "1m", "1m", "2m")
	if !hasLegalAction(rules.LegalActions(kongGame, "0"), CommandKong, "1m") {
		t.Fatal("four identical tiles should expose concealed kong")
	}
}

func TestCompatibilityRuleLegalActionsLimitClaimantToActiveClaim(t *testing.T) {
	g := gameWithPendingPong(t)
	rules := NewCompatibilityRuleSet(ModeCompatibility, RuleConfig{})

	actions := rules.LegalActions(g, "1")

	if len(actions) != 2 || actions[0].Kind != CommandPass || actions[1].Kind != CommandPong {
		t.Fatalf("claim actions = %#v, want pass and pong", actions)
	}
	if got := rules.LegalActions(g, "2"); len(got) != 0 {
		t.Fatalf("non-current actions = %#v, want none", got)
	}
}

func TestCompatibilityRuleRejectsIllegalCommandWithoutMutation(t *testing.T) {
	g := NewGame(9)
	before := g.Snapshot()

	result := g.ApplyCommand(GameCommand{PlayerID: "0", Kind: CommandKong, Tile: "1m"})

	if result.OK || result.Error != "command is not legal" {
		t.Fatalf("result = %#v, want legal-action rejection", result)
	}
	after := g.Snapshot()
	if len(after.Events) != len(before.Events) || after.WallCount != before.WallCount || len(after.Players[0].Hand) != len(before.Players[0].Hand) {
		t.Fatalf("illegal command mutated state: before=%#v after=%#v", before, after)
	}
}

func TestCompatibilitySetupRemainsDeterministic(t *testing.T) {
	before := NewGame(31)
	after, err := NewGameWithRules(31, NewCompatibilityRuleSet(ModeCompatibility, RuleConfig{}))
	if err != nil {
		t.Fatal(err)
	}
	if before.ShuffleProof != after.ShuffleProof || FormatTiles(before.Players[0].Hand) != FormatTiles(after.Players[0].Hand) {
		t.Fatalf("compatibility setup changed: before=%#v after=%#v", before.ShuffleProof, after.ShuffleProof)
	}
}

func hasLegalAction(actions []LegalAction, kind CommandKind, tile string) bool {
	for _, action := range actions {
		if action.Kind == kind && action.Tile == tile {
			return true
		}
	}
	return false
}
