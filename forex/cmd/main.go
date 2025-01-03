package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strconv"
	"time"

	server "github.com/gpr3211/seer/forex/internal/handlers/http"
	"github.com/gpr3211/seer/forex/internal/handlers/websocket"
	"github.com/gpr3211/seer/pkg/discovery"
	"github.com/gpr3211/seer/pkg/discovery/consul"
)

const serviceName = "forex"

func main() {

	log.Println("Starting rating service")
	var port int
	flag.IntVar(&port, "port", 6971, "API handler port")

	flag.Parse()
	porto := strconv.Itoa(port)
	log.Printf("Starting rating service on port %s", porto)

	registry, err := consul.NewRegistry("localhost:8500")
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	instanceID := discovery.GenerateInstanceID("forex")

	if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("localhost:%d", port)); err != nil {
		panic(err)
	}

	go func() {
		for {
			if err := registry.ReportHealthyState(instanceID, serviceName); err != nil {
				log.Println("Failed to report healthy state: " + err.Error())
			}
			time.Sleep(1 * time.Second)
		}
	}()
	defer registry.Deregister(ctx, instanceID, serviceName)

	cfg := websocket.NewConfig()
	srv := server.NewServer(porto, cfg)

	go srv.StartServer()
	websocket.StartForex(cfg)
}
