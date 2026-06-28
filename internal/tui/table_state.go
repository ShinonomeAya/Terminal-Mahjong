package tui

import "mahjong/internal/game"

type tableViewState struct {
	Snapshot   game.GameSnapshot
	Match      game.MatchSnapshot
	ViewerSeat int
	Mode       game.RuleMode
	Online     bool
	Started    bool
	RoomCode   string
}

func tableStateFor(m Model) tableViewState {
	if m.Online {
		mode := m.OnlineMatch.Mode
		if mode == "" {
			mode = m.SelectedMode
		}
		return tableViewState{
			Snapshot:   m.OnlineSnapshot,
			Match:      m.OnlineMatch,
			ViewerSeat: m.OnlineSeat,
			Mode:       mode,
			Online:     true,
			Started:    m.OnlineStarted,
			RoomCode:   m.OnlineRoomCode,
		}
	}
	if m.Game == nil {
		return tableViewState{ViewerSeat: -1}
	}
	snapshot := m.Game.SnapshotFor("0")
	return tableViewState{
		Snapshot:   snapshot,
		Match:      game.MatchSnapshot{Mode: m.Game.Mode, RuleConfig: m.Game.RuleConfig, Round: snapshot},
		ViewerSeat: 0,
		Mode:       m.Game.Mode,
		Started:    true,
	}
}
