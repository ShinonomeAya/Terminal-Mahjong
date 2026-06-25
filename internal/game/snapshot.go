package game

import "io"

type PlayerView struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Human     bool   `json:"human"`
	Hand      []Tile `json:"hand"`
	HandCount int    `json:"hand_count"`
	Flowers   []Tile `json:"flowers"`
	Melds     []Meld `json:"melds"`
	Discards  []Tile `json:"discards"`
}

type GameSnapshot struct {
	Seed         int64              `json:"seed"`
	RNGMode      RNGMode            `json:"rng_mode"`
	ShuffleProof ShuffleProof       `json:"shuffle_proof"`
	Players      []PlayerView       `json:"players"`
	WallCount    int                `json:"wall_count"`
	Current      int                `json:"current"`
	Winner       int                `json:"winner"`
	Discarder    int                `json:"discarder"`
	Dealer       int                `json:"dealer"`
	HandNumber   int                `json:"hand_number"`
	MCRScore     *MCRScoreBreakdown `json:"mcr_score,omitempty"`
	Reason       string             `json:"reason"`
	Over         bool               `json:"over"`
	Events       []GameEvent        `json:"events"`
	Phase        TurnPhase          `json:"phase"`
	PendingClaim *PendingClaim      `json:"pending_claim,omitempty"`
	LegalActions []LegalAction      `json:"legal_actions,omitempty"`
}

type CommandKind string

const (
	CommandDiscard  CommandKind = "discard"
	CommandWin      CommandKind = "win"
	CommandKong     CommandKind = "kong"
	CommandQuit     CommandKind = "quit"
	CommandPass     CommandKind = "pass"
	CommandClaimWin CommandKind = "claim_win"
	CommandPong     CommandKind = "pong"
	CommandChow     CommandKind = "chow"
)

type GameCommand struct {
	PlayerID  string      `json:"player_id"`
	Kind      CommandKind `json:"kind"`
	TileIndex int         `json:"tile_index"`
	Tile      string      `json:"tile"`
}

type CommandResult struct {
	OK       bool         `json:"ok"`
	Error    string       `json:"error,omitempty"`
	Command  GameCommand  `json:"command"`
	Snapshot GameSnapshot `json:"snapshot"`
	Tile     string       `json:"tile,omitempty"`
}

func (g *Game) Snapshot() GameSnapshot {
	players := make([]PlayerView, len(g.Players))
	for i, player := range g.Players {
		players[i] = PlayerView{
			ID:        playerID(i),
			Name:      player.Name,
			Human:     player.Human,
			Hand:      append([]Tile(nil), player.Hand...),
			HandCount: len(player.Hand),
			Flowers:   append([]Tile(nil), player.Flowers...),
			Melds:     copyMelds(player.Melds),
			Discards:  append([]Tile(nil), player.Discards...),
		}
	}
	return GameSnapshot{
		Seed:         g.Seed,
		RNGMode:      g.RNGMode,
		ShuffleProof: g.ShuffleProof,
		Players:      players,
		WallCount:    len(g.Wall),
		Current:      g.Current,
		Winner:       g.Winner,
		Discarder:    g.Discarder,
		Dealer:       g.Dealer,
		HandNumber:   g.HandNumber,
		MCRScore:     copyMCRScore(g.MCRScore),
		Reason:       g.Reason,
		Over:         g.Over,
		Events:       append([]GameEvent(nil), g.Events...),
		Phase:        g.Phase,
		PendingClaim: copyPendingClaim(g.PendingClaim),
		LegalActions: copyLegalActions(g.rules.LegalActions(g, playerID(g.Current))),
	}
}

func (g *Game) SnapshotFor(id string) GameSnapshot {
	snapshot := g.Snapshot()
	if g.Over || g.Phase == PhaseRoundOver {
		return snapshot
	}

	snapshot.Seed = 0
	snapshot.ShuffleProof.Seed = 0
	viewer := -1
	for index := range snapshot.Players {
		if snapshot.Players[index].ID == id {
			viewer = index
		} else {
			snapshot.Players[index].Hand = nil
		}
	}
	for index := range snapshot.Events {
		if (snapshot.Events[index].Kind == EventDraw || snapshot.Events[index].Kind == EventReplacementDraw) && snapshot.Events[index].Player != viewer {
			snapshot.Events[index].Tile = -1
		}
	}
	if viewer != snapshot.Current {
		snapshot.PendingClaim = nil
		snapshot.LegalActions = nil
	}
	return snapshot
}

func (g *Game) ApplyCommand(command GameCommand) CommandResult {
	if g.Over || g.Phase == PhaseRoundOver {
		return g.commandError(command, "game is over")
	}
	if command.PlayerID != playerID(g.Current) {
		return g.commandError(command, "not the current player")
	}
	if !g.rules.Allows(g, command) {
		return g.commandError(command, "command is not legal")
	}
	if g.Phase == PhaseAwaitingClaim {
		return g.applyClaimCommand(command)
	}
	switch command.Kind {
	case CommandDiscard:
		tile, err := g.discardCurrent(command.TileIndex)
		if err != nil {
			return g.commandError(command, err.Error())
		}
		return g.commandOK(command, tile.String())
	case CommandWin:
		if _, mcr := g.rules.(*MCRRuleSet); !mcr && !CanWin(g.Players[g.Current].Hand) {
			return g.commandError(command, "hand is not complete")
		}
		if rules, ok := g.rules.(*MCRRuleSet); ok {
			if score, valid := rules.selfDrawScore(g, g.Current); valid {
				g.MCRScore = &score
				g.Discarder = -1
			}
		}
		g.finish(g.Current, "self-draw", WinSelfDraw)
		return g.commandOK(command, "")
	case CommandKong:
		if !g.tryCurrentKong(command.Tile, io.Discard) {
			return g.commandError(command, "kong is not available")
		}
		return g.commandOK(command, command.Tile)
	case CommandQuit:
		g.Quit("quit")
		return g.commandOK(command, "")
	default:
		return g.commandError(command, "unknown command")
	}
}

func (g *Game) EnsureCurrentTurnDraw() (Tile, bool) {
	if g.Over || g.Phase != PhaseAwaitingDiscard || g.Current < 0 || g.Current >= len(g.Players) {
		return -1, false
	}
	if len(g.Players[g.Current].Hand)%3 != 1 {
		return -1, false
	}
	if len(g.Wall) == 0 {
		g.Over = true
		g.Reason = "draw: wall exhausted"
		g.RecordEvent(EventWallExhausted, g.Current, -1, g.Reason)
		return -1, false
	}
	tile := g.draw(g.Current)
	return tile, tile >= 0
}

func (g *Game) discardCurrent(index int) (Tile, error) {
	if g.Over {
		return -1, io.ErrClosedPipe
	}
	discard, err := g.Players[g.Current].RemoveAt(index)
	if err != nil {
		return -1, err
	}
	g.Players[g.Current].Discards = append(g.Players[g.Current].Discards, discard)
	g.RecordEvent(EventDiscard, g.Current, discard, "")
	g.beginDiscardClaims(g.Current, discard)
	return discard, nil
}

func (g *Game) tryCurrentKong(tileText string, out io.Writer) bool {
	if rules, ok := g.rules.(*MCRRuleSet); ok {
		return rules.declareKong(g, tileText)
	}
	if rules, ok := g.rules.(*RiichiRuleSet); ok {
		return rules.declareKong(g, tileText)
	}
	tile, ok := ParseTile(tileText)
	if !ok || g.Players[g.Current].Count(tile) < 4 {
		return false
	}
	for i := 0; i < 4; i++ {
		g.Players[g.Current].RemoveTile(tile)
	}
	g.Players[g.Current].AddMeld(MeldKong, []Tile{tile, tile, tile, tile})
	g.RecordEvent(EventKong, g.Current, tile, "concealed kong")
	if len(g.Wall) == 0 {
		g.Over = true
		g.Reason = "draw: wall exhausted after kong"
		return true
	}
	g.drawReplacement(g.Current)
	return true
}

func (g *Game) commandOK(command GameCommand, tile string) CommandResult {
	return CommandResult{OK: true, Command: command, Snapshot: g.Snapshot(), Tile: tile}
}

func (g *Game) commandError(command GameCommand, message string) CommandResult {
	return CommandResult{OK: false, Error: message, Command: command, Snapshot: g.Snapshot()}
}

func copyMelds(melds []Meld) []Meld {
	out := make([]Meld, len(melds))
	for i, meld := range melds {
		out[i] = Meld{Kind: meld.Kind, Tiles: append([]Tile(nil), meld.Tiles...)}
	}
	return out
}

func copyLegalActions(actions []LegalAction) []LegalAction {
	out := make([]LegalAction, len(actions))
	for i, action := range actions {
		out[i] = action
		out[i].Consumed = append([]Tile(nil), action.Consumed...)
	}
	return out
}

func copyMCRScore(score *MCRScoreBreakdown) *MCRScoreBreakdown {
	if score == nil {
		return nil
	}
	copyValue := *score
	copyValue.Fans = append([]FanMatch(nil), score.Fans...)
	copyValue.WinningGrouping = copyMelds(score.WinningGrouping)
	return &copyValue
}

func playerID(index int) string {
	return string(rune('0' + index))
}
