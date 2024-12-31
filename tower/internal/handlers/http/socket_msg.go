package http

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/gpr3211/seer/pkg/batcher"
	"github.com/gpr3211/seer/pkg/clog"
	"github.com/gpr3211/seer/tower"
	"log"
)

func fmap[A, B any](mapFunc func(A) B, sliceA []A) []B {
	sliceB := make([]B, len(sliceA))
	for i, a := range sliceA {
		sliceB[i] = mapFunc(a)
	}
	return sliceB
}

type SocketMessage interface {
	IsWebsocket()
}

type APIMsg struct {
	StatusCode int    `json:"status_code"`
	Msg        string `json:"msg"`
}

type SubMsg struct {
	Action   string `json:"action"`
	Exchange string `json:"exchange"`
	Symbol   string `json:"symbol"`
}

func (s SubMsg) IsWebsocket()     {}
func (s APIMsg) IsWebsocket()     {}
func (s StatUpdate) IsWebsocket() {}

type StatUpdate struct {
	Exchange      string  `json:"exchange"`
	Symbol        string  `json:"symbol"`
	BatchSequence int64   `json:"sequence"`
	StartTime     int64   `json:"start"`
	EndTime       int64   `json:"end"`
	Open          float64 `json:"open"`
	High          float64 `json:"high"`
	Low           float64 `json:"low"`
	Close         float64 `json:"close"`
	Volume        float64 `json:"volume"`
	Period        int32   `json:"period"`
	Normal        float64 `json:"-"`
}

func statToBatch(s StatUpdate) batcher.BatchStats {
	return batcher.BatchStats{
		Symbol:        s.Symbol,
		BatchSequence: s.BatchSequence,
		StartTime:     s.StartTime,
		EndTime:       s.EndTime,
		Open:          s.Open,
		High:          s.High,
		Low:           s.Low,
		Close:         s.Close,
		Volume:        s.Volume,
		Period:        s.Period,
	}
}

func (s *Server) manageSubMsg(v SubMsg, sub *Subscriber, c *websocket.Conn) {
	if v.Action == "subscribe" {
		log.Printf("Received subscribe action from user %d: %+v", sub.ID, v)

		s.mu.RLock()
		symbols := s.Client.Buffer.GetSymbols(v.Exchange)
		s.mu.RUnlock()

		found := false
		for _, sym := range symbols {
			if sym == v.Symbol {
				log.Printf("Symbol %s found in exchange %s", v.Symbol, v.Exchange)
				s.mu.Lock()
				sub.Subs[v.Exchange] = append(sub.Subs[v.Exchange], v.Symbol)
				s.mu.Unlock()
				found = true
				break
			}
		}

		if found {
			log.Printf("User %d successfully subscribed to %s:%s", sub.ID, v.Exchange, v.Symbol)
			if err := c.WriteJSON(tower.EzError(200)("Subscribed successfully")); err != nil {
				clog.Printf("Error writing success message: %v", err)
			}
		} else {
			log.Printf("Symbol %s not found in exchange %s, adding new entry", v.Symbol, v.Exchange)
			s.mu.Lock()
			if _, ok := s.Client.Buffer[v.Exchange]; !ok {
				s.Client.Buffer[v.Exchange] = make(map[string]batcher.BatchStats)
			}
			s.Client.Buffer[v.Exchange][v.Symbol] = batcher.BatchStats{}
			sub.Subs[v.Exchange] = append(sub.Subs[v.Exchange], v.Symbol) // Ensure symbol is added here too
			s.mu.Unlock()
			if err := c.WriteJSON(tower.EzError(200)("Subscribed")); err != nil {
				clog.Printf("Error writing subscribe confirmation: %v", err)
			}
		}

		log.Printf("Current subscriptions for user %d: %+v", sub.ID, sub.Subs)
		return
	}

	if v.Action == "unsubscribe" {
		log.Printf("Received unsubscribe action from user %d: %+v", sub.ID, v)
		s.mu.Lock()
		defer s.mu.Unlock()

		if subs, ok := sub.Subs[v.Exchange]; ok {
			newSubs := []string{}
			for _, item := range subs {
				if item != v.Symbol {
					newSubs = append(newSubs, item)
				}
			}
			sub.Subs[v.Exchange] = newSubs
			log.Printf("User %d unsubscribed from %s:%s", sub.ID, v.Exchange, v.Symbol)
			c.WriteJSON(tower.EzError(200)("Unsubbed from " + v.Symbol))
		}
	} else {
		log.Printf("Invalid action from user %d: %+v", sub.ID, v)
		c.WriteJSON(tower.EzError(407)("Invalid action input"))
	}
}

// manageStatUpdate handles stat update from an exchange and sends out msg to users
func (s *Server) manageStatUpdate(v StatUpdate) {
	params := statToBatch(v)
	fmt.Println(params)

	s.mu.Lock()
	s.Client.Buffer[v.Exchange][v.Symbol] = params
	fmt.Printf("Stats updated for %s %s", v.Exchange, v.Symbol)
	s.mu.Unlock()

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, user := range s.User {
		// Check connection health
		if err := user.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
			s.mu.Lock()
			clog.Printf("User %d connection dead, removing subscriber", user.ID)
			s.User = removeSubscriber(s.User, user.ID)
			s.mu.Unlock()
			continue
		}

		// Check if user is subscribed to THIS specific symbol on THIS exchange
		if symbols, ok := user.Subs[v.Exchange]; ok {
			for _, symbol := range symbols {
				if symbol == v.Symbol {
					if err := user.Conn.WriteJSON(params); err != nil {
						clog.Printf("Error sending data to user %d: %v", user.ID, err)
					} else {
						clog.Printf("Sent %s:%s data to user %d", v.Exchange, v.Symbol, user.ID)
					}
					break // Found and sent the matching symbol, no need to continue checking
				}
			}
		}
	}
} // Helper function to remove a subscriber by ID
func removeSubscriber(subscribers []*Subscriber, id int) []*Subscriber {
	newSubscribers := []*Subscriber{}
	for _, user := range subscribers {
		if user.ID != id {
			newSubscribers = append(newSubscribers, user)
		}
	}
	return newSubscribers
}
