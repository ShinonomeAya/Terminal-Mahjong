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
	Seat           int    `json:"seat"`
}

type Client struct {
	serverURL string
	name      string
	conn      *websocket.Conn
	session   ClientSession
}

type ReconnectPolicy struct {
	MaxAttempts int
	BaseDelay   time.Duration
	OnAttempt   func(attempt int, max int)
	OnSuccess   func()
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
	return c.CreateRoomWithRules(ctx, game.ModeCompatibility, game.RuleConfig{})
}

func (c *Client) CreateRoomWithRules(ctx context.Context, mode game.RuleMode, config game.RuleConfig) (protocol.Message, error) {
	if err := c.connect(ctx); err != nil {
		return protocol.Message{}, err
	}
	if err := c.write(protocol.Message{Type: protocol.MsgCreateRoom, Name: c.name, Mode: mode, RuleConfig: config}); err != nil {
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

func (c *Client) ListRooms(ctx context.Context) ([]protocol.RoomSummary, error) {
	if err := c.connect(ctx); err != nil {
		return nil, err
	}
	if err := c.write(protocol.Message{Type: protocol.MsgListRooms}); err != nil {
		return nil, err
	}
	message, err := c.ReadUntil(ctx, 2*time.Second, protocol.MsgRoomList, protocol.MsgError)
	if err != nil {
		return nil, err
	}
	return append([]protocol.RoomSummary(nil), message.Rooms...), nil
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

func (c *Client) SendReady(ctx context.Context) error {
	if err := c.connect(ctx); err != nil {
		return err
	}
	return c.write(protocol.Message{Type: protocol.MsgReady})
}

func (c *Client) RequestReplay(ctx context.Context) error {
	if err := c.connect(ctx); err != nil {
		return err
	}
	return c.write(protocol.Message{Type: protocol.MsgRequestReplay})
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
		if err := c.conn.SetReadDeadline(deadline); err != nil {
			return protocol.Message{}, err
		}
		var message protocol.Message
		err := c.conn.ReadJSON(&message)
		if err != nil {
			return protocol.Message{}, err
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

func (c *Client) ReadUntilWithReconnect(ctx context.Context, timeout time.Duration, policy ReconnectPolicy, messageTypes ...protocol.MessageType) (protocol.Message, error) {
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
		if err := c.conn.SetReadDeadline(deadline); err != nil {
			reconnected, reconnectErr := c.reconnectWithBackoff(ctx, policy)
			if reconnectErr != nil {
				return protocol.Message{}, reconnectErr
			}
			for _, messageType := range messageTypes {
				if reconnected.Type == messageType {
					return reconnected, nil
				}
			}
			continue
		}
		var message protocol.Message
		err := c.conn.ReadJSON(&message)
		if err == nil {
			for _, messageType := range messageTypes {
				if message.Type == messageType {
					return message, nil
				}
			}
			continue
		}
		if time.Now().After(deadline) {
			return protocol.Message{}, fmt.Errorf("timeout waiting for %v", messageTypes)
		}
		reconnected, reconnectErr := c.reconnectWithBackoff(ctx, policy)
		if reconnectErr != nil {
			return protocol.Message{}, reconnectErr
		}
		for _, messageType := range messageTypes {
			if reconnected.Type == messageType {
				return reconnected, nil
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

func (c *Client) reconnectWithBackoff(ctx context.Context, policy ReconnectPolicy) (protocol.Message, error) {
	if policy.MaxAttempts <= 0 {
		policy.MaxAttempts = 5
	}
	if policy.BaseDelay <= 0 {
		policy.BaseDelay = 100 * time.Millisecond
	}
	session := c.session
	if session.PlayerID == "" || session.ReconnectToken == "" {
		return protocol.Message{}, fmt.Errorf("missing reconnect session")
	}
	var lastErr error
	for attempt := 0; attempt < policy.MaxAttempts; attempt++ {
		if policy.OnAttempt != nil {
			policy.OnAttempt(attempt+1, policy.MaxAttempts)
		}
		c.Close()
		message, err := c.Reconnect(ctx, session)
		if err == nil {
			if policy.OnSuccess != nil {
				policy.OnSuccess()
			}
			return message, nil
		}
		lastErr = err
		delay := policy.BaseDelay << attempt
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return protocol.Message{}, ctx.Err()
		case <-timer.C:
		}
	}
	return protocol.Message{}, fmt.Errorf("reconnect failed after %d attempts: %w", policy.MaxAttempts, lastErr)
}

func (c *Client) acceptSessionMessage(message protocol.Message) (protocol.Message, error) {
	if message.Type == protocol.MsgError {
		return protocol.Message{}, errors.New(message.Error)
	}
	c.session.PlayerID = message.PlayerID
	c.session.ReconnectToken = message.ReconnectToken
	c.session.RoomCode = message.RoomCode
	c.session.Seat = message.Seat
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
