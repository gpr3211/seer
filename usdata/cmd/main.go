package main

import (
	"flag"
	server "github.com/gpr3211/seer/usdata/internal/handlers/http"
	"github.com/gpr3211/seer/usdata/internal/handlers/websocket"
	"log"
	"strconv"
)

func main() {

	log.Println("Starting rating service")
	var port int
	flag.IntVar(&port, "port", 6970, "API handler port")

	flag.Parse()
	porto := strconv.Itoa(port)
	log.Printf("Starting rating service on port %s", porto)

	cfg := websocket.NewConfig()
	srv := server.NewServer(porto, cfg)

	go srv.StartServer()
	websocket.StartUS(cfg)

}
