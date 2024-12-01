package main

import (
	"log"

	"github.com/gpr3211/seer/crypto/internal/handlers/websocket"
)

func main() {
	err := websocket.StartCrypto()
	if err != nil {
		log.Fatalln("Failed to start Forex")
	}

}
