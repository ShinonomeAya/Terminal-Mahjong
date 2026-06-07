package game

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

func (g *Game) StartHumanTurn() (Tile, bool) {
	if g.Over || g.Current != 0 || len(g.Wall) == 0 {
		return -1, false
	}
	return g.draw(0), true
}

func (g *Game) HumanDiscardSelected(index int) (Tile, error) {
	if g.Over {
		return -1, fmt.Errorf("game is over")
	}
	if g.Current != 0 {
		return -1, fmt.Errorf("not the human turn")
	}
	discard, err := g.Players[0].RemoveAt(index)
	if err != nil {
		return -1, err
	}
	g.Players[0].Discards = append(g.Players[0].Discards, discard)
	g.RecordEvent(EventDiscard, 0, discard, "")
	g.Current = 1
	return discard, nil
}

func (g *Game) Quit(reason string) {
	g.Over = true
	g.Reason = reason
	g.RecordEvent(EventQuit, 0, -1, reason)
}

func (g *Game) AdvanceAIUntilHumanTurn() {
	declineReader := bufio.NewReader(strings.NewReader(""))
	for !g.Over && g.Current != 0 {
		if len(g.Wall) == 0 {
			g.Over = true
			g.Reason = "draw: wall exhausted"
			g.RecordEvent(EventWallExhausted, g.Current, -1, g.Reason)
			return
		}
		g.draw(g.Current)
		if CanWin(g.Players[g.Current].Hand) {
			g.finish(g.Current, "self-draw", WinSelfDraw)
			return
		}
		g.resolveAIKongs(io.Discard, g.Current)
		if g.Over {
			return
		}
		discard, ok := g.takeDiscardTurn(declineReader, io.Discard, g.Current)
		if !ok {
			return
		}
		if g.resolveDiscardClaims(declineReader, io.Discard, g.Current, discard) {
			continue
		}
		g.Current = (g.Current + 1) % len(g.Players)
	}
	if !g.Over && g.Current == 0 && len(g.Players[0].Hand)%3 == 1 {
		g.StartHumanTurn()
	}
}
