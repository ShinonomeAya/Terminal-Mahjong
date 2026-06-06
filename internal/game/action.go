package game

import (
	"fmt"
	"strconv"
	"strings"
)

type ActionKind int

const (
	ActionUnknown ActionKind = iota
	ActionDiscard
	ActionWin
	ActionKong
	ActionChow
	ActionQuit
)

type Action struct {
	Kind  ActionKind
	Index int
	Tiles []Tile
}

func ParseAction(line string) (Action, error) {
	text := strings.TrimSpace(strings.ToLower(line))
	if text == "q" {
		return Action{Kind: ActionQuit}, nil
	}
	if text == "h" {
		return Action{Kind: ActionWin}, nil
	}
	if strings.HasPrefix(text, "k ") {
		tile, ok := ParseTile(text[2:])
		if !ok {
			return Action{}, fmt.Errorf("unknown kong tile")
		}
		return Action{Kind: ActionKong, Tiles: []Tile{tile}}, nil
	}
	if strings.HasPrefix(text, "c ") {
		parts := strings.Fields(text)
		if len(parts) != 3 {
			return Action{}, fmt.Errorf("chow needs two hand tiles")
		}
		left, okLeft := ParseTile(parts[1])
		right, okRight := ParseTile(parts[2])
		if !okLeft || !okRight {
			return Action{}, fmt.Errorf("unknown chow tile")
		}
		return Action{Kind: ActionChow, Tiles: []Tile{left, right}}, nil
	}
	indexText := strings.TrimPrefix(text, "d ")
	index, err := strconv.Atoi(strings.TrimSpace(indexText))
	if err != nil {
		return Action{}, fmt.Errorf("unknown action")
	}
	return Action{Kind: ActionDiscard, Index: index - 1}, nil
}
