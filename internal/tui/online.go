package tui

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"mahjong/internal/game"
	"mahjong/internal/online"
	"mahjong/internal/protocol"
)

type onlineConnectedMsg struct {
	Message protocol.Message
	Client  *online.Client
}

type onlineSnapshotMsg struct {
	Message protocol.Message
}

type onlineRoomsMsg struct {
	Rooms []protocol.RoomSummary
}

type onlineErrorMsg struct {
	Err error
}

type onlineCommandSentMsg struct{}

type onlineReconnectAttemptMsg struct {
	Attempt int
	Max     int
}

type onlineReconnectSuccessMsg struct{}

func listenOnlineEvents(events <-chan tea.Msg) tea.Cmd {
	if events == nil {
		return nil
	}
	return func() tea.Msg {
		msg, ok := <-events
		if !ok {
			return nil
		}
		return msg
	}
}

func sendOnlineEvent(events chan<- tea.Msg, msg tea.Msg) {
	if events == nil {
		return
	}
	select {
	case events <- msg:
	default:
	}
}

func createOnlineRoomCmd(m Model) tea.Cmd {
	serverURL := m.OnlineServerURL
	name := m.OnlineName
	sessionPath := m.OnlineSession
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		client := online.NewClient(serverURL, name)
		message, err := client.CreateRoom(ctx)
		if err != nil {
			client.Close()
			return onlineErrorMsg{Err: err}
		}
		if err := online.SaveClientSession(sessionPath, client.Session()); err != nil {
			client.Close()
			return onlineErrorMsg{Err: err}
		}
		return onlineConnectedMsg{Message: message, Client: client}
	}
}

func joinOnlineRoomCmd(m Model) tea.Cmd {
	serverURL := m.OnlineServerURL
	name := m.OnlineName
	sessionPath := m.OnlineSession
	roomCode := m.JoinRoomCode
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		client := online.NewClient(serverURL, name)
		message, err := client.JoinRoom(ctx, roomCode)
		if err != nil {
			client.Close()
			return onlineErrorMsg{Err: err}
		}
		if err := online.SaveClientSession(sessionPath, client.Session()); err != nil {
			client.Close()
			return onlineErrorMsg{Err: err}
		}
		return onlineConnectedMsg{Message: message, Client: client}
	}
}

func reconnectOnlineCmd(m Model) tea.Cmd {
	sessionPath := m.OnlineSession
	return func() tea.Msg {
		session, err := online.LoadClientSession(sessionPath)
		if err != nil {
			return onlineErrorMsg{Err: err}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		client := online.NewClient(session.ServerURL, session.Name)
		message, err := client.Reconnect(ctx, session)
		if err != nil {
			client.Close()
			return onlineErrorMsg{Err: err}
		}
		return onlineConnectedMsg{Message: message, Client: client}
	}
}

func listOnlineRoomsCmd(m Model) tea.Cmd {
	serverURL := m.OnlineServerURL
	name := m.OnlineName
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		client := online.NewClient(serverURL, name)
		defer client.Close()
		rooms, err := client.ListRooms(ctx)
		if err != nil {
			return onlineErrorMsg{Err: err}
		}
		return onlineRoomsMsg{Rooms: rooms}
	}
}

func waitOnlineSnapshot(client *online.Client, events chan<- tea.Msg) tea.Cmd {
	if client == nil {
		return nil
	}
	return func() tea.Msg {
		message, err := client.ReadUntilWithReconnect(
			context.Background(),
			24*time.Hour,
			online.ReconnectPolicy{
				MaxAttempts: 5,
				BaseDelay:   200 * time.Millisecond,
				OnAttempt: func(attempt int, max int) {
					sendOnlineEvent(events, onlineReconnectAttemptMsg{Attempt: attempt, Max: max})
				},
				OnSuccess: func() {
					sendOnlineEvent(events, onlineReconnectSuccessMsg{})
				},
			},
			protocol.MsgGameSnapshot,
			protocol.MsgRoomState,
			protocol.MsgReconnected,
			protocol.MsgError,
		)
		if err != nil {
			return onlineErrorMsg{Err: err}
		}
		return onlineSnapshotMsg{Message: message}
	}
}

func sendOnlineReadyCmd(client *online.Client) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := client.SendReady(ctx); err != nil {
			return onlineErrorMsg{Err: err}
		}
		return onlineCommandSentMsg{}
	}
}

func sendOnlineDiscardCmd(client *online.Client, tileIndex int) tea.Cmd {
	return sendOnlineGameCommandCmd(client, game.GameCommand{Kind: game.CommandDiscard, TileIndex: tileIndex})
}

func sendOnlineWinCmd(client *online.Client) tea.Cmd {
	return sendOnlineGameCommandCmd(client, game.GameCommand{Kind: game.CommandWin})
}

func sendOnlineKongCmd(client *online.Client, tile string) tea.Cmd {
	return sendOnlineGameCommandCmd(client, game.GameCommand{Kind: game.CommandKong, Tile: tile})
}

func sendOnlineGameCommandCmd(client *online.Client, command game.GameCommand) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := client.SendCommand(ctx, command); err != nil {
			return onlineErrorMsg{Err: err}
		}
		return onlineCommandSentMsg{}
	}
}

func applyOnlineConnected(m Model, msg onlineConnectedMsg) Model {
	m.Screen = ScreenTable
	m.Online = true
	m.OnlineClient = msg.Client
	m.OnlinePlayerID = msg.Message.PlayerID
	m.OnlineRoomCode = msg.Message.RoomCode
	m.OnlineSeat = msg.Message.Seat
	m.OnlineReadySeats = append([]int(nil), msg.Message.ReadySeats...)
	m.OnlineOccupiedSeats = append([]int(nil), msg.Message.OccupiedSeats...)
	m.OnlineStarted = msg.Message.Started
	if m.OnlineEvents == nil {
		m.OnlineEvents = make(chan tea.Msg, 8)
	}
	m.OnlineSnapshot = msg.Message.Snapshot
	m.Game = nil
	m.StatusLine = fmt.Sprintf("Room:%s Seat:%d", m.OnlineRoomCode, m.OnlineSeat+1)
	m.NetworkStatus = networkStatusForOnlineSnapshot(m)
	return m
}

func applyOnlineSnapshot(m Model, message protocol.Message) Model {
	if message.Type == protocol.MsgError {
		m.StatusLine = message.Error
		return m
	}
	m.Online = true
	if message.PlayerID != "" {
		m.OnlinePlayerID = message.PlayerID
	}
	if message.RoomCode != "" {
		m.OnlineRoomCode = message.RoomCode
	}
	if message.ReadySeats != nil {
		m.OnlineReadySeats = append([]int(nil), message.ReadySeats...)
	}
	if message.OccupiedSeats != nil {
		m.OnlineOccupiedSeats = append([]int(nil), message.OccupiedSeats...)
	}
	if message.Started {
		m.OnlineStarted = true
	}
	if len(message.Snapshot.Players) > 0 {
		m.OnlineSnapshot = message.Snapshot
	}
	m.NetworkStatus = networkStatusForOnlineSnapshot(m)
	if m.OnlineSnapshot.Over {
		m.Screen = ScreenGameOver
		m.StatusLine = "Online round ended"
	}
	return m
}

func networkStatusForOnlineSnapshot(m Model) NetworkStatus {
	if len(m.OnlineSnapshot.Players) == 0 {
		return NetworkWaiting
	}
	if !m.OnlineStarted {
		return NetworkWaiting
	}
	if m.OnlineSnapshot.Current == m.OnlineSeat {
		return NetworkYourTurn
	}
	return NetworkWaiting
}
