package websocket

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/gpr3211/seer/forex/pkg/model"
	"github.com/gpr3211/seer/pkg/batcher"
	"github.com/gpr3211/seer/pkg/clog"
	"github.com/gpr3211/seer/pkg/database"
	"github.com/gpr3211/seer/pkg/writer"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
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
		OutChan: make(chan model.ForexTick, 250),
		ErrChan: make(chan error, 10),
		Done:    make(chan struct{}),
		Closed:  false,
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
		Symbols: []string{"EURUSD"},
		Buffer:  (map[string]batcher.BatchStats{}),
	}
}

func StartForex(cfg *Config) error {
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
	dbQueries := database.New(dab)
	cfg.DB = dbQueries
	_ = os.Getenv("KEY")

	return cfg.startSocket()
}

func (cfg *Config) startSocket() error {
	cfg.initSocketChannels()

	var w = writer.NewPeriodicDataWriter(
		time.Minute, // Write interval
		10000,       // Max buffer size
		"FOREX",
		func(symbolBuffers map[string][]batcher.SocketMsg) error {
			for symbol, buffer := range symbolBuffers {
				fmt.Printf("Writing %d Forex ticks for symbol %s\n", len(buffer), symbol)
				batches, err := batcher.BatchTicks(buffer, 1)
				if err == -1 {
					return errors.New("Failed to batch ticks")
				}
				for _, batch := range batches {
					stats := batcher.GetBatchStatistics(batch, 1)

					cfg.Buffer[stats.Symbol] = stats
					batcher.InsertBatch(stats, cfg.DB, "FOREX")
					cfg.SendStats(stats)
					// TODO insert post to tower
				}
			}
			return nil
		},
	)

	path := "wss://ws.eodhistoricaldata.com/ws/forex?api_token=demo"
	c, _, err := websocket.DefaultDialer.Dial(path, nil)
	if err != nil {
		return fmt.Errorf("websocket connection error: %v", err)
	}
	cfg.Socket = c
	fmt.Println("Starting Forex Client ... ")
	fmt.Println("Subscribing ...")

	for _, s := range cfg.Symbols {
		err := cfg.Subscribe(s)
		if err != nil {
			log.Printf("Failed to sub")
			return err
		}
		fmt.Printf("Forex :: %s  Sub complete", s)
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
				log.Printf("Status MSG:  Code: %s  Type: %v Msg: %s Time: %v", v.Code, v.GetType(), v.Message, v.Time)
			case model.ForexTick:
				//		cfg.SaveForexToDB(v)
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

func (cfg *Config) SendStats(stats batcher.BatchStats) {

	path := "http//localhost:4269/seer/tower/update/forex"

	data, err := json.Marshal(stats)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", path, bytes.NewBuffer(data))
	if err != nil {
		return
	}

	req.Header.Set("Content-type", "application/json")
	resp, err := cfg.Client.httpClient.Do(req)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode > 399 {
		log.Println("bad request")
		return
	}
	log.Println("Batch stat post success.")

}

// SaveForexToDB saves a forex tick to the forex_tick table
// -- hard-coded symbol id bc demo account only allows for one.
func (cfg *Config) SaveForexToDB(v model.ForexTick) error {
	aP := strconv.FormatFloat(v.AskPrice, 'f', 64, 64)
	bP := strconv.FormatFloat(v.BidPrice, 'f', 64, 64)
	id, _ := uuid.FromBytes([]byte(`cbf0abfc-58ec-4de3-96f5-3da99939d732`))
	_, err := cfg.DB.AddForexTick(context.Background(), database.AddForexTickParams{
		ID:          uuid.New(),
		Time:        v.Timestamp,
		SymID:       id,
		AskPrice:    aP,
		BidPrice:    bP,
		DailyDiff:   v.DailyDiff,
		DailyChange: v.DailyChange,
	})
	if err != nil {
		clog.Println("failed to save tick")
		return err
	}
	return nil

}
