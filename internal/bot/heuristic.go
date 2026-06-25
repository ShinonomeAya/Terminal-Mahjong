package bot

import (
	"context"

	"mahjong/internal/game"
)

type BotEngine interface {
	Decide(ctx context.Context, snapshot game.GameSnapshot, playerID string) game.GameCommand
}

type HeuristicBot struct{}

func NewHeuristicBot() *HeuristicBot {
	return &HeuristicBot{}
}

func (b *HeuristicBot) Decide(_ context.Context, snapshot game.GameSnapshot, playerID string) game.GameCommand {
	player, ok := playerByID(snapshot, playerID)
	if !ok {
		return game.GameCommand{PlayerID: playerID, Kind: game.CommandQuit}
	}
	if snapshot.Phase == game.PhaseAwaitingClaim {
		if snapshot.PendingClaim != nil {
			return decideClaim(snapshot, player, playerID)
		}
		if len(snapshot.LegalActions) > 0 {
			return decideLegal(snapshot.LegalActions, player, playerID)
		}
		return decideClaim(snapshot, player, playerID)
	}
	if len(snapshot.LegalActions) > 0 {
		return decideLegal(snapshot.LegalActions, player, playerID)
	}
	if game.CanWin(player.Hand) {
		return game.GameCommand{PlayerID: playerID, Kind: game.CommandWin}
	}
	if tile, ok := concealedKongTile(player.Hand); ok {
		return game.GameCommand{PlayerID: playerID, Kind: game.CommandKong, Tile: tile.String()}
	}
	index := game.ChooseAIDiscard(player.Hand)
	return game.GameCommand{PlayerID: playerID, Kind: game.CommandDiscard, TileIndex: index}
}

func decideLegal(actions []game.LegalAction, player game.PlayerView, playerID string) game.GameCommand {
	for _, kind := range []game.CommandKind{game.CommandClaimWin, game.CommandWin, game.CommandPong, game.CommandChow, game.CommandRiichi, game.CommandKong} {
		if action, ok := firstLegalAction(actions, kind); ok {
			return commandFromLegalAction(playerID, action)
		}
	}
	if action, ok := bestLegalDiscard(actions, player.Hand); ok {
		return commandFromLegalAction(playerID, action)
	}
	if action, ok := firstLegalAction(actions, game.CommandPass); ok {
		return commandFromLegalAction(playerID, action)
	}
	if action, ok := firstLegalAction(actions, game.CommandQuit); ok {
		return commandFromLegalAction(playerID, action)
	}
	return game.GameCommand{PlayerID: playerID, Kind: game.CommandQuit}
}

func firstLegalAction(actions []game.LegalAction, kind game.CommandKind) (game.LegalAction, bool) {
	for _, action := range actions {
		if action.Kind == kind {
			return action, true
		}
	}
	return game.LegalAction{}, false
}

func bestLegalDiscard(actions []game.LegalAction, hand []game.Tile) (game.LegalAction, bool) {
	bestIndex := game.ChooseAIDiscard(hand)
	var fallback game.LegalAction
	found := false
	for _, action := range actions {
		if action.Kind != game.CommandDiscard {
			continue
		}
		if !found {
			fallback = action
			found = true
		}
		if action.TileIndex == bestIndex {
			return action, true
		}
	}
	return fallback, found
}

func commandFromLegalAction(playerID string, action game.LegalAction) game.GameCommand {
	return game.GameCommand{
		PlayerID:  playerID,
		Kind:      action.Kind,
		TileIndex: action.TileIndex,
		Tile:      action.Tile,
	}
}

func decideClaim(snapshot game.GameSnapshot, player game.PlayerView, playerID string) game.GameCommand {
	pass := game.GameCommand{PlayerID: playerID, Kind: game.CommandPass}
	pending := snapshot.PendingClaim
	if pending == nil || pending.Active < 0 || pending.Active >= len(pending.Options) {
		return pass
	}
	first := pending.Options[pending.Active]
	if first.Player < 0 || first.Player >= len(snapshot.Players) || snapshot.Players[first.Player].ID != playerID {
		return pass
	}
	options := []game.ClaimOption{first}
	for i := pending.Active + 1; i < len(pending.Options); i++ {
		option := pending.Options[i]
		if option.Player != first.Player || option.Kind != first.Kind {
			break
		}
		options = append(options, option)
	}
	switch first.Kind {
	case game.ClaimWin:
		return game.GameCommand{PlayerID: playerID, Kind: game.CommandClaimWin}
	case game.ClaimPong:
		return game.GameCommand{PlayerID: playerID, Kind: game.CommandPong}
	case game.ClaimChow:
		return game.GameCommand{PlayerID: playerID, Kind: game.CommandChow, TileIndex: bestChowOption(player.Hand, options)}
	default:
		return pass
	}
}

func bestChowOption(hand []game.Tile, options []game.ClaimOption) int {
	bestIndex := 0
	bestShanten := 99
	for i, option := range options {
		remaining := append([]game.Tile(nil), hand...)
		for _, tile := range option.Consumed {
			remaining = removeTile(remaining, tile)
		}
		if shanten := game.ShantenStandard(remaining); shanten < bestShanten {
			bestIndex = i
			bestShanten = shanten
		}
	}
	return bestIndex
}

func removeTile(tiles []game.Tile, target game.Tile) []game.Tile {
	for i, tile := range tiles {
		if tile == target {
			return append(tiles[:i], tiles[i+1:]...)
		}
	}
	return tiles
}

func playerByID(snapshot game.GameSnapshot, playerID string) (game.PlayerView, bool) {
	for _, player := range snapshot.Players {
		if player.ID == playerID {
			return player, true
		}
	}
	return game.PlayerView{}, false
}

func concealedKongTile(hand []game.Tile) (game.Tile, bool) {
	counts := game.TileCounts(hand)
	for tile, count := range counts {
		if count >= 4 {
			return game.Tile(tile), true
		}
	}
	return 0, false
}
