package models

import "time"

type Signal struct {
	ID        int64     `db:"id"`
	N         int       `db:"n"`
	BinSize   BinSize   `db:"bin"`
	Timestamp time.Time `db:"timestamp"`
	SMA       float64   `db:"sma"`  // simple moving average
	WMA       float64   `db:"wma"`  // weight moving average
	EMA       float64   `db:"ema"`  // exponential moving average
	BBTL      float64   `db:"bbtl"` // bolinger band top line
	BBML      float64   `db:"bbml"` // bolinger band middle line
	BBBL      float64   `db:"bbbl"` // bolinger band bottom line
	CreatedAt int64     `db:"created_at"`
	UpdatedAt int64     `db:"updated_at"`
}
