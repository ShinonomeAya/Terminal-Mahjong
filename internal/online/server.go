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
	replaystore "mahjong/internal/replay"
)

type Server struct {
	mu       sync.Mutex
	rooms    map[string]*room
	sessions map[string]*session
	bots     bot.BotEngine
	upgrader websocket.Upgrader
	nextRoom int
	options  ServerOptions
}

type room struct {
	code      string
	match     *game.Match
	seats     [4]string
	ready     map[string]bool
	started   bool
	updatedAt time.Time
	replay    *game.ReplayFile
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
	return NewServerWithOptions(ServerOptions{})
}

func NewServerWithOptions(options ServerOptions) *Server {
	options = options.withDefaults()
	return &Server{
		rooms:    make(map[string]*room),
		sessions: make(map[string]*session),
		bots:     bot.NewHeuristicBot(),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		options: options,
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
		session, created, err := s.createRoom(conn, message.Name, message.Mode, message.RuleConfig)
		if err != nil {
			writeError(conn, err.Error())
			return current
		}
		writeJSON(conn, stateMessageForSession(created, session, protocol.Message{
			Type:           protocol.MsgRoomCreated,
			PlayerID:       session.playerID,
			ReconnectToken: session.reconnectToken,
			RoomCode:       created.code,
			Seat:           session.seat,
			ReadySeats:     readySeats(created),
			Started:        created.started,
			OccupiedSeats:  occupiedSeats(created),
		}))
		return session
	case protocol.MsgJoinRoom:
		session, joined, err := s.joinRoom(conn, message.RoomCode, message.Name, message.Mode, message.RuleConfig)
		if err != nil {
			writeError(conn, err.Error())
			return current
		}
		writeJSON(conn, stateMessageForSession(joined, session, protocol.Message{
			Type:           protocol.MsgRoomJoined,
			PlayerID:       session.playerID,
			ReconnectToken: session.reconnectToken,
			RoomCode:       joined.code,
			Seat:           session.seat,
			ReadySeats:     readySeats(joined),
			Started:        joined.started,
			OccupiedSeats:  occupiedSeats(joined),
		}))
		s.broadcastRoomStateExcept(joined, session.playerID)
		return session
	case protocol.MsgListRooms:
		writeJSON(conn, protocol.Message{
			Type:  protocol.MsgRoomList,
			Rooms: s.roomSummaries(),
		})
	case protocol.MsgReady:
		s.setReady(current)
	case protocol.MsgPlayCommand:
		s.playCommand(conn, current, message.Command)
	case protocol.MsgReconnect:
		session, room, err := s.reconnect(conn, message.PlayerID, message.ReconnectToken)
		if err != nil {
			writeError(conn, err.Error())
			return current
		}
		writeJSON(conn, stateMessageForSession(room, session, protocol.Message{
			Type:           protocol.MsgReconnected,
			PlayerID:       session.playerID,
			ReconnectToken: session.reconnectToken,
			RoomCode:       session.roomCode,
			Seat:           session.seat,
			ReadySeats:     readySeats(room),
			Started:        room.started,
			OccupiedSeats:  occupiedSeats(room),
			ReplayID:       completedReplayID(room),
		}))
		return session
	case protocol.MsgRequestReplay:
		s.sendCompletedReplay(conn, current)
	default:
		writeError(conn, "unknown message")
	}
	return current
}

func (s *Server) roomSummaries() []protocol.RoomSummary {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.pruneExpiredRoomsLocked(time.Now())
	rooms := make([]protocol.RoomSummary, 0, len(s.rooms))
	for _, room := range s.rooms {
		if room.started {
			continue
		}
		rooms = append(rooms, protocol.RoomSummary{
			Code:       room.code,
			Occupied:   len(occupiedSeats(room)),
			Ready:      len(readySeats(room)),
			Started:    room.started,
			Wall:       room.match.Round.Snapshot().WallCount,
			Mode:       room.match.Mode,
			RuleConfig: room.match.RuleConfig,
		})
	}
	return rooms
}

func (s *Server) createRoom(conn *websocket.Conn, name string, mode game.RuleMode, config game.RuleConfig) (*session, *room, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if mode == "" {
		mode = game.ModeCompatibility
	}
	match, err := game.NewMatch(0, roomRuleSet(mode, config))
	if err != nil {
		return nil, nil, err
	}
	s.nextRoom++
	code := fmt.Sprintf("%06d", s.nextRoom)
	created := &room{
		code:      code,
		match:     match,
		ready:     make(map[string]bool),
		updatedAt: time.Now(),
	}
	session := newSession(name, code, 0, conn)
	created.seats[0] = session.playerID
	s.rooms[code] = created
	s.sessions[session.playerID] = session
	return session, created, nil
}

func roomRuleSet(mode game.RuleMode, config game.RuleConfig) game.RuleSet {
	if mode == game.ModeMCR {
		return game.NewMCRRuleSet(config.MCR)
	}
	if mode == game.ModeRiichi {
		return game.NewRiichiRuleSet(config.Riichi)
	}
	return game.NewCompatibilityRuleSet(mode, config)
}

func (s *Server) joinRoom(conn *websocket.Conn, code string, name string, mode game.RuleMode, config game.RuleConfig) (*session, *room, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	joined := s.rooms[code]
	if joined == nil {
		return nil, nil, fmt.Errorf("room not found")
	}
	if joined.started {
		return nil, nil, fmt.Errorf("room already started")
	}
	if mode != "" && mode != joined.match.Mode {
		return nil, nil, fmt.Errorf("rule mode does not match room")
	}
	if mode != "" && config != joined.match.RuleConfig {
		return nil, nil, fmt.Errorf("rule configuration does not match room")
	}
	seat := firstOpenSeat(joined)
	if seat < 0 {
		return nil, nil, fmt.Errorf("room is full")
	}
	session := newSession(name, code, seat, conn)
	joined.seats[seat] = session.playerID
	s.touchRoomLocked(joined)
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
		s.touchRoomLocked(room)
		if room.started && !wasStarted {
			room.match.EnsureCurrentTurnDraw()
		}
		s.broadcastRoomStateLocked(room, "")
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
	result := room.match.ApplyCommand(command)
	if !result.OK {
		s.mu.Unlock()
		writeError(conn, result.Error)
		return
	}
	s.touchRoomLocked(room)
	s.advanceUnoccupiedBotsLocked(room)
	replayErr := s.ensureCompletedReplayLocked(room)
	s.broadcastGameSnapshotLocked(room, result)
	if replayErr != nil {
		writeError(conn, replayErr.Error())
	} else if room.replay != nil {
		s.broadcastCompletedReplayLocked(room)
	}
	s.mu.Unlock()
}

func (s *Server) advanceUnoccupiedBotsLocked(room *room) {
	const maxBotActions = 200
	for actions := 0; actions < maxBotActions && !room.match.Round.Over; actions++ {
		current := room.match.Round.Current
		if current < 0 || current >= len(room.seats) || room.seats[current] != "" {
			room.match.EnsureCurrentTurnDraw()
			return
		}
		room.match.EnsureCurrentTurnDraw()
		command := s.bots.Decide(context.Background(), room.match.Round.Snapshot(), fmt.Sprintf("%d", current))
		command.PlayerID = fmt.Sprintf("%d", current)
		result := room.match.ApplyCommand(command)
		if !result.OK {
			fallback := game.GameCommand{PlayerID: fmt.Sprintf("%d", current), Kind: game.CommandDiscard, TileIndex: 0}
			if fallbackResult := room.match.ApplyCommand(fallback); !fallbackResult.OK {
				return
			}
		}
	}
}

func (s *Server) reconnect(conn *websocket.Conn, playerID string, token string) (*session, *room, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session := s.sessions[playerID]
	if session == nil || session.reconnectToken != token {
		return nil, nil, fmt.Errorf("invalid reconnect token")
	}
	if !session.offlineAt.IsZero() && time.Since(session.offlineAt) > s.options.ReconnectWindow {
		return nil, nil, fmt.Errorf("reconnect window expired")
	}
	session.conn = conn
	session.offlineAt = time.Time{}
	room := s.rooms[session.roomCode]
	if room == nil {
		return nil, nil, fmt.Errorf("room not found")
	}
	return session, room, nil
}

func (s *Server) touchRoomLocked(room *room) {
	if room != nil {
		room.updatedAt = time.Now()
	}
}

func (s *Server) pruneExpiredRoomsLocked(now time.Time) {
	for code, room := range s.rooms {
		if room.started {
			continue
		}
		if now.Sub(room.updatedAt) <= s.options.RoomIdleTTL {
			continue
		}
		for _, playerID := range room.seats {
			if playerID != "" {
				delete(s.sessions, playerID)
			}
		}
		delete(s.rooms, code)
	}
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

func (s *Server) broadcastRoomStateExcept(room *room, excludedPlayerID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.broadcastRoomStateLocked(room, excludedPlayerID)
}

func (s *Server) broadcastRoomStateLocked(room *room, excludedPlayerID string) {
	for _, playerID := range room.seats {
		if playerID == "" || playerID == excludedPlayerID {
			continue
		}
		session := s.sessions[playerID]
		if session != nil && session.conn != nil {
			writeJSON(session.conn, roomStateMessage(room, session))
		}
	}
}

func (s *Server) broadcastGameSnapshotLocked(room *room, result game.CommandResult) {
	for _, playerID := range room.seats {
		if playerID == "" {
			continue
		}
		session := s.sessions[playerID]
		if session == nil || session.conn == nil {
			continue
		}
		message := stateMessageForSession(room, session, protocol.Message{Type: protocol.MsgGameSnapshot})
		result.Snapshot = message.Snapshot
		message.Result = result
		writeJSON(session.conn, message)
	}
}

func (s *Server) ensureCompletedReplayLocked(room *room) error {
	if room == nil || room.replay != nil || !room.match.Complete {
		return nil
	}
	file, err := room.match.CompletedReplay(
		replaystore.ApplicationVersion(),
		time.Now().UTC(),
		s.replayParticipantsLocked(room),
	)
	if err != nil {
		return err
	}
	room.replay = &file
	return nil
}

func (s *Server) replayParticipantsLocked(room *room) []game.ReplayParticipant {
	participants := make([]game.ReplayParticipant, 0, len(room.seats))
	for seat, playerID := range room.seats {
		id := fmt.Sprintf("bot-%d", seat)
		name := room.match.Round.Players[seat].Name
		if playerID != "" {
			id = playerID
			if session := s.sessions[playerID]; session != nil {
				name = session.name
			}
		}
		participants = append(participants, game.ReplayParticipant{
			Seat: seat,
			ID:   id,
			Name: name,
		})
	}
	return participants
}

func (s *Server) broadcastCompletedReplayLocked(room *room) {
	for _, playerID := range room.seats {
		if playerID == "" {
			continue
		}
		session := s.sessions[playerID]
		if session == nil || session.conn == nil {
			continue
		}
		writeJSON(session.conn, protocol.Message{
			Type:     protocol.MsgReplayReady,
			ReplayID: room.replay.ReplayID,
		})
		writeJSON(session.conn, protocol.Message{
			Type:     protocol.MsgReplayData,
			ReplayID: room.replay.ReplayID,
			Replay:   room.replay,
		})
	}
}

func (s *Server) sendCompletedReplay(conn *websocket.Conn, session *session) {
	if session == nil {
		writeError(conn, "not joined")
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	room := s.rooms[session.roomCode]
	if room == nil || room.replay == nil {
		writeError(conn, "replay is not available")
		return
	}
	writeJSON(conn, protocol.Message{
		Type:     protocol.MsgReplayData,
		ReplayID: room.replay.ReplayID,
		Replay:   room.replay,
	})
}

func completedReplayID(room *room) string {
	if room == nil || room.replay == nil {
		return ""
	}
	return room.replay.ReplayID
}

func firstOpenSeat(room *room) int {
	for i, playerID := range room.seats {
		if playerID == "" {
			return i
		}
	}
	return -1
}

func roomStateMessage(room *room, session *session) protocol.Message {
	return stateMessageForSession(room, session, protocol.Message{
		Type:          protocol.MsgRoomState,
		RoomCode:      room.code,
		ReadySeats:    readySeats(room),
		Started:       room.started,
		OccupiedSeats: occupiedSeats(room),
	})
}

func stateMessageForSession(room *room, session *session, message protocol.Message) protocol.Message {
	id := fmt.Sprintf("%d", session.seat)
	message.Mode = room.match.Mode
	message.RuleConfig = room.match.RuleConfig
	message.Snapshot = room.match.Round.SnapshotFor(id)
	message.Match = room.match.SnapshotFor(id)
	return message
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
