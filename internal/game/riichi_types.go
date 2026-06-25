package game

type RiichiDeclarationState string

const (
	RiichiNone     RiichiDeclarationState = "none"
	RiichiDeclared RiichiDeclarationState = "declared"
	RiichiAccepted RiichiDeclarationState = "accepted"
)

type RiichiShapeKind string

const (
	RiichiShapeStandard        RiichiShapeKind = "standard"
	RiichiShapeSevenPairs      RiichiShapeKind = "seven_pairs"
	RiichiShapeThirteenOrphans RiichiShapeKind = "thirteen_orphans"
)

type RiichiWaitKind string

const (
	RiichiWaitUnknown RiichiWaitKind = ""
	RiichiWaitRyanmen RiichiWaitKind = "ryanmen"
	RiichiWaitKanchan RiichiWaitKind = "kanchan"
	RiichiWaitPenchan RiichiWaitKind = "penchan"
	RiichiWaitTanki   RiichiWaitKind = "tanki"
	RiichiWaitShanpon RiichiWaitKind = "shanpon"
	RiichiWaitKokushi RiichiWaitKind = "kokushi"
)

type RiichiDecomposition struct {
	Kind   RiichiShapeKind `json:"kind"`
	Groups []MCRGroup      `json:"groups"`
	Tiles  []Tile          `json:"tiles"`
	Wait   RiichiWaitKind  `json:"wait,omitempty"`
}

type RiichiYakuMatch struct {
	ID      string `json:"id"`
	NameZH  string `json:"name_zh"`
	NameEN  string `json:"name_en"`
	Han     int    `json:"han"`
	Yakuman int    `json:"yakuman,omitempty"`
}

type RiichiYakuContext struct {
	Decomposition  RiichiDecomposition
	WinningTile    Tile
	WinType        WinType
	Closed         bool
	SeatWind       Tile
	PrevalentWind  Tile
	Riichi         RiichiDeclarationState
	Ippatsu        bool
	DoubleRiichi   bool
	Rinshan        bool
	Chankan        bool
	Haitei         bool
	Houtei         bool
	Renhou         bool
	Tenhou         bool
	Chiihou        bool
	DoraIndicators []Tile
	UraIndicators  []Tile
}

type RiichiScoreBreakdown struct {
	Yaku          []RiichiYakuMatch `json:"yaku"`
	Fu            int               `json:"fu"`
	YakuHan       int               `json:"yaku_han"`
	BonusHan      int               `json:"bonus_han"`
	Yakuman       int               `json:"yakuman"`
	BasePoints    int               `json:"base_points"`
	LimitName     string            `json:"limit_name,omitempty"`
	HasYaku       bool              `json:"has_yaku"`
	WinningGroups []Meld            `json:"winning_groups"`
}

type RiichiRoundState struct {
	DeadWall         []Tile                    `json:"dead_wall"`
	DoraIndicators   []Tile                    `json:"dora_indicators"`
	UraIndicators    []Tile                    `json:"ura_indicators,omitempty"`
	RinshanDraws     int                       `json:"rinshan_draws"`
	KanCount         int                       `json:"kan_count"`
	Honba            int                       `json:"honba"`
	RiichiSticks     int                       `json:"riichi_sticks"`
	Declarations     [4]RiichiDeclarationState `json:"declarations"`
	Ippatsu          [4]bool                   `json:"ippatsu"`
	TemporaryFuriten [4]bool                   `json:"temporary_furiten"`
	RiichiFuriten    [4]bool                   `json:"riichi_furiten"`
}

type RiichiRuleSet struct {
	config        RiichiConfig
	compatibility *CompatibilityRuleSet
}

func NewRiichiRuleSet(config RiichiConfig) *RiichiRuleSet {
	ruleConfig := RuleConfig{Riichi: config}
	return &RiichiRuleSet{
		config:        config,
		compatibility: NewCompatibilityRuleSet(ModeRiichi, ruleConfig),
	}
}

func (rules *RiichiRuleSet) Mode() RuleMode {
	return ModeRiichi
}

func (rules *RiichiRuleSet) Config() RuleConfig {
	return RuleConfig{Riichi: rules.config}
}

func (rules *RiichiRuleSet) InitialPoints() [4]int {
	points := rules.config.StartingPoints
	return [4]int{points, points, points, points}
}

func (rules *RiichiRuleSet) LegalActions(round *Game, playerID string) []LegalAction {
	return rules.legalActions(round, playerID)
}

func (rules *RiichiRuleSet) Allows(round *Game, command GameCommand) bool {
	return rules.allows(round, command)
}
