package game

import (
	"fmt"
	"io"
)

func (g *Game) StartHumanTurn() (Tile, bool) {
	if g.Over || g.Phase != PhaseAwaitingDiscard || g.Current != 0 || len(g.Wall) == 0 {
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
	return g.discardCurrent(index)
}

func (g *Game) Quit(reason string) {
	g.Over = true
	g.Reason = reason
	g.Phase = PhaseRoundOver
	g.PendingClaim = nil
	g.RecordEvent(EventQuit, 0, -1, reason)
}

func (g *Game) AdvanceAIUntilHumanTurn() {
	for !g.Over {
		if g.Phase == PhaseAwaitingClaim {
			if g.Current == 0 {
				return
			}
			result := g.ApplyCommand(g.aiClaimCommand())
			if !result.OK {
				return
			}
			continue
		}
		if g.Current == 0 {
			g.EnsureCurrentTurnDraw()
			return
		}
		if _, ok := g.EnsureCurrentTurnDraw(); !ok && g.Over {
			return
		}
		if g.rules.Allows(g, GameCommand{PlayerID: playerID(g.Current), Kind: CommandWin}) {
			g.finish(g.Current, "self-draw", WinSelfDraw)
			return
		}
		g.resolveAIKongs(io.Discard, g.Current)
		if g.Over {
			return
		}
		index := ChooseAIDiscard(g.Players[g.Current].Hand)
		if _, err := g.discardCurrent(index); err != nil {
			return
		}
	}
}

func (g *Game) aiClaimCommand() GameCommand {
	command := GameCommand{PlayerID: playerID(g.Current), Kind: CommandPass}
	options := g.activeClaimOptions()
	if len(options) == 0 {
		return command
	}
	switch options[0].Kind {
	case ClaimWin:
		command.Kind = CommandClaimWin
	case ClaimPong:
		if shouldAIPong(g.Players[g.Current], g.PendingClaim.Tile) {
			command.Kind = CommandPong
		}
	case ClaimChow:
		melds := make([][]Tile, len(options))
		for i, option := range options {
			melds[i] = append(append([]Tile(nil), option.Consumed...), g.PendingClaim.Tile)
			SortTiles(melds[i])
		}
		if index, ok := shouldAIChow(g.Players[g.Current], g.PendingClaim.Tile, melds); ok {
			command.Kind = CommandChow
			command.TileIndex = index
		}
	}
	return command
}
