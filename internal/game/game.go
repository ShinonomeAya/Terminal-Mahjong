package game

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"strings"
	"time"
)

type Game struct {
	Players []Player
	Wall    []Tile
	Current int
	Winner  int
	Reason  string
	WinType WinType
	Over    bool
	rng     *rand.Rand
}

func NewGame(seed int64) *Game {
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	rng := rand.New(rand.NewSource(seed))
	wall := BuildWall()
	rng.Shuffle(len(wall), func(i, j int) {
		wall[i], wall[j] = wall[j], wall[i]
	})
	players := NewPlayers()
	for round := 0; round < 13; round++ {
		for i := range players {
			players[i].AddTile(wall[0])
			wall = wall[1:]
		}
	}
	return &Game{
		Players: players,
		Wall:    wall,
		Winner:  -1,
		rng:     rng,
	}
}

func (g *Game) Play(in io.Reader, out io.Writer) {
	reader := bufio.NewReader(in)
	fmt.Fprintln(out, "Terminal Mahjong - simplified pushdown rules")
	fmt.Fprintln(out, "Commands: <number> or d <number> to discard, h to win, k <tile> for concealed kong, q to quit. Answer y to claim win, pong, or chow prompts.")
	for !g.Over {
		if len(g.Wall) == 0 {
			g.Over = true
			g.Reason = "draw: wall exhausted"
			break
		}
		drawn := g.draw(g.Current)
		player := &g.Players[g.Current]
		fmt.Fprintf(out, "\n%s %s a tile. Wall: %d\n", player.Name, drawVerb(player), len(g.Wall))
		if CanWin(player.Hand) {
			if player.Human {
				if askYes(reader, out, "You can win by self-draw. Win? [y/N] ") {
					g.finish(g.Current, "self-draw", WinSelfDraw)
					break
				}
			} else {
				g.finish(g.Current, "self-draw", WinSelfDraw)
				fmt.Fprintf(out, "%s wins by self-draw with %s.\n", player.Name, drawn)
				break
			}
		}
		if !player.Human {
			g.resolveAIKongs(out, g.Current)
			if g.Over {
				break
			}
		}
		discard, ok := g.takeDiscardTurn(reader, out, g.Current)
		if !ok {
			break
		}
		if g.resolveDiscardClaims(reader, out, g.Current, discard) {
			continue
		}
		g.Current = (g.Current + 1) % len(g.Players)
	}
	g.printResult(out)
}

func (g *Game) draw(playerIndex int) Tile {
	tile := g.Wall[0]
	g.Wall = g.Wall[1:]
	g.Players[playerIndex].AddTile(tile)
	return tile
}

func (g *Game) takeDiscardTurn(reader *bufio.Reader, out io.Writer, playerIndex int) (Tile, bool) {
	player := &g.Players[playerIndex]
	if player.Human {
		return g.humanDiscard(reader, out)
	}
	index := ChooseAIDiscard(player.Hand)
	discard, _ := player.RemoveAt(index)
	player.Discards = append(player.Discards, discard)
	fmt.Fprintf(out, "%s discards %s.\n", player.Name, discard)
	return discard, true
}

func (g *Game) humanDiscard(reader *bufio.Reader, out io.Writer) (Tile, bool) {
	for {
		g.printTable(out)
		fmt.Fprint(out, "Your action: ")
		line, err := reader.ReadString('\n')
		if err != nil && len(line) == 0 {
			g.Over = true
			g.Reason = "input closed"
			return 0, false
		}
		action, err := ParseAction(line)
		if err != nil {
			fmt.Fprintln(out, "Use a hand number, d <number>, h, k <tile>, or q.")
			continue
		}
		if action.Kind == ActionQuit {
			g.Over = true
			g.Reason = "quit"
			return 0, false
		}
		if action.Kind == ActionWin {
			if CanWin(g.Players[0].Hand) {
				g.finish(0, "self-draw", WinSelfDraw)
				return 0, false
			}
			fmt.Fprintln(out, "You cannot win with this hand yet.")
			continue
		}
		if action.Kind == ActionKong {
			if g.tryHumanKong(action.Tiles[0].String(), out) {
				continue
			}
			continue
		}
		if action.Kind != ActionDiscard {
			fmt.Fprintln(out, "Use a hand number, d <number>, h, k <tile>, or q.")
			continue
		}
		discard, err := g.Players[0].RemoveAt(action.Index)
		if err != nil {
			fmt.Fprintln(out, err)
			continue
		}
		g.Players[0].Discards = append(g.Players[0].Discards, discard)
		fmt.Fprintf(out, "You discard %s.\n", discard)
		return discard, true
	}
}

func (g *Game) tryHumanKong(tileText string, out io.Writer) bool {
	tile, ok := ParseTile(tileText)
	if !ok {
		fmt.Fprintln(out, "Unknown tile. Examples: 1m, 9p, 3s, E, S, W, N, Z, F, B.")
		return false
	}
	if g.Players[0].Count(tile) < 4 {
		fmt.Fprintf(out, "You do not have four %s tiles.\n", tile)
		return false
	}
	for i := 0; i < 4; i++ {
		g.Players[0].RemoveTile(tile)
	}
	g.Players[0].AddMeld(MeldKong, []Tile{tile, tile, tile, tile})
	fmt.Fprintf(out, "You declare a concealed kong of %s.\n", tile)
	if len(g.Wall) == 0 {
		g.Over = true
		g.Reason = "draw: wall exhausted after kong"
		return true
	}
	replacement := g.draw(0)
	fmt.Fprintf(out, "Replacement draw: %s.\n", replacement)
	return true
}

func (g *Game) resolveAIKongs(out io.Writer, playerIndex int) {
	for {
		kongTile, ok := concealedKongTile(g.Players[playerIndex])
		if !ok {
			return
		}
		for i := 0; i < 4; i++ {
			g.Players[playerIndex].RemoveTile(kongTile)
		}
		g.Players[playerIndex].AddMeld(MeldKong, []Tile{kongTile, kongTile, kongTile, kongTile})
		fmt.Fprintf(out, "%s declares a concealed kong of %s.\n", g.Players[playerIndex].Name, kongTile)
		if len(g.Wall) == 0 {
			g.Over = true
			g.Reason = "draw: wall exhausted after kong"
			return
		}
		replacement := g.draw(playerIndex)
		fmt.Fprintf(out, "%s draws a replacement tile.\n", g.Players[playerIndex].Name)
		if CanWin(g.Players[playerIndex].Hand) {
			g.finish(playerIndex, fmt.Sprintf("self-draw after kong with %s", replacement), WinSelfDraw)
			return
		}
	}
}

func concealedKongTile(player Player) (Tile, bool) {
	counts := TileCounts(player.Hand)
	for tile, count := range counts {
		if count == 4 {
			return Tile(tile), true
		}
	}
	return 0, false
}

func ChowOptions(player Player, discard Tile) [][]Tile {
	if !discard.IsSuit() {
		return nil
	}
	options := make([][]Tile, 0, 3)
	discardIndex := int(discard)
	for start := discardIndex - 2; start <= discardIndex; start++ {
		if start < 0 || start+2 >= 27 {
			continue
		}
		first := Tile(start)
		second := Tile(start + 1)
		third := Tile(start + 2)
		if first.Rank() > third.Rank() {
			continue
		}
		if first != discard && player.Count(first) == 0 {
			continue
		}
		if second != discard && player.Count(second) == 0 {
			continue
		}
		if third != discard && player.Count(third) == 0 {
			continue
		}
		option := []Tile{first, second, third}
		SortTiles(option)
		options = append(options, option)
	}
	return options
}

func (g *Game) resolveDiscardClaims(reader *bufio.Reader, out io.Writer, discarder int, discard Tile) bool {
	for offset := 1; offset < len(g.Players); offset++ {
		claimer := (discarder + offset) % len(g.Players)
		claimHand := append([]Tile(nil), g.Players[claimer].Hand...)
		claimHand = append(claimHand, discard)
		SortTiles(claimHand)
		if CanWin(claimHand) {
			if g.Players[claimer].Human {
				if askYes(reader, out, fmt.Sprintf("You can win on %s. Win? [y/N] ", discard)) {
					g.finish(claimer, fmt.Sprintf("discard-win on %s from %s", discard, g.Players[discarder].Name), WinDiscard)
					return true
				}
			} else {
				g.finish(claimer, fmt.Sprintf("discard-win on %s from %s", discard, g.Players[discarder].Name), WinDiscard)
				fmt.Fprintf(out, "%s wins on %s.\n", g.Players[claimer].Name, discard)
				return true
			}
		}
	}
	for offset := 1; offset < len(g.Players); offset++ {
		claimer := (discarder + offset) % len(g.Players)
		if g.Players[claimer].Count(discard) < 2 {
			continue
		}
		if g.Players[claimer].Human {
			if !askYes(reader, out, fmt.Sprintf("Pong %s? [y/N] ", discard)) {
				continue
			}
		} else if !shouldAIPong(g.Players[claimer], discard) {
			continue
		}
		g.claimPong(claimer, discard)
		fmt.Fprintf(out, "%s pongs %s.\n", g.Players[claimer].Name, discard)
		claimedDiscard, ok := g.takeDiscardTurn(reader, out, claimer)
		if !ok {
			return true
		}
		g.Current = claimer
		if g.resolveDiscardClaims(reader, out, claimer, claimedDiscard) {
			return true
		}
		g.Current = (claimer + 1) % len(g.Players)
		return true
	}
	nextPlayer := (discarder + 1) % len(g.Players)
	chowOptions := ChowOptions(g.Players[nextPlayer], discard)
	if len(chowOptions) == 0 {
		return false
	}
	chowIndex := -1
	if g.Players[nextPlayer].Human {
		for i, option := range chowOptions {
			handTiles := chowHandTiles(option, discard)
			if askYes(reader, out, fmt.Sprintf("Chow %s with %s? [y/N] ", discard, FormatTiles(handTiles))) {
				chowIndex = i
				break
			}
		}
	} else if index, ok := shouldAIChow(g.Players[nextPlayer], discard, chowOptions); ok {
		chowIndex = index
	}
	if chowIndex < 0 {
		return false
	}
	g.claimChow(nextPlayer, discard, chowOptions[chowIndex])
	fmt.Fprintf(out, "%s chows %s.\n", g.Players[nextPlayer].Name, discard)
	claimedDiscard, ok := g.takeDiscardTurn(reader, out, nextPlayer)
	if !ok {
		return true
	}
	g.Current = nextPlayer
	if g.resolveDiscardClaims(reader, out, nextPlayer, claimedDiscard) {
		return true
	}
	g.Current = (nextPlayer + 1) % len(g.Players)
	return true
}

func shouldAIChow(player Player, discard Tile, options [][]Tile) (int, bool) {
	if len(options) == 0 {
		return 0, false
	}
	counts := TileCounts(player.Hand)
	bestIndex := -1
	bestScore := 999
	for i, option := range options {
		score := 0
		for _, tile := range chowHandTiles(option, discard) {
			score += tileUsefulness(tile, counts)
		}
		if score < bestScore {
			bestScore = score
			bestIndex = i
		}
	}
	return bestIndex, bestIndex >= 0 && bestScore <= 12
}

func chowHandTiles(option []Tile, discard Tile) []Tile {
	handTiles := make([]Tile, 0, 2)
	removedDiscard := false
	for _, tile := range option {
		if tile == discard && !removedDiscard {
			removedDiscard = true
			continue
		}
		handTiles = append(handTiles, tile)
	}
	return handTiles
}

func (g *Game) claimChow(playerIndex int, discard Tile, option []Tile) {
	player := &g.Players[playerIndex]
	for _, tile := range chowHandTiles(option, discard) {
		player.RemoveTile(tile)
	}
	player.AddMeld(MeldChow, option)
}

func (g *Game) claimPong(playerIndex int, tile Tile) {
	player := &g.Players[playerIndex]
	player.RemoveTile(tile)
	player.RemoveTile(tile)
	player.AddMeld(MeldPong, []Tile{tile, tile, tile})
}

func shouldAIPong(player Player, tile Tile) bool {
	if !tile.IsSuit() {
		return true
	}
	counts := TileCounts(player.Hand)
	return tileUsefulness(tile, counts) >= 6
}

func (g *Game) finish(winner int, reason string, winType WinType) {
	g.Winner = winner
	g.Reason = reason
	g.WinType = winType
	g.Over = true
}

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
		return
	}
	fmt.Fprintf(out, "Result: %s\n", g.Reason)
}

func askYes(reader *bufio.Reader, out io.Writer, prompt string) bool {
	fmt.Fprint(out, prompt)
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(strings.ToLower(line))
	return line == "y" || line == "yes"
}

func drawVerb(player *Player) string {
	if player.Human {
		return "draw"
	}
	return "draws"
}
