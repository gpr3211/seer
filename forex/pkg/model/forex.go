package model

type ForexType string

const (
	ForexTypeTick   = ForexType("tick")
	ForexTypeSub    = ForexType("sub")
	ForexTypeUnsub  = ForexType("unsub")
	ForexTypeStatus = ForexType("status")
)

type Forex interface {
	GetType() ForexType
}

type ForexTick struct {
	Symbol      string  `json:"s"`  // ex ETH-USD, BTC-USD
	BidPrice    float64 `json:"b"`  // float64
	AskPrice    float64 `json:"a"`  // float64
	Quantity    string  `json:"q"`  // float64
	DailyChange string  `json:"dc"` // float64 (Percentage)
	DailyDiff   string  `json:"dd"` // float64 (Percentage)
	Timestamp   int64   `json:"t"`  // int64
	ForexType
}

func (f ForexTick) GetType() ForexType {
	return ForexTypeTick
}

type StatusMsg struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Time    int64  `json:"-"`
	ForexType
}

func (f StatusMsg) GetType() ForexType {
	return ForexTypeStatus
}

type SubMsgs struct {
	Action  string `json:"action"`
	Symbols string `json:"symbols"` // Changed to string from []string
	ForexType
}

func (f SubMsgs) GetType() ForexType {
	return f.ForexType
}
