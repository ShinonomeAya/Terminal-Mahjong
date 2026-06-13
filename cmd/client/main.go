package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"mahjong/internal/game"
	"mahjong/internal/online"
	"mahjong/internal/protocol"
)

func main() {
	if err := run(context.Background(), os.Args[1:], os.Stdout); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, args []string, out io.Writer) error {
	flags := flag.NewFlagSet("client", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	serverURL := flags.String("server", "ws://127.0.0.1:8080/ws", "server websocket URL")
	name := flags.String("name", "Player", "player name")
	joinRoom := flags.String("join", "", "room code to join")
	reconnect := flags.Bool("reconnect", false, "reconnect using the saved session file")
	sessionPath := flags.String("session", ".mahjong-session.json", "session file path")
	discard := flags.Int("discard", 0, "discard a 1-based hand tile index after connecting")
	ready := flags.Bool("ready", false, "send ready after connecting")
	win := flags.Bool("win", false, "declare win after connecting")
	kong := flags.String("kong", "", "declare a concealed kong tile after connecting")
	watch := flags.Bool("watch", false, "keep reading snapshots and reconnect if the connection drops")
	reconnectAttempts := flags.Int("reconnect-attempts", 5, "maximum reconnect attempts in watch mode")
	if err := flags.Parse(args); err != nil {
		return err
	}

	connectCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	client := online.NewClient(*serverURL, *name)
	defer client.Close()

	message, err := connect(connectCtx, client, *sessionPath, *reconnect, *joinRoom)
	if err != nil {
		return err
	}
	if err := online.SaveClientSession(*sessionPath, client.Session()); err != nil {
		return err
	}
	printMessage(out, message)

	if *ready {
		if err := client.SendReady(connectCtx); err != nil {
			return err
		}
		update, err := client.ReadUntil(connectCtx, 2*time.Second, protocol.MsgRoomState, protocol.MsgError)
		if err != nil {
			return err
		}
		printMessage(out, update)
	}

	if *discard > 0 {
		command := game.GameCommand{Kind: game.CommandDiscard, TileIndex: *discard - 1}
		if err := sendCommandAndPrint(connectCtx, client, out, command); err != nil {
			return err
		}
	}
	if *win {
		if err := sendCommandAndPrint(connectCtx, client, out, game.GameCommand{Kind: game.CommandWin}); err != nil {
			return err
		}
	}
	if *kong != "" {
		command := game.GameCommand{Kind: game.CommandKong, Tile: *kong}
		if err := sendCommandAndPrint(connectCtx, client, out, command); err != nil {
			return err
		}
	}
	if *watch {
		return watchClient(ctx, client, *reconnectAttempts, out)
	}
	return nil
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

func sendCommandAndPrint(ctx context.Context, client *online.Client, out io.Writer, command game.GameCommand) error {
	if err := client.SendCommand(ctx, command); err != nil {
		return err
	}
	update, err := client.ReadUntil(ctx, 2*time.Second, protocol.MsgGameSnapshot, protocol.MsgError)
	if err != nil {
		return err
	}
	printMessage(out, update)
	return nil
}

func printMessage(out io.Writer, message protocol.Message) {
	fmt.Fprintf(out, "type=%s room=%s player=%s wall=%d current=%d started=%t\n",
		message.Type,
		message.RoomCode,
		message.PlayerID,
		message.Snapshot.WallCount,
		message.Snapshot.Current,
		message.Started,
	)
}

func watchClient(ctx context.Context, client *online.Client, reconnectAttempts int, out io.Writer) error {
	policy := online.ReconnectPolicy{MaxAttempts: reconnectAttempts, BaseDelay: 200 * time.Millisecond}
	for {
		message, err := client.ReadUntilWithReconnect(ctx, 24*time.Hour, policy, protocol.MsgGameSnapshot, protocol.MsgReconnected)
		if err != nil {
			return err
		}
		printMessage(out, message)
	}
}
