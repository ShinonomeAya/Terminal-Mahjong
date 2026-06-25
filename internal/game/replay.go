package game

import (
	"encoding/json"
	"fmt"
	"strings"
)

const ReplaySchemaVersion = 3

type ReplayLog struct {
	SchemaVersion     int                   `json:"schema_version"`
	Mode              RuleMode              `json:"mode"`
	RuleConfig        RuleConfig            `json:"rule_config"`
	Seed              int64                 `json:"seed"`
	ShuffleProof      ShuffleProof          `json:"shuffle_proof"`
	Winner            string                `json:"winner"`
	Result            string                `json:"result"`
	Score             string                `json:"score"`
	Events            []GameEvent           `json:"events"`
	Dealer            int                   `json:"dealer,omitempty"`
	HandNumber        int                   `json:"hand_number,omitempty"`
	Points            [4]int                `json:"points,omitempty"`
	MCRScore          *MCRScoreBreakdown    `json:"mcr_score,omitempty"`
	MCRSettlements    []MCRSettlement       `json:"mcr_settlements,omitempty"`
	RiichiScore       *RiichiScoreBreakdown `json:"riichi_score,omitempty"`
	RiichiSettlements []RiichiSettlement    `json:"riichi_settlements,omitempty"`
	DoraIndicators    []Tile                `json:"dora_indicators,omitempty"`
	UraIndicators     []Tile                `json:"ura_indicators,omitempty"`
}

func (g *Game) ReplayLog() ReplayLog {
	winner := ""
	scoreLabel := ""
	if g.Winner >= 0 {
		winner = g.Players[g.Winner].Name
		if g.Mode == ModeMCR && g.MCRScore != nil {
			scoreLabel = fmt.Sprintf("%d points", g.MCRScore.TotalPoints)
		} else {
			score := ScoreRound(WinContext{
				WinType: g.WinType,
				Melds:   g.Players[g.Winner].Melds,
				Pattern: WinPatternOf(g.Players[g.Winner].Hand),
			})
			scoreLabel = score.Label
		}
	}
	log := ReplayLog{
		SchemaVersion: ReplaySchemaVersion,
		Mode:          g.Mode,
		RuleConfig:    g.RuleConfig,
		Seed:          g.Seed,
		ShuffleProof:  g.ShuffleProof,
		Winner:        winner,
		Result:        g.Reason,
		Score:         scoreLabel,
		Events:        append([]GameEvent(nil), g.Events...),
		Dealer:        g.Dealer,
		HandNumber:    g.HandNumber,
		MCRScore:      copyMCRScore(g.MCRScore),
	}
	if g.Riichi != nil {
		log.DoraIndicators = append([]Tile(nil), g.Riichi.DoraIndicators...)
		if g.Over || g.Phase == PhaseRoundOver {
			log.UraIndicators = append([]Tile(nil), g.Riichi.UraIndicators...)
			log.RiichiScore = copyRiichiScore(g.RiichiScore)
		}
	}
	return log
}

func (match *Match) ReplayLog() ReplayLog {
	log := match.Round.ReplayLog()
	log.Dealer = match.Dealer
	log.HandNumber = match.RoundNumber
	log.Points = match.Points
	log.MCRSettlements = copyMCRSettlements(match.MCRSettlements)
	log.RiichiSettlements = copyRiichiSettlements(match.RiichiSettlements)
	if match.LastMCRSettlement != nil {
		log.MCRScore = copyMCRScore(&match.LastMCRSettlement.Score)
		log.Score = fmt.Sprintf("%d points", match.LastMCRSettlement.Score.TotalPoints)
	}
	if match.LastRiichiSettlement != nil && len(match.LastRiichiSettlement.Scores) > 0 {
		log.RiichiScore = copyRiichiScore(&match.LastRiichiSettlement.Scores[0])
		log.Score = fmt.Sprintf("%d base points", match.LastRiichiSettlement.Scores[0].BasePoints)
	}
	return log
}

func (r ReplayLog) ToJSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

func ReplaySummary(log ReplayLog) string {
	lines := []string{
		"Replay",
		fmt.Sprintf("Mode: %s", log.Mode),
		fmt.Sprintf("Seed: %d", log.Seed),
		fmt.Sprintf("Wall Hash: %s", emptyDash(log.ShuffleProof.WallHash)),
		fmt.Sprintf("Winner: %s", emptyDash(log.Winner)),
		fmt.Sprintf("Result: %s", emptyDash(log.Result)),
		fmt.Sprintf("Score: %s", emptyDash(log.Score)),
		fmt.Sprintf("Events: %d", len(log.Events)),
		sectionTitle("Recent Events"),
	}
	recent := RecentEvents(log.Events, 5)
	if len(recent) == 0 {
		lines = append(lines, "-")
	} else {
		for _, event := range recent {
			lines = append(lines, event.String())
		}
	}
	return strings.Join(lines, "\n")
}

func emptyDash(text string) string {
	if text == "" {
		return "-"
	}
	return text
}
