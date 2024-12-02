package main

import (
	"time"

	server "github.com/gpr3211/seer/crypto/internal/handlers/http"
	"github.com/gpr3211/seer/crypto/internal/handlers/websocket"
)

func main() {

	cfg := websocket.NewConfig()
	srv := server.NewServer(cfg)

	go srv.StartServer()
	for { // big brain retry
		websocket.StartCrypto(cfg)
		time.Sleep(time.Second)
	}
}
