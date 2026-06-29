package tui

import (
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"mahjong/internal/game"
	"mahjong/internal/replay"
)

type replaySavedMsg struct {
	Path     string
	ReplayID string
}

type replaySaveErrorMsg struct {
	ReplayID string
	Err      error
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

func saveReplayFileCmd(file game.ReplayFile, dir string) tea.Cmd {
	return func() tea.Msg {
		path, err := replay.Save(dir, file)
		if err != nil {
			return replaySaveErrorMsg{ReplayID: file.ReplayID, Err: err}
		}
		return replaySavedMsg{Path: path, ReplayID: file.ReplayID}
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
