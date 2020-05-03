package models

import "time"

type Candle struct {
	ID              int64     `db:"id",json:"id"`
	Theme           string    `db:"theme",json:"theme"`
	Symbol          string    `db:"symbol"`
	Timestamp       time.Time `db:"timestamp"`
	Open            float64   `db:"open"`
	High            float64   `db:"high"`
	Low             float64   `db:"low"`
	Close           float64   `db:"close"`
	Trades          int       `db:"trades"`
	Volume          int64     `db:"volume"`
	Vwap            float64   `db:"vwap"`
	LastSize        int       `db:"lastsize"`
	Turnover        int64     `db:"turnover"`
	HomeNotional    float64   `db:"homenotional"`
	ForeignNotional float64   `db:"foreignnotional"`
	CreatedAt       int64     `db:"created_at"`
	UpdatedAt       int64     `db:"updated_at"`
}
