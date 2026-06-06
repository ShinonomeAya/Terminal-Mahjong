package main

import (
	"os"

	"mahjong/internal/game"
)

func main() {
	game.NewGame(0).Play(os.Stdin, os.Stdout)
}
