package protocol

import "mahjong/internal/game"

type MessageType string

const (
	MsgHello        MessageType = "hello"
	MsgCreateRoom   MessageType = "create_room"
	MsgJoinRoom     MessageType = "join_room"
	MsgListRooms    MessageType = "list_rooms"
	MsgReady        MessageType = "ready"
	MsgRoomList     MessageType = "room_list"
	MsgRoomState    MessageType = "room_state"
	MsgGameSnapshot MessageType = "game_snapshot"
	MsgPlayCommand  MessageType = "play_command"
	MsgReconnect    MessageType = "reconnect"
	MsgReconnected  MessageType = "reconnected"
	MsgRoomCreated  MessageType = "room_created"
	MsgRoomJoined   MessageType = "room_joined"
	MsgError        MessageType = "error"
)

type RoomSummary struct {
	Code     string `json:"code"`
	Occupied int    `json:"occupied"`
	Ready    int    `json:"ready"`
	Started  bool   `json:"started"`
	Wall     int    `json:"wall"`
}

type Message struct {
	Type           MessageType        `json:"type"`
	PlayerID       string             `json:"player_id,omitempty"`
	ReconnectToken string             `json:"reconnect_token,omitempty"`
	RoomCode       string             `json:"room_code,omitempty"`
	Name           string             `json:"name,omitempty"`
	Seat           int                `json:"seat,omitempty"`
	ReadySeats     []int              `json:"ready_seats,omitempty"`
	Started        bool               `json:"started,omitempty"`
	OccupiedSeats  []int              `json:"occupied_seats,omitempty"`
	Command        game.GameCommand   `json:"command,omitempty"`
	Result         game.CommandResult `json:"result,omitempty"`
	Snapshot       game.GameSnapshot  `json:"snapshot,omitempty"`
	Rooms          []RoomSummary      `json:"rooms,omitempty"`
	Error          string             `json:"error,omitempty"`
}
