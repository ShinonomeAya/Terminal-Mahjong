package game

import (
	"fmt"
	"io"
)

func (g *Game) printTable(out io.Writer) {
	fmt.Fprintf(out, "\nWall: %d\n", len(g.Wall))
	for i, player := range g.Players {
		if player.Human {
			fmt.Fprintf(out, "%s melds: %s | discards: %s\n", player.Name, player.MeldSummary(), FormatTiles(player.Discards))
			continue
		}
		fmt.Fprintf(out, "%s hand: %d tiles | melds: %s | discards: %s\n", player.Name, len(player.Hand), player.MeldSummary(), FormatTiles(player.Discards))
		if i == len(g.Players)-1 {
			fmt.Fprint(out, "")
		}
	}
	fmt.Fprintln(out, "Your hand:")
	for i, tile := range g.Players[0].Hand {
		fmt.Fprintf(out, "%2d:%s ", i+1, tile)
	}
	fmt.Fprintln(out)
	fmt.Fprintf(out, "Tips: %s\n", HandTips(g.Players[0].Hand))
}

func (g *Game) printResult(out io.Writer) {
	fmt.Fprintln(out, "\nGame over.")
	if g.Winner >= 0 {
		score := ScoreRound(WinContext{
			WinType: g.WinType,
			Melds:   g.Players[g.Winner].Melds,
		})
		fmt.Fprintf(out, "Winner: %s\n", g.Players[g.Winner].Name)
		fmt.Fprintf(out, "Win: %s\n", g.Reason)
		fmt.Fprintf(out, "Score: %s\n", score.Label)
		fmt.Fprintf(out, "Events: %d\n", len(g.Events))
		return
	}
	fmt.Fprintf(out, "Result: %s\n", g.Reason)
	fmt.Fprintf(out, "Events: %d\n", len(g.Events))
}

func drawVerb(player *Player) string {
	if player.Human {
		return "draw"
	}
	return "draws"
}
