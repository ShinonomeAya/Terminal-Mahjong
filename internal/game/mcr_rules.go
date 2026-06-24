package game

func (rules *MCRRuleSet) legalActions(round *Game, id string) []LegalAction {
	if round == nil || round.Over || round.Phase == PhaseRoundOver || id != playerID(round.Current) {
		return nil
	}
	if round.Phase == PhaseAwaitingClaim {
		options := round.activeClaimOptions()
		if len(options) == 0 {
			return nil
		}
		actions := []LegalAction{{Kind: CommandPass}}
		for index, option := range options {
			action := LegalAction{Kind: commandKindForClaim(option.Kind), Consumed: append([]Tile(nil), option.Consumed...)}
			if option.Kind == ClaimChow {
				action.TileIndex = index
			}
			actions = append(actions, action)
		}
		return actions
	}
	if round.Phase != PhaseAwaitingDiscard {
		return nil
	}

	player := round.Players[round.Current]
	actions := make([]LegalAction, 0, len(player.Hand)+4)
	for index := range player.Hand {
		actions = append(actions, LegalAction{Kind: CommandDiscard, TileIndex: index})
	}
	if score, ok := rules.selfDrawScore(round, round.Current); ok && score.MeetsMinimum {
		actions = append(actions, LegalAction{Kind: CommandWin})
	}
	counts := TileCounts(player.Hand)
	for tile, count := range counts {
		if count == 4 || (count >= 1 && playerHasPong(player, Tile(tile))) {
			actions = append(actions, LegalAction{Kind: CommandKong, Tile: Tile(tile).String()})
		}
	}
	return append(actions, LegalAction{Kind: CommandQuit})
}

func (rules *MCRRuleSet) allows(round *Game, command GameCommand) bool {
	for _, action := range rules.legalActions(round, command.PlayerID) {
		if action.Kind != command.Kind {
			continue
		}
		switch action.Kind {
		case CommandDiscard, CommandChow:
			if action.TileIndex != command.TileIndex {
				continue
			}
		case CommandKong:
			if action.Tile != command.Tile && action.Tile != "" {
				continue
			}
		}
		return true
	}
	return false
}

func (rules *MCRRuleSet) buildPendingClaim(round *Game, discarder int, discard Tile) *PendingClaim {
	options := make([]ClaimOption, 0)
	for offset := 1; offset < len(round.Players); offset++ {
		playerIndex := (discarder + offset) % len(round.Players)
		if rules.discardWinScore(round, playerIndex, discarder, discard).MeetsMinimum {
			options = append(options, ClaimOption{Kind: ClaimWin, Player: playerIndex})
		}
	}
	for offset := 1; offset < len(round.Players); offset++ {
		playerIndex := (discarder + offset) % len(round.Players)
		count := round.Players[playerIndex].Count(discard)
		if count >= 3 {
			options = append(options, ClaimOption{Kind: ClaimKong, Player: playerIndex, Consumed: []Tile{discard, discard, discard}})
		}
		if count >= 2 {
			options = append(options, ClaimOption{Kind: ClaimPong, Player: playerIndex, Consumed: []Tile{discard, discard}})
		}
	}
	nextPlayer := (discarder + 1) % len(round.Players)
	for _, meld := range ChowOptions(round.Players[nextPlayer], discard) {
		options = append(options, ClaimOption{Kind: ClaimChow, Player: nextPlayer, Consumed: chowHandTiles(meld, discard)})
	}
	if len(options) == 0 {
		return nil
	}
	return &PendingClaim{Discarder: discarder, Tile: discard, Options: options}
}

func (rules *MCRRuleSet) selfDrawScore(round *Game, playerIndex int) (MCRScoreBreakdown, bool) {
	winningTile, ok := lastPrivateDraw(round.Events, playerIndex)
	if !ok {
		return MCRScoreBreakdown{}, false
	}
	hand := append([]Tile(nil), round.Players[playerIndex].Hand...)
	if !removeOneMCRTile(&hand, winningTile) {
		return MCRScoreBreakdown{}, false
	}
	player := round.Players[playerIndex]
	return ScoreMCR(hand, player.Melds, MCRScoreContext{
		Winner:          playerIndex,
		Discarder:       -1,
		WinningTile:     winningTile,
		WinType:         WinSelfDraw,
		Flowers:         len(player.Flowers),
		ReplacementDraw: lastDrawWasReplacement(round.Events, playerIndex),
	}), true
}

func (rules *MCRRuleSet) discardWinScore(round *Game, playerIndex, discarder int, discard Tile) MCRScoreBreakdown {
	return rules.discardWinScoreWithContext(round, playerIndex, discarder, discard, false)
}

func (rules *MCRRuleSet) discardWinScoreWithContext(round *Game, playerIndex, discarder int, discard Tile, robbingKong bool) MCRScoreBreakdown {
	player := round.Players[playerIndex]
	return ScoreMCR(player.Hand, player.Melds, MCRScoreContext{
		Winner:      playerIndex,
		Discarder:   discarder,
		WinningTile: discard,
		WinType:     WinDiscard,
		Flowers:     len(player.Flowers),
		RobbingKong: robbingKong,
	})
}

func lastPrivateDraw(events []GameEvent, player int) (Tile, bool) {
	for index := len(events) - 1; index >= 0; index-- {
		event := events[index]
		if event.Player == player && (event.Kind == EventDraw || event.Kind == EventReplacementDraw) {
			return event.Tile, true
		}
	}
	return -1, false
}

func lastDrawWasReplacement(events []GameEvent, player int) bool {
	for index := len(events) - 1; index >= 0; index-- {
		event := events[index]
		if event.Player == player && (event.Kind == EventDraw || event.Kind == EventReplacementDraw) {
			return event.Kind == EventReplacementDraw
		}
	}
	return false
}

func removeOneMCRTile(hand *[]Tile, tile Tile) bool {
	for index, value := range *hand {
		if value == tile {
			*hand = append((*hand)[:index], (*hand)[index+1:]...)
			return true
		}
	}
	return false
}

func playerHasPong(player Player, tile Tile) bool {
	for _, meld := range player.Melds {
		if meld.Kind == MeldPong && len(meld.Tiles) == 3 && meld.Tiles[0] == tile {
			return true
		}
	}
	return false
}

func (rules *MCRRuleSet) declareKong(round *Game, tileText string) bool {
	tile, ok := ParseTile(tileText)
	if !ok {
		return false
	}
	player := &round.Players[round.Current]
	if player.Count(tile) >= 4 {
		for count := 0; count < 4; count++ {
			player.RemoveTile(tile)
		}
		player.AddMeld(MeldKong, []Tile{tile, tile, tile, tile})
		round.RecordEvent(EventKong, round.Current, tile, "concealed kong")
		round.drawReplacement(round.Current)
		return true
	}

	meldIndex := pongMeldIndex(*player, tile)
	if player.Count(tile) < 1 || meldIndex < 0 {
		return false
	}
	pending := rules.buildRobbingKongClaim(round, round.Current, tile, meldIndex)
	if pending != nil {
		round.PendingClaim = pending
		round.Phase = PhaseAwaitingClaim
		round.Current = pending.Options[0].Player
		return true
	}
	round.PendingClaim = &PendingClaim{RobbingKong: true, Tile: tile, KongPlayer: round.Current, KongMeld: meldIndex}
	rules.completeAddedKong(round)
	return true
}

func (rules *MCRRuleSet) buildRobbingKongClaim(round *Game, player int, tile Tile, meldIndex int) *PendingClaim {
	var options []ClaimOption
	for offset := 1; offset < len(round.Players); offset++ {
		candidate := (player + offset) % len(round.Players)
		if rules.discardWinScoreWithContext(round, candidate, player, tile, true).MeetsMinimum {
			options = append(options, ClaimOption{Kind: ClaimWin, Player: candidate})
		}
	}
	if len(options) == 0 {
		return nil
	}
	return &PendingClaim{
		Discarder:   player,
		Tile:        tile,
		Options:     options,
		RobbingKong: true,
		KongPlayer:  player,
		KongMeld:    meldIndex,
	}
}

func (rules *MCRRuleSet) completeAddedKong(round *Game) {
	pending := round.PendingClaim
	if pending == nil || !pending.RobbingKong || pending.KongPlayer < 0 || pending.KongPlayer >= len(round.Players) {
		return
	}
	playerIndex, meldIndex, tile := pending.KongPlayer, pending.KongMeld, pending.Tile
	player := &round.Players[playerIndex]
	if meldIndex < 0 || meldIndex >= len(player.Melds) || player.Melds[meldIndex].Kind != MeldPong || !player.RemoveTile(tile) {
		return
	}
	player.Melds[meldIndex].Kind = MeldKong
	player.Melds[meldIndex].Tiles = append(player.Melds[meldIndex].Tiles, tile)
	round.RecordEvent(EventKong, playerIndex, tile, "added kong")
	round.PendingClaim = nil
	round.Phase = PhaseAwaitingDiscard
	round.Current = playerIndex
	round.drawReplacement(playerIndex)
}

func pongMeldIndex(player Player, tile Tile) int {
	for index, meld := range player.Melds {
		if meld.Kind == MeldPong && len(meld.Tiles) == 3 && meld.Tiles[0] == tile {
			return index
		}
	}
	return -1
}
