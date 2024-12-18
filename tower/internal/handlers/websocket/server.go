package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/gpr3211/seer/pkg/batcher"
	"github.com/gpr3211/seer/tower"
)

type SocketServer struct {
	Srv        *http.Server
	Client     *http.Client
	context    context.Context
	wg         *sync.WaitGroup
	mu         *sync.RWMutex
	cancelChan context.CancelFunc
	batchChan  chan batcher.BatchStats
}

func NewServer(port string) *SocketServer {

	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  time.Second * 30,
		WriteTimeout: time.Second * 30,
	}
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	return &SocketServer{
		Srv:        srv,
		context:    ctx,
		wg:         &wg,
		mu:         &sync.RWMutex{},
		cancelChan: cancel,
	}
}
func (s *SocketServer) StartServer() {
	s.wg.Add(1)
	defer s.wg.Done()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /seer/tower/v1/health", s.HandleReady)
	//	mux.HandleFunc("POST /seer/crypto/v1/subscribe", s.HandleSubscriptions)
	//	mux.HandleFunc("GET /seer/crypto/v1/buff", s.HandleStats)

	s.Srv.Handler = mux

	// Channel to catch server errors
	errChan := make(chan error, 1)

	// Start server in goroutine
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

func (s *SocketServer) HandleReady(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, tower.EzError(405)("Wrong Request Method"))
		return
	}
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func respondWithError(w http.ResponseWriter, Err tower.APIError) {
	w.Header().Set("Content-type", "application/json")
	dat, err := json.Marshal(Err)
	if err != nil {
		log.Printf("error marshalling JSON @respondWithJSON")
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(int(Err.StatusCode))
	w.Write(dat)
}

//
//
//

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-type", "application/json")
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("error marshalling JSON @respondWithJSON")
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}
