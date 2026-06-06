package game

import "fmt"

type WinType int

const (
	WinNone WinType = iota
	WinSelfDraw
	WinDiscard
)

type WinContext struct {
	WinType WinType
	Melds   []Meld
	Pattern WinPattern
}

type ScoreResult struct {
	Points int
	Label  string
}

func ScoreRound(context WinContext) ScoreResult {
	base := 0
	baseLabel := "no win"
	switch context.WinType {
	case WinSelfDraw:
		base = 2
		baseLabel = "self-draw +2"
	case WinDiscard:
		base = 1
		baseLabel = "discard-win +1"
	}
	bonus := 0
	for _, meld := range context.Melds {
		if meld.Kind == MeldPong || meld.Kind == MeldKong {
			bonus++
		}
	}
	patternLabel := ""
	if context.Pattern == WinPatternSevenPairs {
		bonus += 2
		patternLabel = "seven pairs +2"
	}
	points := base + bonus
	if bonus == 0 {
		return ScoreResult{Points: points, Label: baseLabel}
	}
	if patternLabel != "" && bonus == 2 {
		return ScoreResult{
			Points: points,
			Label:  fmt.Sprintf("%s, %s = %d", baseLabel, patternLabel, points),
		}
	}
	if patternLabel != "" {
		return ScoreResult{
			Points: points,
			Label:  fmt.Sprintf("%s, meld/pattern bonus +%d (%s) = %d", baseLabel, bonus, patternLabel, points),
		}
	}
	return ScoreResult{
		Points: points,
		Label:  fmt.Sprintf("%s, meld bonus +%d = %d", baseLabel, bonus, points),
	}
}
