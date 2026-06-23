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
