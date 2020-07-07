package types

type Exchange string

const (
	Bitmex  Exchange = "bitmex"
	Binance Exchange = "binance"
)

type Symbol string

const (
	XBTUSD Symbol = "XBTUSD"
)

type Side string

const (
	SideEmpty Side = "Empty"
	SideSell  Side = "Sell"
	SideBuy   Side = "Buy"
)

type OrderType string

const (
	Limit           OrderType = "Limit"
	Market          OrderType = "Market"
	StopLimit       OrderType = "StopLimit"
	LimitIfTouched  OrderType = "LimitIfTouched"
	MarketIfTouched OrderType = "MarketIfTouched"
)

type OrdStatus string

const (
	OrdNew             OrdStatus = "New"
	OrdFilled          OrdStatus = "Filled"
	OrdPartiallyFilled OrdStatus = "PartiallyFilled"
	ordCanceled        OrdStatus = "Canceled"
)

type PriceType string

const (
	TrailingStopPeg PriceType = "TrailingStopPeg"
)

type Theme string

const (
	Instrument Theme = "instrument"
	Trade      Theme = "trade"
	TradeBin1m Theme = "tradeBin1m"
	TradeBin5m Theme = "tradeBin5m"
	TradeBin1h Theme = "tradeBin1h"
	TradeBin1d Theme = "tradeBin1d"
)

type Operation string

const (
	SubscribeAct   Operation = "subscribe"
	UnsubscribeAct Operation = "unsubscribe"
)

type SubscribeMsg struct {
	Op   Operation `json:"op"` // operation
	Args []Theme   `json:"args"`
}

func NewTemeWithPair(t Theme, p Symbol) Theme {
	return t + ":" + Theme(p)
}

func NewSubscribeMsg(op Operation, args []Theme) *SubscribeMsg {
	return &SubscribeMsg{
		Op:   op,
		Args: args,
	}
}
