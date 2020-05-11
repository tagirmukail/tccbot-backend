package utils

import (
	"fmt"
	"math/rand"
	"time"
)

func FromTime(now time.Time, binSize string, count int) (from time.Time, err error) {
	var (
		toTime   time.Time
		duration time.Duration
	)
	switch binSize {
	case "1m":
		duration = time.Duration(count) * time.Minute
		toTime = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, time.UTC)
	case "5m":
		duration = time.Duration(count) * (5 * time.Minute)
		toTime = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, time.UTC)
	case "1h":
		duration = time.Duration(count) * (1 * time.Hour)
		toTime = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, time.UTC)
	case "1d":
		duration = time.Duration(count) * (24 * time.Hour)
		toTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	default:
		err = fmt.Errorf("unsupported bin size:%s", binSize)
		return
	}
	from = toTime.Add(-duration)
	return
}

func RandomRange(min, max float64) float64 {
	rand.Seed(time.Now().UnixNano())
	return min + rand.Float64()*(max-min)
}
