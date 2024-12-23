package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/gpr3211/seer/crypto"
	"github.com/gpr3211/seer/crypto/pkg/model"
	"github.com/gpr3211/seer/pkg/batcher"
	_ "github.com/lib/pq"
)

func (s *Server) HandleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, crypto.EzError(405)("Wrong Request Method"))
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	//	last := s.Client.Buffer[param.Symbol]

	out := []batcher.BatchStats{}
	for _, k := range s.Client.Buffer {
		out = append(out, k)
	}
	respondWithJSON(w, 200, out)
	return
}

func (s *Server) HandleSubscriptions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, crypto.EzError(405)("you done goofed kid"))
		return
	}
	type parameters struct {
		Action  string `json:"action"`
		Symbols string `json:"symbols"`
	}
	param := parameters{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&param)
	if err != nil {
		respondWithError(w, crypto.EzError(401)("wrong json format"))
		return
	}
	params := model.SubMsgs{Action: param.Action, Symbols: param.Symbols, CryptoType: "CC"}

	if params.Action != "subscribe" && params.Action != "unsubscribe" {
		respondWithError(w, crypto.BadInput)
		return
	}

	msg, err := json.Marshal(params)
	if err != nil {
		log.Println("Failed to parse sub msg")
		return
	}
	if err := s.Client.Socket.WriteMessage(websocket.TextMessage, msg); err != nil {
		return
	}
	switch params.Action {
	case "subscribe":
		s.Client.Symbols = append(s.Client.Symbols, params.Symbols)
		respondWithJSON(w, 200, fmt.Sprintf("Subbed to %s", params.Symbols))
	case "unsubscribe":
		new := []string{}
		for _, s := range s.Client.Symbols {
			if s != params.Symbols {
				new = append(new, s)
			}
		}
		s.Client.Symbols = new
		respondWithJSON(w, 200, fmt.Sprintf("Unsubbed from %s", params.Symbols))
	}
	return
}

func (s *Server) HandleReady(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, crypto.EzError(405)("Wrong Request Method"))
		return
	}
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func respondWithError(w http.ResponseWriter, Err crypto.APIError) {
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
