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

type LegalAction struct {
	Kind      CommandKind `json:"kind"`
	TileIndex int         `json:"tile_index,omitempty"`
	Tile      string      `json:"tile,omitempty"`
	Consumed  []Tile      `json:"consumed,omitempty"`
}

type RuleSet interface {
	Mode() RuleMode
	Config() RuleConfig
	InitialPoints() [4]int
	LegalActions(round *Game, playerID string) []LegalAction
	Allows(round *Game, command GameCommand) bool
	BuildWall() []Tile
	Deal(round *Game) error
	Draw(round *Game, player int, source DrawSource) (Tile, bool)
}

type CompatibilityRuleSet struct {
	mode   RuleMode
	config RuleConfig
}

func NewCompatibilityRuleSet(mode RuleMode, config RuleConfig) *CompatibilityRuleSet {
	return &CompatibilityRuleSet{mode: mode, config: config}
}

func (rules *CompatibilityRuleSet) Mode() RuleMode {
	return rules.mode
}

func (rules *CompatibilityRuleSet) Config() RuleConfig {
	return rules.config
}

func (rules *CompatibilityRuleSet) InitialPoints() [4]int {
	if rules.mode == ModeRiichi {
		points := rules.config.Riichi.StartingPoints
		return [4]int{points, points, points, points}
	}
	return [4]int{}
}

func (rules *CompatibilityRuleSet) BuildWall() []Tile {
	return BuildWall()
}

func (rules *CompatibilityRuleSet) Deal(round *Game) error {
	const tilesPerPlayer = 13
	required := len(round.Players) * tilesPerPlayer
	if len(round.Wall) < required {
		return fmt.Errorf("wall has %d tiles, need %d to deal", len(round.Wall), required)
	}
	for dealt := 0; dealt < tilesPerPlayer; dealt++ {
		for player := range round.Players {
			round.Players[player].AddTile(round.Wall[0])
			round.Wall = round.Wall[1:]
		}
	}
	return nil
}

func (rules *CompatibilityRuleSet) Draw(round *Game, player int, source DrawSource) (Tile, bool) {
	if len(round.Wall) == 0 {
		return -1, false
	}
	tile := round.Wall[0]
	round.Wall = round.Wall[1:]
	round.Players[player].AddTile(tile)
	round.RecordEvent(EventDraw, player, tile, "")
	return tile, true
}

func (rules *CompatibilityRuleSet) LegalActions(round *Game, id string) []LegalAction {
	if round == nil || round.Over || round.Phase == PhaseRoundOver || id != playerID(round.Current) {
		return nil
	}
	if round.Phase == PhaseAwaitingClaim {
		options := round.activeClaimOptions()
		if len(options) == 0 {
			return nil
		}
		actions := []LegalAction{{Kind: CommandPass}}
		for index, option := range options {
			action := LegalAction{Kind: commandKindForClaim(option.Kind), Consumed: append([]Tile(nil), option.Consumed...)}
			if option.Kind == ClaimChow {
				action.TileIndex = index
			}
			actions = append(actions, action)
		}
		return actions
	}
	if round.Phase != PhaseAwaitingDiscard {
		return nil
	}

	player := round.Players[round.Current]
	actions := make([]LegalAction, 0, len(player.Hand)+3)
	for index := range player.Hand {
		actions = append(actions, LegalAction{Kind: CommandDiscard, TileIndex: index})
	}
	if CanWin(player.Hand) {
		actions = append(actions, LegalAction{Kind: CommandWin})
	}
	counts := TileCounts(player.Hand)
	for tile, count := range counts {
		if count == 4 {
			actions = append(actions, LegalAction{Kind: CommandKong, Tile: Tile(tile).String()})
		}
	}
	return append(actions, LegalAction{Kind: CommandQuit})
}

func (rules *CompatibilityRuleSet) Allows(round *Game, command GameCommand) bool {
	for _, action := range rules.LegalActions(round, command.PlayerID) {
		if action.Kind != command.Kind {
			continue
		}
		switch action.Kind {
		case CommandDiscard, CommandChow:
			if action.TileIndex != command.TileIndex {
				continue
			}
		case CommandKong:
			if action.Tile != command.Tile {
				continue
			}
		}
		return true
	}
	return false
}

func commandKindForClaim(kind ClaimKind) CommandKind {
	switch kind {
	case ClaimWin:
		return CommandClaimWin
	case ClaimPong:
		return CommandPong
	case ClaimChow:
		return CommandChow
	default:
		return ""
	}
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
