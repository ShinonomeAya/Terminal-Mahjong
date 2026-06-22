package game

type TurnPhase string

const (
	PhaseAwaitingDiscard TurnPhase = "awaiting_discard"
	PhaseAwaitingClaim   TurnPhase = "awaiting_claim"
	PhaseRoundOver       TurnPhase = "round_over"
)

type ClaimKind string

const (
	ClaimWin  ClaimKind = "win"
	ClaimPong ClaimKind = "pong"
	ClaimChow ClaimKind = "chow"
)

type ClaimOption struct {
	Kind     ClaimKind `json:"kind"`
	Player   int       `json:"player"`
	Consumed []Tile    `json:"consumed,omitempty"`
}

type PendingClaim struct {
	Discarder int           `json:"discarder"`
	Tile      Tile          `json:"tile"`
	Options   []ClaimOption `json:"options"`
	Active    int           `json:"active"`
}

func (g *Game) buildPendingClaim(discarder int, discard Tile) *PendingClaim {
	options := make([]ClaimOption, 0)
	for offset := 1; offset < len(g.Players); offset++ {
		playerIndex := (discarder + offset) % len(g.Players)
		hand := append([]Tile(nil), g.Players[playerIndex].Hand...)
		hand = append(hand, discard)
		SortTiles(hand)
		if CanWin(hand) {
			options = append(options, ClaimOption{Kind: ClaimWin, Player: playerIndex})
		}
	}
	for offset := 1; offset < len(g.Players); offset++ {
		playerIndex := (discarder + offset) % len(g.Players)
		if g.Players[playerIndex].Count(discard) >= 2 {
			options = append(options, ClaimOption{
				Kind:     ClaimPong,
				Player:   playerIndex,
				Consumed: []Tile{discard, discard},
			})
		}
	}
	nextPlayer := (discarder + 1) % len(g.Players)
	for _, meld := range ChowOptions(g.Players[nextPlayer], discard) {
		options = append(options, ClaimOption{
			Kind:     ClaimChow,
			Player:   nextPlayer,
			Consumed: chowHandTiles(meld, discard),
		})
	}
	if len(options) == 0 {
		return nil
	}
	return &PendingClaim{Discarder: discarder, Tile: discard, Options: options}
}

func copyPendingClaim(pending *PendingClaim) *PendingClaim {
	if pending == nil {
		return nil
	}
	copyValue := &PendingClaim{
		Discarder: pending.Discarder,
		Tile:      pending.Tile,
		Active:    pending.Active,
		Options:   make([]ClaimOption, len(pending.Options)),
	}
	for i, option := range pending.Options {
		copyValue.Options[i] = ClaimOption{
			Kind:     option.Kind,
			Player:   option.Player,
			Consumed: append([]Tile(nil), option.Consumed...),
		}
	}
	return copyValue
}
