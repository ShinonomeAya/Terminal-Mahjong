package tui

import "mahjong/internal/game"

func newStartedGame() *game.Game {
	return newStartedGameWithRules(game.ModeCompatibility, game.RuleConfig{})
}

func newStartedGameWithRules(mode game.RuleMode, config game.RuleConfig) *game.Game {
	return newStartedMatchWithRules(mode, config).Round
}

func newStartedMatchWithRules(mode game.RuleMode, config game.RuleConfig) *game.Match {
	var rules game.RuleSet
	switch mode {
	case game.ModeMCR:
		rules = game.NewMCRRuleSet(config.MCR)
	case game.ModeRiichi:
		rules = game.NewRiichiRuleSet(config.Riichi)
	default:
		rules = game.NewCompatibilityRuleSet(game.ModeCompatibility, game.RuleConfig{})
	}
	match, err := game.NewMatch(0, rules)
	if err != nil {
		panic(err)
	}
	match.EnsureCurrentTurnDraw()
	return match
}

func syncLocalRound(m Model) Model {
	if m.LocalMatch != nil {
		m.Game = m.LocalMatch.Round
	}
	return m
}

func restartLocalMatch(m Model) Model {
	if m.LocalMatch == nil {
		m.Game = newStartedGame()
		return m
	}
	m.LocalMatch = newStartedMatchWithRules(m.LocalMatch.Mode, m.LocalMatch.RuleConfig)
	m.LastReplayPath = ""
	return syncLocalRound(m)
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
