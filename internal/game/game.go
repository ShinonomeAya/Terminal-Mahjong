package game

import (
	"bufio"
	cryptoRand "crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"strings"
)

type RNGMode string

const (
	CryptoSeeded RNGMode = "crypto_seeded"
	FixedSeed    RNGMode = "fixed_seed"
)

type ShuffleProof struct {
	Seed     int64  `json:"seed"`
	WallHash string `json:"wall_hash"`
}

type Game struct {
	Players      []Player
	Wall         []Tile
	Current      int
	Seed         int64
	RNGMode      RNGMode
	ShuffleProof ShuffleProof
	Winner       int
	Reason       string
	WinType      WinType
	Over         bool
	Events       []GameEvent
	Phase        TurnPhase
	PendingClaim *PendingClaim
	rng          *rand.Rand
}

func NewGame(seed int64) *Game {
	mode := FixedSeed
	if seed == 0 {
		seed = cryptoSeed()
		mode = CryptoSeeded
	}
	rng := rand.New(rand.NewSource(seed))
	wall := BuildWall()
	rng.Shuffle(len(wall), func(i, j int) {
		wall[i], wall[j] = wall[j], wall[i]
	})
	proof := ShuffleProof{Seed: seed, WallHash: wallHash(wall)}
	players := NewPlayers()
	for round := 0; round < 13; round++ {
		for i := range players {
			players[i].AddTile(wall[0])
			wall = wall[1:]
		}
	}
	return &Game{
		Players:      players,
		Wall:         wall,
		Seed:         seed,
		RNGMode:      mode,
		ShuffleProof: proof,
		Winner:       -1,
		Phase:        PhaseAwaitingDiscard,
		rng:          rng,
	}
}

func cryptoSeed() int64 {
	var bytes [8]byte
	if _, err := cryptoRand.Read(bytes[:]); err != nil {
		panic(fmt.Sprintf("crypto random seed failed: %v", err))
	}
	seed := int64(binary.LittleEndian.Uint64(bytes[:]) & ^(uint64(1) << 63))
	if seed == 0 {
		return 1
	}
	return seed
}

func (g *Game) Play(in io.Reader, out io.Writer) {
	reader := bufio.NewReader(in)
	fmt.Fprintln(out, "Terminal Mahjong - simplified pushdown rules")
	fmt.Fprintln(out, "Commands: <number> or d <number> to discard, h to win, k <tile> for concealed kong, q to quit. Answer y to claim win, pong, or chow prompts.")
	for !g.Over {
		if len(g.Wall) == 0 {
			g.Over = true
			g.Reason = "draw: wall exhausted"
			g.RecordEvent(EventWallExhausted, g.Current, -1, g.Reason)
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
	g.RecordEvent(EventDraw, playerIndex, tile, "")
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
	g.RecordEvent(EventDiscard, playerIndex, discard, "")
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
			g.RecordEvent(EventQuit, 0, -1, "quit")
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
		g.RecordEvent(EventDiscard, 0, discard, "")
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
	g.RecordEvent(EventKong, 0, tile, "concealed kong")
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
		g.RecordEvent(EventKong, playerIndex, kongTile, "concealed kong")
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

func (g *Game) finish(winner int, reason string, winType WinType) {
	g.Winner = winner
	g.Reason = reason
	g.WinType = winType
	g.Over = true
	g.RecordEvent(EventWin, winner, -1, reason)
}

func askYes(reader *bufio.Reader, out io.Writer, prompt string) bool {
	fmt.Fprint(out, prompt)
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(strings.ToLower(line))
	return line == "y" || line == "yes"
}
