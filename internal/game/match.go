package game

import "fmt"

type Match struct {
	Mode        RuleMode
	RuleConfig  RuleConfig
	Points      [4]int
	Dealer      int
	RoundNumber int
	Complete    bool
	Round       *Game
	rules       RuleSet
}

type MatchSnapshot struct {
	Mode        RuleMode     `json:"mode"`
	RuleConfig  RuleConfig   `json:"rule_config"`
	Points      [4]int       `json:"points"`
	Dealer      int          `json:"dealer"`
	RoundNumber int          `json:"round_number"`
	Complete    bool         `json:"complete"`
	Round       GameSnapshot `json:"round"`
}

func NewMatch(seed int64, rules RuleSet) (*Match, error) {
	if rules == nil {
		return nil, fmt.Errorf("rule set is required")
	}
	round, err := NewGameWithRules(seed, rules)
	if err != nil {
		return nil, err
	}
	return &Match{
		Mode:        rules.Mode(),
		RuleConfig:  rules.Config(),
		Points:      rules.InitialPoints(),
		RoundNumber: 1,
		Round:       round,
		rules:       rules,
	}, nil
}

func (match *Match) Snapshot() MatchSnapshot {
	return match.snapshotWithRound(match.Round.Snapshot())
}

func (match *Match) SnapshotFor(playerID string) MatchSnapshot {
	return match.snapshotWithRound(match.Round.Snapshot())
}

func (match *Match) ApplyCommand(command GameCommand) CommandResult {
	result := match.Round.ApplyCommand(command)
	match.Complete = match.Round.Over
	return result
}

func (match *Match) EnsureCurrentTurnDraw() (Tile, bool) {
	tile, ok := match.Round.EnsureCurrentTurnDraw()
	match.Complete = match.Round.Over
	return tile, ok
}

func (match *Match) snapshotWithRound(round GameSnapshot) MatchSnapshot {
	return MatchSnapshot{
		Mode:        match.Mode,
		RuleConfig:  match.RuleConfig,
		Points:      match.Points,
		Dealer:      match.Dealer,
		RoundNumber: match.RoundNumber,
		Complete:    match.Complete,
		Round:       round,
	}
}
