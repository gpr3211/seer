package http

import (
	"net/http"
	"sync"
	"time"

	"github.com/gpr3211/seer/tower/internal/metadata"
)

type Client struct {
	httpClient http.Client
	Nurse	metadata.ServiceHealth
	BufferData	metadata.ServiceData
	mu *sync.RWMutex
}

func NewClient() Client {
	return Client{
		httpClient: http.Client{
			Timeout: time.Second * 30,
		},
	}
}

func (c *Client) FetchLatest(){}




}



