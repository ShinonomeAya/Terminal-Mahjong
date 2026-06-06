package game

import (
	"fmt"
	"io"
)

func (g *Game) printTable(out io.Writer) {
	fmt.Fprintf(out, "\n%s\n", sectionTitle("Terminal Mahjong"))
	fmt.Fprintf(out, "Wall: %d | Events: %d\n", len(g.Wall), len(g.Events))
	fmt.Fprintln(out, sectionTitle("Opponents"))
	for _, player := range g.Players {
		if !player.Human {
			fmt.Fprintln(out, formatOpponentLine(player))
		}
	}
	fmt.Fprintln(out, sectionTitle("Your Hand"))
	fmt.Fprintf(out, "Melds: %s | Discards: %s\n", g.Players[0].MeldSummary(), FormatTiles(g.Players[0].Discards))
	for i, tile := range g.Players[0].Hand {
		fmt.Fprintf(out, "%2d:%s ", i+1, tile)
	}
	fmt.Fprintln(out)
	fmt.Fprintf(out, "Tips: %s\n", HandTips(g.Players[0].Hand))
	fmt.Fprintln(out, sectionTitle("Commands"))
	fmt.Fprintln(out, commandHelp())
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

func sectionTitle(title string) string {
	return "== " + title + " =="
}

func commandHelp() string {
	return "<number>/d <number>: discard | h: win | k <tile>: kong | q: quit | y/Enter: claim/decline"
}

func formatOpponentLine(player Player) string {
	return fmt.Sprintf("%s hand: %d tiles | melds: %s | discards: %s", player.Name, len(player.Hand), player.MeldSummary(), FormatTiles(player.Discards))
}
