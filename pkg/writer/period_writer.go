package writer

import (
	"errors"
	"github.com/gpr3211/seer/pkg/batcher"
	"log"
	"sync"
	"time"
)

type PeriodWriter func(map[string][]batcher.SocketMsg) error
type PeriodicDataWriter struct {
	buffer        map[string][]batcher.SocketMsg
	bufferMutex   sync.RWMutex
	writeInterval time.Duration
	maxBufferSize int
	writeFn       func(map[string][]batcher.SocketMsg) error
	socketType    string
}

// NewPeriodicDataWriter creates a new periodic data writer
func NewPeriodicDataWriter(
	writeInterval time.Duration,
	maxBufferSize int,
	exchange string,
	writeFn func(map[string][]batcher.SocketMsg) error) *PeriodicDataWriter {
	pw := &PeriodicDataWriter{
		buffer:        make(map[string][]batcher.SocketMsg),
		writeInterval: writeInterval,
		maxBufferSize: maxBufferSize,
		writeFn:       writeFn,
		socketType:    exchange,
	}

	go pw.startPeriodicWrite()
	return pw
}

// AddData adds a new data item to the buffer
func (pw *PeriodicDataWriter) AddData(data batcher.SocketMsg) error {
	//	fmt.Println("Adding data to period buff", data.GetSym())
	pw.bufferMutex.Lock()
	defer pw.bufferMutex.Unlock()

	symbol := data.GetSym()

	pw.buffer[symbol] = append(pw.buffer[symbol], data)

	// Check if buffer should be written
	if len(pw.buffer[symbol]) >= pw.maxBufferSize {
		return pw.writeBufferForSymbol(symbol)
	}

	return nil
}

// writeBuffer writes the current buffer contents
func (pw *PeriodicDataWriter) writeBufferForSymbol(symbol string) error {
	if len(pw.buffer[symbol]) == 0 {
		return nil
	}
	bufferToWrite := make(map[string][]batcher.SocketMsg)
	bufferToWrite[symbol] = make([]batcher.SocketMsg, len(pw.buffer[symbol]))
	copy(bufferToWrite[symbol], pw.buffer[symbol])

	// Clear the original buffer for this symbol
	pw.buffer[symbol] = pw.buffer[symbol][:0]

	// Call the write function with the specific symbol's buffer
	return pw.writeFn(bufferToWrite)
}

func (pw *PeriodicDataWriter) startPeriodicWrite() {
	ticker := time.NewTicker(pw.writeInterval)
	defer ticker.Stop()

	for range ticker.C {
		pw.bufferMutex.Lock()

		// Prepare buffers to write
		buffersToWrite := make(map[string][]batcher.SocketMsg)
		for symbol, buffer := range pw.buffer {
			if len(buffer) > 0 {
				buffersToWrite[symbol] = make([]batcher.SocketMsg, len(buffer))
				copy(buffersToWrite[symbol], buffer)
				pw.buffer[symbol] = pw.buffer[symbol][:0]
			}
		}
		pw.bufferMutex.Unlock()

		// Write buffers outside of the mutex lock
		if len(buffersToWrite) > 0 {
			if err := pw.writeFn(buffersToWrite); err != nil {
				log.Printf("Error writing buffers: %v", err)
			}
		}
	}
} // SaveBatchedStats processes incoming ticks, batches them, and saves the statistics
func SaveBatchedStats(tickChan <-chan interface{}, saveFn func(batcher.BatchStats) error) {
	// Collect ticks to batch
	var ticks []batcher.SocketMsg

	// Create a ticker to periodically process and save batches
	batchTicker := time.NewTicker(1 * time.Minute)
	defer batchTicker.Stop()

	for {
		select {
		case tick, ok := <-tickChan:
			if !ok {
				// Channel closed, process remaining ticks
				processBatchesAndSave(ticks, saveFn)
				return
			}

			// Type assert and add to ticks if it's a SocketMsg
			if socketMsg, ok := tick.(batcher.SocketMsg); ok {
				ticks = append(ticks, socketMsg)
			}

		case <-batchTicker.C:
			// Periodic batch processing
			processBatchesAndSave(ticks, saveFn)
			// Reset ticks after processing
			ticks = nil
		}
	}
}

// processBatchesAndSave processes ticks into batches and saves their statistics
func processBatchesAndSave(ticks []batcher.SocketMsg, saveFn func(batcher.BatchStats) error) error {
	// If no ticks, do nothing
	if len(ticks) == 0 {
		return errors.New("len is 0")
	}

	// Batch ticks into 1-minute intervals
	batches, ok := batcher.BatchTicks(ticks, 1)
	if ok == -1 {
		log.Println("Failed to batch ticks", "Time: ", time.Now().UnixMilli())

		return errors.New("Failed to batch stats")
	}

	// Process and save statistics for each batch
	for _, batch := range batches {
		// Calculate batch statistics
		stats := batcher.GetBatchStatistics(batch, 1)

		// Save the statistics using provided save function
		if err := saveFn(stats); err != nil {
			log.Printf("Error saving batch stats: %v", err)
		}
	}
	return nil
}

// TOOD EXAMPLESDSADASD
/*
writer := NewPeriodicDataWriter(
    time.Minute,  // Write interval
    100,          // Max buffer size
    func(symbolBuffers map[string][]SocketMsg) error {
        for symbol, buffer := range symbolBuffers {
            fmt.Printf("Writing %d ticks for symbol %s\n", len(buffer), symbol)
            // Actual write logic (e.g., database insert)
            for _, tick := range buffer {
                // Process each tick
            }
        }
        return nil
    },
)


*/
