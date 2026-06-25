package tui

import "mahjong/internal/game"

func newStartedGame() *game.Game {
	return newStartedGameWithRules(game.ModeCompatibility, game.RuleConfig{})
}

func newStartedGameWithRules(mode game.RuleMode, config game.RuleConfig) *game.Game {
	var rules game.RuleSet
	switch mode {
	case game.ModeMCR:
		rules = game.NewMCRRuleSet(config.MCR)
	case game.ModeRiichi:
		rules = game.NewRiichiRuleSet(config.Riichi)
	default:
		rules = game.NewCompatibilityRuleSet(game.ModeCompatibility, game.RuleConfig{})
	}
	g, err := game.NewGameWithRules(0, rules)
	if err != nil {
		panic(err)
	}
	g.StartHumanTurn()
	return g
}

func selectedRuleConfig(m Model) game.RuleConfig {
	switch m.SelectedMode {
	case game.ModeMCR:
		return game.DefaultRuleConfig(game.ModeMCR)
	case game.ModeRiichi:
		config := game.DefaultRuleConfig(game.ModeRiichi)
		config.Riichi.RedFives = m.SelectedRiichiRedFives
		return config
	default:
		return game.RuleConfig{}
	}
}
