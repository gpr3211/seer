package batcher

import (
	"context"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gpr3211/seer/pkg/clog"
	"github.com/gpr3211/seer/pkg/database"
)

type SocketMsg interface {
	IsWebsocket()
	GetTime() int64
	GetPrice() float64
	GetSym() string
	GetVol() float64
}

// TimeBatch represents a batch of tick data within a time window
type TimeBatch struct {
	StartTime time.Time
	EndTime   time.Time
	Ticks     []SocketMsg
}

type BatchStats struct {
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
}

func (b BatchStats) UniqueKey() string {
	return fmt.Sprintf("%s_%d_%d", b.Symbol, b.StartTime, b.EndTime)
}

// Maintain a set of processed batch keys
var processedBatches = make(map[string]bool)
var processedBatchesMutex sync.Mutex

type Interface interface {
	// Len is the number of elements in the collection.
	Len() int
	// Less reports whether the element with
	// index i should sort before the element with index j.
	Less(i, j int) bool
	// Swap swaps the elements with indexes i and j.
	Swap(i, j int)
}

func (t TimeBatch) Len() int {
	return len(t.Ticks)
}
func (t TimeBatch) Less(i, j int) bool {
	return t.Ticks[i].GetTime() < t.Ticks[j].GetTime()
}
func (t TimeBatch) Swap(i, j int) {
	t.Ticks[i], t.Ticks[j] = t.Ticks[j], t.Ticks[i]
}

// BatchTicks groups tick data into specified minute intervals
func BatchTicks(ticks []SocketMsg, intervalMinutes int) ([]TimeBatch, int) {
	if len(ticks) == 0 {
		return nil, -1
	}

	// Sort ticks by timestamp to ensure chronological processing
	sort.Slice(ticks, func(i, j int) bool {
		return ticks[i].GetTime() < ticks[j].GetTime()
	})

	var batches []TimeBatch
	currentTime := time.UnixMilli(ticks[0].GetTime()).UTC()

	// Ensure clean, non-overlapping time windows
	currentBatchStart := currentTime.Truncate(time.Duration(intervalMinutes) * time.Minute)
	currentBatch := TimeBatch{
		StartTime: currentBatchStart,
		EndTime:   currentBatchStart.Add(time.Duration(intervalMinutes) * time.Minute),
		Ticks:     make([]SocketMsg, 0),
	}

	for _, tick := range ticks {
		tickTime := time.UnixMilli(tick.GetTime()).UTC()

		// Validate tick is within the current batch time window
		if tickTime.Before(currentBatch.StartTime) {
			// Skip ticks that are too early
			continue
		}

		// Handle batch transitions
		for tickTime.After(currentBatch.EndTime) || tickTime.Equal(currentBatch.EndTime) {
			if len(currentBatch.Ticks) > 0 {
				batches = append(batches, currentBatch)
			}

			currentBatchStart = currentBatch.EndTime
			currentBatch = TimeBatch{
				StartTime: currentBatchStart,
				EndTime:   currentBatchStart.Add(time.Duration(intervalMinutes) * time.Minute),
				Ticks:     make([]SocketMsg, 0),
			}
		}

		currentBatch.Ticks = append(currentBatch.Ticks, tick)
	}

	// Add final batch if it contains ticks
	if len(currentBatch.Ticks) > 0 {
		batches = append(batches, currentBatch)
	}

	return batches, intervalMinutes
}

// GetBatchStatistics remains unchanged as it works with the TimeBatch structure
func GetBatchStatistics(batch TimeBatch, period int) BatchStats {
	if len(batch.Ticks) == 0 {
		return BatchStats{}
	}
	stats := BatchStats{
		Symbol:    batch.Ticks[0].GetSym(),
		StartTime: batch.StartTime.UnixMilli(),
		EndTime:   batch.EndTime.UnixMilli(),
		Open:      batch.Ticks[0].GetPrice(),
		Close:     batch.Ticks[len(batch.Ticks)-1].GetPrice(),
		High:      batch.Ticks[0].GetPrice(),
		Low:       batch.Ticks[0].GetPrice(),
		Volume:    0,
		Period:    int32(period),
	}
	for _, tick := range batch.Ticks {
		price := tick.GetPrice()
		if price > stats.High {
			stats.High = price
		}
		if price < stats.Low {
			stats.Low = price
		}
		stats.Volume += tick.GetVol()
	}
	return stats
}

// InsertBatch Insters a batchStat to DB
// - InsertBatch(BatchStats, *databse.Queries)

func InsertBatch(b BatchStats, db *database.Queries, exchange string) error {
	processedBatchesMutex.Lock()
	key := b.UniqueKey()
	if processedBatches[key] {
		processedBatchesMutex.Unlock()
		return fmt.Errorf("batch already processed: %s", key)
	}
	processedBatches[key] = true
	processedBatchesMutex.Unlock()
	symId, err := db.GetTickerId(context.Background(), b.Symbol)
	if err != nil {
		db.CreateTicker(context.Background(), database.CreateTickerParams{
			ID:           uuid.New(),
			Code:         b.Symbol,
			Name:         b.Symbol,
			ExchangeName: exchange,
		})
	}

	symId, err = db.GetTickerId(context.Background(), b.Symbol)

	_, err = db.CreateBatchStat(context.Background(), database.CreateBatchStatParams{
		ID:            uuid.New(),
		SymID:         symId,
		StartTime:     b.StartTime,
		EndTime:       b.EndTime,
		Open:          b.Open,
		High:          b.High,
		Low:           b.Low,
		Close:         b.Close,
		Volume:        b.Volume,
		PeriodMinutes: b.Period,
	})
	if err != nil {
		log.Println("Failed to add batch to DB", err)
		clog.Println("Failed to add batch")
		return err
	} else {
		//		fmt.Printf("Batch Stat Added Symbol: %s | Timestamp: %v | Period %v", b.Symbol, b.EndTime, b.Period)
	}
	return nil
}
