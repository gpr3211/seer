package main

import (
	server "github.com/gpr3211/seer/tower/internal/handlers/http"
)

func main() {

	srv := server.NewServer("4269")

	srv.StartServer()

}
