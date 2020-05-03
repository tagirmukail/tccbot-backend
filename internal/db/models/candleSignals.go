package models

import "time"

type CandleWithSignals struct {
	Candle  *Candle  `json:"candle"`
	Signals []Signal `json:"signals"`
}

type CandleData struct {
	X time.Time  `json:"x"`
	Y [4]float64 `json:"y"` // 0:open, 1:high, 2:low, 3:close
}

type LineData struct {
	X time.Time `json:"x"`
	Y float64   `json:"y"`
}

type CandleWithSignalsData struct {
	Candles []CandleData `json:"candles"`
	SMA20   []LineData   `json:"sma_20"`
	EMA20   []LineData   `json:"ema_20"`
	WMA20   []LineData   `json:"wma_20"`
	SMA50   []LineData   `json:"sma_50"`
	EMA50   []LineData   `json:"ema_50"`
	WMA50   []LineData   `json:"wma_50"`
	BBTL    []LineData   `json:"bbtl"`
	BBML    []LineData   `json:"bbml"`
	BBBL    []LineData   `json:"bbbl"`
}
