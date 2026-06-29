package tui

import (
	"fmt"

	"mahjong/internal/game"
	"mahjong/internal/online"
	"mahjong/internal/protocol"
	"mahjong/internal/replay"

	tea "github.com/charmbracelet/bubbletea"
)

type Screen int

const (
	ScreenMenu Screen = iota
	ScreenJoinOnline
	ScreenOnlineRooms
	ScreenHelp
	ScreenTable
	ScreenGameOver
	ScreenReplayBrowser
	ScreenReplayViewer
)

type Model struct {
	Screen                 Screen
	MenuIndex              int
	GameOverIndex          int
	SelectedIndex          int
	ClaimOptionIndex       int
	UnicodeTiles           bool
	Language               Language
	SelectedMode           game.RuleMode
	SelectedRiichiRedFives int
	ShowTactical           bool
	Game                   *game.Game
	LocalMatch             *game.Match
	ReplayDir              string
	LastReplayPath         string
	ReplayRequestedID      string
	ReplaySavingID         string
	ReplaySavedID          string
	ReplayEntries          []replay.Entry
	ReplayIssues           []replay.FileIssue
	ReplayIndex            int
	ReplayFile             *game.ReplayFile
	ReplayFrame            int
	ReplayPlaying          bool
	ReplayShowDetails      bool
	Online                 bool
	OnlineClient           *online.Client
	OnlineSnapshot         game.GameSnapshot
	OnlineMatch            game.MatchSnapshot
	OnlinePlayerID         string
	OnlineRoomCode         string
	OnlineSeat             int
	OnlineReadySeats       []int
	OnlineOccupiedSeats    []int
	OnlineStarted          bool
	OnlineEvents           chan tea.Msg
	OnlineRooms            []protocol.RoomSummary
	RoomIndex              int
	OnlineServerURL        string
	OnlineName             string
	OnlineSession          string
	JoinRoomCode           string
	HandHitBoxes           []TileHitBox
	StatusLine             string
	NetworkStatus          NetworkStatus
	ReconnectAttempt       int
	ReconnectMax           int
	Width                  int
	Height                 int
}

func NewModel() Model {
	return Model{
		Screen:                 ScreenMenu,
		UnicodeTiles:           true,
		Language:               LanguageChinese,
		SelectedMode:           game.ModeRiichi,
		SelectedRiichiRedFives: 3,
		ReplayDir:              "replays",
		OnlineServerURL:        "ws://127.0.0.1:8080/ws",
		OnlineName:             "Player",
		OnlineSession:          ".mahjong-session.json",
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.Screen {
		case ScreenMenu:
			return updateMenu(m, msg)
		case ScreenJoinOnline:
			return updateJoinOnline(m, msg)
		case ScreenOnlineRooms:
			return updateOnlineRooms(m, msg)
		case ScreenHelp:
			return updateHelp(m, msg)
		case ScreenTable:
			return updateTable(m, msg)
		case ScreenGameOver:
			return updateGameOver(m, msg)
		case ScreenReplayBrowser:
			return updateReplayBrowser(m, msg)
		case ScreenReplayViewer:
			return updateReplayViewer(m, msg)
		default:
			return m, nil
		}
	case tea.MouseMsg:
		if m.Screen == ScreenTable {
			return updateTableMouse(m, msg)
		}
		return m, nil
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil
	case onlineConnectedMsg:
		updated := applyOnlineConnected(m, msg)
		updated, replayCmd, _ := applyOnlineReplayMessage(updated, msg.Message)
		return updated, tea.Batch(replayCmd, waitOnlineSnapshot(msg.Client, updated.OnlineEvents), listenOnlineEvents(updated.OnlineEvents))
	case onlineSnapshotMsg:
		if updated, replayCmd, handled := applyOnlineReplayMessage(m, msg.Message); handled {
			return updated, tea.Batch(replayCmd, waitOnlineSnapshot(updated.OnlineClient, updated.OnlineEvents))
		}
		updated := applyOnlineSnapshot(m, msg.Message)
		return updated, waitOnlineSnapshot(updated.OnlineClient, updated.OnlineEvents)
	case onlineRoomsMsg:
		m.Screen = ScreenOnlineRooms
		m.OnlineRooms = append([]protocol.RoomSummary(nil), msg.Rooms...)
		m.RoomIndex = 0
		m.StatusLine = fmt.Sprintf("Rooms found: %d", len(msg.Rooms))
		return m, nil
	case onlineReconnectAttemptMsg:
		m.NetworkStatus = NetworkReconnecting
		m.ReconnectAttempt = msg.Attempt
		m.ReconnectMax = msg.Max
		return m, listenOnlineEvents(m.OnlineEvents)
	case onlineReconnectSuccessMsg:
		m.NetworkStatus = NetworkReconnected
		m.StatusLine = "Reconnected"
		return m, listenOnlineEvents(m.OnlineEvents)
	case onlineErrorMsg:
		m.NetworkStatus = NetworkOffline
		m.StatusLine = msg.Err.Error()
		return m, nil
	case replaySavedMsg:
		m.LastReplayPath = msg.Path
		if msg.ReplayID != "" {
			m.ReplaySavingID = ""
			m.ReplaySavedID = msg.ReplayID
		}
		m.StatusLine = "Replay saved: " + msg.Path
		return m, nil
	case replaySaveErrorMsg:
		if msg.ReplayID != "" && m.ReplaySavingID == msg.ReplayID {
			m.ReplaySavingID = ""
		}
		m.StatusLine = "Replay save failed: " + msg.Err.Error()
		return m, nil
	case replayListMsg:
		m.ReplayEntries = append([]replay.Entry(nil), msg.Entries...)
		m.ReplayIssues = append([]replay.FileIssue(nil), msg.Issues...)
		if len(m.ReplayEntries) == 0 {
			m.ReplayIndex = 0
		} else if m.ReplayIndex >= len(m.ReplayEntries) {
			m.ReplayIndex = len(m.ReplayEntries) - 1
		}
		m.StatusLine = replayListStatus(m)
		return m, nil
	case replayLoadedMsg:
		file := msg.File
		m.ReplayFile = &file
		m.ReplayFrame = 0
		m.ReplayPlaying = false
		m.ReplayShowDetails = false
		m.Screen = ScreenReplayViewer
		m.StatusLine = ""
		return m, nil
	case replayListErrorMsg:
		m.StatusLine = replayErrorStatus(m, msg.Err)
		return m, nil
	case replayLoadErrorMsg:
		m.StatusLine = replayErrorStatus(m, msg.Err)
		return m, nil
	case replayTickMsg:
		if m.Screen != ScreenReplayViewer || !m.ReplayPlaying {
			return m, nil
		}
		return applyReplayTick(m)
	default:
		return m, nil
	}
}

func (m Model) View() string {
	switch m.Screen {
	case ScreenMenu:
		return renderMenu(m)
	case ScreenJoinOnline:
		return renderJoinOnline(m)
	case ScreenOnlineRooms:
		return renderOnlineRooms(m)
	case ScreenHelp:
		return renderHelp(m)
	case ScreenTable:
		return renderTable(m)
	case ScreenGameOver:
		return renderGameOver(m)
	case ScreenReplayBrowser:
		return renderReplayBrowser(m)
	case ScreenReplayViewer:
		return renderReplayViewer(m)
	default:
		return ""
	}
}
