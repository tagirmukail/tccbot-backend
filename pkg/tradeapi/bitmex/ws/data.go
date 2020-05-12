package ws

import (
	"errors"
	"fmt"

	"github.com/tagirmukail/tccbot-backend/internal/types"
)

type BitmexData struct {
	Table  string               `json:"table"`
	Action string               `json:"action"`
	Data   []BitmexIncomingData `json:"data"`
}

type BitmexIncomingData struct {
	Table           string       `json:"table"`
	Symbol          types.Symbol `json:"symbol"`
	Timestamp       string       `json:"timestamp"`
	HomeNotional    float64      `json:"homeNotional"`
	ForeignNotional float64      `json:"foreignNotional"`

	// tradeBin
	Open     float64 `json:"open"`
	High     float64 `json:"high"`
	Low      float64 `json:"low"`
	Close    float64 `json:"close"`
	Trades   int     `json:"trades"`
	Volume   int64   `json:"volume"`
	LastSize int     `json:"lastSize"`
	Turnover int64   `json:"turnover"`
	Vwap     float64 `json:"vwap"`

	// exchanges
	Side          types.Side `json:"side"`
	Size          int        `json:"size"`
	Price         float64    `json:"price"`
	TickDirection string     `json:"tickDirection"`
	TrdMatchID    string     `json:"trdMatchID"`
	GrossValue    int64      `json:"grossValue"`
}

func (b *BitmexData) Validate() error {
	if b.Table != string(types.TradeBin5m) &&
		b.Table != string(types.TradeBin1h) &&
		b.Table != string(types.TradeBin1d) {
		return fmt.Errorf("bad table:%v", b.Table)
	}

	if b.Action != "update" && b.Action != "insert" {
		return fmt.Errorf("bad action: %v", b.Action)
	}

	if len(b.Data) == 0 {
		return errors.New("empty data")
	}

	return nil
}
