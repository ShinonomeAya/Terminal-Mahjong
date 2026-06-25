package game

import "fmt"

func (rules *RiichiRuleSet) riichiDeclarationDiscardIndexes(round *Game, playerIndex int) []int {
	if !rules.canDeclareRiichi(round, playerIndex) {
		return nil
	}
	player := round.Players[playerIndex]
	indexes := make([]int, 0, len(player.Hand))
	for index := range player.Hand {
		hand := append([]Tile(nil), player.Hand...)
		hand = append(hand[:index], hand[index+1:]...)
		if RiichiTenpai(hand, player.Melds) {
			indexes = append(indexes, index)
		}
	}
	return indexes
}

func (rules *RiichiRuleSet) canDeclareRiichi(round *Game, playerIndex int) bool {
	if round == nil || round.Riichi == nil || playerIndex < 0 || playerIndex >= len(round.Players) {
		return false
	}
	if !riichiDeclarationIsNone(round.Riichi.Declarations[playerIndex]) {
		return false
	}
	player := round.Players[playerIndex]
	if len(player.Melds) != 0 || len(round.Wall) < 4 || len(player.Hand)%3 != 2 {
		return false
	}
	return true
}

func (rules *RiichiRuleSet) declareRiichi(round *Game, tileIndex int) (Tile, error) {
	if !containsInt(rules.riichiDeclarationDiscardIndexes(round, round.Current), tileIndex) {
		return -1, fmt.Errorf("riichi declaration is not legal")
	}
	round.Riichi.Declarations[round.Current] = RiichiDeclared
	round.Riichi.RiichiSticks++
	tile, err := round.discardCurrent(tileIndex)
	if err != nil {
		round.Riichi.Declarations[round.Current] = RiichiNone
		round.Riichi.RiichiSticks--
		return -1, err
	}
	return tile, nil
}

func (rules *RiichiRuleSet) acceptRiichiDeclaration(round *Game, playerIndex int) {
	if round == nil || round.Riichi == nil || playerIndex < 0 || playerIndex >= len(round.Players) {
		return
	}
	if round.Riichi.Declarations[playerIndex] != RiichiDeclared {
		return
	}
	round.Riichi.Declarations[playerIndex] = RiichiAccepted
	round.Riichi.Ippatsu[playerIndex] = true
}

func (rules *RiichiRuleSet) recordClaimPass(round *Game, playerIndex int, options []ClaimOption) {
	if round == nil || round.Riichi == nil || playerIndex < 0 || playerIndex >= len(round.Players) {
		return
	}
	passedRon := false
	for _, option := range options {
		if option.Kind == ClaimWin && option.Player == playerIndex {
			passedRon = true
			break
		}
	}
	if !passedRon {
		return
	}
	if round.Riichi.Declarations[playerIndex] == RiichiAccepted {
		round.Riichi.RiichiFuriten[playerIndex] = true
		round.Riichi.TemporaryFuriten[playerIndex] = false
		return
	}
	round.Riichi.TemporaryFuriten[playerIndex] = true
}

func (rules *RiichiRuleSet) isFuritenForRon(round *Game, playerIndex int, discard Tile) bool {
	if round == nil || round.Riichi == nil || playerIndex < 0 || playerIndex >= len(round.Players) {
		return true
	}
	if round.Riichi.TemporaryFuriten[playerIndex] || round.Riichi.RiichiFuriten[playerIndex] {
		return true
	}
	player := round.Players[playerIndex]
	waits := RiichiWaits(player.Hand, player.Melds)
	for _, ownDiscard := range player.Discards {
		for _, wait := range waits {
			if ownDiscard.Base() == wait.Base() {
				return true
			}
		}
	}
	return false
}

func (rules *RiichiRuleSet) cancelIppatsu(round *Game) {
	if round == nil || round.Riichi == nil {
		return
	}
	for index := range round.Riichi.Ippatsu {
		round.Riichi.Ippatsu[index] = false
	}
}

func containsInt(values []int, needle int) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func riichiDeclarationIsNone(state RiichiDeclarationState) bool {
	return state == "" || state == RiichiNone
}
