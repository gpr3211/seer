package main

import (
	server "github.com/gpr3211/seer/usdata/internal/handlers/http"
	"github.com/gpr3211/seer/usdata/internal/handlers/websocket"
)

func main() {
	cfg := websocket.NewConfig()
	srv := server.NewServer(cfg)

	go srv.StartServer()
	websocket.StartUS(cfg)

}
