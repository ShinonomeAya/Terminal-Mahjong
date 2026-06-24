package game

type DrawSource string

const (
	DrawNormal      DrawSource = "normal"
	DrawReplacement DrawSource = "replacement"
)

type MCRRuleSet struct {
	config MCRConfig
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

type FanMatch struct {
	ID     FanID  `json:"id"`
	NameZH string `json:"name_zh"`
	NameEN string `json:"name_en"`
	Points int    `json:"points"`
	Count  int    `json:"count"`
}

type MCRScoreContext struct {
	Winner          int     `json:"winner"`
	Discarder       int     `json:"discarder"`
	WinningTile     Tile    `json:"winning_tile"`
	WinType         WinType `json:"win_type"`
	SeatWind        Tile    `json:"seat_wind"`
	PrevalentWind   Tile    `json:"prevalent_wind"`
	Flowers         int     `json:"flowers"`
	LastTileDraw    bool    `json:"last_tile_draw"`
	LastTileClaim   bool    `json:"last_tile_claim"`
	LastOfKind      bool    `json:"last_of_kind"`
	ReplacementDraw bool    `json:"replacement_draw"`
	RobbingKong     bool    `json:"robbing_kong"`
}

type MCRScoreBreakdown struct {
	Fans            []FanMatch `json:"fans"`
	NonFlowerPoints int        `json:"non_flower_points"`
	FlowerPoints    int        `json:"flower_points"`
	TotalPoints     int        `json:"total_points"`
	MeetsMinimum    bool       `json:"meets_minimum"`
	WinningGrouping []Meld     `json:"winning_grouping"`
}

type MCRSettlement struct {
	Winner    int               `json:"winner"`
	Discarder int               `json:"discarder"`
	Deltas    [4]int            `json:"deltas"`
	Score     MCRScoreBreakdown `json:"score"`
}

func NewMCRRuleSet(config MCRConfig) *MCRRuleSet {
	return &MCRRuleSet{config: config}
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
	return rules.legalActions(round, playerID)
}

func (rules *MCRRuleSet) Allows(round *Game, command GameCommand) bool {
	return rules.allows(round, command)
}
