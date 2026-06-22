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
		return decideClaim(snapshot, player, playerID)
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
