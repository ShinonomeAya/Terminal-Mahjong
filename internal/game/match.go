package game

import "fmt"

type Match struct {
	Mode              RuleMode
	RuleConfig        RuleConfig
	Points            [4]int
	Dealer            int
	RoundNumber       int
	Complete          bool
	Round             *Game
	LastMCRSettlement *MCRSettlement
	rules             RuleSet
}

type MatchSnapshot struct {
	Mode              RuleMode       `json:"mode"`
	RuleConfig        RuleConfig     `json:"rule_config"`
	Points            [4]int         `json:"points"`
	Dealer            int            `json:"dealer"`
	RoundNumber       int            `json:"round_number"`
	Complete          bool           `json:"complete"`
	Round             GameSnapshot   `json:"round"`
	LastMCRSettlement *MCRSettlement `json:"last_mcr_settlement,omitempty"`
}

func NewMatch(seed int64, rules RuleSet) (*Match, error) {
	if rules == nil {
		return nil, fmt.Errorf("rule set is required")
	}
	round, err := NewGameWithRules(seed, rules)
	if err != nil {
		return nil, err
	}
	match := &Match{
		Mode:        rules.Mode(),
		RuleConfig:  rules.Config(),
		Points:      rules.InitialPoints(),
		RoundNumber: 1,
		Round:       round,
		rules:       rules,
	}
	match.Round.Dealer = match.Dealer
	match.Round.HandNumber = match.RoundNumber
	match.Round.Current = match.Dealer
	return match, nil
}

func (match *Match) Snapshot() MatchSnapshot {
	return match.snapshotWithRound(match.Round.Snapshot())
}

func (match *Match) SnapshotFor(playerID string) MatchSnapshot {
	return match.snapshotWithRound(match.Round.SnapshotFor(playerID))
}

func (match *Match) ApplyCommand(command GameCommand) CommandResult {
	result := match.Round.ApplyCommand(command)
	if result.OK && match.Round.Over {
		if match.Mode == ModeMCR {
			match.completeMCRRound()
		} else {
			match.Complete = true
		}
	}
	return result
}

func (match *Match) EnsureCurrentTurnDraw() (Tile, bool) {
	tile, ok := match.Round.EnsureCurrentTurnDraw()
	if match.Round.Over {
		if match.Mode == ModeMCR {
			match.completeMCRRound()
		} else {
			match.Complete = true
		}
	}
	return tile, ok
}

func (match *Match) snapshotWithRound(round GameSnapshot) MatchSnapshot {
	return MatchSnapshot{
		Mode:              match.Mode,
		RuleConfig:        match.RuleConfig,
		Points:            match.Points,
		Dealer:            match.Dealer,
		RoundNumber:       match.RoundNumber,
		Complete:          match.Complete,
		Round:             round,
		LastMCRSettlement: copyMCRSettlement(match.LastMCRSettlement),
	}
}

func (match *Match) completeMCRRound() {
	if match.Complete || match.Mode != ModeMCR || match.Round == nil || !match.Round.Over {
		return
	}
	if match.Round.Winner >= 0 && match.Round.MCRScore != nil {
		settlement := SettleMCR(*match.Round.MCRScore, match.Round.Winner, match.Round.Discarder, match.Round.WinType)
		match.LastMCRSettlement = &settlement
		for player, delta := range settlement.Deltas {
			match.Points[player] += delta
		}
	} else {
		match.LastMCRSettlement = nil
	}

	match.Dealer = (match.Dealer + 1) % 4
	if match.RoundNumber >= 16 {
		match.Complete = true
		return
	}
	match.RoundNumber++
	next, err := NewGameWithRules(match.Round.Seed+1, match.rules)
	if err != nil {
		match.Complete = true
		return
	}
	next.Dealer = match.Dealer
	next.HandNumber = match.RoundNumber
	next.Current = match.Dealer
	match.Round = next
}

func copyMCRSettlement(settlement *MCRSettlement) *MCRSettlement {
	if settlement == nil {
		return nil
	}
	copyValue := *settlement
	copyValue.Score = *copyMCRScore(&settlement.Score)
	return &copyValue
}
