package http

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/gpr3211/seer/crypto"
	"github.com/gpr3211/seer/crypto/pkg/model"
	_ "github.com/lib/pq"
)

func (s *Server) HandleReady(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, crypto.EzError(405)("Wrong Request Method"))
		return
	}
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (s *Server) HandleSubscriptions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, crypto.EzError(405)("you done goofed kid"))
		return
	}
	type parameters struct {
		Action  string `json:"a"`
		Symbols string `json:"s"`
	}
	param := parameters{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&param)
	if err != nil {
		respondWithError(w, crypto.EzError(401)("wrong json format"))
	}
	params := model.SubMsgs{Action: param.Action, Symbols: param.Symbols, CryptoType: "CC"}
	msg, err := json.Marshal(params)
	if err != nil {
		log.Println("Failed to parse sub msg")
		return
	}
	if err := s.Client.Socket.WriteMessage(websocket.TextMessage, msg); err != nil {
		return
	}
	respondWithJSON(w, 200, "Action complete >>>")
	return
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
