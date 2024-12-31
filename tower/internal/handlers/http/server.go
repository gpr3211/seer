package http

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
	"log"
	"math/rand/v2"
	"net/http"
	"sync"
	"time"
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

	for {
		var rawMsg map[string]interface{}
		err := c.ReadJSON(&rawMsg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("Unexpected websocket error: %v", err)
			}
			// Normal closure or going away - just return from handler
			return
		} // Check if it's a SubMsg based on the "action" field
		if action, ok := rawMsg["action"].(string); ok {
			switch action {
			case "subscribe", "unsubscribe":
				var subMsg SubMsg
				if err := mapToStruct(rawMsg, &subMsg); err == nil {
					s.manageSubMsg(subMsg, sub, c)
				} else {
					c.WriteJSON(APIMsg{StatusCode: 400, Msg: "Invalid subscription message"})
				}
			default:
				c.WriteJSON(APIMsg{StatusCode: 407, Msg: "Unknown action"})
			}
		} else if isStatUpdate(rawMsg) {
			// If it's not an action, check if it's likely a StatUpdate
			var statUpdate StatUpdate
			if err := mapToStruct(rawMsg, &statUpdate); err == nil {
				s.manageStatUpdate(statUpdate)
			} else {
				c.WriteJSON(APIMsg{StatusCode: 400, Msg: "Invalid stats update message"})
			}
		} else {
			// Default case for unknown message formats
			c.WriteJSON(APIMsg{StatusCode: 407, Msg: "Unknown message format"})
		}
	}
}
func isStatUpdate(data map[string]interface{}) bool {
	// Check for essential fields that would indicate it's a StatUpdate
	_, hasSymbol := data["symbol"]
	_, hasSequence := data["close"]
	return hasSymbol && hasSequence
}

// Helper function to map raw JSON into a struct
func mapToStruct(data map[string]interface{}, target interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonData, target)
}
