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
	ClaimKong ClaimKind = "kong"
	ClaimPong ClaimKind = "pong"
	ClaimChow ClaimKind = "chow"
)

type ClaimOption struct {
	Kind     ClaimKind `json:"kind"`
	Player   int       `json:"player"`
	Consumed []Tile    `json:"consumed,omitempty"`
}

type PendingClaim struct {
	Discarder   int           `json:"discarder"`
	Tile        Tile          `json:"tile"`
	Options     []ClaimOption `json:"options"`
	Active      int           `json:"active"`
	RobbingKong bool          `json:"robbing_kong,omitempty"`
	KongPlayer  int           `json:"kong_player,omitempty"`
	KongMeld    int           `json:"kong_meld,omitempty"`
}

func (g *Game) buildPendingClaim(discarder int, discard Tile) *PendingClaim {
	if rules, ok := g.rules.(*MCRRuleSet); ok {
		return rules.buildPendingClaim(g, discarder, discard)
	}
	if rules, ok := g.rules.(*RiichiRuleSet); ok {
		return rules.buildPendingClaim(g, discarder, discard)
	}
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
		Discarder:   pending.Discarder,
		Tile:        pending.Tile,
		Active:      pending.Active,
		RobbingKong: pending.RobbingKong,
		KongPlayer:  pending.KongPlayer,
		KongMeld:    pending.KongMeld,
		Options:     make([]ClaimOption, len(pending.Options)),
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

func (g *Game) beginDiscardClaims(discarder int, discard Tile) {
	g.PendingClaim = g.buildPendingClaim(discarder, discard)
	if g.PendingClaim == nil {
		g.completeUnclaimedDiscard(discarder)
		return
	}
	g.Phase = PhaseAwaitingClaim
	g.Current = g.PendingClaim.Options[0].Player
}

func (g *Game) applyClaimCommand(command GameCommand) CommandResult {
	options := g.activeClaimOptions()
	if len(options) == 0 {
		return g.commandError(command, "no active claim")
	}
	if command.Kind == CommandPass {
		g.passActiveClaim(len(options))
		return g.commandOK(command, "")
	}

	wanted := claimKindForCommand(command.Kind)
	if wanted == "" || wanted != options[0].Kind {
		return g.commandError(command, "claim is not available")
	}
	optionIndex := 0
	if wanted == ClaimChow {
		optionIndex = command.TileIndex
	}
	if optionIndex < 0 || optionIndex >= len(options) {
		return g.commandError(command, "claim option is not available")
	}
	option := options[optionIndex]
	switch option.Kind {
	case ClaimWin:
		reason := "discard-win"
		robbingKong := g.PendingClaim.RobbingKong
		if robbingKong {
			reason = "robbing-kong"
		}
		if rules, ok := g.rules.(*MCRRuleSet); ok {
			score := rules.discardWinScoreWithContext(g, option.Player, g.PendingClaim.Discarder, g.PendingClaim.Tile, robbingKong)
			g.MCRScore = &score
			g.Discarder = g.PendingClaim.Discarder
		}
		g.finish(option.Player, reason, WinDiscard)
	case ClaimKong:
		tile := g.PendingClaim.Tile
		g.removeClaimedDiscard()
		g.claimExposedKong(option.Player, tile)
		g.completeAcceptedClaim(option.Player)
		if rules, ok := g.rules.(*RiichiRuleSet); ok {
			rules.afterAcceptedKong(g)
		}
		g.drawReplacement(option.Player)
	case ClaimPong:
		g.removeClaimedDiscard()
		g.claimPong(option.Player, g.PendingClaim.Tile)
		g.completeAcceptedClaim(option.Player)
	case ClaimChow:
		g.removeClaimedDiscard()
		meld := append([]Tile(nil), option.Consumed...)
		meld = append(meld, g.PendingClaim.Tile)
		SortTiles(meld)
		g.claimChow(option.Player, g.PendingClaim.Tile, meld)
		g.completeAcceptedClaim(option.Player)
	}
	return g.commandOK(command, "")
}

func (g *Game) activeClaimOptions() []ClaimOption {
	if g.PendingClaim == nil || g.PendingClaim.Active < 0 || g.PendingClaim.Active >= len(g.PendingClaim.Options) {
		return nil
	}
	first := g.PendingClaim.Options[g.PendingClaim.Active]
	end := g.PendingClaim.Active + 1
	for end < len(g.PendingClaim.Options) {
		option := g.PendingClaim.Options[end]
		if option.Player != first.Player || option.Kind != first.Kind {
			break
		}
		end++
	}
	return g.PendingClaim.Options[g.PendingClaim.Active:end]
}

func (g *Game) passActiveClaim(activeCount int) {
	g.PendingClaim.Active += activeCount
	if g.PendingClaim.Active >= len(g.PendingClaim.Options) {
		if g.PendingClaim.RobbingKong {
			if rules, ok := g.rules.(*MCRRuleSet); ok {
				rules.completeAddedKong(g)
				return
			}
			if rules, ok := g.rules.(*RiichiRuleSet); ok {
				rules.completeAddedKong(g)
				return
			}
		}
		discarder := g.PendingClaim.Discarder
		g.completeUnclaimedDiscard(discarder)
		return
	}
	g.Current = g.PendingClaim.Options[g.PendingClaim.Active].Player
}

func (g *Game) completeUnclaimedDiscard(discarder int) {
	g.PendingClaim = nil
	g.Phase = PhaseAwaitingDiscard
	g.Current = (discarder + 1) % len(g.Players)
}

func (g *Game) completeAcceptedClaim(player int) {
	g.PendingClaim = nil
	g.Phase = PhaseAwaitingDiscard
	g.Current = player
}

func (g *Game) removeClaimedDiscard() {
	if g.PendingClaim == nil {
		return
	}
	discards := g.Players[g.PendingClaim.Discarder].Discards
	for i := len(discards) - 1; i >= 0; i-- {
		if discards[i] == g.PendingClaim.Tile {
			g.Players[g.PendingClaim.Discarder].Discards = append(discards[:i], discards[i+1:]...)
			return
		}
	}
}

func claimKindForCommand(command CommandKind) ClaimKind {
	switch command {
	case CommandClaimWin:
		return ClaimWin
	case CommandKong:
		return ClaimKong
	case CommandPong:
		return ClaimPong
	case CommandChow:
		return ClaimChow
	default:
		return ""
	}
}
