package tui

import (
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"mahjong/internal/game"
	"mahjong/internal/replay"
)

type replaySavedMsg struct {
	Path string
}

type replaySaveErrorMsg struct {
	Err error
}

func saveCompletedReplayCmd(match *game.Match, dir string) tea.Cmd {
	return func() tea.Msg {
		file, err := match.CompletedReplay(
			replay.ApplicationVersion(),
			time.Now().UTC(),
			replayParticipants(match),
		)
		if err != nil {
			return replaySaveErrorMsg{Err: err}
		}
		path, err := replay.Save(dir, file)
		if err != nil {
			return replaySaveErrorMsg{Err: err}
		}
		return replaySavedMsg{Path: path}
	}
}

func replayParticipants(match *game.Match) []game.ReplayParticipant {
	participants := make([]game.ReplayParticipant, 0, len(match.Round.Players))
	for seat, player := range match.Round.Players {
		participants = append(participants, game.ReplayParticipant{
			Seat: seat,
			ID:   strconv.Itoa(seat),
			Name: player.Name,
		})
	}
	return participants
}
