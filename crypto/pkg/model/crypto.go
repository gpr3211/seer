package model

import "strconv"

type CryptoType string

const (
	CryptoTypeTick   = CryptoType("tick")
	CryptoTypeSub    = CryptoType("sub")
	CryptoTypeUnsub  = CryptoType("unsub")
	CryptoTypeStatus = CryptoType("status")
)

type Crypto interface {
	GetType() CryptoType
}

// CryptoTick is a response from a crypto exch
type CryptoTick struct {
	//	Exchange    Exchange `json:"e"`
	Symbol      string `json:"s"`  // ex ETH-USD, BTC-USD
	Price       string `json:"p"`  // float64
	Quantity    string `json:"q"`  // float64
	DailyChange string `json:"dc"` // float
	DailyDiff   string `json:"dd"` // float
	Timestamp   int64  `json:"t"`  // int64
	CryptoType  `json:"-"`
}

func (f CryptoTick) GetPrice() float64 {
	out, _ := strconv.ParseFloat(f.Quantity, 64)
	return out
}

func (f CryptoTick) GetTime() int64 {
	return f.Timestamp
}
func (f CryptoTick) IsWebsocket() {}
func (f CryptoTick) GetVol() float64 {
	out, _ := strconv.ParseFloat(f.Quantity, 64)
	return out
}
func (f CryptoTick) GetSym() string {
	return f.Symbol
}

func (f CryptoTick) GetType() CryptoType {
	return CryptoTypeTick
}

type StatusMsg struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Time       int64  `json:"-"`
	CryptoType `json:"-"`
}

func (f StatusMsg) GetType() CryptoType {
	return CryptoTypeStatus
}

type SubMsgs struct {
	Action     string `json:"action"`
	Symbols    string `json:"symbols"` // Changed to string from []string
	CryptoType `json:"-"`
}

func (f SubMsgs) GetType() CryptoType {
	return f.CryptoType
}
