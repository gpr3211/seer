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
)

type FullData map[string]map[string]batcher.BatchStats

type Client struct {
	httpClient http.Client
	Buffer     FullData
	mu         *sync.RWMutex
}

// GetExchanges returns an array of the exchanges
func (f FullData) GetExchanges() []string {

	out := make([]string, 0, len(f))
	for k := range f {
		out = append(out, k)
	}
	return out
}
func (f FullData) GetSymbols(exchange string) []string {
	syms := []string{}
	for k := range f[exchange] {
		syms = append(syms, k)
	}
	return syms
}

func NewClient() Client {

	data := make(FullData)
	data["crypto"] = make(map[string]batcher.BatchStats)

	//	data["forex"] = make(map[string]batcher.BatchStats)

	data["usdata"] = make(map[string]batcher.BatchStats)

	return Client{
		httpClient: http.Client{
			Timeout: time.Second * 30,
		},
		Buffer: data,
		mu:     &sync.RWMutex{},
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
		return
		//		path = "http://localhost:6971/seer/forex/v1/buff"

	case "usdata":
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
	for _, k := range stats {
		c.Buffer[service][k.Symbol] = k
	}
	c.mu.Unlock()

	fmt.Printf("Latests stats for %s Updated at %v\n", service, time.Now())
	return
}

func (c *Client) FetchAllStats() {

	for k := range c.Buffer {
		c.FetchLatest(k)
	}
	log.Println("Full Fetch complete")
	log.Printf("Printing full stats === \n")

	for k, v := range c.Buffer {
		fmt.Printf("\nPRITNING STATS FOR  %s\n", k)
		for _, stat := range v {
			t := time.UnixMilli(stat.EndTime)
			fmt.Printf("\nSymbol: %s\n Open: %v\n Close: %v\nLow: %v\nHigh: %v Time: %v\n", stat.Symbol, stat.Open, stat.Close, stat.Low, stat.High, t)
		}
	}
	return

}
