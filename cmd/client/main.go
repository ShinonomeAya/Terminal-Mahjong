package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/gorilla/websocket"

	"mahjong/internal/protocol"
)

func main() {
	addr := flag.String("server", "ws://127.0.0.1:8080/ws", "server websocket URL")
	name := flag.String("name", "Player", "player name")
	flag.Parse()

	conn, _, err := websocket.DefaultDialer.Dial(*addr, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	if err := conn.WriteJSON(protocol.Message{Type: protocol.MsgCreateRoom, Name: *name}); err != nil {
		log.Fatal(err)
	}
	var response protocol.Message
	if err := conn.ReadJSON(&response); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("type=%s room=%s player=%s token=%s\n", response.Type, response.RoomCode, response.PlayerID, response.ReconnectToken)
}
