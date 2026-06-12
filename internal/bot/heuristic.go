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
	if game.CanWin(player.Hand) {
		return game.GameCommand{PlayerID: playerID, Kind: game.CommandWin}
	}
	if tile, ok := concealedKongTile(player.Hand); ok {
		return game.GameCommand{PlayerID: playerID, Kind: game.CommandKong, Tile: tile.String()}
	}
	index := game.ChooseAIDiscard(player.Hand)
	return game.GameCommand{PlayerID: playerID, Kind: game.CommandDiscard, TileIndex: index}
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
