package models

import (
	"database/sql/driver"
	"fmt"
)

type BinSize uint8

const (
	_ BinSize = iota
	Bin1m
	Bin5m
	Bin1h
	Bin1d
)

func (b BinSize) String() string {
	switch b {
	case Bin1m:
		return "1m"
	case Bin5m:
		return "5m"
	case Bin1h:
		return "1h"
	case Bin1d:
		return "1d"
	default:
		return ""
	}
}

func (b *BinSize) Scan(value interface{}) error {
	bin, ok := value.(string)
	if !ok {
		return fmt.Errorf("value type must be only string, but current:%T", value)
	}
	var err error
	*b, err = ToBinSize(bin)
	return err
}

func (b *BinSize) Value() (driver.Value, error) {
	return b.String(), nil
}

func ToBinSize(binSize string) (BinSize, error) {
	switch binSize {
	case "1m":
		return Bin1m, nil
	case "5m":
		return Bin5m, nil
	case "1h":
		return Bin1h, nil
	case "1d":
		return Bin1d, nil
	default:
		return 0, fmt.Errorf("unknown bin_size:%s", binSize)
	}
}
