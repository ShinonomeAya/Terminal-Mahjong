package online

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"mahjong/internal/bot"
	"mahjong/internal/game"
	"mahjong/internal/protocol"
)

const reconnectWindow = 2 * time.Minute

type Server struct {
	mu       sync.Mutex
	rooms    map[string]*room
	sessions map[string]*session
	bots     bot.BotEngine
	upgrader websocket.Upgrader
	nextRoom int
}

type room struct {
	code    string
	game    *game.Game
	seats   [4]string
	ready   map[string]bool
	started bool
}

type session struct {
	playerID       string
	reconnectToken string
	name           string
	roomCode       string
	seat           int
	conn           *websocket.Conn
	offlineAt      time.Time
}

func NewServer() *Server {
	return &Server{
		rooms:    make(map[string]*room),
		sessions: make(map[string]*session),
		bots:     bot.NewHeuristicBot(),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/ws" {
		http.NotFound(w, r)
		return
	}
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	go s.handleConn(conn)
}

func (s *Server) handleConn(conn *websocket.Conn) {
	var current *session
	defer func() {
		s.markOffline(current)
		_ = conn.Close()
	}()

	for {
		var message protocol.Message
		if err := conn.ReadJSON(&message); err != nil {
			return
		}
		next := s.handleMessage(conn, current, message)
		if next != nil {
			current = next
		}
	}
}

func (s *Server) handleMessage(conn *websocket.Conn, current *session, message protocol.Message) *session {
	switch message.Type {
	case protocol.MsgCreateRoom:
		session, created := s.createRoom(conn, message.Name)
		writeJSON(conn, protocol.Message{
			Type:           protocol.MsgRoomCreated,
			PlayerID:       session.playerID,
			ReconnectToken: session.reconnectToken,
			RoomCode:       created.code,
			Seat:           session.seat,
			ReadySeats:     readySeats(created),
			Started:        created.started,
			OccupiedSeats:  occupiedSeats(created),
			Snapshot:       created.game.Snapshot(),
		})
		return session
	case protocol.MsgJoinRoom:
		session, joined, err := s.joinRoom(conn, message.RoomCode, message.Name)
		if err != nil {
			writeError(conn, err.Error())
			return current
		}
		writeJSON(conn, protocol.Message{
			Type:           protocol.MsgRoomJoined,
			PlayerID:       session.playerID,
			ReconnectToken: session.reconnectToken,
			RoomCode:       joined.code,
			Seat:           session.seat,
			ReadySeats:     readySeats(joined),
			Started:        joined.started,
			OccupiedSeats:  occupiedSeats(joined),
			Snapshot:       joined.game.Snapshot(),
		})
		return session
	case protocol.MsgReady:
		s.setReady(current)
	case protocol.MsgPlayCommand:
		s.playCommand(conn, current, message.Command)
	case protocol.MsgReconnect:
		session, room, snapshot, err := s.reconnect(conn, message.PlayerID, message.ReconnectToken)
		if err != nil {
			writeError(conn, err.Error())
			return current
		}
		writeJSON(conn, protocol.Message{
			Type:           protocol.MsgReconnected,
			PlayerID:       session.playerID,
			ReconnectToken: session.reconnectToken,
			RoomCode:       session.roomCode,
			Seat:           session.seat,
			ReadySeats:     readySeats(room),
			Started:        room.started,
			OccupiedSeats:  occupiedSeats(room),
			Snapshot:       snapshot,
		})
		return session
	default:
		writeError(conn, "unknown message")
	}
	return current
}

func (s *Server) createRoom(conn *websocket.Conn, name string) (*session, *room) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextRoom++
	code := fmt.Sprintf("%06d", s.nextRoom)
	created := &room{
		code:  code,
		game:  game.NewGame(0),
		ready: make(map[string]bool),
	}
	session := newSession(name, code, 0, conn)
	created.seats[0] = session.playerID
	s.rooms[code] = created
	s.sessions[session.playerID] = session
	return session, created
}

func (s *Server) joinRoom(conn *websocket.Conn, code string, name string) (*session, *room, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	joined := s.rooms[code]
	if joined == nil {
		return nil, nil, fmt.Errorf("room not found")
	}
	seat := firstOpenSeat(joined)
	if seat < 0 {
		return nil, nil, fmt.Errorf("room is full")
	}
	session := newSession(name, code, seat, conn)
	joined.seats[seat] = session.playerID
	s.sessions[session.playerID] = session
	return session, joined, nil
}

func (s *Server) setReady(session *session) {
	if session == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if room := s.rooms[session.roomCode]; room != nil {
		room.ready[session.playerID] = true
		wasStarted := room.started
		room.started = allOccupiedReady(room)
		if room.started && !wasStarted {
			room.game.EnsureCurrentTurnDraw()
		}
		s.broadcastRoomLocked(room, roomStateMessage(room))
	}
}

func (s *Server) playCommand(conn *websocket.Conn, session *session, command game.GameCommand) {
	if session == nil {
		writeError(conn, "not joined")
		return
	}
	s.mu.Lock()
	room := s.rooms[session.roomCode]
	if room == nil {
		s.mu.Unlock()
		writeError(conn, "room not found")
		return
	}
	if !room.started {
		s.mu.Unlock()
		writeError(conn, "room has not started")
		return
	}
	command.PlayerID = fmt.Sprintf("%d", session.seat)
	result := room.game.ApplyCommand(command)
	if !result.OK {
		s.mu.Unlock()
		writeError(conn, result.Error)
		return
	}
	s.advanceUnoccupiedBotsLocked(room)
	result.Snapshot = room.game.Snapshot()
	s.broadcastRoomLocked(room, protocol.Message{Type: protocol.MsgGameSnapshot, Result: result, Snapshot: result.Snapshot})
	s.mu.Unlock()
}

func (s *Server) advanceUnoccupiedBotsLocked(room *room) {
	const maxBotActions = 200
	for actions := 0; actions < maxBotActions && !room.game.Over; actions++ {
		current := room.game.Current
		if current < 0 || current >= len(room.seats) || room.seats[current] != "" {
			room.game.EnsureCurrentTurnDraw()
			return
		}
		room.game.EnsureCurrentTurnDraw()
		command := s.bots.Decide(context.Background(), room.game.Snapshot(), fmt.Sprintf("%d", current))
		command.PlayerID = fmt.Sprintf("%d", current)
		result := room.game.ApplyCommand(command)
		if !result.OK {
			fallback := game.GameCommand{PlayerID: fmt.Sprintf("%d", current), Kind: game.CommandDiscard, TileIndex: 0}
			if fallbackResult := room.game.ApplyCommand(fallback); !fallbackResult.OK {
				return
			}
		}
	}
}

func (s *Server) reconnect(conn *websocket.Conn, playerID string, token string) (*session, *room, game.GameSnapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session := s.sessions[playerID]
	if session == nil || session.reconnectToken != token {
		return nil, nil, game.GameSnapshot{}, fmt.Errorf("invalid reconnect token")
	}
	if !session.offlineAt.IsZero() && time.Since(session.offlineAt) > reconnectWindow {
		return nil, nil, game.GameSnapshot{}, fmt.Errorf("reconnect window expired")
	}
	session.conn = conn
	session.offlineAt = time.Time{}
	room := s.rooms[session.roomCode]
	if room == nil {
		return nil, nil, game.GameSnapshot{}, fmt.Errorf("room not found")
	}
	return session, room, room.game.Snapshot(), nil
}

func (s *Server) markOffline(session *session) {
	if session == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if stored := s.sessions[session.playerID]; stored != nil && stored.conn == session.conn {
		stored.conn = nil
		stored.offlineAt = time.Now()
	}
}

func (s *Server) broadcastRoomLocked(room *room, message protocol.Message) {
	for _, playerID := range room.seats {
		if playerID == "" {
			continue
		}
		session := s.sessions[playerID]
		if session != nil && session.conn != nil {
			writeJSON(session.conn, message)
		}
	}
}

func firstOpenSeat(room *room) int {
	for i, playerID := range room.seats {
		if playerID == "" {
			return i
		}
	}
	return -1
}

func roomStateMessage(room *room) protocol.Message {
	return protocol.Message{
		Type:          protocol.MsgRoomState,
		RoomCode:      room.code,
		ReadySeats:    readySeats(room),
		Started:       room.started,
		OccupiedSeats: occupiedSeats(room),
		Snapshot:      room.game.Snapshot(),
	}
}

func readySeats(room *room) []int {
	seats := make([]int, 0, len(room.ready))
	for seat, playerID := range room.seats {
		if playerID != "" && room.ready[playerID] {
			seats = append(seats, seat)
		}
	}
	return seats
}

func occupiedSeats(room *room) []int {
	seats := make([]int, 0, len(room.seats))
	for seat, playerID := range room.seats {
		if playerID != "" {
			seats = append(seats, seat)
		}
	}
	return seats
}

func allOccupiedReady(room *room) bool {
	occupied := 0
	for _, playerID := range room.seats {
		if playerID == "" {
			continue
		}
		occupied++
		if !room.ready[playerID] {
			return false
		}
	}
	return occupied > 0
}

func newSession(name string, roomCode string, seat int, conn *websocket.Conn) *session {
	if name == "" {
		name = fmt.Sprintf("Player-%d", seat+1)
	}
	return &session{
		playerID:       randomToken(8),
		reconnectToken: randomToken(32),
		name:           name,
		roomCode:       roomCode,
		seat:           seat,
		conn:           conn,
	}
}

func randomToken(bytesLen int) string {
	bytes := make([]byte, bytesLen)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}

func writeJSON(conn *websocket.Conn, message protocol.Message) {
	_ = conn.WriteJSON(message)
}

func writeError(conn *websocket.Conn, message string) {
	writeJSON(conn, protocol.Message{Type: protocol.MsgError, Error: message})
}
