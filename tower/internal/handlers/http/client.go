package http

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gpr3211/seer/pkg/batcher"
	"github.com/gpr3211/seer/tower/internal/metadata"
)

type Client struct {
	httpClient http.Client
	Nurse      *metadata.ServiceHealth
	BufferData metadata.ServiceData
	mu         *sync.RWMutex
}

func NewClient() Client {
	health := metadata.ServiceHealth{"forex": true, "usdata": true, "crypto": true}
	buffer := metadata.ServiceData{"forex": []batcher.BatchStats{}, "usdata": []batcher.BatchStats{}, "crypto": []batcher.BatchStats{}}

	return Client{
		httpClient: http.Client{
			Timeout: time.Second * 30,
		},
		Nurse:      &health,
		BufferData: buffer,
		mu:         &sync.RWMutex{},
	}
}

//	forexPath := "http://localhost:6971/seer/forex/v1/"
//	usPath := "http://localhost:6970/seer/usdata/v1/"
//	cryptoPath := "http://localhost:6969/seer/crypto/v1/"

func (c *Client) FetchLatest(service string) {
	//	usPath := "http://localhost:6970/seer/usdata/v1/"
	//	cryptoPath := "http://localhost:6969/seer/crypto/v1/"

	path := ""
	switch service {
	case "forex":
		path = "http://localhost:6971/seer/forex/v1/buff"

	case "ustrade":
		path = "http://localhost:6970/seer/usdata/v1/buff"

	case "crypto":
		path = "http://localhost:6969/seer/crypto/v1/buff"

	}

	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		log.Printf("Failed to fetch %s stats ", service)
		return
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("Failed to fetch %s stats ", service)
		return
	}
	defer resp.Body.Close()
	stats := []batcher.BatchStats{}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Failed to read resp body for ", service)
		return
	}
	err = json.Unmarshal(data, &stats)
	if err != nil {
		return
	}
	c.mu.Lock()
	c.BufferData[service] = stats
	c.mu.Unlock()

	fmt.Printf("Latests stats for %s Updated at %v", service, time.Now())
	return
}

func (c *Client) FetchAllStats() {

	for k := range c.BufferData {
		c.FetchLatest(k)
	}
	log.Println("Full Fetch complete")
	log.Printf("Printing full stats === \n")

	for k, v := range c.BufferData {
		fmt.Printf("\nPRITNING STATS FOR  %s", k)
		for _, stat := range v {
			fmt.Printf("Symbol: %s\n Open: %v\n Close: %v\n Time: %v", stat.Symbol, stat.Open, stat.Close, stat.EndTime)
		}
	}
	return

}
