package websocket

import (
	"encoding/json"
	"github.com/gpr3211/seer/crypto/pkg/model"
)

func UnmarshalMsg(data []byte) (interface{}, error) {

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	// Check for distinctive fields to determine type
	_, hasStatusMsg := raw["status_code"]
	_, hasMessaMsg := raw["message"]
	_, hasDailyChange := raw["dc"]
	if hasStatusMsg && hasMessaMsg {
		var status model.StatusMsg
		err := json.Unmarshal(data, &status)
		return status, err
	}

	if hasDailyChange {
		var crypto model.CryptoTick
		err := json.Unmarshal(data, &crypto)
		return crypto, err
	}
	var forex model.CryptoTick
	// CLEAN ? yes
	err := json.Unmarshal(data, &forex)
	return forex, err
}
