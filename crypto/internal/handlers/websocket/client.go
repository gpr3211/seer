package websocket

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	_ "github.com/lib/pq"

	"github.com/gorilla/websocket"
	"github.com/gpr3211/seer/crypto/pkg/model"
	"github.com/gpr3211/seer/pkg/batcher"
	"github.com/gpr3211/seer/pkg/database"
	"github.com/gpr3211/seer/pkg/discovery/consul"
	"github.com/gpr3211/seer/pkg/writer"
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
		OutChan: make(chan model.CryptoTick, 250),
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
type Gateway struct {
	registry *consul.Registry
}

func New(registry *consul.Registry) *Gateway {
	return &Gateway{registry}
}

type Config struct {
	DB      *database.Queries
	Client  *Client
	Symbols []string
	key     string
	*SocketChannels
	Socket      *websocket.Conn
	towerSocket *websocket.Conn
	Buffer      map[string]batcher.BatchStats
}

func NewConfig() *Config {
	return &Config{
		Client:  NewClient(2 * time.Minute),
		Symbols: []string{"BTC-USD", "ETH-USD"},
		Buffer:  (map[string]batcher.BatchStats{}),
	}
}

type StatUpdate struct {
	Exchange      string  `json:"exchange"`
	Symbol        string  `json:"symbol"`
	BatchSequence int64   `json:"sequence"`
	StartTime     int64   `json:"start"`
	EndTime       int64   `json:"end"`
	Open          float64 `json:"open"`
	High          float64 `json:"high"`
	Low           float64 `json:"low"`
	Close         float64 `json:"close"`
	Volume        float64 `json:"volume"`
	Period        int32   `json:"period"`
	Normal        float64 `json:"-"`
}

func batchToStat(s batcher.BatchStats) StatUpdate {
	return StatUpdate{
		Exchange:      "crypto",
		Symbol:        s.Symbol,
		BatchSequence: s.BatchSequence,
		StartTime:     s.StartTime,
		EndTime:       s.EndTime,
		Open:          s.Open,
		High:          s.High,
		Low:           s.Low,
		Close:         s.Close,
		Volume:        s.Volume,
		Period:        s.Period,
	}
}

// TODO add ping from client to server to check client health

func StartCrypto(cfg *Config) error {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("failed to load")
	}
	dbUrl := os.Getenv("CONN_STRING")

	dab, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatalf("%v", err)
	} else {
		fmt.Println("DB OPEN SUCC")
	}
	defer dab.Close()
	dbQueries := database.New(dab)
	cfg.DB = dbQueries
	//	fmt.Println(dbUrl)
	_ = os.Getenv("KEY")

	return cfg.startSocket()
}

func (cfg *Config) startSocket() error {
	cfg.initSocketChannels()

	path := "wss://ws.eodhistoricaldata.com/ws/crypto?api_token=demo"
	towerPath := "ws://localhost:4269/seer/tower/ws"

	fmt.Println("Starting Crypto Client ... ")
	c, _, err := websocket.DefaultDialer.Dial(path, nil)
	if err != nil {
		return fmt.Errorf("websocket connection error: %v", err)
	}
	cfg.Socket = c

	fmt.Println("Starting Tower Client ... ")
	a, _, err := websocket.DefaultDialer.Dial(towerPath, nil)
	if err != nil {
		return fmt.Errorf("websocket connection error: %v", err)
	}
	cfg.towerSocket = a

	fmt.Println("Subscribing ...")

	for _, s := range cfg.Symbols {
		err := cfg.Subscribe(s)
		if err != nil {
			log.Printf("Failed to sub")
			return err
		}
		fmt.Printf("Forex :: %s  Sub complete\n", s)
	}
	go func() {
		var w = writer.NewPeriodicDataWriter(
			time.Minute, // Write interval
			"CC",
			func(symbolBuffers map[string][]batcher.SocketMsg) error {
				for symbol, buffer := range symbolBuffers {
					fmt.Printf("Writing %d Crypto ticks for symbol %s Time: %v\n", len(buffer), symbol, time.Now().Local())
					batches, err := batcher.BatchTicks(buffer, 1)
					if err == -1 {
						return errors.New("Failed to batch ticks")
					}
					for _, batch := range batches {
						stats := batcher.GetBatchStatistics(batch, 1)
						updateMsg := batchToStat(stats)

						cfg.Buffer[stats.Symbol] = stats
						err := cfg.towerSocket.WriteJSON(updateMsg)
						if err != nil {
							return err
							// TODO special insert signal to reset tower and reconnect !?
						}

						batcher.InsertBatch(stats, cfg.DB, "Crypto")
						fmt.Println("stats sent to socket")
					}
				}
				return nil
			},
		)

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
				log.Printf("Status MSG:  Code: %s  Type: %v Msg: %s Time: %v", v.Code, v.GetType(), v.Message, v.Time)
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
