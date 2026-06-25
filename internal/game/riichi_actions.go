package game

func (rules *RiichiRuleSet) legalActions(round *Game, id string) []LegalAction {
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
	if len(player.Hand)%3 != 2 {
		return nil
	}
	actions := make([]LegalAction, 0, len(player.Hand)+4)
	for index := range player.Hand {
		actions = append(actions, LegalAction{Kind: CommandDiscard, TileIndex: index})
	}
	if riichiCanSelfDrawWin(round, round.Current) {
		actions = append(actions, LegalAction{Kind: CommandWin})
	}
	for _, index := range rules.riichiDeclarationDiscardIndexes(round, round.Current) {
		actions = append(actions, LegalAction{Kind: CommandRiichi, TileIndex: index})
	}
	counts := TileCounts(player.Hand)
	for tile, count := range counts {
		base := Tile(tile)
		if count == 4 || (count >= 1 && playerHasBasePong(player, base)) {
			actions = append(actions, LegalAction{Kind: CommandKong, Tile: base.String()})
		}
	}
	return append(actions, LegalAction{Kind: CommandQuit})
}

func (rules *RiichiRuleSet) allows(round *Game, command GameCommand) bool {
	if round == nil || round.Over || round.Phase == PhaseRoundOver || command.PlayerID != playerID(round.Current) {
		return false
	}
	if round.Phase == PhaseAwaitingClaim {
		return rules.allowsClaimCommand(round, command)
	}
	if round.Phase != PhaseAwaitingDiscard || round.Current < 0 || round.Current >= len(round.Players) {
		return false
	}
	player := round.Players[round.Current]
	if len(player.Hand)%3 != 2 {
		return command.Kind == CommandQuit
	}
	switch command.Kind {
	case CommandDiscard:
		return command.TileIndex >= 0 && command.TileIndex < len(player.Hand)
	case CommandWin:
		return riichiCanSelfDrawWin(round, round.Current)
	case CommandRiichi:
		for _, index := range rules.riichiDeclarationDiscardIndexes(round, round.Current) {
			if index == command.TileIndex {
				return true
			}
		}
		return false
	case CommandKong:
		tile, ok := ParseTile(command.Tile)
		if !ok {
			return false
		}
		base := tile.Base()
		return player.Count(base) >= 4 || (player.Count(base) >= 1 && playerHasBasePong(player, base))
	case CommandQuit:
		return true
	default:
		return false
	}
}

func (rules *RiichiRuleSet) allowsClaimCommand(round *Game, command GameCommand) bool {
	options := round.activeClaimOptions()
	if len(options) == 0 {
		return command.Kind == CommandPass
	}
	if command.Kind == CommandPass {
		return true
	}
	for index, option := range options {
		if commandKindForClaim(option.Kind) != command.Kind {
			continue
		}
		if option.Kind == ClaimChow && command.TileIndex != index {
			continue
		}
		return true
	}
	return false
}

func (rules *RiichiRuleSet) buildPendingClaim(round *Game, discarder int, discard Tile) *PendingClaim {
	options := make([]ClaimOption, 0)
	for offset := 1; offset < len(round.Players); offset++ {
		playerIndex := (discarder + offset) % len(round.Players)
		if riichiCanWinOnDiscard(round, playerIndex, discard) && !rules.isFuritenForRon(round, playerIndex, discard) {
			options = append(options, ClaimOption{Kind: ClaimWin, Player: playerIndex})
		}
	}
	baseDiscard := discard.Base()
	for offset := 1; offset < len(round.Players); offset++ {
		playerIndex := (discarder + offset) % len(round.Players)
		count := round.Players[playerIndex].Count(baseDiscard)
		if count >= 3 {
			options = append(options, ClaimOption{Kind: ClaimKong, Player: playerIndex, Consumed: []Tile{baseDiscard, baseDiscard, baseDiscard}})
		}
		if count >= 2 {
			options = append(options, ClaimOption{Kind: ClaimPong, Player: playerIndex, Consumed: []Tile{baseDiscard, baseDiscard}})
		}
	}
	nextPlayer := (discarder + 1) % len(round.Players)
	for _, meld := range ChowOptions(round.Players[nextPlayer], baseDiscard) {
		options = append(options, ClaimOption{Kind: ClaimChow, Player: nextPlayer, Consumed: chowHandTiles(meld, baseDiscard)})
	}
	if len(options) == 0 {
		return nil
	}
	return &PendingClaim{Discarder: discarder, Tile: discard, Options: options}
}

func (rules *RiichiRuleSet) declareKong(round *Game, tileText string) bool {
	tile, ok := ParseTile(tileText)
	if !ok {
		return false
	}
	player := &round.Players[round.Current]
	if player.Count(tile) >= 4 {
		meldTiles, ok := removeRiichiTiles(player, tile, 4)
		if !ok {
			return false
		}
		player.AddMeld(MeldKong, meldTiles)
		round.RecordEvent(EventKong, round.Current, tile.Base(), "concealed kong")
		rules.afterAcceptedKong(round)
		round.drawReplacement(round.Current)
		return true
	}

	meldIndex := riichiPongMeldIndex(*player, tile)
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
	round.PendingClaim = &PendingClaim{RobbingKong: true, Tile: tile.Base(), KongPlayer: round.Current, KongMeld: meldIndex}
	rules.completeAddedKong(round)
	return true
}

func (rules *RiichiRuleSet) buildRobbingKongClaim(round *Game, player int, tile Tile, meldIndex int) *PendingClaim {
	var options []ClaimOption
	for offset := 1; offset < len(round.Players); offset++ {
		candidate := (player + offset) % len(round.Players)
		if riichiCanWinOnDiscard(round, candidate, tile) {
			options = append(options, ClaimOption{Kind: ClaimWin, Player: candidate})
		}
	}
	if len(options) == 0 {
		return nil
	}
	return &PendingClaim{
		Discarder:   player,
		Tile:        tile.Base(),
		Options:     options,
		RobbingKong: true,
		KongPlayer:  player,
		KongMeld:    meldIndex,
	}
}

func (rules *RiichiRuleSet) completeAddedKong(round *Game) {
	pending := round.PendingClaim
	if pending == nil || !pending.RobbingKong || pending.KongPlayer < 0 || pending.KongPlayer >= len(round.Players) {
		return
	}
	playerIndex, meldIndex, tile := pending.KongPlayer, pending.KongMeld, pending.Tile.Base()
	player := &round.Players[playerIndex]
	if meldIndex < 0 || meldIndex >= len(player.Melds) || player.Melds[meldIndex].Kind != MeldPong {
		return
	}
	removed, ok := removeRiichiTiles(player, tile, 1)
	if !ok {
		return
	}
	player.Melds[meldIndex].Kind = MeldKong
	player.Melds[meldIndex].Tiles = append(player.Melds[meldIndex].Tiles, removed[0])
	round.RecordEvent(EventKong, playerIndex, tile, "added kong")
	round.PendingClaim = nil
	round.Phase = PhaseAwaitingDiscard
	round.Current = playerIndex
	rules.afterAcceptedKong(round)
	round.drawReplacement(playerIndex)
}

func (rules *RiichiRuleSet) afterAcceptedKong(round *Game) {
	rules.cancelIppatsu(round)
	revealRiichiKanDora(round)
}

func riichiCanWinOnDiscard(round *Game, playerIndex int, discard Tile) bool {
	if playerIndex < 0 || playerIndex >= len(round.Players) {
		return false
	}
	player := round.Players[playerIndex]
	return len(RiichiDecompose(player.Hand, player.Melds, discard)) > 0
}

func riichiCanSelfDrawWin(round *Game, playerIndex int) bool {
	winningTile, ok := lastPrivateDraw(round.Events, playerIndex)
	if !ok || playerIndex < 0 || playerIndex >= len(round.Players) {
		return false
	}
	hand := append([]Tile(nil), round.Players[playerIndex].Hand...)
	if !removeOneRiichiTile(&hand, winningTile) {
		return false
	}
	return len(RiichiDecompose(hand, round.Players[playerIndex].Melds, winningTile)) > 0
}

func removeOneRiichiTile(hand *[]Tile, tile Tile) bool {
	base := tile.Base()
	for index, value := range *hand {
		if value.Base() == base {
			*hand = append((*hand)[:index], (*hand)[index+1:]...)
			return true
		}
	}
	return false
}

func removeRiichiTiles(player *Player, tile Tile, count int) ([]Tile, bool) {
	removed := make([]Tile, 0, count)
	base := tile.Base()
	for len(removed) < count {
		index := -1
		for handIndex, handTile := range player.Hand {
			if handTile.Base() == base {
				index = handIndex
				break
			}
		}
		if index < 0 {
			for _, tile := range removed {
				player.AddTile(tile)
			}
			return nil, false
		}
		removed = append(removed, player.Hand[index])
		player.Hand = append(player.Hand[:index], player.Hand[index+1:]...)
	}
	SortTiles(removed)
	return removed, true
}

func playerHasBasePong(player Player, tile Tile) bool {
	base := tile.Base()
	for _, meld := range player.Melds {
		if meld.Kind == MeldPong && len(meld.Tiles) == 3 && meld.Tiles[0].Base() == base {
			return true
		}
	}
	return false
}

func riichiPongMeldIndex(player Player, tile Tile) int {
	base := tile.Base()
	for index, meld := range player.Melds {
		if meld.Kind == MeldPong && len(meld.Tiles) == 3 && meld.Tiles[0].Base() == base {
			return index
		}
	}
	return -1
}
