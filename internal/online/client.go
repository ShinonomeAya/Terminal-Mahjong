package online

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/websocket"

	"mahjong/internal/game"
	"mahjong/internal/protocol"
)

type ClientSession struct {
	ServerURL      string `json:"server_url"`
	Name           string `json:"name"`
	PlayerID       string `json:"player_id"`
	ReconnectToken string `json:"reconnect_token"`
	RoomCode       string `json:"room_code"`
}

type Client struct {
	serverURL string
	name      string
	conn      *websocket.Conn
	session   ClientSession
}

func NewClient(serverURL string, name string) *Client {
	return &Client{
		serverURL: serverURL,
		name:      name,
		session: ClientSession{
			ServerURL: serverURL,
			Name:      name,
		},
	}
}

func (c *Client) Close() {
	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
	}
}

func (c *Client) Session() ClientSession {
	return c.session
}

func (c *Client) CreateRoom(ctx context.Context) (protocol.Message, error) {
	if err := c.connect(ctx); err != nil {
		return protocol.Message{}, err
	}
	if err := c.write(protocol.Message{Type: protocol.MsgCreateRoom, Name: c.name}); err != nil {
		return protocol.Message{}, err
	}
	message, err := c.ReadUntil(ctx, 2*time.Second, protocol.MsgRoomCreated, protocol.MsgError)
	if err != nil {
		return protocol.Message{}, err
	}
	return c.acceptSessionMessage(message)
}

func (c *Client) JoinRoom(ctx context.Context, roomCode string) (protocol.Message, error) {
	if err := c.connect(ctx); err != nil {
		return protocol.Message{}, err
	}
	if err := c.write(protocol.Message{Type: protocol.MsgJoinRoom, RoomCode: roomCode, Name: c.name}); err != nil {
		return protocol.Message{}, err
	}
	message, err := c.ReadUntil(ctx, 2*time.Second, protocol.MsgRoomJoined, protocol.MsgError)
	if err != nil {
		return protocol.Message{}, err
	}
	return c.acceptSessionMessage(message)
}

func (c *Client) Reconnect(ctx context.Context, session ClientSession) (protocol.Message, error) {
	c.session = session
	c.serverURL = session.ServerURL
	c.name = session.Name
	if err := c.connect(ctx); err != nil {
		return protocol.Message{}, err
	}
	if err := c.write(protocol.Message{
		Type:           protocol.MsgReconnect,
		PlayerID:       session.PlayerID,
		ReconnectToken: session.ReconnectToken,
	}); err != nil {
		return protocol.Message{}, err
	}
	message, err := c.ReadUntil(ctx, 2*time.Second, protocol.MsgReconnected, protocol.MsgError)
	if err != nil {
		return protocol.Message{}, err
	}
	return c.acceptSessionMessage(message)
}

func (c *Client) SendCommand(ctx context.Context, command game.GameCommand) error {
	if err := c.connect(ctx); err != nil {
		return err
	}
	return c.write(protocol.Message{Type: protocol.MsgPlayCommand, Command: command})
}

func (c *Client) ReadUntil(ctx context.Context, timeout time.Duration, messageTypes ...protocol.MessageType) (protocol.Message, error) {
	if err := c.connect(ctx); err != nil {
		return protocol.Message{}, err
	}
	deadline := time.Now().Add(timeout)
	for {
		if err := ctx.Err(); err != nil {
			return protocol.Message{}, err
		}
		if time.Now().After(deadline) {
			return protocol.Message{}, fmt.Errorf("timeout waiting for %v", messageTypes)
		}
		if err := c.conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond)); err != nil {
			return protocol.Message{}, err
		}
		var message protocol.Message
		err := c.conn.ReadJSON(&message)
		if err != nil {
			continue
		}
		for _, messageType := range messageTypes {
			if message.Type == messageType {
				if message.Type == protocol.MsgError {
					return protocol.Message{}, errors.New(message.Error)
				}
				return message, nil
			}
		}
	}
}

func (c *Client) connect(ctx context.Context) error {
	if c.conn != nil {
		return nil
	}
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, c.serverURL, nil)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

func (c *Client) write(message protocol.Message) error {
	if c.conn == nil {
		return fmt.Errorf("client is not connected")
	}
	return c.conn.WriteJSON(message)
}

func (c *Client) acceptSessionMessage(message protocol.Message) (protocol.Message, error) {
	if message.Type == protocol.MsgError {
		return protocol.Message{}, errors.New(message.Error)
	}
	c.session.PlayerID = message.PlayerID
	c.session.ReconnectToken = message.ReconnectToken
	c.session.RoomCode = message.RoomCode
	return message, nil
}

func SaveClientSession(path string, session ClientSession) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func LoadClientSession(path string) (ClientSession, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ClientSession{}, err
	}
	var session ClientSession
	if err := json.Unmarshal(data, &session); err != nil {
		return ClientSession{}, err
	}
	return session, nil
}
