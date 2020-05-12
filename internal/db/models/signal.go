package models

import (
	"database/sql/driver"
	"fmt"
	"time"
)

type SignalType uint32

const (
	_ SignalType = iota
	SMA
	EMA
	WMA
	RSI
	BolingerBand
	MACD
)

type Signal struct {
	ID                 int64      `db:"id"`
	N                  int        `db:"n"`
	MACDFast           int        `db:"macd_fast"`
	MACDSlow           int        `db:"macd_slow"`
	MACDSig            int        `db:"macd_sig"`
	BinSize            BinSize    `db:"bin"`
	Timestamp          time.Time  `db:"timestamp"`
	SignalType         SignalType `db:"signal_t"`
	SignalValue        float64    `db:"signal_v"`
	BBTL               float64    `db:"bbtl"` // bolinger band top line
	BBML               float64    `db:"bbml"` // bolinger band middle line
	BBBL               float64    `db:"bbbl"` // bolinger band bottom line
	MACDValue          float64    `db:"macd_v"`
	MACDHistogramValue float64    `db:"macd_h_v"`
	CreatedAt          int64      `db:"created_at"`
	UpdatedAt          int64      `db:"updated_at"`
}

func (t SignalType) String() string {
	switch t {
	case SMA:
		return "SMA"
	case EMA:
		return "EMA"
	case WMA:
		return "WMA"
	case RSI:
		return "RSI"
	case MACD:
		return "MACD"
	case BolingerBand:
		return "BolingerBand"
	default:
		return ""
	}
}

func (t *SignalType) Scan(value interface{}) error {
	sigType, ok := value.(string)
	if !ok {
		return fmt.Errorf("value type must be only string, but current: %T", value)
	}
	var err error
	*t, err = ToSignalType(sigType)
	return err
}

func (t *SignalType) Value() (driver.Value, error) {
	return t, nil
}

func ToSignalType(sigType string) (SignalType, error) {
	switch sigType {
	case "SMA":
		return SMA, nil
	case "EMA":
		return EMA, nil
	case "WMA":
		return WMA, nil
	case "RSI":
		return RSI, nil
	case "MACD":
		return MACD, nil
	case "BolingerBand":
		return BolingerBand, nil
	default:
		return 0, fmt.Errorf("unknown signal_type: %s", sigType)
	}
}
