package websocket

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
	"github.com/gpr3211/seer/usdata/pkg/model"
)

func (c *Config) Subscribe(conn *websocket.Conn, s string) error {
	msg := model.SubMsgs{
		Action:  "subscribe",
		Symbols: s,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		log.Println("Failed to parse sub msg")
		return fmt.Errorf("error marshaling subscription: %v", err)
	}
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return fmt.Errorf("error subscribing to %s: %v", s, err)
	}
	return nil
}

func (c Config) Unsub(symbol string) error {
	msg := model.SubMsgs{
		Action:  "unsubscribe",
		Symbols: symbol,
	}
	out, err := json.Marshal(msg)
	if err != nil {
		log.Println("Failed to parse sub msg")
		return fmt.Errorf("error marshaling subscription: %v", err)
	}
	if err := c.socket.WriteMessage(websocket.TextMessage, out); err != nil {
		return fmt.Errorf("error subscribing to %s: %v", symbol, err)
	}
	return nil
}
