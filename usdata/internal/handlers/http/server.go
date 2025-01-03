package http

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gpr3211/seer/usdata/internal/handlers/websocket"
	_ "github.com/lib/pq"
)

type Server struct {
	Srv        *http.Server
	Client     *websocket.Config
	context    context.Context
	wg         *sync.WaitGroup
	mu         *sync.RWMutex
	cancelChan context.CancelFunc
}

func NewServer(port string, cfg *websocket.Config) *Server {

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
		Client:     cfg,
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
	mux.HandleFunc("GET /seer/usdata/v1/health", s.HandleReady)
	mux.HandleFunc("POST /seer/usdata/v1/subscribe", s.HandleSubscriptions)
	mux.HandleFunc("GET /seer/usdata/v1/buff", s.HandleStats)

	s.Srv.Handler = mux

	// Channel to catch server errors
	errChan := make(chan error, 1)

	// Start server in goroutine
	go func() {
		fmt.Println("Starting Server on :6970")
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
