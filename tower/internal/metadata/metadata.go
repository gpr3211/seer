package metadata

import "github.com/gpr3211/seer/pkg/batcher"

type ServiceHealth map[string]bool
type ServiceData map[string][]batcher.BatchStats

const (
	FOREX string = "forex"
	CC    string = "crypto"
	US    string = "usdata"
)

func init() {
    SvcHealth := 

}
