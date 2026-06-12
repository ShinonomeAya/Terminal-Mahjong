package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"mahjong/internal/online"
)

func main() {
	addr := flag.String("addr", ":8080", "server listen address")
	flag.Parse()

	server := online.NewServer()
	fmt.Printf("Terminal Mahjong server listening on %s\n", *addr)
	log.Fatal(http.ListenAndServe(*addr, server))
}
