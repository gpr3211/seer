package websocket

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/gpr3211/seer/crypto/pkg/model"
	"github.com/gpr3211/seer/pkg/batcher"
	"github.com/gpr3211/seer/pkg/database"
	"github.com/gpr3211/seer/pkg/writer"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
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
		OutChan: make(chan model.CryptoTick),
		ErrChan: make(chan error, 10),
		Done:    make(chan struct{}),
		Closed:  false,
	}

}

type SocketChannels struct {
	ErrChan chan error
	OutChan chan model.CryptoTick
	Done    chan struct{}
	Closed  bool
	Mutex   sync.Mutex
}

type Config struct {
	DB      *database.Queries
	Client  *Client
	Symbols []string
	key     string
	*SocketChannels
	socket *websocket.Conn
}

func NewConfig() *Config {
	return &Config{
		Client:  NewClient(1),
		Symbols: []string{"ETH-USD", "BTC-USD"},
	}
}

var w = writer.NewPeriodicDataWriter(
	time.Minute, // Write interval
	10000,       // Max buffer size
	"CC",
	func(symbolBuffers map[string][]batcher.SocketMsg) error {
		for symbol, buffer := range symbolBuffers {
			fmt.Printf("Writing %d Crypto ticks for symbol %s\n", len(buffer), symbol)
			batches, err := batcher.BatchTicks(buffer, 1)
			if err == -1 {
				return errors.New("Failed to batch ticks")
			}
			for _, batch := range batches {
				stats := batcher.GetBatchStatistics(batch, 1)
				fmt.Println("INSERT ADDING Crypto STATS ")
				//	InsertBatch(stats, cfg.DB, exhange())
				fmt.Println("Insert complete Crypto:", stats.Symbol, stats.EndTime)
			}
		}
		return nil
	},
)

func StartCrypto() error {

	err := godotenv.Load()
	if err != nil {
		fmt.Println("failed to load")
	}

	dbUrl := os.Getenv("CONN_STRING")
	//fmt.Println(dbUrl)

	_, err = sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatalf("%v", err)
	} else {
		fmt.Println("DB OPEN SUCC")
	}
	//	dbQueries := database.New(dab)
	//	fmt.Println(dbUrl)

	_ = os.Getenv("KEY")

	cfg := NewConfig()
	cfg.startSocket()

	return nil
}

func (cfg *Config) startSocket() error {
	cfg.initSocketChannels()

	path := "wss://ws.eodhistoricaldata.com/ws/crypto?api_token=demo"
	c, _, err := websocket.DefaultDialer.Dial(path, nil)
	if err != nil {
		return fmt.Errorf("websocket connection error: %v", err)
	}
	cfg.socket = c
	fmt.Println("Starting Crypto Client ... ")
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
			case model.CryptoTick:
				w.AddData(v)
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
