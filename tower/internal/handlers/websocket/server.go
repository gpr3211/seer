package websocket

import (
	"encoding/json"
	"github.com/gpr3211/seer/tower"
	_ "github.com/lib/pq"
	"log"
	"net/http"
)

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
