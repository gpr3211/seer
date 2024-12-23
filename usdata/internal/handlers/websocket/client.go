package websocket

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/gpr3211/seer/pkg/batcher"
	"github.com/gpr3211/seer/pkg/database"
	"github.com/gpr3211/seer/pkg/writer"
	"github.com/gpr3211/seer/usdata/pkg/model"
	"github.com/joho/godotenv"
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
	DB      *database.Queries
	Client  *Client
	Symbols []string
	key     string
	*SocketChannels
	Socket *websocket.Conn
	Buffer map[string]batcher.BatchStats
}

func NewConfig() *Config {
	return &Config{
		Client:  NewClient(1),
		Symbols: []string{"TSLA", "AAPL", "MSFT"},
		Buffer:  (map[string]batcher.BatchStats{}),
	}
}

// StartCrypto starts the Crypto websocket Top
func StartUS(cfg *Config) error {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("failed to load")
	}
	dbUrl := os.Getenv("CONN_STRING")
	fmt.Println(dbUrl)
	_ = os.Getenv("KEY")

	dab, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatalf("%v", err)
	} else {
		fmt.Println("DB OPEN SUCC")
	}
	dbQueries := database.New(dab)
	cfg.DB = dbQueries
	//	fmt.Println(dbUrl)
	_ = os.Getenv("KEY")

	return cfg.startSocket()
}

// startSocket used in startCrypto
func (cfg *Config) startSocket() error {
	cfg.initSocketChannels()

	var w = writer.NewPeriodicDataWriter(
		time.Minute, // Write interval
		10000,       // Max buffer size
		"US",
		func(symbolBuffers map[string][]batcher.SocketMsg) error {
			for symbol, buffer := range symbolBuffers {
				fmt.Printf("Writing %d US-Trade ticks for symbol %s\n", len(buffer), symbol)
				batches, err := batcher.BatchTicks(buffer, 1)
				if err == -1 {
					return errors.New("Failed to batch ticks")
				}
				for _, batch := range batches {
					stats := batcher.GetBatchStatistics(batch, 1)
					cfg.Buffer[stats.Symbol] = stats
					batcher.InsertBatch(stats, cfg.DB, "US")
				}
			}
			return nil
		},
	)

	path := "wss://ws.eodhistoricaldata.com/ws/us?api_token=demo"
	c, _, err := websocket.DefaultDialer.Dial(path, nil)
	if err != nil {
		return fmt.Errorf("websocket connection error: %v", err)
	}
	cfg.Socket = c
	fmt.Println("Starting US-Trade Client ... ")
	fmt.Println("Subscribing ...")

	for _, s := range cfg.Symbols {
		err := cfg.Subscribe(s)
		if err != nil {
			log.Printf("Failed to sub")
			return err
		}
		fmt.Printf("USTRADE:: %s  Sub complete\n", s)
	}
	go func() {
		defer close(cfg.Done)
		for {
			_, msg, err := cfg.Socket.ReadMessage()
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
				w.AddData(v)
				//	fmt.Println("Crypto in")
				//	cfg.OutChan <- v
			}
		}
	}()

	// Block and wait
	<-cfg.Done
	return nil
}
