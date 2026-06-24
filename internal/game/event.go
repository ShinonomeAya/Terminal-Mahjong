package game

import (
	"fmt"
	"strings"
)

type EventKind string

const (
	EventDraw            EventKind = "draw"
	EventFlower          EventKind = "flower"
	EventReplacementDraw EventKind = "replacement-draw"
	EventDiscard         EventKind = "discard"
	EventChow            EventKind = "chow"
	EventPong            EventKind = "pong"
	EventKong            EventKind = "kong"
	EventWin             EventKind = "win"
	EventQuit            EventKind = "quit"
	EventWallExhausted   EventKind = "wall-exhausted"
)

type GameEvent struct {
	Turn   int
	Kind   EventKind
	Player int
	Tile   Tile
	Note   string
}

func (g *Game) RecordEvent(kind EventKind, player int, tile Tile, note string) {
	g.Events = append(g.Events, GameEvent{
		Turn:   len(g.Events) + 1,
		Kind:   kind,
		Player: player,
		Tile:   tile,
		Note:   note,
	})
}

func EventSummary(events []GameEvent) string {
	lines := make([]string, 0, len(events))
	for _, event := range events {
		lines = append(lines, event.String())
	}
	return strings.Join(lines, "\n")
}

func RecentEvents(events []GameEvent, limit int) []GameEvent {
	if limit <= 0 {
		return nil
	}
	if len(events) <= limit {
		return append([]GameEvent(nil), events...)
	}
	return append([]GameEvent(nil), events[len(events)-limit:]...)
}

func (e GameEvent) String() string {
	player := playerEventName(e.Player)
	tileText := ""
	if e.Tile >= 0 && e.Tile <= FlowerWinter {
		tileText = " " + e.Tile.String()
	}
	note := ""
	if e.Note != "" {
		note = " - " + e.Note
	}
	return fmt.Sprintf("%02d. %s %s%s%s", e.Turn, player, e.Kind, tileText, note)
}

func playerEventName(player int) string {
	switch player {
	case 0:
		return "You"
	case 1:
		return "AI-1"
	case 2:
		return "AI-2"
	case 3:
		return "AI-3"
	default:
		return fmt.Sprintf("Player-%d", player)
	}
}
