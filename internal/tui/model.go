package tui

import (
	"mahjong/internal/game"
	"mahjong/internal/online"

	tea "github.com/charmbracelet/bubbletea"
)

type Screen int

const (
	ScreenMenu Screen = iota
	ScreenJoinOnline
	ScreenHelp
	ScreenTable
	ScreenGameOver
)

type Model struct {
	Screen              Screen
	MenuIndex           int
	GameOverIndex       int
	SelectedIndex       int
	UnicodeTiles        bool
	Game                *game.Game
	Online              bool
	OnlineClient        *online.Client
	OnlineSnapshot      game.GameSnapshot
	OnlinePlayerID      string
	OnlineRoomCode      string
	OnlineSeat          int
	OnlineReadySeats    []int
	OnlineOccupiedSeats []int
	OnlineStarted       bool
	OnlineEvents        chan tea.Msg
	OnlineServerURL     string
	OnlineName          string
	OnlineSession       string
	JoinRoomCode        string
	HandHitBoxes        []TileHitBox
	StatusLine          string
	NetworkStatus       NetworkStatus
	ReconnectAttempt    int
	ReconnectMax        int
	Width               int
	Height              int
}

func NewModel() Model {
	return Model{
		Screen:          ScreenMenu,
		UnicodeTiles:    true,
		OnlineServerURL: "ws://127.0.0.1:8080/ws",
		OnlineName:      "Player",
		OnlineSession:   ".mahjong-session.json",
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
		case ScreenHelp:
			return updateHelp(m, msg)
		case ScreenTable:
			return updateTable(m, msg)
		case ScreenGameOver:
			return updateGameOver(m, msg)
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
		return updated, tea.Batch(waitOnlineSnapshot(msg.Client, updated.OnlineEvents), listenOnlineEvents(updated.OnlineEvents))
	case onlineSnapshotMsg:
		updated := applyOnlineSnapshot(m, msg.Message)
		return updated, waitOnlineSnapshot(updated.OnlineClient, updated.OnlineEvents)
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
	case ScreenHelp:
		return renderHelp()
	case ScreenTable:
		return renderTable(m)
	case ScreenGameOver:
		return renderGameOver(m)
	default:
		return ""
	}
}
