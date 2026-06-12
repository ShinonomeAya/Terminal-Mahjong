package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"mahjong/internal/game"
	"mahjong/internal/online"
	"mahjong/internal/protocol"
)

func main() {
	serverURL := flag.String("server", "ws://127.0.0.1:8080/ws", "server websocket URL")
	name := flag.String("name", "Player", "player name")
	joinRoom := flag.String("join", "", "room code to join")
	reconnect := flag.Bool("reconnect", false, "reconnect using the saved session file")
	sessionPath := flag.String("session", ".mahjong-session.json", "session file path")
	discard := flag.Int("discard", 0, "discard a 1-based hand tile index after connecting")
	watch := flag.Bool("watch", false, "keep reading snapshots and reconnect if the connection drops")
	reconnectAttempts := flag.Int("reconnect-attempts", 5, "maximum reconnect attempts in watch mode")
	flag.Parse()

	connectCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client := online.NewClient(*serverURL, *name)
	defer client.Close()

	message, err := connect(connectCtx, client, *sessionPath, *reconnect, *joinRoom)
	if err != nil {
		log.Fatal(err)
	}
	if err := online.SaveClientSession(*sessionPath, client.Session()); err != nil {
		log.Fatal(err)
	}
	printMessage(message)

	if *discard > 0 {
		command := game.GameCommand{Kind: game.CommandDiscard, TileIndex: *discard - 1}
		if err := client.SendCommand(connectCtx, command); err != nil {
			log.Fatal(err)
		}
		update, err := client.ReadUntil(connectCtx, 2*time.Second, protocol.MsgGameSnapshot)
		if err != nil {
			log.Fatal(err)
		}
		printMessage(update)
	}
	if *watch {
		watchClient(context.Background(), client, *reconnectAttempts)
	}
}

func connect(ctx context.Context, client *online.Client, sessionPath string, reconnect bool, joinRoom string) (protocol.Message, error) {
	if reconnect {
		session, err := online.LoadClientSession(sessionPath)
		if err != nil {
			return protocol.Message{}, err
		}
		return client.Reconnect(ctx, session)
	}
	if joinRoom != "" {
		return client.JoinRoom(ctx, joinRoom)
	}
	return client.CreateRoom(ctx)
}

func printMessage(message protocol.Message) {
	fmt.Printf("type=%s room=%s player=%s wall=%d current=%d\n",
		message.Type,
		message.RoomCode,
		message.PlayerID,
		message.Snapshot.WallCount,
		message.Snapshot.Current,
	)
}

func watchClient(ctx context.Context, client *online.Client, reconnectAttempts int) {
	policy := online.ReconnectPolicy{MaxAttempts: reconnectAttempts, BaseDelay: 200 * time.Millisecond}
	for {
		message, err := client.ReadUntilWithReconnect(ctx, 24*time.Hour, policy, protocol.MsgGameSnapshot, protocol.MsgReconnected)
		if err != nil {
			log.Fatal(err)
		}
		printMessage(message)
	}
}
