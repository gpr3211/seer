package websocket

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/gpr3211/seer/forex/pkg/model"
	"github.com/joho/godotenv"
	"net/http"
	"os"
	"sync"
	"time"
)

type Client struct {
	httpClient http.Client
}

func NewClient(timeout time.Duration) Client {
	return *&Client{
		httpClient: http.Client{
			Timeout: timeout,
		},
	}
}

type SocketChannels struct {
	ErrChan chan error
	OutChan chan model.ForexTick
	Done    chan struct{}
	Closed  bool
	Mutex   sync.Mutex
}

type Config struct {
	Client  *Client
	Symbols []string
	key     string
	SocketChannels
	socket *websocket.Conn
}

func StartForex() {

	err := godotenv.Load()
	if err != nil {
		fmt.Println("failed to load")
	}

	dbUrl := os.Getenv("CONN_STRING")
	fmt.Println(dbUrl)
	key := os.Getenv("KEY")

}
