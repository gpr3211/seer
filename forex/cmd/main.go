package main

import (
	"log"

	"flag"
	"strconv"

	server "github.com/gpr3211/seer/forex/internal/handlers/http"
	"github.com/gpr3211/seer/forex/internal/handlers/websocket"
)

func main() {

	log.Println("Starting rating service")
	var port int
	flag.IntVar(&port, "port", 6971, "API handler port")

	flag.Parse()
	porto := strconv.Itoa(port)
	log.Printf("Starting rating service on port %s", porto)

	cfg := websocket.NewConfig()
	srv := server.NewServer(porto, cfg)

	go srv.StartServer()
	websocket.StartForex(cfg)
}
