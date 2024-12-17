package model

type USTradeType string

const (
	USTradeTypeTick   = USTradeType("tick")
	USTradeTypeSub    = USTradeType("sub")
	USTradeTypeUnsub  = USTradeType("unsub")
	USTradeTypeStatus = USTradeType("status")
)

type USTrade interface {
	GetType() USTradeType
}

// USTradeTick is a response from a crypto exch
type USTradeTick struct {
	//	Exchange    Exchange `json:"e"`
	Symbol       string  `json:"s"`  // ex TSLA, BRB.K
	Price        float64 `json:"p"`  // float64
	Condition    []int   `json:"c"`  // status code ? TODO
	Quantity     int     `json:"v"`  // float64
	DarkPool     bool    `json:"dp"` // bool
	MarketStatus string  `json:"ms"` // open/close/???
	Timestamp    int64   `json:"t"`  // bool
	USTradeType  `json:"-"`
}

func (f USTradeTick) GetPrice() float64 {
	return f.Price
}

func (f USTradeTick) GetTime() int64 {
	return f.Timestamp
}
func (f USTradeTick) IsWebsocket() {}
func (f USTradeTick) GetVol() float64 {
	return float64(f.Quantity)
}
func (f USTradeTick) GetSym() string {
	return f.Symbol
}

func (f USTradeTick) GetType() USTradeType {
	return USTradeTypeTick
}

type StatusMsg struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	Time        int64  `json:"-"`
	USTradeType `json:"-"`
}

func (f StatusMsg) GetType() USTradeType {
	return USTradeTypeStatus
}

type SubMsgs struct {
	Action      string `json:"action"`
	Symbols     string `json:"symbols"` // Changed to string from []string
	USTradeType `json:"-"`
}

func (f SubMsgs) GetType() USTradeType {
	return f.USTradeType
}
