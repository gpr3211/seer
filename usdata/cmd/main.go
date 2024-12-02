package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gpr3211/seer/pkg/discovery"
	"github.com/gpr3211/seer/pkg/discovery/consul"
	server "github.com/gpr3211/seer/usdata/internal/handlers/http"
	"github.com/gpr3211/seer/usdata/internal/handlers/websocket"
)

const serviceName = "us-trade"

func main() {

	log.Println("Starting rating service")
	var port int
	flag.IntVar(&port, "port", 6970, "API handler port")

	flag.Parse()
	porto := strconv.Itoa(port)
	log.Printf("Starting rating service on port %s", porto)

	registry, err := consul.NewRegistry("localhost:8500")
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	instanceID := discovery.GenerateInstanceID("us-trade")

	if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("localhost:%d", port)); err != nil {
		panic(err)
	}

	go func() {
		for {
			if err := registry.ReportHealthyState(instanceID, serviceName); err != nil {
				log.Println("Failed to report healthy state: " + err.Error())
			}
			time.Sleep(15 * time.Second)
		}
	}()
	defer registry.Deregister(ctx, instanceID, serviceName)

	cfg := websocket.NewConfig()
	srv := server.NewServer(porto, cfg)

	go srv.StartServer()
	websocket.StartUS(cfg)

}
