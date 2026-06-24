package game

type DrawSource string

const (
	DrawNormal      DrawSource = "normal"
	DrawReplacement DrawSource = "replacement"
)

type MCRRuleSet struct {
	config        MCRConfig
	compatibility *CompatibilityRuleSet
}

type FanID string

type MCRFanContext struct {
	Decomposition   MCRDecomposition
	WinningTile     Tile
	WinType         WinType
	SeatWind        Tile
	PrevalentWind   Tile
	Flowers         []Tile
	LastTileDraw    bool
	LastTileClaim   bool
	LastOfKind      bool
	ReplacementDraw bool
	RobbingKong     bool
}

type MCRFanOccurrence struct {
	ID     FanID `json:"id"`
	Points int   `json:"points"`
	Count  int   `json:"count"`
	Groups []int `json:"groups,omitempty"`
}

func NewMCRRuleSet(config MCRConfig) *MCRRuleSet {
	ruleConfig := RuleConfig{MCR: config}
	return &MCRRuleSet{
		config:        config,
		compatibility: NewCompatibilityRuleSet(ModeMCR, ruleConfig),
	}
}

func (rules *MCRRuleSet) Mode() RuleMode {
	return ModeMCR
}

func (rules *MCRRuleSet) Config() RuleConfig {
	return RuleConfig{MCR: rules.config}
}

func (rules *MCRRuleSet) InitialPoints() [4]int {
	return [4]int{}
}

func (rules *MCRRuleSet) LegalActions(round *Game, playerID string) []LegalAction {
	return rules.compatibility.LegalActions(round, playerID)
}

func (rules *MCRRuleSet) Allows(round *Game, command GameCommand) bool {
	return rules.compatibility.Allows(round, command)
}
