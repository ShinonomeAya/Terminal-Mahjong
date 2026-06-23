package game

import "fmt"

type RuleMode string

const (
	ModeCompatibility RuleMode = "compatibility"
	ModeMCR           RuleMode = "mcr"
	ModeRiichi        RuleMode = "riichi"
)

type MatchLength string

const MatchEastSouth MatchLength = "east_south"

type MCRConfig struct {
	MinimumPoints int `json:"minimum_points"`
}

type RiichiConfig struct {
	StartingPoints int         `json:"starting_points"`
	MatchLength    MatchLength `json:"match_length"`
	OpenTanyao     bool        `json:"open_tanyao"`
	RedFives       int         `json:"red_fives"`
}

type RuleConfig struct {
	MCR    MCRConfig    `json:"mcr,omitempty"`
	Riichi RiichiConfig `json:"riichi,omitempty"`
}

func DefaultRuleConfig(mode RuleMode) RuleConfig {
	switch mode {
	case ModeMCR:
		return RuleConfig{MCR: MCRConfig{MinimumPoints: 8}}
	case ModeRiichi:
		return RuleConfig{Riichi: RiichiConfig{
			StartingPoints: 25000,
			MatchLength:    MatchEastSouth,
			OpenTanyao:     true,
			RedFives:       3,
		}}
	default:
		return RuleConfig{}
	}
}

func ParseRuleMode(text string) (RuleMode, error) {
	mode := RuleMode(text)
	switch mode {
	case ModeCompatibility, ModeMCR, ModeRiichi:
		return mode, nil
	default:
		return "", fmt.Errorf("unknown rule mode %q", text)
	}
}

func (config RuleConfig) Validate(mode RuleMode) error {
	switch mode {
	case ModeCompatibility:
		if config != (RuleConfig{}) {
			return fmt.Errorf("compatibility mode does not accept rule options")
		}
	case ModeMCR:
		if config.Riichi != (RiichiConfig{}) {
			return fmt.Errorf("MCR mode does not accept riichi options")
		}
		if config.MCR.MinimumPoints != 8 {
			return fmt.Errorf("MCR minimum points must be 8")
		}
	case ModeRiichi:
		if config.MCR != (MCRConfig{}) {
			return fmt.Errorf("riichi mode does not accept MCR options")
		}
		if config.Riichi.StartingPoints < 1000 || config.Riichi.StartingPoints > 100000 {
			return fmt.Errorf("riichi starting points must be between 1000 and 100000")
		}
		if config.Riichi.MatchLength != MatchEastSouth {
			return fmt.Errorf("unsupported riichi match length %q", config.Riichi.MatchLength)
		}
		if config.Riichi.RedFives != 0 && config.Riichi.RedFives != 3 {
			return fmt.Errorf("riichi red fives must be 0 or 3")
		}
	default:
		return fmt.Errorf("unknown rule mode %q", mode)
	}
	return nil
}
