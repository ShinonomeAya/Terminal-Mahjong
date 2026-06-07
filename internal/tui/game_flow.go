package tui

import "mahjong/internal/game"

func newStartedGame() *game.Game {
	g := game.NewGame(0)
	g.StartHumanTurn()
	return g
}
