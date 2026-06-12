package protocol

import "mahjong/internal/game"

type MessageType string

const (
	MsgHello        MessageType = "hello"
	MsgCreateRoom   MessageType = "create_room"
	MsgJoinRoom     MessageType = "join_room"
	MsgReady        MessageType = "ready"
	MsgGameSnapshot MessageType = "game_snapshot"
	MsgPlayCommand  MessageType = "play_command"
	MsgReconnect    MessageType = "reconnect"
	MsgReconnected  MessageType = "reconnected"
	MsgRoomCreated  MessageType = "room_created"
	MsgRoomJoined   MessageType = "room_joined"
	MsgError        MessageType = "error"
)

type Message struct {
	Type           MessageType        `json:"type"`
	PlayerID       string             `json:"player_id,omitempty"`
	ReconnectToken string             `json:"reconnect_token,omitempty"`
	RoomCode       string             `json:"room_code,omitempty"`
	Name           string             `json:"name,omitempty"`
	Seat           int                `json:"seat,omitempty"`
	Command        game.GameCommand   `json:"command,omitempty"`
	Result         game.CommandResult `json:"result,omitempty"`
	Snapshot       game.GameSnapshot  `json:"snapshot,omitempty"`
	Error          string             `json:"error,omitempty"`
}
