package game

import (
	"encoding/json"
	"fmt"
	"strings"
)

type ReplayLog struct {
	Seed   int64       `json:"seed"`
	Winner string      `json:"winner"`
	Result string      `json:"result"`
	Score  string      `json:"score"`
	Events []GameEvent `json:"events"`
}

func (g *Game) ReplayLog() ReplayLog {
	winner := ""
	scoreLabel := ""
	if g.Winner >= 0 {
		winner = g.Players[g.Winner].Name
		score := ScoreRound(WinContext{
			WinType: g.WinType,
			Melds:   g.Players[g.Winner].Melds,
			Pattern: WinPatternOf(g.Players[g.Winner].Hand),
		})
		scoreLabel = score.Label
	}
	return ReplayLog{
		Seed:   g.Seed,
		Winner: winner,
		Result: g.Reason,
		Score:  scoreLabel,
		Events: append([]GameEvent(nil), g.Events...),
	}
}

func (r ReplayLog) ToJSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

func ReplaySummary(log ReplayLog) string {
	lines := []string{
		"Replay",
		fmt.Sprintf("Seed: %d", log.Seed),
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
