package http

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/gpr3211/seer/tower"
	_ "github.com/lib/pq"
)

type Parameters struct {
	Action   string `json:"action"`
	Symbol   string `json:"symbol"`
	Exchange string `json:"exchange"`
}

type Subscriber struct {
	Conn *websocket.Conn
	ID   int
	Subs map[string][]string // map of services and symbols subbed for
}

func NewSubscriber(c *websocket.Conn) *Subscriber {
	return &Subscriber{
		Conn: c,
		ID:   rand.IntN(13011),
		Subs: make(map[string][]string),
	}
}

type Server struct {
	Srv        *http.Server
	Client     Client
	User       []*Subscriber
	context    context.Context
	wg         *sync.WaitGroup
	mu         *sync.RWMutex
	cancelChan context.CancelFunc
}

func NewServer(port string) *Server {

	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  time.Second * 30,
		WriteTimeout: time.Second * 30,
	}
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		Srv:        srv,
		Client:     NewClient(),
		context:    ctx,
		wg:         &wg,
		mu:         &sync.RWMutex{},
		cancelChan: cancel,
	}
}
func (s *Server) StartServer() {
	s.wg.Add(1)
	defer s.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				fmt.Println("Tick")
				s.Client.FetchAllStats()

				s.mu.RLock()
				for _, user := range s.User {
					// Check if connection is still alive
					// TODO FIX THIS SHIT
					if err := user.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
						s.mu.Lock()
						log.Printf("User %d connection dead, skipping", user.ID)
						newUsers := []*Subscriber{}
						for _, us := range s.User {
							if us.ID != user.ID {
								newUsers = append(newUsers, us)
							}
						}
						s.User = newUsers
						s.mu.Unlock()
						continue
					}

					for exchange, symbols := range user.Subs {
						for _, symbol := range symbols {
							if stats, ok := s.Client.Buffer[exchange][symbol]; ok {
								if err := user.Conn.WriteJSON(stats); err != nil {
									log.Printf("Error sending data to user %d: %v", user.ID)
								} else {
									log.Printf("Sent %s:%s data to user %d", exchange, symbol, user.ID)
								}
							}
						}
					}
				}
				s.mu.RUnlock()
			}
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /seer/tower/ws", s.handleSubscribe)
	s.Srv.Handler = mux

	errChan := make(chan error, 1)

	go func() {
		fmt.Printf("Starting Server on %v\n", s.Srv.Addr)
		if err := s.Srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()
	// Wait for context cancellation or server error
	select {
	case <-s.context.Done():
		fmt.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.Srv.Shutdown(ctx)
		return
	case err := <-errChan:
		fmt.Println(err)
		return
	}
}

func (s *Server) handleSubscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer c.Close()

	log.Println("Upgraded to websocket")

	sub := NewSubscriber(c)
	s.mu.Lock()
	s.User = append(s.User, sub)
	s.mu.Unlock()

	params := Parameters{}

	for {
		err := c.ReadJSON(&params)
		if err != nil {
			log.Printf("Error reading JSON: %v", err)
			c.WriteJSON(tower.EzError(407)("bad json"))
			continue
		}

		switch params.Action {
		case "subscribe":
			s.mu.RLock()
			symbols := s.Client.Buffer.GetSymbols(params.Exchange)
			s.mu.RUnlock()

			found := false
			for _, sym := range symbols {
				if sym == params.Symbol {
					s.mu.Lock()
					sub.Subs[params.Exchange] = append(sub.Subs[params.Exchange], params.Symbol)
					s.mu.Unlock()
					found = true
					break
				}
			}

			if found {
				if err := c.WriteJSON(tower.EzError(200)("Subscribed successfully")); err != nil {
					log.Printf("Error writing success message: %v", err)
				}
			} else {
				if err := c.WriteJSON(tower.EzError(407)("symbol not found")); err != nil {
					log.Printf("Error writing error message: %v", err)
				}
			}
		case "unsubscribe":
			s.mu.RLock()
			new := []string{}
			for _, item := range sub.Subs[params.Exchange] {
				if item != params.Symbol {
					new = append(new, item)
				}
			}
			sub.Subs[params.Exchange] = new
			s.mu.RUnlock()
			c.WriteJSON(tower.EzError(200)("Unsubbed from " + params.Symbol))
		}
	}
}
