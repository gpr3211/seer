package websocket

import (
	"encoding/json"
	"github.com/gpr3211/seer/usdata/pkg/model"
)

func UnmarshalMsg(data []byte) (interface{}, error) {

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	// Check for distinctive fields to determine type
	_, hasStatusMsg := raw["status_code"]
	_, hasMessaMsg := raw["message"]
	if hasStatusMsg && hasMessaMsg {
		var status model.StatusMsg
		err := json.Unmarshal(data, &status)
		return status, err
	}
	var trade model.USTradeTick
	err := json.Unmarshal(data, &trade)
	return trade, err
}
