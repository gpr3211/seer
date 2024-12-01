package websocket

import (
	"fmt"
	"log"
	"net/http"
	//	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/gpr3211/seer/pkg/writer"
	"github.com/gpr3211/seer/usdata/pkg/model"
	// "github.com/joho/godotenv"
)

type Client struct {
	httpClient http.Client
}

func NewClient(timeout time.Duration) *Client {
	return &Client{
		httpClient: http.Client{
			Timeout: timeout,
		},
	}
}

func (cfg *Config) initSocketChannels() {
	cfg.SocketChannels = &SocketChannels{
		OutChan: make(chan model.USTradeTick),
		ErrChan: make(chan error, 10),
		Done:    make(chan struct{}),
		Closed:  false,
	}

}

type SocketChannels struct {
	ErrChan chan error
	OutChan chan model.USTradeTick
	Done    chan struct{}
	Closed  bool
	Mutex   sync.Mutex
}

type Config struct {
	Client  *Client
	Symbols []string
	key     string
	*SocketChannels
	socket *websocket.Conn
}

func NewConfig() *Config {
	return &Config{
		Client:  NewClient(1),
		Symbols: []string{"TSLA", "AAPL", "MSFT"},
	}
}

func StartCrypto() error {

	//	err := godotenv.Load()
	//	if err != nil {
	//		fmt.Println("failed to load")
	//	}

	//	dbUrl := os.Getenv("CONN_STRING")
	//	fmt.Println(dbUrl)
	//	key := os.Getenv("KEY")
	//	fmt.Println(key)

	cfg := NewConfig()
	cfg.startSocket()

	return nil
}

func (cfg *Config) startSocket() error {
	cfg.initSocketChannels()

	path := "wss://ws.eodhistoricaldata.com/ws/us?api_token=demo"
	c, _, err := websocket.DefaultDialer.Dial(path, nil)
	if err != nil {
		return fmt.Errorf("websocket connection error: %v", err)
	}
	cfg.socket = c
	fmt.Println("Starting US-Trade Client ... ")
	fmt.Println("Subscribing ...")

	for _, s := range cfg.Symbols {
		err := cfg.Subscribe(c, s)
		if err != nil {
			log.Printf("Failed to sub")
			return err
		}
		fmt.Printf("Crypto:: %s  Sub complete\n", s)
	}
	go func() {
		defer close(cfg.Done)
		for {
			_, msg, err := cfg.socket.ReadMessage()
			if err != nil {
				cfg.ErrChan <- fmt.Errorf("read error: %v", err)
				return
			}

			tick, err := UnmarshalMsg(msg)
			if err != nil {
				cfg.ErrChan <- fmt.Errorf("unmarshal error: %v", err)
				continue
			}

			switch v := tick.(type) {
			case model.StatusMsg:
				log.Printf("Status MSG: %s -- %s", v.Code, v.Message)
			case model.USTradeTick:
				writer.Writer.AddData(v)
				//	fmt.Println("Crypto in")
				//	cfg.OutChan <- v
				//	fmt.Println(v.Symbol, v.Quantity, v.DailyChange, v.Price, v.Timestamp)
			}
		}
	}()

	// Block and wait
	<-cfg.Done
	return nil
}
