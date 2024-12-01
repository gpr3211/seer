package main

import (
	"log"

	"github.com/gpr3211/seer/forex/internal/handlers/websocket"
)

func main() {
	err := websocket.StartForex()
	if err != nil {
		log.Fatalln("Failed to start Forex")
	}

}
