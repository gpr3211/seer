package model

import "strconv"

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
	Symbol      string `json:"s"`  // ex ETH-USD, BTC-USD
	Price       string `json:"p"`  // float64
	Quantity    string `json:"q"`  // float64
	DailyChange string `json:"dc"` // float
	DailyDiff   string `json:"dd"` // float
	Timestamp   int64  `json:"t"`  // int64
	USTradeType `json:"-"`
}

func (f USTradeTick) GetPrice() float64 {
	out, _ := strconv.ParseFloat(f.Quantity, 64)
	return out
}

func (f USTradeTick) GetTime() int64 {
	return f.Timestamp
}
func (f USTradeTick) IsWebsocket() {}
func (f USTradeTick) GetVol() float64 {
	out, _ := strconv.ParseFloat(f.Quantity, 64)
	return out
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
